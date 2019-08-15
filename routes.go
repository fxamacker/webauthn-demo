// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"net/http"
)

func (s *server) routes() {
	s.router.HandleFunc("/attestation/options", s.handleAuthnSession(s.handleAttestationOptions())).Methods("POST")

	s.router.HandleFunc("/assertion/options", s.handleAuthnSession(s.handleAssertionOptions())).Methods("POST")

	s.router.HandleFunc("/attestation/result", s.handleAuthnSession(s.handleAttestationResult)).Methods("POST")

	s.router.HandleFunc("/assertion/result", s.handleAuthnSession(s.handleAssertionResult)).Methods("POST")

	s.router.HandleFunc("/logout", s.handleAuthnSession(s.handleLogout)).Methods("GET")

	s.router.HandleFunc("/user", s.loggedInUserOnly(s.handleSession(nil, []string{sessionNameLoginSession}, s.handleUser()))).Methods("GET")

	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))
}
