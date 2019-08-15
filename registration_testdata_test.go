// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"net/http"
)

const (
	attestationOptionsRequest1 = `{
		"username": "johndoe@example.com",
		"displayName": "John Doe",
		"authenticatorSelection": {
			"requireResidentKey": false,
			"residentKey": "preferred",
			"authenticatorAttachment": "cross-platform",
			"userVerification": "preferred"
		},
		"attestation": "direct"
	}`

	attestationOptionsRequest2 = `{
		"username": "johndoe@example.com",
		"displayName": "John Doe",
		"authenticatorSelection": {
			"requireResidentKey": true,
			"residentKey": "required",
			"authenticatorAttachment": "platform",
			"userVerification": "preferred"
		},
		"attestation": "none"
	}`

	attestationOptionsRequestMissingUserName = `{
		"username": "",
		"displayName": "John Doe",
		"authenticatorSelection": {
			"requireResidentKey": false,
			"residentKey": "preferred",
			"authenticatorAttachment": "cross-platform",
			"userVerification": "preferred"
		},
		"attestation": "direct"
	}`

	attestationOptionsRequestMissingDisplayName = `{
		"username": "johndoe@example.com",
		"displayName": "",
		"authenticatorSelection": {
			"requireResidentKey": false,
			"residentKey": "preferred",
			"authenticatorAttachment": "cross-platform",
			"userVerification": "preferred"
		},
		"attestation": "direct"
	}`

	attestationOptionsSuccessResponse1 = `{
		"status": "ok",
		"errorMessage": "",
		"rp": {
			"id": "localhost",
			"name": "WebAuthn local server"
		},
		"user": {
			"id": "S3932ee31vKEC0JtJMIQ",
			"name": "johndoe@example.com",
			"displayName": "John Doe"
		},
	
		"challenge": "uhUjPNlZfvn7onwuhNdsLPkkE5Fv-lUN",
		"pubKeyCredParams": [
			{
				"type": "public-key",
				"alg": -7
			}
		],
		"timeout": 10000,
		"authenticatorSelection": {
			"requireResidentKey": false,
			"residentKey": "preferred",
			"authenticatorAttachment": "cross-platform",
			"userVerification": "preferred"
		},
		"attestation": "direct"
	}`

	attestationOptionsSuccessResponse2 = `{
		"status": "ok",
		"errorMessage": "",
		"rp": {
			"id": "localhost",
			"name": "WebAuthn local server"
		},
		"user": {
			"id": "S3932ee31vKEC0JtJMIQ",
			"name": "johndoe@example.com",
			"displayName": "John Doe"
		},
	
		"challenge": "uhUjPNlZfvn7onwuhNdsLPkkE5Fv-lUN",
		"pubKeyCredParams": [
			{
				"type": "public-key",
				"alg": -7
			}
		],
		"timeout": 10000,
		"authenticatorSelection": {
			"requireResidentKey": true,
			"residentKey": "required",
			"authenticatorAttachment": "platform",
			"userVerification": "preferred"
		},
		"attestation": "none"
	}`

	attestationOptionsSuccessResponseExistingUser = `{
		"status": "ok",
		"errorMessage": "",
		"rp": {
			"id": "localhost",
			"name": "WebAuthn local server"
		},
		"user": {
			"id": "AQID",
			"name": "johndoe@example.com",
			"displayName": "John Doe"
		},

		"challenge": "uhUjPNlZfvn7onwuhNdsLPkkE5Fv-lUN",
		"pubKeyCredParams": [
			{
				"type": "public-key",
				"alg": -7
			}
		],
		"excludeCredentials": [
			{
				"type": "public-key",
				"id": "LFdoCFJTyB82ZzSJUHc-c72yraRc_1mPvGX8ToE8su39xX26Jcqd31LUkKOS36FIAWgWl6itMKqmDvruha6ywA"
			}
		],
		"timeout": 10000,
		"authenticatorSelection": {
			"requireResidentKey": false,
			"residentKey": "preferred",
			"authenticatorAttachment": "cross-platform",
			"userVerification": "preferred"
		},
		"attestation": "direct"
	}`

	attestationOptionsErrorResponseMissingUserName = `{
		"status": "failed",
		"errorMessage": "Missing username"
	}`

	attestationOptionsErrorResponseMissingDisplayName = `{
		"status": "failed",
		"errorMessage": "Missing displayName"
	}`

	attestationResultRequest = `{
		"id": "LFdoCFJTyB82ZzSJUHc-c72yraRc_1mPvGX8ToE8su39xX26Jcqd31LUkKOS36FIAWgWl6itMKqmDvruha6ywA",
		"rawId": "LFdoCFJTyB82ZzSJUHc-c72yraRc_1mPvGX8ToE8su39xX26Jcqd31LUkKOS36FIAWgWl6itMKqmDvruha6ywA",
		"response": {
			"clientDataJSON": "eyJjaGFsbGVuZ2UiOiJOeHlab3B3VktiRmw3RW5uTWFlXzVGbmlyN1FKN1FXcDFVRlVLakZIbGZrIiwiY2xpZW50RXh0ZW5zaW9ucyI6e30sImhhc2hBbGdvcml0aG0iOiJTSEEtMjU2Iiwib3JpZ2luIjoiaHR0cDovL2xvY2FsaG9zdDozMDAwIiwidHlwZSI6IndlYmF1dGhuLmNyZWF0ZSJ9",
			"attestationObject": "o2NmbXRoZmlkby11MmZnYXR0U3RtdKJjc2lnWEcwRQIgVzzvX3Nyp_g9j9f2B-tPWy6puW01aZHI8RXjwqfDjtQCIQDLsdniGPO9iKr7tdgVV-FnBYhvzlZLG3u28rVt10YXfGN4NWOBWQJOMIICSjCCATKgAwIBAgIEVxb3wDANBgkqhkiG9w0BAQsFADAuMSwwKgYDVQQDEyNZdWJpY28gVTJGIFJvb3QgQ0EgU2VyaWFsIDQ1NzIwMDYzMTAgFw0xNDA4MDEwMDAwMDBaGA8yMDUwMDkwNDAwMDAwMFowLDEqMCgGA1UEAwwhWXViaWNvIFUyRiBFRSBTZXJpYWwgMjUwNTY5MjI2MTc2MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEZNkcVNbZV43TsGB4TEY21UijmDqvNSfO6y3G4ytnnjP86ehjFK28-FdSGy9MSZ-Ur3BVZb4iGVsptk5NrQ3QYqM7MDkwIgYJKwYBBAGCxAoCBBUxLjMuNi4xLjQuMS40MTQ4Mi4xLjUwEwYLKwYBBAGC5RwCAQEEBAMCBSAwDQYJKoZIhvcNAQELBQADggEBAHibGMqbpNt2IOL4i4z96VEmbSoid9Xj--m2jJqg6RpqSOp1TO8L3lmEA22uf4uj_eZLUXYEw6EbLm11TUo3Ge-odpMPoODzBj9aTKC8oDFPfwWj6l1O3ZHTSma1XVyPqG4A579f3YAjfrPbgj404xJns0mqx5wkpxKlnoBKqo1rqSUmonencd4xanO_PHEfxU0iZif615Xk9E4bcANPCfz-OLfeKXiT-1msixwzz8XGvl2OTMJ_Sh9G9vhE-HjAcovcHfumcdoQh_WM445Za6Pyn9BZQV3FCqMviRR809sIATfU5lu86wu_5UGIGI7MFDEYeVGSqzpzh6mlcn8QSIZoYXV0aERhdGFYxEmWDeWIDoxodDQXD2R2YFuP5K65ooYyx5lc87qDHZdjQQAAAAAAAAAAAAAAAAAAAAAAAAAAAEAsV2gIUlPIHzZnNIlQdz5zvbKtpFz_WY-8ZfxOgTyy7f3Ffbolyp3fUtSQo5LfoUgBaBaXqK0wqqYO-u6FrrLApQECAyYgASFYIPr9-YH8DuBsOnaI3KJa0a39hyxh9LDtHErNvfQSyxQsIlgg4rAuQQ5uy4VXGFbkiAt0uwgJJodp-DymkoBcrGsLtkI"
		},
		"type": "public-key"
	}`

	attestationResultRequestMissingID = `{
		"response": {
			"clientDataJSON": "eyJjaGFsbGVuZ2UiOiJOeHlab3B3VktiRmw3RW5uTWFlXzVGbmlyN1FKN1FXcDFVRlVLakZIbGZrIiwiY2xpZW50RXh0ZW5zaW9ucyI6e30sImhhc2hBbGdvcml0aG0iOiJTSEEtMjU2Iiwib3JpZ2luIjoiaHR0cDovL2xvY2FsaG9zdDozMDAwIiwidHlwZSI6IndlYmF1dGhuLmNyZWF0ZSJ9",
			"attestationObject": "o2NmbXRoZmlkby11MmZnYXR0U3RtdKJjc2lnWEcwRQIgVzzvX3Nyp_g9j9f2B-tPWy6puW01aZHI8RXjwqfDjtQCIQDLsdniGPO9iKr7tdgVV-FnBYhvzlZLG3u28rVt10YXfGN4NWOBWQJOMIICSjCCATKgAwIBAgIEVxb3wDANBgkqhkiG9w0BAQsFADAuMSwwKgYDVQQDEyNZdWJpY28gVTJGIFJvb3QgQ0EgU2VyaWFsIDQ1NzIwMDYzMTAgFw0xNDA4MDEwMDAwMDBaGA8yMDUwMDkwNDAwMDAwMFowLDEqMCgGA1UEAwwhWXViaWNvIFUyRiBFRSBTZXJpYWwgMjUwNTY5MjI2MTc2MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEZNkcVNbZV43TsGB4TEY21UijmDqvNSfO6y3G4ytnnjP86ehjFK28-FdSGy9MSZ-Ur3BVZb4iGVsptk5NrQ3QYqM7MDkwIgYJKwYBBAGCxAoCBBUxLjMuNi4xLjQuMS40MTQ4Mi4xLjUwEwYLKwYBBAGC5RwCAQEEBAMCBSAwDQYJKoZIhvcNAQELBQADggEBAHibGMqbpNt2IOL4i4z96VEmbSoid9Xj--m2jJqg6RpqSOp1TO8L3lmEA22uf4uj_eZLUXYEw6EbLm11TUo3Ge-odpMPoODzBj9aTKC8oDFPfwWj6l1O3ZHTSma1XVyPqG4A579f3YAjfrPbgj404xJns0mqx5wkpxKlnoBKqo1rqSUmonencd4xanO_PHEfxU0iZif615Xk9E4bcANPCfz-OLfeKXiT-1msixwzz8XGvl2OTMJ_Sh9G9vhE-HjAcovcHfumcdoQh_WM445Za6Pyn9BZQV3FCqMviRR809sIATfU5lu86wu_5UGIGI7MFDEYeVGSqzpzh6mlcn8QSIZoYXV0aERhdGFYxEmWDeWIDoxodDQXD2R2YFuP5K65ooYyx5lc87qDHZdjQQAAAAAAAAAAAAAAAAAAAAAAAAAAAEAsV2gIUlPIHzZnNIlQdz5zvbKtpFz_WY-8ZfxOgTyy7f3Ffbolyp3fUtSQo5LfoUgBaBaXqK0wqqYO-u6FrrLApQECAyYgASFYIPr9-YH8DuBsOnaI3KJa0a39hyxh9LDtHErNvfQSyxQsIlgg4rAuQQ5uy4VXGFbkiAt0uwgJJodp-DymkoBcrGsLtkI"
		},
		"type": "public-key"
	}`

	attestationResultSuccessResponse = `{
		"status": "ok",
		"errorMessage": ""
	}`

	attestationResultErrorResponseBadContextData = `{
		"status": "failed",
		"errorMessage": "Session doesn't have PublicKeyCredentialCreationOptions data"
	}`

	attestationResultErrorResponseFailedToParse = `{
		"status": "failed",
		"errorMessage": "Failed to parse attestation: webauthn/attestation: missing credential id and raw id"
	}`

	attestationResultErrorResponseFailedToVerify = `{
		"status": "failed",
		"errorMessage": "Failed to verify attestation: webauthn/attestation: failed to verify client data challenge: client data challenge does not match expected challenge"
	}`
)

var (
	attestationOptionsTests = handlerTest{
		requestMethod:     "POST",
		requestURL:        "/attestation/options",
		equalResponseBody: equalAttestationOptionsResponse(),
		testcases: []handlerTestData{
			{
				name:                 "success",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreGetUserNone,
				initMockSessionStore: initSessionStore(getEmptySession, getAttestationOptionsNewUserSession1),
				requestBody:          attestationOptionsRequest1,
				wantStatusCode:       http.StatusOK,
				wantResponseBody:     attestationOptionsSuccessResponse1,
			},
			{
				name:                 "request overrides webauthn config settings",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreGetUserNone,
				initMockSessionStore: initSessionStore(getEmptySession, getAttestationOptionsNewUserSession2),
				requestBody:          attestationOptionsRequest2,
				wantStatusCode:       http.StatusOK,
				wantResponseBody:     attestationOptionsSuccessResponse2,
			},
			{
				name:                 "user exists",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreGetUser,
				initMockSessionStore: initSessionStore(getEmptySession, getAttestationOptionsExistingUserSession),
				requestBody:          attestationOptionsRequest1,
				wantStatusCode:       http.StatusOK,
				wantResponseBody:     attestationOptionsSuccessResponseExistingUser,
			},
			{
				name:                 "request missing user name",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreGetUserNone,
				initMockSessionStore: initSessionStore(getEmptySession, getEmptySession),
				requestBody:          attestationOptionsRequestMissingUserName,
				wantStatusCode:       http.StatusBadRequest,
				wantResponseBody:     attestationOptionsErrorResponseMissingUserName,
			},
			{
				name:                 "request missing display name",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreGetUserNone,
				initMockSessionStore: initSessionStore(getEmptySession, getEmptySession),
				requestBody:          attestationOptionsRequestMissingDisplayName,
				wantStatusCode:       http.StatusBadRequest,
				wantResponseBody:     attestationOptionsErrorResponseMissingDisplayName,
			},
		},
	}

	attestationResultTests = handlerTest{
		requestMethod:     "POST",
		requestURL:        "/attestation/result",
		equalResponseBody: equalServerResponse,
		testcases: []handlerTestData{
			{
				name:                 "success",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreAddUserCredential,
				initMockSessionStore: initSessionStore(getAttestationOptionsNewUserSession1, getUserSession),
				requestBody:          attestationResultRequest,
				wantStatusCode:       http.StatusOK,
				wantResponseBody:     attestationResultSuccessResponse,
			},
			{
				name:                 "wrong/empty session data in context",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreAddUserCredentialNotCalled,
				initMockSessionStore: initSessionStore(getEmptySession, getEmptySession),
				requestBody:          attestationResultRequest,
				wantStatusCode:       http.StatusUnauthorized,
				wantResponseBody:     attestationResultErrorResponseBadContextData,
			},
			{
				name:                 "can't parse attestation",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreAddUserCredentialNotCalled,
				initMockSessionStore: initSessionStore(getAttestationOptionsNewUserSession1, getUserSession),
				requestBody:          attestationResultRequestMissingID,
				wantStatusCode:       http.StatusBadRequest,
				wantResponseBody:     attestationResultErrorResponseFailedToParse,
			},
			{
				name:                 "can't verify attestation",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreAddUserCredentialNotCalled,
				initMockSessionStore: initSessionStore(getAttestationOptionsNewUserWrongChallengeSession, getUserSession),
				requestBody:          attestationResultRequest,
				wantStatusCode:       http.StatusBadRequest,
				wantResponseBody:     attestationResultErrorResponseFailedToVerify,
			},
		},
	}
)
