// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type DBTestSuite struct {
	suite.Suite
	dbStore
}

type credentialsByID []credential

func (c credentialsByID) Len() int { return len(c) }

func (c credentialsByID) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (c credentialsByID) Less(i, j int) bool {
	return bytes.Compare(c[i].CredentialID, c[j].CredentialID) <= 0
}

type usersByID []user

func (u usersByID) Len() int { return len(u) }

func (u usersByID) Swap(i, j int) { u[i], u[j] = u[j], u[i] }

func (u usersByID) Less(i, j int) bool {
	return bytes.Compare(u[i].UserID, u[j].UserID) <= 0
}

var (
	userNotExist = user{
		UserName: "user_not_exist",
	}
	// User with one credential
	user1 = user{
		UserID:      []byte{117, 115, 101, 104, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		UserName:    "User1",
		DisplayName: "User1 display name",
		CredentialIDs: [][]byte{
			[]byte{99, 114, 101, 100, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		},
	}
	// User with two credentials
	user2 = user{
		UserID:      []byte{117, 115, 101, 104, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
		UserName:    "User2",
		DisplayName: "User2 display name",
		CredentialIDs: [][]byte{
			[]byte{99, 114, 101, 100, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			[]byte{99, 114, 101, 100, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3},
		},
	}
	credentialNotExist = credential{
		CredentialID: []byte{99, 114, 101, 100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		UserID:       []byte{117, 115, 101, 104, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	credential1 = credential{
		CredentialID: []byte{99, 114, 101, 100, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		UserID:       []byte{117, 115, 101, 104, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		Counter:      1,
		CoseKey:      []byte{1, 2, 3},
	}
	credential2 = credential{
		CredentialID: []byte{99, 114, 101, 100, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
		UserID:       []byte{117, 115, 101, 104, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
		Counter:      2,
		CoseKey:      []byte{1, 2, 3},
	}
	credential3 = credential{
		CredentialID: []byte{99, 114, 101, 100, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3},
		UserID:       []byte{117, 115, 101, 104, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
		Counter:      3,
		CoseKey:      []byte{1, 2, 3},
	}
	users       = []user{user1, user2}
	credentials = []credential{credential1, credential2, credential3}
)

func (suite *DBTestSuite) SetupSuite() {
	connString := os.Getenv("DB_CONNSTRING")
	if connString == "" {
		panic("Missing env variable DB_CONNSTRING")
	}
	db, err := sql.Open("postgres", connString)
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}
	suite.dbStore.DB = db
}

func (suite *DBTestSuite) TearDownSuite() {
	suite.dbStore.Close()
}

func (suite *DBTestSuite) SetupTest() {
	_, err := suite.dbStore.Exec("DELETE FROM credentials")
	if err != nil {
		panic(err)
	}
	_, err = suite.dbStore.Exec("DELETE FROM users")
	if err != nil {
		panic(err)
	}
}

func (suite *DBTestSuite) seedUserCredentialTables(ctx context.Context) {
	insertUserStmt, err := suite.dbStore.PrepareContext(ctx, "INSERT INTO users (id, username, display_name) VALUES ($1, $2, $3)")
	if err != nil {
		panic(err)
	}
	insertCredentialStmt, err := suite.dbStore.PrepareContext(ctx, "INSERT INTO credentials (id, user_id, counter, cose_key, registered_at, loggedin_at) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		panic(err)
	}
	for _, u := range users {
		_, err := insertUserStmt.ExecContext(ctx, u.UserID, u.UserName, u.DisplayName)
		if err != nil {
			panic(err)
		}
	}
	for _, c := range credentials {
		_, err := insertCredentialStmt.ExecContext(ctx, c.CredentialID, c.UserID, c.Counter, c.CoseKey, time.Now(), time.Now())
		if err != nil {
			panic(err)
		}
	}
}

func (suite *DBTestSuite) queryUserCredentialTables(ctx context.Context) ([]user, []credential) {
	var users []user
	rows, err := suite.dbStore.Query("SELECT id, username, display_name FROM users")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var u user
		if err := rows.Scan(&u.UserID, &u.UserName, &u.DisplayName); err != nil {
			panic(err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	var credentials []credential
	rows, err = suite.dbStore.Query("SELECT id, user_id, counter, cose_key FROM credentials")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var c credential
		if err := rows.Scan(&c.CredentialID, &c.UserID, &c.Counter, &c.CoseKey); err != nil {
			panic(err)
		}
		credentials = append(credentials, c)
		for i := 0; i < len(users); i++ {
			if bytes.Equal(c.UserID, users[i].UserID) {
				users[i].CredentialIDs = append(users[i].CredentialIDs, c.CredentialID)
			}
		}
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	return users, credentials
}

func (suite *DBTestSuite) TestGetUser() {
	ctx := context.Background()

	suite.seedUserCredentialTables(ctx)

	user, err := suite.dbStore.getUser(ctx, userNotExist.UserName)
	if err == nil || err != errNoRecords {
		suite.T().Errorf("(*dbstore).getUser(%s) returns error %q, want error %q", userNotExist.UserName, err, errNoRecords)
	}
	if user != nil {
		suite.T().Errorf("(*dbstore).getUser(%s) returns user %+v, want nil", userNotExist.UserName, user)
	}

	for _, expectedUser := range users {
		user, err := suite.dbStore.getUser(ctx, expectedUser.UserName)
		if err != nil {
			suite.T().Errorf("(*dbstore).getUser(%s) returns error %q", expectedUser.UserName, err)
		}
		if !reflect.DeepEqual(*user, expectedUser) {
			suite.T().Errorf("(*dbstore).getUser(%s) returns user %+v, want %+v", expectedUser.UserName, user, expectedUser)
		}
	}
}

func (suite *DBTestSuite) TestGetCredential() {
	ctx := context.Background()

	suite.seedUserCredentialTables(ctx)

	c, err := suite.dbStore.getCredential(ctx, credentialNotExist.UserID, credentialNotExist.CredentialID)
	if err == nil || err != errNoRecords {
		suite.T().Errorf("(*dbstore).getCredential(%v, %v) returns error %q, want error %q", credentialNotExist.UserID, credentialNotExist.CredentialID, err, errNoRecords)
	}
	if c != nil {
		suite.T().Errorf("(*dbstore).getCredential(%v, %v) returns credential %+v, want nil", credentialNotExist.UserID, credentialNotExist.CredentialID, c)
	}

	for _, expectedCredential := range credentials {
		c, err := suite.dbStore.getCredential(ctx, expectedCredential.UserID, expectedCredential.CredentialID)
		if err != nil {
			suite.T().Errorf("(*dbstore).getCredential(%v, %v) returns error %q", expectedCredential.UserID, expectedCredential.CredentialID, err)
		}
		if !reflect.DeepEqual(*c, expectedCredential) {
			suite.T().Errorf("(*dbstore).getCredential(%v, %v) returns credential %+v, want %+v", expectedCredential.UserID, expectedCredential.CredentialID, c, expectedCredential)
		}
	}
}

func (suite *DBTestSuite) TestAddUserCredential() {
	ctx := context.Background()

	// User does not exist, add user record and credential record
	if err := suite.dbStore.addUserCredential(ctx, &user2, &credential2); err != nil {
		suite.T().Errorf("(*dbstore).addUserAndCredential(%+v, %+v) returns error %q", user2, credential2, err)
		return
	}
	// User exists, add credential record
	if err := suite.dbStore.addUserCredential(ctx, &user2, &credential3); err != nil {
		suite.T().Errorf("(*dbstore).addUserAndCredential(%+v, %+v) returns error %q", user2, credential3, err)
		return
	}
	// User and credential exist, return error
	if err := suite.dbStore.addUserCredential(ctx, &user2, &credential3); err == nil || err != errRecordExists {
		suite.T().Errorf("(*dbstore).addUserAndCredential(%+v, %+v) returns error %q, want error %q", user2, credential3, err, errRecordExists)
		return
	}

	usersFromDB, credentialsFromDB := suite.queryUserCredentialTables(ctx)

	c := []credential{credential2, credential3}

	sort.Sort(credentialsByID(credentialsFromDB))
	sort.Sort(credentialsByID(c))

	if len(usersFromDB) != 1 {
		suite.T().Errorf("Got %d user records, want %d records", len(usersFromDB), 1)
	}
	if !reflect.DeepEqual(user2, usersFromDB[0]) {
		suite.T().Errorf("Got user %+v, want %+v", usersFromDB[0], user2)
	}
	if len(credentialsFromDB) != 2 {
		suite.T().Errorf("Got %d credential records, want %d records", len(credentialsFromDB), 2)
	}
	if !reflect.DeepEqual(c, credentialsFromDB) {
		suite.T().Errorf("Got credential %+v, want %+v", credentialsFromDB, c)
	}
}

func (suite *DBTestSuite) TestUpdateCredential() {
	ctx := context.Background()

	suite.seedUserCredentialTables(ctx)

	err := suite.dbStore.updateCredential(ctx, &credentialNotExist)
	if err == nil || err != errNoRecords {
		suite.T().Errorf("(*dbstore).updateCredential(%v, %v) returns error %q, want error %q", credentialNotExist.UserID, credentialNotExist.CredentialID, err, errNoRecords)
	}

	newCredentials := make([]credential, len(credentials))
	copy(newCredentials, credentials)
	for i := 0; i < len(newCredentials); i++ {
		newCredentials[i].Counter++
		err := suite.dbStore.updateCredential(ctx, &newCredentials[i])
		if err != nil {
			suite.T().Errorf("(*dbstore).updateCredential(%v) returns error %q", credentials[i], err)
		}
	}

	_, credentialsFromDB := suite.queryUserCredentialTables(ctx)

	sort.Sort(credentialsByID(credentialsFromDB))
	sort.Sort(credentialsByID(newCredentials))

	if !reflect.DeepEqual(credentialsFromDB, newCredentials) {
		suite.T().Errorf("Got credential %+v, want %+v", credentialsFromDB, newCredentials)
	}
}

func TestDBTestSuite(t *testing.T) {
	suite.Run(t, new(DBTestSuite))
}
