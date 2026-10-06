// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package pprofvizcollector

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/quic-go/quic-go/http3"
)

const readHeaderTimeout = 5 * time.Second
const readTimeout = 10 * time.Second
const idleTimeout = 120 * time.Second

func newHTTPServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		IdleTimeout:       idleTimeout,
	}
}

// Server is a running instance started by New. The zero value is not usable;
// obtain one from New.
type Server struct {
	udpConn      net.PacketConn // nil unless HTTP3 was enabled
	shutdownErr  error
	httpServer   *http.Server
	h3Server     *http3.Server // nil unless HTTP3 was enabled
	addr         string
	shutdownOnce sync.Once
}

// Addr returns the address the server is listening on, e.g. "127.0.0.1:57621".
// Useful when New was called with port 0 to let the OS pick a free port.
func (s *Server) Addr() string {
	return s.addr
}

// Shutdown gracefully stops the server: it stops accepting new connections
// and waits for in-flight requests to finish, or for ctx to be canceled.
// Safe to call more than once; only the first call does any work.
func (s *Server) Shutdown(ctx context.Context) error {
	s.shutdownOnce.Do(func() {
		var errs []error
		if s.h3Server != nil {
			errs = append(errs, s.h3Server.Shutdown(ctx))
		}
		errs = append(errs, s.httpServer.Shutdown(ctx))
		if s.udpConn != nil {
			// http3.Server.Serve documents that closing the server does not
			// close the connection it was given; close it ourselves.
			errs = append(errs, s.udpConn.Close())
		}
		s.shutdownErr = errors.Join(errs...)
	})
	return s.shutdownErr
}

// selfDestruct shuts the whole Server down in the background after an
// unexpected listener failure.
func (s *Server) selfDestruct() {
	go func() {
		// #nosec G104 -- best-effort cleanup; no caller to report the error to
		_ = s.Shutdown(context.Background())
	}()
}
