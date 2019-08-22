// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/fxamacker/webauthn/androidkeystore"
	_ "github.com/fxamacker/webauthn/androidsafetynet"
	_ "github.com/fxamacker/webauthn/fidou2f"
	_ "github.com/fxamacker/webauthn/packed"
	_ "github.com/fxamacker/webauthn/tpm"

	_ "github.com/lib/pq"
)

type contextKey string

const (
	sessionNameLoginSession              string = "LoginSession"            // session store name for login session
	sessionMapKeyUserSession             string = "UserSession"             // session map key for *userSession
	sessionMapKeyWebAuthnCreationOptions string = "WebAuthnCreationOptions" // session map key for *webauthn.PublicKeyCredentialCreationOptions
	sessionMapKeyWebAuthnRequestOptions  string = "WebAuthnRequestOptions"  // session map key for *webauthn.PublicKeyCredentialRequestOptions

	contextKeyLoginSession contextKey = contextKey(sessionNameLoginSession) // context key for login session
)

func main() {
	var configFilePath, certFilePath, keyFilePath string
	flag.StringVar(&configFilePath, "config", "", "config file path")
	flag.StringVar(&certFilePath, "cert", "", "cert file path")
	flag.StringVar(&keyFilePath, "key", "", "key file path")

	flag.Parse()

	if configFilePath == "" || certFilePath == "" || keyFilePath == "" {
		flag.Usage()
		return
	}

	configFile, err := os.Open(configFilePath)
	if err != nil {
		fmt.Println("Failed to open config file: " + err.Error())
		flag.Usage()
		return
	}

	if _, err := os.Stat(certFilePath); os.IsNotExist(err) {
		fmt.Println("Cert file " + certFilePath + " doesn't exist.")
		flag.Usage()
		return
	}

	if _, err := os.Stat(keyFilePath); os.IsNotExist(err) {
		fmt.Println("Key file " + keyFilePath + " doesn't exist.")
		flag.Usage()
		return
	}

	c, err := newConfig(configFile)
	if err != nil {
		panic(err)
	}

	s, err := newServer(c)
	if err != nil {
		panic(err)
	}
	defer s.close()

	s.routes()

	server := &http.Server{
		Addr:         s.httpServerAddr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      s.router,
	}

	go func() {
		if err := server.ListenAndServeTLS(certFilePath, keyFilePath); err != http.ErrServerClosed {
			// Error starting or closing listener
			log.Printf("HTTP server ListenAndServeTLS: %v\n", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		// Error from closing listeners, or context timeout
		log.Printf("HTTP server Shutdown: %v\n", err)
	}
}
