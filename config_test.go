// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"encoding/base64"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/fxamacker/webauthn"
)

type configTest struct {
	name              string
	configFileContent string
	configEnv         map[string]string
	wantConfig        config
}

type configErrorTest struct {
	name              string
	configFileContent string
	configEnv         map[string]string
	wantErrorMsg      string
}

const (
	configFileContent = `{
		"WebAuthn": {
			"RPID": "localhost",
			"RPName": "WebAuthn local host",
			"RPIcon": "",
			"Timeout": 30000,
			"ChallengeLength": 32,
			"AuthenticatorAttachment": "cross-platform",
			"ResidentKey": "preferred",
			"UserVerification": "preferred",
			"Attestation": "direct",
			"CredentialAlgs": [ -7, -37, -257 ]
		},
		"Origin": "https://localhost:8443"
	}`
	invalidWebAuthnConfigFileContent = `{
		"WebAuthn": {
			"RPID": "",
			"RPName": "WebAuthn local host",
			"RPIcon": "",
			"Timeout": 30000,
			"ChallengeLength": 32,
			"AuthenticatorAttachment": "cross-platform",
			"ResidentKey": "preferred",
			"UserVerification": "preferred",
			"Attestation": "direct",
			"CredentialAlgs": [ -7, -37, -257 ]
		},
		"Origin": "https://localhost:8443"
	}`
	invalidOriginConfigFileContent = `{
		"WebAuthn": {
			"RPID": "localhost",
			"RPName": "WebAuthn local host",
			"RPIcon": "",
			"Timeout": 30000,
			"ChallengeLength": 32,
			"AuthenticatorAttachment": "cross-platform",
			"ResidentKey": "preferred",
			"UserVerification": "preferred",
			"Attestation": "direct",
			"CredentialAlgs": [ -7, -37, -257 ]
		},
		"Origin": ""
	}`
)

var (
	webAuthnConfig = &webauthn.Config{
		RPID:                    "localhost",
		RPName:                  "WebAuthn local host",
		RPIcon:                  "",
		Timeout:                 uint64(30000),
		ChallengeLength:         32,
		AuthenticatorAttachment: "cross-platform",
		ResidentKey:             "preferred",
		UserVerification:        "preferred",
		Attestation:             "direct",
		CredentialAlgs:          []int{-7, -37, -257},
	}

	configTests = []configTest{
		{
			name:              "success",
			configFileContent: configFileContent,
			configEnv: map[string]string{
				"SESSION_KEY":   base64.RawStdEncoding.EncodeToString([]byte("secure_session_key")),
				"DB_CONNSTRING": "user=testuser password=testpassword host=localhost dbname=testdb",
				"REDIS_NETWORK": "tcp",
				"REDIS_ADDR":    "redis15.localnet.org:6390",
				"REDIS_PWD":     "redis_password",
			},
			wantConfig: config{
				WebAuthn:     webAuthnConfig,
				Origin:       "https://localhost:8443",
				SessionKey:   []byte("secure_session_key"),
				DBConnString: "user=testuser password=testpassword host=localhost dbname=testdb",
				RedisNetwork: "tcp",
				RedisAddr:    "redis15.localnet.org:6390",
				RedisPwd:     "redis_password",
			},
		},
		{
			name:              "config using default values",
			configFileContent: configFileContent,
			configEnv: map[string]string{
				"SESSION_KEY":   base64.RawStdEncoding.EncodeToString([]byte("secure_session_key")),
				"DB_CONNSTRING": "user=testuser password=testpassword host=localhost dbname=testdb",
			},
			wantConfig: config{
				WebAuthn:     webAuthnConfig,
				Origin:       "https://localhost:8443",
				SessionKey:   []byte("secure_session_key"),
				DBConnString: "user=testuser password=testpassword host=localhost dbname=testdb",
				RedisNetwork: "tcp",
				RedisAddr:    "localhost:6379",
				RedisPwd:     "",
			},
		},
	}

	configErrorTests = []configErrorTest{
		{
			name:              "empty config file content",
			configFileContent: "",
			configEnv: map[string]string{
				"SESSION_KEY":   base64.RawStdEncoding.EncodeToString([]byte("secure_session_key")),
				"DB_CONNSTRING": "user=testuser password=testpassword host=localhost dbname=testdb",
			},
			wantErrorMsg: "failed to decode config file: EOF",
		},
		{
			name:              "invalid webauthn config",
			configFileContent: invalidWebAuthnConfigFileContent,
			configEnv: map[string]string{
				"SESSION_KEY":   base64.RawStdEncoding.EncodeToString([]byte("secure_session_key")),
				"DB_CONNSTRING": "user=testuser password=testpassword host=localhost dbname=testdb",
			},
			wantErrorMsg: "rp id is required",
		},
		{
			name:              "invalid origin",
			configFileContent: invalidOriginConfigFileContent,
			configEnv: map[string]string{
				"SESSION_KEY":   base64.RawStdEncoding.EncodeToString([]byte("secure_session_key")),
				"DB_CONNSTRING": "user=testuser password=testpassword host=localhost dbname=testdb",
			},
			wantErrorMsg: "origin is empty",
		},
		{
			name:              "empty session key",
			configFileContent: configFileContent,
			configEnv: map[string]string{
				"SESSION_KEY":   "",
				"DB_CONNSTRING": "user=testuser password=testpassword host=localhost dbname=testdb",
			},
			wantErrorMsg: "SESSION_KEY is empty",
		},
		{
			name:              "empty db connection string",
			configFileContent: configFileContent,
			configEnv: map[string]string{
				"SESSION_KEY":   base64.RawStdEncoding.EncodeToString([]byte("secure_session_key")),
				"DB_CONNSTRING": "",
			},
			wantErrorMsg: "DB_CONNSTRING is empty",
		},
	}
)

func TestNewConfig(t *testing.T) {
	for _, tc := range configTests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.configEnv {
				previousV, hasEnv := os.LookupEnv(k)
				if v != "" {
					os.Setenv(k, v)
				} else {
					os.Unsetenv(k)
				}
				if hasEnv {
					defer os.Setenv(k, previousV)
				} else {
					defer os.Unsetenv(k)
				}
			}
			c, err := newConfig(strings.NewReader(tc.configFileContent))
			if err != nil {
				t.Errorf("newConfig returns error %s", err)
			}
			if !reflect.DeepEqual(*c, tc.wantConfig) {
				t.Errorf("newConfig returns %+v, want %+v", *c, tc.wantConfig)
			}
		})
	}
}

func TestNewConfigError(t *testing.T) {
	for _, tc := range configErrorTests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.configEnv {
				previousV, hasEnv := os.LookupEnv(k)
				if v != "" {
					os.Setenv(k, v)
				} else {
					os.Unsetenv(k)
				}
				if hasEnv {
					defer os.Setenv(k, previousV)
				} else {
					defer os.Unsetenv(k)
				}
			}
			if _, err := newConfig(strings.NewReader(tc.configFileContent)); err == nil {
				t.Errorf("newConfig returns no error, want error containing substring %q", tc.wantErrorMsg)
			} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
				t.Errorf("newConfig returns error %q, want error containing substring %q", err, tc.wantErrorMsg)
			}
		})
	}
}
