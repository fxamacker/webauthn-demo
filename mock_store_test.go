// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"context"
	"net/http"
	"time"

	"github.com/fxamacker/webauthn"

	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/mock"
)

type MockDataStore struct {
	mock.Mock
}

func (m *MockDataStore) getUser(ctx context.Context, username string) (*user, error) {
	args := m.Called(ctx, username)
	if args.Get(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user), args.Error(1)
}

func (m *MockDataStore) getCredential(ctx context.Context, userID []byte, credentialID []byte) (*credential, error) {
	args := m.Called(ctx, userID, credentialID)
	if args.Get(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*credential), args.Error(1)
}

func (m *MockDataStore) getCredentialTimestamp(ctx context.Context, userID []byte, credentialID []byte) (registeredAt time.Time, loggedInAt time.Time, err error) {
	args := m.Called(ctx, userID, credentialID)
	if args.Get(2) != nil {
		return time.Time{}, time.Time{}, args.Error(1)
	}
	return args.Get(0).(time.Time), args.Get(1).(time.Time), args.Error(2)
}

func (m *MockDataStore) addUserCredential(ctx context.Context, u *user, c *credential) error {
	args := m.Called(ctx, u, c)
	return args.Error(0)
}

func (m *MockDataStore) updateCredential(ctx context.Context, c *credential) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

type MockSessionStore struct {
	mock.Mock
}

func (m *MockSessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	args := m.Called(r, name)
	return args.Get(0).(*sessions.Session), args.Error(1)
}

func (m *MockSessionStore) New(r *http.Request, name string) (*sessions.Session, error) {
	args := m.Called(r, name)
	return args.Get(0).(*sessions.Session), args.Error(1)
}

func (m *MockSessionStore) Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	args := m.Called(r, w, removeSessionRandomData(s))
	return args.Error(0)
}

func initDataStoreGetUserNone(mockDataStore *MockDataStore) {
	mockDataStore.On("getUser", mock.Anything, mock.Anything).Return(nil, errNoRecords)
}

func initDataStoreGetUser(mockDataStore *MockDataStore) {
	mockDataStore.On("getUser", mock.Anything, mockExistingUser.UserName).Return(mockExistingUser, nil).Once()
}

func initDataStoreAddUserCredentialNotCalled(mockDataStore *MockDataStore) {
	mockDataStore.On("addUserCredential", mock.Anything, mock.Anything, mock.Anything).Times(0)
}

func initDataStoreAddUserCredential(mockDataStore *MockDataStore) {
	mockDataStore.On("addUserCredential", mock.Anything, mockNewUser, mockCredential).Return(nil).Once()
}

func initDataStoreGetAndUpdateCredential(mockDataStore *MockDataStore) {
	c2 := *mockCredential // make a copy of credentialMock
	c2.Counter = 0
	mockDataStore.On("getCredential", mock.Anything, mockCredential.UserID, mockCredential.CredentialID).Return(mockCredential, nil).Once()
	mockDataStore.On("updateCredential", mock.Anything, &c2).Return(nil).Maybe()
}

func initDataStoreGetAndUpdateCredentialNotCalled(mockDataStore *MockDataStore) {
	mockDataStore.On("getCredential", mock.Anything, mock.Anything, mock.Anything).Times(0)
	mockDataStore.On("updateCredential", mock.Anything, mock.Anything).Times(0)
}

func initDataStoreGetCredentialTimestamp(mockDataStore *MockDataStore) {
	t1 := time.Date(2009, time.January, 1, 1, 0, 0, 0, time.UTC)
	t2 := time.Date(2009, time.February, 1, 1, 0, 0, 0, time.UTC)
	mockDataStore.On("getCredentialTimestamp", mock.Anything, mock.Anything, mock.Anything).Return(t1, t2, nil).Once()
}

func initSessionStore(getSession getSessionFunc, saveSession getSessionFunc) func(mockSessionStore *MockSessionStore) {
	return func(mockSessionStore *MockSessionStore) {
		mockSessionStore.On("Get", mock.Anything, sessionNameLoginSession).Return(getSession(mockSessionStore), nil)
		mockSessionStore.On("Save", mock.Anything, mock.Anything, removeSessionRandomData(saveSession(mockSessionStore))).Return(nil)
	}
}

func removeSessionRandomData(session *sessions.Session) *sessions.Session {
	if s, ok := session.Values[sessionMapKeyUserSession].(*userSession); ok {
		if s.User.CredentialIDs == nil { // new user
			s.User.UserID = nil
		}
	}
	if s, ok := session.Values[sessionMapKeyWebAuthnCreationOptions].(*webauthn.PublicKeyCredentialCreationOptions); ok {
		s.Challenge = nil
		s.User.ID = nil
	}
	if s, ok := session.Values[sessionMapKeyWebAuthnRequestOptions].(*webauthn.PublicKeyCredentialRequestOptions); ok {
		s.Challenge = nil
	}
	return session
}
