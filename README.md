# WebAuthn demo in Go

`webauthn demo` is a web application that supports [WebAuthn](https://w3c.github.io/webauthn/) registration and authentication.  It implements [proposed REST API](https://fidoalliance.org/specs/fido-v2.0-rd-20180702/fido-server-v2.0-rd-20180702.html#transport-binding-profile) for FIDO2 servers.

This application uses the following components:
* [`webauthn`](https://www.github.com/fxamacker/webauthn) library to parse and validate registration and authentication requests.
* Bootstrap and jQuery for web interface.
* gorilla/mux for routing and gorilla/sessions for session management.
* Redis for session storage. 
* PostgreSQL for data persistence.  

This application is a working demo, not for production.

## Installation 

```
go get github.com/fxamacker/webauthn-demo
```

## Set up and configuration 

* Set up [Redis](https://redis.io) and [PostgreSQL](https://www.postgresql.org) servers.
* Create a database in PostgreSQL and run [createtables.sql](createtables.sql) script.
* Edit [config.json](config.json) as needed.
* Build the application.
* Set environment variables: 
  * SESSION_KEY: base64 encoded session encryption key.
  * DB_CONNSTRING: PostgreSQL connection string.
  * REDIS_NETWORK: Redis network, "tcp" by default.
  * REDIS_ADDR: Redis address, "localhost:6379" by default.
  * REDIS_PWD: Redis password.
* Run the application and use command line switches to specify config file, cert file, and key file. 

For example:
```
export SESSION_KEY="U0VTU0lPTl9LRVk" 
export DB_CONNSTRING="user=testuser password=testpwd host=localhost dbname=webauthn sslmode=disable"
export REDIS_NETWORK="tcp"
export REDIS_ADDR="localhost:6379"
export REDIS_PWD=""
./webauthn-demo -config=config.json -cert=cert.pem -key=key.pem
```

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

## License 

Copyright (c) 2019 [Faye Amacker](https://github.com/fxamacker)

Licensed under [Apache License 2.0](LICENSE)
