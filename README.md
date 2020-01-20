[![Build Status](https://travis-ci.com/fxamacker/webauthn-demo.svg?branch=master)](https://travis-ci.com/fxamacker/webauthn-demo)
[![Go Report Card](https://goreportcard.com/badge/github.com/fxamacker/webauthn-demo)](https://goreportcard.com/report/github.com/fxamacker/webauthn-demo)
[![GitHub](https://img.shields.io/github/license/fxamacker/webauthn-demo)](https://github.com/fxamacker/webauthn-demo/blob/master/LICENSE)

# WebAuthn Server Demo (Go/Golang)

This web app is a demo for my [WebAuthn server library](https://www.github.com/fxamacker/webauthn) (fxamacker/webauthn).  It supports [WebAuthn](https://w3c.github.io/webauthn/) registration and authentication.  It implements [proposed REST API](https://fidoalliance.org/specs/fido-v2.0-rd-20180702/fido-server-v2.0-rd-20180702.html#transport-binding-profile) for FIDO2 servers.

<p align="center">
  <img src="https://user-images.githubusercontent.com/57072051/68431219-4e066780-0177-11ea-8a3f-5a137cc76cf1.png" alt="Picture of FIDO U2F key">
</p>

## What's WebAuthn?
WebAuthn (Web Authentication) is a [W3C web standard](https://www.w3.org/TR/webauthn/) for authenticating users to web-based apps and services.  It's a core component of [FIDO2](https://en.wikipedia.org/wiki/FIDO2_Project), the successor of FIDO U2F legacy protocol.

## Demo WebAuthn Server Components

* [fxamacker/webauthn](https://www.github.com/fxamacker/webauthn) to parse and validate registration and authentication requests.
* Bootstrap and jQuery for web interface.
* gorilla/mux for routing and gorilla/sessions for session management.
* Redis for session storage. 
* PostgreSQL for data persistence.  

## Current Status

This demo is not for production use because it's designed to be a demo.

## System Requirements

* Go 1.12 (or newer)
* Tested on x86_64 but it should work on other little-endian systems supported by Go.

## Installation 

```
go get github.com/fxamacker/webauthn-demo
```

## Running WebAuthn Demo Using Docker

```
$ CERTS_DIR=[folder containing cert.pem and key.pem] docker-compose up
```

WebAuthn demo runs at https://localhost:8443 on your Docker host.

## Customizing WebAuthn Demo Using Docker 

* Edit [config.json](config.json) to change WebAuthn server settings as needed.
* Edit [.env](.env) as needed:
  * CERTS_DIR: folder containing cert.pem and key.pem.
  * DB_NAME: database name (default: webauthn).
  * DB_PASSWORD: database password (default: dockerpwd).
  * DB_USER: database user (default: docker).
  * DB_DATA_DIR: database storage folder.
  * CACHE_DATA_DIR: cache storage folder.
  * SESSION_KEY: base64 encoded session encryption key.
* Run WebAuthn demo: 

```
$ docker-compose up
```

WebAuthn demo runs at https://localhost:8443 on your Docker host.

## Registration

[Registration](https://fidoalliance.org/specs/fido-v2.0-rd-20180702/fido-server-v2.0-rd-20180702.html#registration-overview) process consists of two steps: create credential creation options and register credentials.  See [signup.html](static/signup.html), [webauthn.register.js](static/js/webauthn.register.js), and [registration_handlers.go](registration_handlers.go).

**Create credential creation options:**

Server handles `/attestation/options` request by returning credential creation options (PublicKeyCredentialCreationOptions) to client.  Client then uses those options with `navigator.credentials.create()` to create new credentials.  

```
// Simplified `/attestation/options` handler from registration_handlers.go
func (s *server) handleAttestationOptions(w http.ResponseWriter, r *http.Request) {
    // Get user from datastore by username.
    u, _ := s.dataStore.getUser(r.Context(), optionsRequest.Username)
    
    // Create PublicKeyCredentialCreationOptions using webauthn library.
    creationOptions, _ := webauthn.NewAttestationOptions(s.webAuthnConfig, &webauthn.User{ID: u.UserID, Name: u.UserName, DisplayName: u.DisplayName, CredentialIDs: u.CredentialIDs})

    // Save creationOptions and user info in session to verify new credential later.
    session.Values[WebAuthnCreationOptions] = creationOptions
    session.Values[UserSession] = &userSession{User: u}

    // Write creationOptions to response.
}
```

**Register credentials:**

Server verifies and registers new credentials received via `/attestation/result`.

```
// Simplified `/attestation/result` handler from registration_handlers.go
func (s *server) handleAttestationResult(w http.ResponseWriter, r *http.Request) {
    // Get saved creationOptions and user info from session.

    // Parse and verify credential in request body.
    credentialAttestation, _ := webauthn.ParseAttestation(r.Body)
    expected := &webauthn.AttestationExpectedData{
	Origin:           s.rpOrigin,
	RPID:             savedCreationOptions.RP.ID,
	CredentialAlgs:   credentialAlgs,
	Challenge:        base64.RawURLEncoding.EncodeToString(savedCreationOptions.Challenge),
	UserVerification: savedCreationOptions.AuthenticatorSelection.UserVerification,
    }    
    _, _, err = webauthn.VerifyAttestation(credentialAttestation, expected)

   // Save user credential in datastore.
   c := &credential{
	CredentialID: credentialAttestation.RawID,
	UserID:       uSession.User.UserID,
	Counter:      credentialAttestation.AuthnData.Counter,
	CoseKey:      credentialAttestation.AuthnData.Credential.Raw,
   }    
   err = s.dataStore.addUserCredential(r.Context(), uSession.User, c)

   // Write "ok" response. 
}
```

## Authentication

[Authentication](https://fidoalliance.org/specs/fido-v2.0-rd-20180702/fido-server-v2.0-rd-20180702.html#authentication-overview) process requires two steps: create credential request options and verify credentials.  See [signin.html](static/signin.html), [webauthn.authn.js](static/js/webauthn.authn.js), and [authentication_handlers.go](authentication_handlers.go).

**Create credential request options:**

Server handles `/assertion/options` request by returning credential request options (PublicKeyCredentialRequestOptions) to client.  Client then uses those options with `navigator.credentials.get()` to get existing credentials.  

```
// Simplified `/assertion/options` handler from authentication_handlers.go
func (s *server) handleAssertionOptions(w http.ResponseWriter, r *http.Request) {
    // Get user from datastore by username.
    u, _ := s.dataStore.getUser(r.Context(), optionsRequest.Username)
    
    // Create PublicKeyCredentialRequestOptions using webauthn library.
    requestOptions, _ := webauthn.NewAssertionOptions(s.webAuthnConfig, &webauthn.User{ID: u.UserID, Name: u.UserName, DisplayName: u.DisplayName, CredentialIDs: u.CredentialIDs})

    // Save requestOptions and user info in session to verify credential later.
    session.Values[WebAuthnRequestOptions] = requestOptions
    session.Values[UserSession] = &userSession{User: u}

    // Write requestOptions to response.
}
```

**Verify credentials:**

Server verifies credentials received via `/asssertion/result`.

```
// Simplified `/assertion/result` handler from authentication_handlers.go
func (s *server) handleAssertionResult(w http.ResponseWriter, r *http.Request) {
    // Get saved requestOptions and user info.

    // Parse credential in request body.
    credentialAssertion, _ := webauthn.ParseAssertion(r.Body)

    // Get credential from datastore by received credential ID.
    c, _ := s.dataStore.getCredential(r.Context(), uSession.User.UserID, credentialAssertion.RawID)

    // Verify credential.
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
    err = webauthn.VerifyAssertion(credentialAssertion, expected)

    // Update authenticator counter in datastore.
    c.Counter = credentialAssertion.AuthnData.Counter
    err = s.dataStore.updateCredential(r.Context(), c)

    // Write "ok" response. 
}
```

## Security Policy

Security fixes are provided for the latest released version.

To report security vulnerabilities, please email faye.github@gmail.com and allow time for the problem to be resolved before reporting it to the public.

## License 

Copyright (c) 2019-present [Faye Amacker](https://github.com/fxamacker)

fxamacker/webauthn-demo is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.
