// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"database/sql"
	"encoding/gob"
	"errors"
	"net/url"
	"strings"

	"github.com/fxamacker/webauthn"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	redistore "gopkg.in/boj/redistore.v1"
)

type server struct {
	webAuthnConfig *webauthn.Config
	rpOrigin       string
	dataStore      dataStore
	sessionStore   sessions.Store
	router         *mux.Router
}

func newServer(c *config) (*server, error) {
	// Initialize origin.
	origin := c.Origin
	u, err := url.Parse(origin)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(u.Scheme) != "https" {
		return nil, errors.New("WebAuthn origin must be https")
	}

	// Initialize data store.
	db, err := sql.Open("postgres", c.DBConnString)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	dataStore := &dbStore{db}

	// Initialize session store.
	rediStore, err := redistore.NewRediStore(10, c.RedisNetwork, c.RedisAddr, c.RedisPwd, c.SessionKey)
	if err != nil {
		return nil, err
	}
	rediStore.SetMaxAge(60 * 5) // expires after 5 minutes
	gob.Register(&userSession{})
	gob.Register(&webauthn.PublicKeyCredentialCreationOptions{})
	gob.Register(&webauthn.PublicKeyCredentialRequestOptions{})

	return &server{
		webAuthnConfig: c.WebAuthn,
		rpOrigin:       origin,
		dataStore:      dataStore,
		sessionStore:   rediStore,
		router:         mux.NewRouter(),
	}, nil
}

func (s *server) close() {
	if dbStore, ok := s.dataStore.(*dbStore); ok {
		dbStore.Close()
	}
	if rediStore, ok := s.sessionStore.(*redistore.RediStore); ok {
		rediStore.Close()
	}
}
