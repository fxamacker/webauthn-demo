// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
)

// serverResponse represents server's "ok" or "failed" response.
type serverResponse struct {
	Status       string `json:"status"`       // Status of the response, either "ok" or "failed".
	ErrorMessage string `json:"errorMessage"` // Error message if status is set to "failed".
}

const (
	statusOK                   = "ok"
	statusFailed               = "failed"
	serverResponseOKJSONString = `{ "status": "ok" }`
)

func writeFailedServerResponse(w http.ResponseWriter, httpStatusCode int, errMsg string) (int, error) {
	b, err := json.Marshal(serverResponse{statusFailed, errMsg})
	if err != nil {
		return 0, err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	return w.Write(b)
}

func writeOKServerResponse(w http.ResponseWriter) (int, error) {
	w.Header().Set("Content-Type", "application/json")
	return w.Write([]byte(serverResponseOKJSONString))
}
