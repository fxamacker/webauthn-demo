// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/fxamacker/webauthn"
	"github.com/gorilla/sessions"
)

func (s *server) handleAttestationOptions() http.HandlerFunc {
	type request struct {
		Username               string                                   `json:"username"`
		DisplayName            string                                   `json:"displayName"`
		AuthenticatorSelection webauthn.AuthenticatorSelectionCriteria  `json:"authenticatorSelection"`
		Attestation            webauthn.AttestationConveyancePreference `json:"attestation"`
	}
	type response struct {
		serverResponse
		*webauthn.PublicKeyCredentialCreationOptions
	}
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := r.Context().Value(contextKeyLoginSession).(*sessions.Session)
		if !ok {
			panic("Failed to get session data from context")
		}

		// Parse and verify request.
		var optionsRequest request
		if err := json.NewDecoder(r.Body).Decode(&optionsRequest); err != nil {
			writeFailedServerResponse(w, http.StatusBadRequest, "Failed to json decode request body: "+err.Error())
			return
		}
		if optionsRequest.Username == "" {
			writeFailedServerResponse(w, http.StatusBadRequest, "Missing username")
			return
		}
		if optionsRequest.DisplayName == "" {
			writeFailedServerResponse(w, http.StatusBadRequest, "Missing displayName")
			return
		}
		if optionsRequest.AuthenticatorSelection.UserVerification == "" {
			optionsRequest.AuthenticatorSelection.UserVerification = webauthn.UserVerificationPreferred
		}
		if optionsRequest.Attestation == "" {
			optionsRequest.Attestation = webauthn.AttestationNone
		}

		// Get user from datastore.
		u, err := s.dataStore.getUser(r.Context(), optionsRequest.Username)
		if err == errNoRecords {
			u = &user{
				UserName:    optionsRequest.Username,
				DisplayName: optionsRequest.DisplayName,
			}
		} else if err != nil {
			writeFailedServerResponse(w, http.StatusInternalServerError, "Failed to query user in database: "+err.Error())
			return
		}

		// Generate user ID for new user.
		if u.UserID == nil {
			u.UserID = make([]byte, 64) // user ID is 64 random bytes
			if n, err := rand.Read(u.UserID); err != nil {
				writeFailedServerResponse(w, http.StatusInternalServerError, "Failed to generate user ID: "+err.Error())
				return
			} else if n != 64 {
				writeFailedServerResponse(w, http.StatusInternalServerError, "Failed to generate requested random bytes")
				return
			}
		}

		// Generate PublicKeyCredentialCreationOptions from WebAuthn config and user input.
		creationOptions, err := webauthn.NewAttestationOptions(s.webAuthnConfig, &webauthn.User{ID: u.UserID, Name: u.UserName, DisplayName: u.DisplayName, CredentialIDs: u.CredentialIDs})
		if err != nil {
			writeFailedServerResponse(w, http.StatusInternalServerError, "Failed to generate PublicKeyCredentialCreationOptions: "+err.Error())
			return
		}
		creationOptions.AuthenticatorSelection = optionsRequest.AuthenticatorSelection
		creationOptions.Attestation = optionsRequest.Attestation

		// Save creationOptions and user info in session to verify new credential later.
		session.Values[sessionMapKeyWebAuthnCreationOptions] = creationOptions
		session.Values[sessionMapKeyUserSession] = &userSession{User: u}

		// Write response.
		creationOptionsResponse := &response{
			serverResponse:                     serverResponse{Status: statusOK},
			PublicKeyCredentialCreationOptions: creationOptions,
		}
		b, err := json.Marshal(creationOptionsResponse)
		if err != nil {
			writeFailedServerResponse(w, http.StatusInternalServerError, "Failed to json encode response body: "+err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}
}

func (s *server) handleAttestationResult(w http.ResponseWriter, r *http.Request) {
	// Get saved creationOptions and user info.
	session, ok := r.Context().Value(contextKeyLoginSession).(*sessions.Session)
	if !ok {
		panic("Failed to get session data from context")
	}
	savedCreationOptions, ok := session.Values[sessionMapKeyWebAuthnCreationOptions].(*webauthn.PublicKeyCredentialCreationOptions)
	if !ok {
		delete(session.Values, sessionMapKeyWebAuthnCreationOptions)
		writeFailedServerResponse(w, http.StatusUnauthorized, "Session doesn't have PublicKeyCredentialCreationOptions data")
		return
	}
	uSession, ok := session.Values[sessionMapKeyUserSession].(*userSession)
	if !ok {
		delete(session.Values, sessionMapKeyWebAuthnCreationOptions)
		writeFailedServerResponse(w, http.StatusUnauthorized, "Session doesn't have user data")
		return
	}

	// Parse and verify request.
	credentialAttestation, err := webauthn.ParseAttestation(r.Body)
	if err != nil {
		delete(session.Values, sessionMapKeyWebAuthnCreationOptions)
		writeFailedServerResponse(w, http.StatusBadRequest, "Failed to parse attestation: "+err.Error())
		return
	}
	var credentialAlgs []int
	for _, param := range savedCreationOptions.PubKeyCredParams {
		credentialAlgs = append(credentialAlgs, param.Alg)
	}
	expected := &webauthn.AttestationExpectedData{
		Origin:           s.rpOrigin,
		RPID:             savedCreationOptions.RP.ID,
		CredentialAlgs:   credentialAlgs,
		Challenge:        base64.RawURLEncoding.EncodeToString(savedCreationOptions.Challenge),
		UserVerification: savedCreationOptions.AuthenticatorSelection.UserVerification,
	}
	// todo: VerifyAttestation returns attestationType and trustPath.  Need to verify that
	// attestation type is acceptable and trust path can be trusted.
	_, _, err = webauthn.VerifyAttestation(credentialAttestation, expected)
	if err != nil {
		delete(session.Values, sessionMapKeyWebAuthnCreationOptions)
		writeFailedServerResponse(w, http.StatusBadRequest, "Failed to verify attestation: "+err.Error())
		return
	}

	// Save user credential in datastore.
	c := &credential{
		CredentialID: credentialAttestation.RawID,
		UserID:       uSession.User.UserID,
		Counter:      credentialAttestation.AuthnData.Counter,
		CoseKey:      credentialAttestation.AuthnData.Credential.Raw,
	}
	if err = s.dataStore.addUserCredential(r.Context(), uSession.User, c); err == errRecordExists {
		delete(session.Values, sessionMapKeyWebAuthnCreationOptions)
		writeFailedServerResponse(w, http.StatusInternalServerError, "User credential exists in the system")
		return
	} else if err != nil {
		delete(session.Values, sessionMapKeyWebAuthnCreationOptions)
		writeFailedServerResponse(w, http.StatusInternalServerError, "Failed to save user credential: "+err.Error())
		return
	}

	// Delete creationOptions and update user info in session.
	delete(session.Values, sessionMapKeyWebAuthnCreationOptions)
	uSession.User.CredentialIDs = append(uSession.User.CredentialIDs, credentialAttestation.RawID)
	if len(uSession.LoggedInCredentialID) == 0 {
		uSession.LoggedInCredentialID = credentialAttestation.RawID
	}

	// Write response.
	writeOKServerResponse(w)
}
