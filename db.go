// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// dataStore interface is implemented by dbStore to query/insert/update user and credential data.
type dataStore interface {
	getUser(ctx context.Context, username string) (*user, error)
	getCredential(ctx context.Context, userID []byte, credentialID []byte) (*credential, error)
	getCredentialTimestamp(ctx context.Context, userID []byte, credentialID []byte) (registeredAt time.Time, loggedInAt time.Time, err error)
	addUserCredential(ctx context.Context, u *user, c *credential) error
	updateCredential(ctx context.Context, c *credential) error
}

type dbStore struct {
	*sql.DB
}

var (
	errNoRecords    = errors.New("webauthn/datastore: no records")
	errRecordExists = errors.New("webauthn/datastore: record exists")
)

// getUser queries user by username.  If user doesn't exist, returns errNoRecords.
func (db *dbStore) getUser(ctx context.Context, username string) (*user, error) {
	query := "SELECT users.id, display_name, credentials.id FROM users, credentials WHERE users.id = credentials.user_id AND username = $1"
	rows, err := db.QueryContext(ctx, query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	u := &user{
		UserName: username,
	}
	for rows.Next() {
		var credentialID []byte
		if err := rows.Scan(&u.UserID, &u.DisplayName, &credentialID); err != nil {
			return nil, err
		}
		u.CredentialIDs = append(u.CredentialIDs, credentialID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if u.UserID == nil {
		return nil, errNoRecords
	}
	return u, nil
}

// getCredential queries credential by user id and credential id.  If credential doesn't exist, returns errNoRecords.
func (db *dbStore) getCredential(ctx context.Context, userID []byte, credentialID []byte) (*credential, error) {
	c := &credential{
		CredentialID: credentialID,
		UserID:       userID,
	}
	query := "SELECT counter, cose_key FROM credentials WHERE user_id = $1 AND id = $2"
	row := db.QueryRowContext(ctx, query, userID, credentialID)
	if err := row.Scan(&c.Counter, &c.CoseKey); err == sql.ErrNoRows {
		return nil, errNoRecords
	} else if err != nil {
		return nil, err
	}
	return c, nil
}

// getCredentialTimestamp queries credential's registered and last logged in timestamp by user id and credential id.  If credential doesn't exist, returns errNoRecords.
func (db *dbStore) getCredentialTimestamp(ctx context.Context, userID []byte, credentialID []byte) (registeredAt time.Time, loggedInAt time.Time, err error) {
	query := "SELECT registered_at, loggedin_at FROM credentials WHERE user_id = $1 AND id = $2"
	row := db.QueryRowContext(ctx, query, userID, credentialID)
	if err = row.Scan(&registeredAt, &loggedInAt); err == sql.ErrNoRows {
		err = errNoRecords
		return
	} else if err != nil {
		return
	}
	return
}

// addUserCredential inserts user and credential.  If user exists, it skips user.  If both user and credential exist, it returns errRecordExists.
func (db *dbStore) addUserCredential(ctx context.Context, u *user, c *credential) error {
	userQuery := "INSERT INTO users (id, username, display_name) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT users_pkey DO NOTHING"
	credentialQuery := "INSERT INTO credentials (id, user_id, counter, cose_key, registered_at, loggedin_at) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT ON CONSTRAINT credentials_pkey DO NOTHING"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	_, err = tx.Exec(userQuery, u.UserID, u.UserName, u.DisplayName)
	if err != nil {
		tx.Rollback()
		return err
	}
	now := time.Now()
	res, err := tx.Exec(credentialQuery, c.CredentialID, c.UserID, c.Counter, c.CoseKey, now, now)
	if err != nil {
		tx.Rollback()
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return errRecordExists
	}
	return nil
}

// updateCredential updates credential by credential id and user id.  If credential doesn't exist, it returns errNoRecords.
func (db *dbStore) updateCredential(ctx context.Context, c *credential) error {
	query := "UPDATE credentials SET counter = $1, loggedin_at = $2 WHERE user_id = $3 and id = $4"
	res, err := db.ExecContext(ctx, query, c.Counter, time.Now(), c.UserID, c.CredentialID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return errNoRecords
	}
	return nil
}
