// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down pprof-viz-collector...")
}
