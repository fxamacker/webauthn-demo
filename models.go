// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

type user struct {
	UserID        []byte
	UserName      string
	DisplayName   string
	CredentialIDs [][]byte
}

type credential struct {
	CredentialID []byte
	UserID       []byte
	Counter      uint32
	CoseKey      []byte
}

type userSession struct {
	User                 *user
	LoggedInCredentialID []byte
}
