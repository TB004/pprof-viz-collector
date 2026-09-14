// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package pprofvizcollector

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"github.com/TB004/pprof-viz-collector/internal"
	"github.com/TB004/pprof-viz-collector/internal/proto/protoconnect"
	"github.com/quic-go/quic-go/http3"
	"github.com/rs/cors"
)

type Config struct {
	// Token, if non-empty, is required on every RPC call via Authorization: Bearer <token>.
	// Applies to both Bind and New.
	Token string
	// AllowedOrigins is the list of origins permitted by CORS (e.g. ["https://app.example.com"]).
	// Use ["*"] to allow all origins. An empty slice disables CORS middleware entirely.
	// Applies to both Bind and New.
	AllowedOrigins []string
	// CertFile is the path to the TLS certificate file. Enables TLS when set. Only used by New.
	CertFile string
	// KeyFile is the path to the TLS private key file. Must be set together with CertFile. Only used by New.
	KeyFile string
	// HTTP3 also serves HTTP/3 over QUIC in parallel with the TLS/TCP server.
	// Requires CertFile and KeyFile. Only used by New.
	HTTP3 bool
	// Host is the network interface to bind to (e.g. "127.0.0.1" or "0.0.0.0").
	// Defaults to "localhost" when empty. Binding to a non-loopback address
	// exposes the agent externally; pair it with TLS and a Token. Only used by New.
	Host string
}

func newCORSHandler(origins []string, h http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins: origins,
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: append(connectcors.AllowedHeaders(), "Authorization"),
		ExposedHeaders: connectcors.ExposedHeaders(),
	}).Handler(h)
}

// Every RPC this service exposes takes either Empty or a two-field message, so
// this is generous headroom, not a real limit.
const maxRequestBytes = 4096

func recoverPanic(_ context.Context, _ connect.Spec, _ http.Header, r any) error {
	return connect.NewError(connect.CodeInternal, fmt.Errorf("panic: %v", r))
}

func handlerOpts(cfg Config) []connect.HandlerOption {
	opts := []connect.HandlerOption{
		connect.WithCompression(internal.CompressionZstd, internal.NewZstdDecompressor, internal.NewZstdCompressor),
		connect.WithCompressMinBytes(256),
		connect.WithReadMaxBytes(maxRequestBytes),
		connect.WithRecover(recoverPanic),
	}
	if cfg.Token != "" {
		opts = append(opts, connect.WithInterceptors(internal.NewTokenInterceptor(cfg.Token)))
	}
	return opts
}

// Bind registers the CollectorService handler to the provided http.ServeMux.
// cfg.CertFile, cfg.KeyFile, and cfg.HTTP3 are ignored; those are server-level concerns managed externally.
func Bind(mux *http.ServeMux, cfg Config) {
	path, handler := protoconnect.NewCollectorServiceHandler(internal.CollectorServiceHandler{}, handlerOpts(cfg)...)
	if len(cfg.AllowedOrigins) > 0 {
		handler = newCORSHandler(cfg.AllowedOrigins, handler)
	}
	mux.Handle(path, handler)
}

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	idleTimeout       = 120 * time.Second
)

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
	httpServer *http.Server
	h3Server   *http3.Server  // nil unless HTTP3 was enabled
	udpConn    net.PacketConn // nil unless HTTP3 was enabled

	addr string

	shutdownOnce sync.Once
	shutdownErr  error
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

// New starts an HTTP server on the specified port serving the CollectorService
// and returns immediately; it does not block. Call Shutdown on the returned
// Server to stop it.
//
// Without TLS, it supports HTTP/2 over cleartext and HTTP/1.1 (necessary for browsers).
// With CertFile and KeyFile set, it supports HTTP/2 over TLS (minimum TLS 1.2).
// With HTTP3 also set, it additionally serves HTTP/3 over QUIC in parallel, with Alt-Svc
// headers on TLS responses so clients can upgrade automatically.
func New(port int, cfg Config) (*Server, error) {
	if cfg.HTTP3 && (cfg.CertFile == "" || cfg.KeyFile == "") {
		return nil, errors.New("HTTP3 requires CertFile and KeyFile to be configured")
	}

	mux := http.NewServeMux()
	path, handler := protoconnect.NewCollectorServiceHandler(internal.CollectorServiceHandler{}, handlerOpts(cfg)...)
	mux.Handle(path, handler)

	host := cfg.Host
	if host == "" {
		host = "localhost"
	}
	addr := net.JoinHostPort(host, strconv.Itoa(port))

	// Wrap mux with CORS middleware if origins are configured.
	var h http.Handler = mux
	if len(cfg.AllowedOrigins) > 0 {
		h = newCORSHandler(cfg.AllowedOrigins, mux)
	}

	// Cleartext HTTP/1.1 and HTTP/2 (h2c) when no certificate is configured.
	// HTTP/1.1 must stay enabled: no browser speaks h2c.
	if cfg.CertFile == "" {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, err
		}

		protos := new(http.Protocols)
		protos.SetHTTP1(true)
		protos.SetHTTP2(false)
		protos.SetUnencryptedHTTP2(true)

		s := newHTTPServer(ln.Addr().String(), h)
		s.Protocols = protos

		srv := &Server{httpServer: s, addr: ln.Addr().String()}
		go func() {
			if err := s.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
				srv.selfDestruct()
			}
		}()
		return srv, nil
	}

	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}

	// HTTP/2-over-TLS.
	if !cfg.HTTP3 {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, err
		}

		s := newHTTPServer(ln.Addr().String(), h)
		s.TLSConfig = tlsConfig

		srv := &Server{httpServer: s, addr: ln.Addr().String()}
		go func() {
			if err := s.ServeTLS(ln, cfg.CertFile, cfg.KeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
				srv.selfDestruct()
			}
		}()
		return srv, nil
	}

	// TLS with HTTP/3: serve HTTP/2 over TLS/TCP and HTTP/3 over QUIC in parallel,
	// advertising QUIC via Alt-Svc headers on the TLS/TCP responses.
	tcpLn, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	udpConn, err := net.ListenPacket("udp", addr)
	if err != nil {
		// #nosec G104 -- best-effort cleanup; the ListenPacket error above is what's returned
		_ = tcpLn.Close()
		return nil, err
	}

	h3 := &http3.Server{
		Handler:   h,
		TLSConfig: tlsConfig.Clone(),
	}
	tcp := newHTTPServer(tcpLn.Addr().String(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// #nosec G104 -- best-effort Alt-Svc advertisement
		_ = h3.SetQUICHeaders(w.Header())
		h.ServeHTTP(w, r)
	}))
	tcp.TLSConfig = tlsConfig.Clone()

	srv := &Server{httpServer: tcp, h3Server: h3, udpConn: udpConn, addr: tcpLn.Addr().String()}
	go func() {
		if err := tcp.ServeTLS(tcpLn, cfg.CertFile, cfg.KeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srv.selfDestruct()
		}
	}()
	go func() {
		if err := h3.Serve(udpConn); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srv.selfDestruct()
		}
	}()

	return srv, nil
}
