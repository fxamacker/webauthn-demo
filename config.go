// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/fxamacker/webauthn"
)

// config has configuration data from config file and environment variables.
type config struct {
	WebAuthn     *webauthn.Config
	Origin       string
	SessionKey   []byte
	DBConnString string
	RedisNetwork string
	RedisAddr    string
	RedisPwd     string
}

func newConfig(configFile io.Reader) (*config, error) {
	var err error
	c := &config{}
	if err := json.NewDecoder(configFile).Decode(c); err != nil {
		return nil, errors.New("failed to decode config file: " + err.Error())
	}
	if err := c.WebAuthn.Valid(); err != nil {
		return nil, err
	}
	if c.Origin == "" {
		return nil, errors.New("origin is empty")
	}
	c.SessionKey, err = base64.RawStdEncoding.DecodeString(os.Getenv("SESSION_KEY"))
	if err != nil {
		return nil, errors.New("failed to base64 decode session key: " + err.Error())
	}
	if len(c.SessionKey) == 0 {
		return nil, errors.New("SESSION_KEY is empty")
	}
	c.DBConnString = os.Getenv("DB_CONNSTRING")
	if c.DBConnString == "" {
		return nil, errors.New("DB_CONNSTRING is empty")
	}
	c.RedisNetwork = os.Getenv("REDIS_NETWORK")
	if c.RedisNetwork == "" {
		c.RedisNetwork = "tcp"
	}
	c.RedisAddr = os.Getenv("REDIS_ADDR")
	if c.RedisAddr == "" {
		c.RedisAddr = "localhost:6379"
	}
	c.RedisPwd = os.Getenv("REDIS_PWD")

	return c, nil
}
