// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"net/http"
)

var (
	logoutSuccessResponse = `{
		"status": "ok",
		"errorMessage": ""
	}`

	userSuccessResponse = `{
		"status": "ok",
		"name": "johndoe@example.com",
		"displayName": "John Doe",
		"CredentialID": "LFdoCFJTyB82ZzSJUHc-c72yraRc_1mPvGX8ToE8su39xX26Jcqd31LUkKOS36FIAWgWl6itMKqmDvruha6ywA",
		"RegisteredAt": "2009-11-10 23:00:00 +0000 UTC",
		"LoggedInAt": "2009-11-10 23:00:00 +0000 UTC"
	}`

	userErrorResponse = `{
		"status": "failed",
		"errorMessage": "User is not logged in"
	}`

	logoutTests = handlerTest{
		requestMethod:     "GET",
		requestURL:        "/logout",
		equalResponseBody: equalServerResponse,
		testcases: []handlerTestData{
			{
				name:                 "user is not logged in",
				server:               getMockServer(),
				initMockDataStore:    nil,
				initMockSessionStore: initSessionStore(getEmptySession, getEmptySession),
				requestBody:          "",
				wantStatusCode:       http.StatusOK,
				wantResponseBody:     logoutSuccessResponse,
			},
			{
				name:                 "user is logged in",
				server:               getMockServer(),
				initMockDataStore:    nil,
				initMockSessionStore: initSessionStore(getUserSession, getEmptySession),
				requestBody:          "",
				wantStatusCode:       http.StatusOK,
				wantResponseBody:     logoutSuccessResponse,
			},
		},
	}

	userTests = handlerTest{
		requestMethod:     "GET",
		requestURL:        "/user",
		equalResponseBody: equalServerResponse,
		testcases: []handlerTestData{
			{
				name:                 "user is not logged in",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreGetCredentialTimestamp,
				initMockSessionStore: initSessionStore(getEmptySession, getEmptySession),
				requestBody:          "",
				wantStatusCode:       http.StatusUnauthorized,
				wantResponseBody:     userErrorResponse,
			},
			{
				name:                 "user is logged in",
				server:               getMockServer(),
				initMockDataStore:    initDataStoreGetCredentialTimestamp,
				initMockSessionStore: initSessionStore(getUserSession, getUserSession),
				requestBody:          "",
				wantStatusCode:       http.StatusOK,
				wantResponseBody:     userSuccessResponse,
			},
		},
	}
)
