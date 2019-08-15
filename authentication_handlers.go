// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/fxamacker/webauthn"
	"github.com/gorilla/sessions"
)

func (s *server) handleAssertionOptions() http.HandlerFunc {
	type request struct {
		Username         string                               `json:"username"`
		UserVerification webauthn.UserVerificationRequirement `json:"userVerification"`
	}
	type response struct {
		serverResponse
		*webauthn.PublicKeyCredentialRequestOptions
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
		if optionsRequest.UserVerification == "" {
			optionsRequest.UserVerification = webauthn.UserVerificationPreferred
		}

		// Get user from datastore.
		u, err := s.dataStore.getUser(r.Context(), optionsRequest.Username)
		if err != nil && err != errNoRecords {
			writeFailedServerResponse(w, http.StatusInternalServerError, "Failed to query user in database: "+err.Error())
			return
		}
		if u == nil || u.UserID == nil {
			writeFailedServerResponse(w, http.StatusBadRequest, optionsRequest.Username+" is not registered")
			return
		}

		// Generate PublicKeyCredentialRequestOptions from WebAuthn config and user input.
		requestOptions, err := webauthn.NewAssertionOptions(s.webAuthnConfig, &webauthn.User{ID: u.UserID, Name: u.UserName, DisplayName: u.DisplayName, CredentialIDs: u.CredentialIDs})
		if err != nil {
			writeFailedServerResponse(w, http.StatusInternalServerError, "Failed to generate PublicKeyCredentialRequestOptions: "+err.Error())
			return
		}
		requestOptions.UserVerification = optionsRequest.UserVerification

		// Save requestOptions and user info in session to verify credential later.
		session.Values[sessionMapKeyWebAuthnRequestOptions] = requestOptions
		session.Values[sessionMapKeyUserSession] = &userSession{User: u}

		// Write response.
		getOptionsResponse := response{
			serverResponse:                    serverResponse{Status: statusOK},
			PublicKeyCredentialRequestOptions: requestOptions,
		}
		b, err := json.Marshal(getOptionsResponse)
		if err != nil {
			writeFailedServerResponse(w, http.StatusInternalServerError, "Failed to json encode response body: "+err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}
}

func (s *server) handleAssertionResult(w http.ResponseWriter, r *http.Request) {
	// Get saved requestOptions and user info.
	session, ok := r.Context().Value(contextKeyLoginSession).(*sessions.Session)
	if !ok {
		panic("Failed to get session data from context")
	}
	savedRequestOptions, ok := session.Values[sessionMapKeyWebAuthnRequestOptions].(*webauthn.PublicKeyCredentialRequestOptions)
	if !ok {
		delete(session.Values, sessionMapKeyWebAuthnRequestOptions)
		writeFailedServerResponse(w, http.StatusUnauthorized, "Session doesn't have PublicKeyCredentialRequestOptions data")
		return
	}
	uSession, ok := session.Values[sessionMapKeyUserSession].(*userSession)
	if !ok {
		delete(session.Values, sessionMapKeyWebAuthnRequestOptions)
		writeFailedServerResponse(w, http.StatusUnauthorized, "Session doesn't have user data")
		return
	}

	// Parse credential.
	credentialAssertion, err := webauthn.ParseAssertion(r.Body)
	if err != nil {
		delete(session.Values, sessionMapKeyWebAuthnRequestOptions)
		writeFailedServerResponse(w, http.StatusBadRequest, "Failed to parse assertion: "+err.Error())
		return
	}

	// Get credential from datastore by received credential ID.
	c, err := s.dataStore.getCredential(r.Context(), uSession.User.UserID, credentialAssertion.RawID)
	if err != nil {
		delete(session.Values, sessionMapKeyWebAuthnRequestOptions)
		writeFailedServerResponse(w, http.StatusInternalServerError, "Failed to find credential: "+err.Error())
		return
	}
	credKey, _, err := webauthn.ParseCredential(c.CoseKey)
	if err != nil {
		delete(session.Values, sessionMapKeyWebAuthnRequestOptions)
		writeFailedServerResponse(w, http.StatusInternalServerError, "Failed to create credential public key: "+err.Error())
		return
	}

	// Verify credential.
	var userCredentialIDs [][]byte
	for _, desc := range savedRequestOptions.AllowCredentials {
		userCredentialIDs = append(userCredentialIDs, desc.ID)
	}
	expected := &webauthn.AssertionExpectedData{
		Origin:            s.rpOrigin,
		RPID:              savedRequestOptions.RPID,
		Challenge:         base64.RawURLEncoding.EncodeToString(savedRequestOptions.Challenge),
		UserVerification:  savedRequestOptions.UserVerification,
		UserID:            uSession.User.UserID,
		UserCredentialIDs: userCredentialIDs,
		PrevCounter:       c.Counter,
		Credential:        credKey,
	}
	if err = webauthn.VerifyAssertion(credentialAssertion, expected); err != nil {
		delete(session.Values, sessionMapKeyWebAuthnRequestOptions)
		writeFailedServerResponse(w, http.StatusBadRequest, "Failed to verify assertion: "+err.Error())
		return
	}

	// Update authenticator counter in datastore.
	c.Counter = credentialAssertion.AuthnData.Counter
	if err = s.dataStore.updateCredential(r.Context(), c); err != nil {
		delete(session.Values, sessionMapKeyWebAuthnRequestOptions)
		writeFailedServerResponse(w, http.StatusInternalServerError, "Failed to update credential: "+err.Error())
		return
	}

	// Delete requestOptions and update user info in session.
	delete(session.Values, sessionMapKeyWebAuthnRequestOptions)
	uSession.LoggedInCredentialID = credentialAssertion.RawID

	// Write response.
	writeOKServerResponse(w)
}
