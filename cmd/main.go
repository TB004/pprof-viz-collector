// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

// This is a simple CLI example application that starts the pprof-viz-agent on a specified port
// to allow testing/developing the library without needing to integrate it into an existing application.

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	pprofvizagent "github.com/TB004/pprof-viz-agent"
)

func main() {
	const port = 9090
	log.Printf("Starting pprof-viz-agent on port %d...", port)

	// Start the agent in a separate goroutine
	go func() {
		config := pprofvizagent.Config{
			AllowedOrigins: []string{"*"},
		}
		if err := pprofvizagent.New(port, config); err != nil {
			log.Fatalf("failed to start agent: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the application
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down pprof-viz-agent...")
}
