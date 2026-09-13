// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	pprofvizcollector "github.com/TB004/pprof-viz-collector"
)

const shutdownTimeout = 3 * time.Second

func main() {
	const port = 9090
	// #nosec G101 -- example app only
	const token = "$argon2id$v=19$m=19456,t=2,p=1$q8phA9hRew+TKOoOeXxQYA$4gjeEyK3gUScgc+wt1It59DsIahCvav7sGcznebV9lo"

	config := pprofvizcollector.Config{
		AllowedOrigins: []string{"*"},
		HTTP3:          false,
		Token:          token,
		CertFile:       "",
		KeyFile:        "",
		Host:           "localhost",
	}

	server, err := pprofvizcollector.New(port, config)
	if err != nil {
		log.Fatalf("failed to start agent: %v", err)
	}
	log.Printf("pprof-viz-collector listening on %s", server.Addr())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down pprof-viz-collector...")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("error during shutdown: %v", err)
	}
}
