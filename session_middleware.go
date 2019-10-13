// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
)

// sessionHandler is a middleware handler that stores session data in request context and
// saves read-write data in sessionStore if response code is 200.
type sessionHandler struct {
	rwSessionNames []string
	rSessionNames  []string
	next           http.HandlerFunc
	server         *server
}

func (m *sessionHandler) storeSessionInContext(ctx context.Context, r *http.Request, sessionNames []string) (context.Context, error) {
	for _, sessionName := range sessionNames {
		// Get session data.
		session, err := m.server.sessionStore.Get(r, sessionName)
		if err != nil {
			return nil, errors.New("failed to retrieve session \"" + sessionName + "\": " + err.Error())
		}
		// Store session in context.
		ctx = context.WithValue(ctx, contextKey(sessionName), session)
	}
	return ctx, nil
}

// ServeHTTP stores session data in request context and saves read-write data in sessionStore
// if response code is 200.
func (m *sessionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	ctx := r.Context()

	// Get session data and store it in context.
	if ctx, err = m.storeSessionInContext(ctx, r, m.rwSessionNames); err != nil {
		writeFailedServerResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if ctx, err = m.storeSessionInContext(ctx, r, m.rSessionNames); err != nil {
		writeFailedServerResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Store context for next handler.
	r = r.WithContext(ctx)

	// Save session by hijacking ResponseWriter.
	if len(m.rwSessionNames) > 0 {
		w = &sessionWriter{ResponseWriter: w, r: r, rwSession: m.rwSessionNames}
	}

	m.next(w, r)
}

// sessionWriter hijacks ResponseWriter and saves read-write context data in sessionStore if response code is 200.
type sessionWriter struct {
	http.ResponseWriter
	r         *http.Request
	status    int
	rwSession []string
}

func (w *sessionWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *sessionWriter) Write(b []byte) (int, error) {
	if w.status != 0 {
		return w.ResponseWriter.Write(b)
	}

	for _, sessionName := range w.rwSession {
		session, ok := w.r.Context().Value(contextKey(sessionName)).(*sessions.Session)
		if !ok {
			panic("Failed to get session " + sessionName + " from context")
		}
		if err := session.Save(w.r, w.ResponseWriter); err != nil {
			return writeFailedServerResponse(w.ResponseWriter, http.StatusInternalServerError, "Failed to save session "+session.Name()+": "+err.Error())
		}
	}

	return w.ResponseWriter.Write(b)
}

// handleSession returns a session middleware handler.
func (s *server) handleSession(rwSessionNames []string, rSessionNames []string, next http.HandlerFunc) http.HandlerFunc {
	h := &sessionHandler{
		rwSessionNames: rwSessionNames,
		rSessionNames:  rSessionNames,
		next:           next,
		server:         s,
	}
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

// handleAuthnSession returns a session middleware handler used by registration, authentication, and logout handlers.
func (s *server) handleAuthnSession(next http.HandlerFunc) http.HandlerFunc {
	return s.handleSession([]string{sessionNameLoginSession}, []string{sessionNameLoginSession}, next)
}

// loggedInUserOnly returns a handler that responds with a 401 unauthorized error if user is not logged in.
func (s *server) loggedInUserOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		loginSession, err := s.sessionStore.Get(r, sessionNameLoginSession)
		if err != nil {
			writeFailedServerResponse(w, http.StatusInternalServerError, "Failed to retrieve session \""+sessionNameLoginSession+"\": "+err.Error())
			return
		}
		if u, ok := loginSession.Values[sessionMapKeyUserSession].(*userSession); !ok || len(u.LoggedInCredentialID) == 0 {
			writeFailedServerResponse(w, http.StatusUnauthorized, "User is not logged in")
			return
		}
		next(w, r)
	}
}
