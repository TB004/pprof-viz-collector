// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

// This is a simple CLI example application that starts the pprof-viz-collector on a specified port
// to allow testing/developing the library without needing to integrate it into an existing application.

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	pprofvizcollector "github.com/TB004/pprof-viz-collector"
)

func main() {
	const port = 9090
	// #nosec G101 -- example app only
	const token = "$argon2id$v=19$m=19456,t=2,p=1$q8phA9hRew+TKOoOeXxQYA$4gjeEyK3gUScgc+wt1It59DsIahCvav7sGcznebV9lo"

	log.Printf("Starting pprof-viz-collector on port %d...", port)

	// Start the agent in a separate goroutine
	go func() {
		config := pprofvizcollector.Config{
			AllowedOrigins: []string{"*"},
			HTTP3:          false,
			Token:          token,
			CertFile:       "",
			KeyFile:        "",
			Host:           "localhost",
		}
		if err := pprofvizcollector.New(port, config); err != nil {
			log.Fatalf("failed to start agent: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the application
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down pprof-viz-collector...")
}
