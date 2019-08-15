// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"
)

func (s *server) handleLogout(w http.ResponseWriter, r *http.Request) {
	session, ok := r.Context().Value(contextKeyLoginSession).(*sessions.Session)
	if !ok {
		panic("Failed to get session data from context")
	}
	delete(session.Values, sessionMapKeyUserSession)
	writeOKServerResponse(w)
}

func (s *server) handleUser() http.HandlerFunc {
	type response struct {
		Status       string `json:"status"`
		Name         string `json:"name"`
		DisplayName  string `json:"displayName"`
		CredentialID string `json:"credentialID"`
		RegisteredAt string `json:"registeredAt"`
		LoggedInAt   string `json:"loggedInAt"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := r.Context().Value(contextKeyLoginSession).(*sessions.Session)
		if !ok {
			panic("Failed to get session data from context")
		}
		uSession, ok := session.Values[sessionMapKeyUserSession].(*userSession)
		if !ok {
			writeFailedServerResponse(w, http.StatusUnauthorized, "Session doesn't have user credential")
			return
		}

		registeredAt, loggedInAt, err := s.dataStore.getCredentialTimestamp(r.Context(), uSession.User.UserID, uSession.LoggedInCredentialID)
		if err != nil {
			writeFailedServerResponse(w, http.StatusInternalServerError, "Failed to find credential: "+err.Error())
			return
		}
		resp := response{
			Status:       statusOK,
			Name:         uSession.User.UserName,
			DisplayName:  uSession.User.DisplayName,
			CredentialID: base64.RawURLEncoding.EncodeToString(uSession.LoggedInCredentialID),
			RegisteredAt: registeredAt.Format("02 Jan 06 15:04 MST"),
			LoggedInAt:   loggedInAt.Format("02 Jan 06 15:04 MST"),
		}
		b, err := json.Marshal(resp)
		if err != nil {
			writeFailedServerResponse(w, http.StatusInternalServerError, "failed to json marshal response body")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}
}
