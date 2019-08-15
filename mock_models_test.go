// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"encoding/base64"

	"github.com/fxamacker/webauthn"

	"github.com/gorilla/sessions"
)

var (
	mockNewUser = &user{
		UserID:      []byte{1, 2, 3},
		UserName:    "johndoe@example.com",
		DisplayName: "John Doe",
	}

	mockExistingUser = &user{
		UserID:      []byte{1, 2, 3},
		UserName:    "johndoe@example.com",
		DisplayName: "John Doe",
		CredentialIDs: [][]byte{
			base64RawURLDecodeString("LFdoCFJTyB82ZzSJUHc-c72yraRc_1mPvGX8ToE8su39xX26Jcqd31LUkKOS36FIAWgWl6itMKqmDvruha6ywA"),
		},
	}

	mockCredential = &credential{
		CredentialID: base64RawURLDecodeString("LFdoCFJTyB82ZzSJUHc-c72yraRc_1mPvGX8ToE8su39xX26Jcqd31LUkKOS36FIAWgWl6itMKqmDvruha6ywA"),
		UserID:       []byte{1, 2, 3},
		Counter:      uint32(0),
		CoseKey:      []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 250, 253, 249, 129, 252, 14, 224, 108, 58, 118, 136, 220, 162, 90, 209, 173, 253, 135, 44, 97, 244, 176, 237, 28, 74, 205, 189, 244, 18, 203, 20, 44, 34, 88, 32, 226, 176, 46, 65, 14, 110, 203, 133, 87, 24, 86, 228, 136, 11, 116, 187, 8, 9, 38, 135, 105, 248, 60, 166, 146, 128, 92, 172, 107, 11, 182, 66},
	}
)

func getEmptySession(store sessions.Store) *sessions.Session {
	return sessions.NewSession(store, sessionNameLoginSession)
}

func getAttestationOptionsNewUserSession1(store sessions.Store) *sessions.Session {
	session := sessions.NewSession(store, sessionNameLoginSession)
	mockNewUserCopy := *mockNewUser
	session.Values[sessionMapKeyUserSession] = &userSession{User: &mockNewUserCopy}
	session.Values[sessionMapKeyWebAuthnCreationOptions] = &webauthn.PublicKeyCredentialCreationOptions{
		RP: webauthn.PublicKeyCredentialRpEntity{
			Name: "WebAuthn local server",
			ID:   "localhost",
		},
		User: webauthn.PublicKeyCredentialUserEntity{
			Name:        mockNewUserCopy.UserName,
			ID:          mockNewUserCopy.UserID,
			DisplayName: mockNewUserCopy.DisplayName,
		},
		Challenge: base64RawURLDecodeString("NxyZopwVKbFl7EnnMae_5Fnir7QJ7QWp1UFUKjFHlfk"),
		PubKeyCredParams: []webauthn.PublicKeyCredentialParameters{
			{Type: webauthn.PublicKeyCredentialTypePublicKey, Alg: webauthn.COSEAlgES256},
		},
		Timeout:            uint64(10000),
		ExcludeCredentials: nil,
		AuthenticatorSelection: webauthn.AuthenticatorSelectionCriteria{
			AuthenticatorAttachment: webauthn.AuthenticatorCrossPlatform,
			RequireResidentKey:      false,
			ResidentKey:             webauthn.ResidentKeyPreferred,
			UserVerification:        webauthn.UserVerificationPreferred,
		},
		Attestation: webauthn.AttestationDirect,
	}
	return session
}

func getAttestationOptionsNewUserSession2(store sessions.Store) *sessions.Session {
	session := sessions.NewSession(store, sessionNameLoginSession)
	mockNewUserCopy := *mockNewUser
	session.Values[sessionMapKeyUserSession] = &userSession{User: &mockNewUserCopy}
	session.Values[sessionMapKeyWebAuthnCreationOptions] = &webauthn.PublicKeyCredentialCreationOptions{
		RP: webauthn.PublicKeyCredentialRpEntity{
			Name: "WebAuthn local server",
			ID:   "localhost",
		},
		User: webauthn.PublicKeyCredentialUserEntity{
			Name:        mockNewUserCopy.UserName,
			ID:          mockNewUserCopy.UserID,
			DisplayName: mockNewUserCopy.DisplayName,
		},
		Challenge: base64RawURLDecodeString("NxyZopwVKbFl7EnnMae_5Fnir7QJ7QWp1UFUKjFHlfk"),
		PubKeyCredParams: []webauthn.PublicKeyCredentialParameters{
			{Type: webauthn.PublicKeyCredentialTypePublicKey, Alg: webauthn.COSEAlgES256},
		},
		Timeout:            uint64(10000),
		ExcludeCredentials: nil,
		AuthenticatorSelection: webauthn.AuthenticatorSelectionCriteria{
			AuthenticatorAttachment: webauthn.AuthenticatorPlatform,
			RequireResidentKey:      true,
			ResidentKey:             webauthn.ResidentKeyRequired,
			UserVerification:        webauthn.UserVerificationPreferred,
		},
		Attestation: webauthn.AttestationNone,
	}
	return session
}

func getAttestationOptionsNewUserWrongChallengeSession(store sessions.Store) *sessions.Session {
	session := getAttestationOptionsNewUserSession1(store)
	v := session.Values[sessionMapKeyWebAuthnCreationOptions].(*webauthn.PublicKeyCredentialCreationOptions)
	v.Challenge[0] = 0
	return session
}

func getAttestationOptionsExistingUserSession(store sessions.Store) *sessions.Session {
	session := sessions.NewSession(store, sessionNameLoginSession)
	mockExistingUserCopy := *mockExistingUser
	var excludeCredentials []webauthn.PublicKeyCredentialDescriptor
	for _, id := range mockExistingUserCopy.CredentialIDs {
		excludeCredentials = append(excludeCredentials, webauthn.PublicKeyCredentialDescriptor{Type: webauthn.PublicKeyCredentialTypePublicKey, ID: id})
	}
	session.Values[sessionMapKeyUserSession] = &userSession{User: &mockExistingUserCopy}
	session.Values[sessionMapKeyWebAuthnCreationOptions] = &webauthn.PublicKeyCredentialCreationOptions{
		RP: webauthn.PublicKeyCredentialRpEntity{
			Name: "WebAuthn local server",
			ID:   "localhost",
		},
		User: webauthn.PublicKeyCredentialUserEntity{
			Name:        mockExistingUserCopy.UserName,
			ID:          mockExistingUserCopy.UserID,
			DisplayName: mockExistingUserCopy.DisplayName,
		},
		Challenge: base64RawURLDecodeString("xdj0CBfX692qsATpy0kNc8533JdvdLUpqYP8wDTX_ZE"),
		PubKeyCredParams: []webauthn.PublicKeyCredentialParameters{
			{Type: webauthn.PublicKeyCredentialTypePublicKey, Alg: webauthn.COSEAlgES256},
		},
		Timeout:            uint64(10000),
		ExcludeCredentials: excludeCredentials,
		AuthenticatorSelection: webauthn.AuthenticatorSelectionCriteria{
			AuthenticatorAttachment: webauthn.AuthenticatorCrossPlatform,
			RequireResidentKey:      false,
			ResidentKey:             webauthn.ResidentKeyPreferred,
			UserVerification:        webauthn.UserVerificationPreferred,
		},
		Attestation: webauthn.AttestationDirect,
	}
	return session
}

func getAssertionOptionsExistingUserSession(store sessions.Store) *sessions.Session {
	session := sessions.NewSession(store, sessionNameLoginSession)
	mockExistingUserCopy := *mockExistingUser
	session.Values[sessionMapKeyUserSession] = &userSession{User: &mockExistingUserCopy}
	session.Values[sessionMapKeyWebAuthnRequestOptions] = &webauthn.PublicKeyCredentialRequestOptions{
		Challenge: base64RawURLDecodeString("xdj0CBfX692qsATpy0kNc8533JdvdLUpqYP8wDTX_ZE"),
		Timeout:   uint64(10000),
		RPID:      "localhost",
		AllowCredentials: []webauthn.PublicKeyCredentialDescriptor{
			{
				Type: webauthn.PublicKeyCredentialTypePublicKey,
				ID:   base64RawURLDecodeString("LFdoCFJTyB82ZzSJUHc-c72yraRc_1mPvGX8ToE8su39xX26Jcqd31LUkKOS36FIAWgWl6itMKqmDvruha6ywA"),
			},
		},
		UserVerification: webauthn.UserVerificationPreferred,
	}
	return session
}

func getAssertionOptionsExistingUserWrongChallengeSession(store sessions.Store) *sessions.Session {
	session := getAssertionOptionsExistingUserSession(store)
	v := session.Values[sessionMapKeyWebAuthnRequestOptions].(*webauthn.PublicKeyCredentialRequestOptions)
	v.Challenge[0] = 255
	return session
}

func getUserSession(store sessions.Store) *sessions.Session {
	session := sessions.NewSession(store, sessionNameLoginSession)
	mockExistingUserCopy := *mockExistingUser
	session.Values[sessionMapKeyUserSession] = &userSession{
		User:                 &mockExistingUserCopy,
		LoggedInCredentialID: base64RawURLDecodeString("LFdoCFJTyB82ZzSJUHc-c72yraRc_1mPvGX8ToE8su39xX26Jcqd31LUkKOS36FIAWgWl6itMKqmDvruha6ywA"),
	}
	return session
}

func base64RawURLDecodeString(s string) []byte {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}
