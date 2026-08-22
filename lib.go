// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package pprofvizcollector

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"strconv"
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

func handlerOpts(cfg Config) []connect.HandlerOption {
	opts := []connect.HandlerOption{
		connect.WithCompression(internal.CompressionZstd, internal.NewZstdDecompressor, internal.NewZstdCompressor),
		connect.WithCompressMinBytes(256),
	}
	if cfg.Token != "" {
		opts = append(opts, connect.WithInterceptors(internal.NewTokenInterceptor(cfg.Token)))
	}
	return opts
}

// Bind registers the AgentService handler to the provided http.ServeMux.
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

// New creates and starts an HTTP server on the specified port serving the AgentService.
// Without TLS, it supports HTTP/2 over cleartext.
// With CertFile and KeyFile set, it supports HTTP/2 over TLS (minimum TLS 1.2).
// With HTTP3 also set, it additionally serves HTTP/3 over QUIC in parallel, with Alt-Svc
// headers on TLS responses so clients can upgrade automatically.
func New(port int, cfg Config) error {
	if cfg.HTTP3 && (cfg.CertFile == "" || cfg.KeyFile == "") {
		return errors.New("HTTP3 requires CertFile and KeyFile to be configured")
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
		protos := new(http.Protocols)
		protos.SetHTTP1(true)
		protos.SetHTTP2(false)
		protos.SetUnencryptedHTTP2(true)

		s := newHTTPServer(addr, h)
		s.Protocols = protos
		return s.ListenAndServe()
	}

	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}

	// HTTP/2-over-TLS.
	if !cfg.HTTP3 {
		s := newHTTPServer(addr, h)
		s.TLSConfig = tlsConfig
		return s.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
	}

	// TLS with HTTP/3: serve HTTP/2 over TLS/TCP and HTTP/3 over QUIC in parallel,
	// advertising QUIC via Alt-Svc headers on the TLS/TCP responses.
	h3 := &http3.Server{
		Addr:      addr,
		Handler:   h,
		TLSConfig: tlsConfig.Clone(),
	}
	tcp := newHTTPServer(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// #nosec G104 -- best-effort Alt-Svc advertisement
		_ = h3.SetQUICHeaders(w.Header())
		h.ServeHTTP(w, r)
	}))
	tcp.TLSConfig = tlsConfig.Clone()

	errc := make(chan error, 2)
	go func() { errc <- tcp.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile) }()
	go func() { errc <- h3.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile) }()

	// Whichever server fails first wins; close the other and surface any
	// shutdown errors alongside the original.
	err := <-errc
	return errors.Join(err, tcp.Close(), h3.Close())
}
