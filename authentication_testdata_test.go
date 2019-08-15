// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"net/http"
)

const (
	assertionOptionsRequest1 = `{
		"username": "johndoe@example.com",
		"userVerification": "preferred"
	}`

	assertionOptionsRequest2 = `{
		"username": "johndoe@example.com",
		"userVerification": ""
	}`

	assertionOptionsRequestMissingUserName = `{
		"username": "",
		"userVerification": "preferred"
	}`

	assertionOptionsSuccessResponse1 = `{
		"status": "ok",
		"errorMessage": "",
		"challenge": "6283u0svT-YIF3pSolzkQHStwkJCaLKx",
		"timeout": 10000,
		"rpId": "localhost",
		"allowCredentials": [
			{
				"id": "LFdoCFJTyB82ZzSJUHc-c72yraRc_1mPvGX8ToE8su39xX26Jcqd31LUkKOS36FIAWgWl6itMKqmDvruha6ywA",
				"type": "public-key"
			}
		],
		"userVerification": "preferred"
	}`

	assertionOptionsSuccessResponse2 = `{
		"status": "ok",
		"errorMessage": "",
		"challenge": "6283u0svT-YIF3pSolzkQHStwkJCaLKx",
		"timeout": 10000,
		"rpId": "localhost",
		"allowCredentials": [
			{
				"id": "LFdoCFJTyB82ZzSJUHc-c72yraRc_1mPvGX8ToE8su39xX26Jcqd31LUkKOS36FIAWgWl6itMKqmDvruha6ywA",
				"type": "public-key"
			}
		],
		"userVerification": "preferred"
	}`

	assertionOptionsErrorResponseUserNotRegistered = `{
		"status": "failed",
		"errorMessage": "johndoe@example.com is not registered"
	}`

	assertionOptionsErrorResponseMissingUserName = `{
		"status": "failed",
		"errorMessage": "Missing username"
	}`

	assertionResultRequest = `{
		"id":"LFdoCFJTyB82ZzSJUHc-c72yraRc_1mPvGX8ToE8su39xX26Jcqd31LUkKOS36FIAWgWl6itMKqmDvruha6ywA",
		"rawId":"LFdoCFJTyB82ZzSJUHc-c72yraRc_1mPvGX8ToE8su39xX26Jcqd31LUkKOS36FIAWgWl6itMKqmDvruha6ywA",
		"response":{
			"authenticatorData":"SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2MBAAAAAA",
			"signature":"MEYCIQCv7EqsBRtf2E4o_BjzZfBwNpP8fLjd5y6TUOLWt5l9DQIhANiYig9newAJZYTzG1i5lwP-YQk9uXFnnDaHnr2yCKXL",
			"userHandle":"",
			"clientDataJSON":"eyJjaGFsbGVuZ2UiOiJ4ZGowQ0JmWDY5MnFzQVRweTBrTmM4NTMzSmR2ZExVcHFZUDh3RFRYX1pFIiwiY2xpZW50RXh0ZW5zaW9ucyI6e30sImhhc2hBbGdvcml0aG0iOiJTSEEtMjU2Iiwib3JpZ2luIjoiaHR0cDovL2xvY2FsaG9zdDozMDAwIiwidHlwZSI6IndlYmF1dGhuLmdldCJ9"
		},
		"type":"public-key"
	}`

	assertionResultRequestMissingID = `{
		"response":{
			"authenticatorData":"SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2MBAAAAAA",
			"signature":"MEYCIQCv7EqsBRtf2E4o_BjzZfBwNpP8fLjd5y6TUOLWt5l9DQIhANiYig9newAJZYTzG1i5lwP-YQk9uXFnnDaHnr2yCKXL",
			"userHandle":"",
			"clientDataJSON":"eyJjaGFsbGVuZ2UiOiJ4ZGowQ0JmWDY5MnFzQVRweTBrTmM4NTMzSmR2ZExVcHFZUDh3RFRYX1pFIiwiY2xpZW50RXh0ZW5zaW9ucyI6e30sImhhc2hBbGdvcml0aG0iOiJTSEEtMjU2Iiwib3JpZ2luIjoiaHR0cDovL2xvY2FsaG9zdDozMDAwIiwidHlwZSI6IndlYmF1dGhuLmdldCJ9"
		},
		"type":"public-key"
	}`

	assertionResultSuccessResponse = `{
		"status": "ok",
		"errorMessage": ""
	}`

	assertionResultErrorResponseBadContextData = `{
		"status": "failed",
		"errorMessage": "Session doesn't have PublicKeyCredentialRequestOptions data"
	}`

	assertionResultErrorResponseFailedToParse = `{
		"status": "failed",
		"errorMessage": "Failed to parse assertion: webauthn/assertion: missing credential id and raw id"
	}`

	assertionResultErrorResponseFailedToVerify = `{
		"status": "failed",
		"errorMessage": "Failed to verify assertion: webauthn/assertion: failed to verify client data challenge: client data challenge does not match expected challenge"
	}`
)

var (
	assertionOptionsTests = handlerTest{
		requestMethod:     "POST",
		requestURL:        "/assertion/options",
		equalResponseBody: equalAssertionOptionsResponse(),
		testcases: []handlerTestData{
			{
				name:                 "success",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreGetUser,
				initMockSessionStore: initSessionStore(getEmptySession, getAssertionOptionsExistingUserSession),
				requestBody:          assertionOptionsRequest1,
				wantStatusCode:       http.StatusOK,
				wantResponseBody:     assertionOptionsSuccessResponse1,
			},
			{
				name:                 "request overrides webauthn config settings",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreGetUser,
				initMockSessionStore: initSessionStore(getEmptySession, getAssertionOptionsExistingUserSession),
				requestBody:          assertionOptionsRequest2,
				wantStatusCode:       http.StatusOK,
				wantResponseBody:     assertionOptionsSuccessResponse2,
			},
			{
				name:                 "user doesn't exist",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreGetUserNone,
				initMockSessionStore: initSessionStore(getEmptySession, getEmptySession),
				requestBody:          assertionOptionsRequest1,
				wantStatusCode:       http.StatusBadRequest,
				wantResponseBody:     assertionOptionsErrorResponseUserNotRegistered,
			},
			{
				name:                 "request missing user name",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreGetUser,
				initMockSessionStore: initSessionStore(getEmptySession, getEmptySession),
				requestBody:          assertionOptionsErrorResponseMissingUserName,
				wantStatusCode:       http.StatusBadRequest,
				wantResponseBody:     assertionOptionsErrorResponseMissingUserName,
			},
		},
	}

	assertionResultTests = handlerTest{
		requestMethod:     "POST",
		requestURL:        "/assertion/result",
		equalResponseBody: equalServerResponse,
		testcases: []handlerTestData{
			{
				name:                 "success",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreGetAndUpdateCredential,
				initMockSessionStore: initSessionStore(getAssertionOptionsExistingUserSession, getUserSession),
				requestBody:          assertionResultRequest,
				wantStatusCode:       http.StatusOK,
				wantResponseBody:     assertionResultSuccessResponse,
			},
			{
				name:                 "wrong session data in context",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreGetAndUpdateCredentialNotCalled,
				initMockSessionStore: initSessionStore(getEmptySession, getEmptySession),
				requestBody:          assertionResultRequest,
				wantStatusCode:       http.StatusUnauthorized,
				wantResponseBody:     assertionResultErrorResponseBadContextData,
			},
			{
				name:                 "can't parse assertion",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreGetAndUpdateCredentialNotCalled,
				initMockSessionStore: initSessionStore(getAssertionOptionsExistingUserSession, getUserSession),
				requestBody:          assertionResultRequestMissingID,
				wantStatusCode:       http.StatusBadRequest,
				wantResponseBody:     assertionResultErrorResponseFailedToParse,
			},
			{
				name:                 "can't verify assertion",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreGetAndUpdateCredential,
				initMockSessionStore: initSessionStore(getAssertionOptionsExistingUserWrongChallengeSession, getUserSession),
				requestBody:          assertionResultRequest,
				wantStatusCode:       http.StatusBadRequest,
				wantResponseBody:     assertionResultErrorResponseFailedToVerify,
			},
		},
	}
)
