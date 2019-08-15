// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/fxamacker/webauthn"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/mock"
)

type (
	initMockDataStoreFunc    func(*MockDataStore)
	initMockSessionStoreFunc func(*MockSessionStore)
	getSessionFunc           func(sessions.Store) *sessions.Session
	equalResponseBodyFunc    func([]byte, []byte) (bool, error)

	handlerTestData struct {
		name                 string
		server               *server
		initMockDataStore    initMockDataStoreFunc
		initMockSessionStore initMockSessionStoreFunc
		requestBody          string
		wantStatusCode       int
		wantResponseBody     string
	}
	handlerTest struct {
		requestMethod     string
		requestURL        string
		equalResponseBody equalResponseBodyFunc
		testcases         []handlerTestData
	}
)

var (
	handlerTests = []handlerTest{
		attestationOptionsTests,
		attestationResultTests,
		assertionOptionsTests,
		assertionResultTests,
		logoutTests,
		userTests,
	}
)

func getWebAuthnConfig() *webauthn.Config {
	config := &webauthn.Config{
		RPID:                    "localhost",
		RPName:                  "WebAuthn local server",
		RPIcon:                  "",
		Timeout:                 uint64(10000),
		ChallengeLength:         32,
		AuthenticatorAttachment: webauthn.AuthenticatorCrossPlatform,
		ResidentKey:             webauthn.ResidentKeyPreferred,
		UserVerification:        webauthn.UserVerificationPreferred,
		Attestation:             webauthn.AttestationDirect,
		CredentialAlgs:          []int{webauthn.COSEAlgES256},
	}
	if err := config.Valid(); err != nil {
		panic(err)
	}
	return config
}

func getMockServer() *server {
	return &server{
		webAuthnConfig: getWebAuthnConfig(),
		dataStore:      &MockDataStore{},
		sessionStore:   &MockSessionStore{},
		router:         mux.NewRouter(),
		rpOrigin:       "http://localhost:3000",
	}
}

func equalAttestationOptionsResponse() equalResponseBodyFunc {
	type response struct {
		serverResponse
		*webauthn.PublicKeyCredentialCreationOptions
	}
	return func(got []byte, want []byte) (bool, error) {
		var gotResponse, wantResponse response
		if err := json.Unmarshal(want, &wantResponse); err != nil {
			return false, err
		}
		if err := json.Unmarshal(got, &gotResponse); err != nil {
			return false, err
		}
		if wantResponse.PublicKeyCredentialCreationOptions != nil {
			wantResponse.User.ID = gotResponse.User.ID
			wantResponse.Challenge = gotResponse.Challenge
		}
		return reflect.DeepEqual(gotResponse, wantResponse), nil
	}
}

func equalAssertionOptionsResponse() equalResponseBodyFunc {
	type response struct {
		serverResponse
		*webauthn.PublicKeyCredentialRequestOptions
	}
	return func(got []byte, want []byte) (bool, error) {
		var gotResponse, wantResponse response
		if err := json.Unmarshal(want, &wantResponse); err != nil {
			return false, err
		}
		if err := json.Unmarshal(got, &gotResponse); err != nil {
			return false, err
		}
		if wantResponse.PublicKeyCredentialRequestOptions != nil {
			wantResponse.Challenge = gotResponse.Challenge
		}
		return reflect.DeepEqual(gotResponse, wantResponse), nil
	}
}

func equalServerResponse(got []byte, want []byte) (bool, error) {
	var gotResponse, wantResponse serverResponse
	if err := json.Unmarshal(want, &wantResponse); err != nil {
		return false, err
	}
	if err := json.Unmarshal(got, &gotResponse); err != nil {
		return false, err
	}
	return reflect.DeepEqual(gotResponse, wantResponse), nil
}

func TestHandlers(t *testing.T) {
	for _, test := range handlerTests {
		requestMethod := test.requestMethod
		requestURL := test.requestURL
		equalResponseBody := test.equalResponseBody

		for _, tc := range test.testcases {
			name := requestMethod + " " + requestURL + " " + tc.name
			t.Run(name, func(t *testing.T) {
				if tc.initMockDataStore != nil {
					tc.initMockDataStore(tc.server.dataStore.(*MockDataStore))
				}
				if tc.initMockSessionStore != nil {
					tc.initMockSessionStore(tc.server.sessionStore.(*MockSessionStore))
				}

				tc.server.routes()

				recorder := httptest.NewRecorder()

				r, err := http.NewRequest(requestMethod, requestURL, strings.NewReader(tc.requestBody))
				if err != nil {
					t.Fatal(err)
				}

				tc.server.router.ServeHTTP(recorder, r)

				// Verify response status
				if recorder.Code != tc.wantStatusCode {
					t.Errorf("%s status code is %d, want %d", requestURL, recorder.Code, tc.wantStatusCode)
				}

				// Verify response body
				if responseEqual, err := equalResponseBody(recorder.Body.Bytes(), []byte(tc.wantResponseBody)); err != nil {
					t.Errorf("Failed to test response body: " + err.Error())
				} else if !responseEqual {
					t.Errorf("%s response is %s, want %s", requestURL, recorder.Body.String(), tc.wantResponseBody)
				}

				mock.AssertExpectationsForObjects(t)
			})
		}
	}
}
