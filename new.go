// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package pprofvizcollector

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"strconv"

	"github.com/TB004/pprof-viz-collector/internal/proto/protoconnect"
	"github.com/TB004/pprof-viz-collector/internal/rpc"
	"github.com/quic-go/quic-go/http3"
)

// NewConfig configures New.
type NewConfig struct {
	// Token, if non-empty, is required on every RPC call via Authorization: Bearer <token>.
	Token string
	// AllowedOrigins is the list of origins permitted by CORS (e.g. ["https://app.example.com"]).
	// Use ["*"] to allow all origins. An empty slice disables CORS middleware entirely.
	AllowedOrigins []string
	// CompressMinBytes is the minimum response size, in bytes, below which
	// responses are sent uncompressed. Defaults to 256 when zero. Lower it on
	// bandwidth-constrained links, where compressing even small frames is
	// worth the CPU cost; raise it if compression's CPU cost outweighs the
	// bandwidth savings for your typical message sizes.
	CompressMinBytes int
	// CertFile is the path to the TLS certificate file. Enables TLS when set.
	CertFile string
	// KeyFile is the path to the TLS private key file. Must be set together with CertFile.
	KeyFile string
	// HTTP3 also serves HTTP/3 over QUIC in parallel with the TLS/TCP server.
	// Requires CertFile and KeyFile.
	HTTP3 bool
	// Host is the network interface to bind to (e.g. "127.0.0.1" or "0.0.0.0").
	// Defaults to "localhost" when empty. Binding to a non-loopback address
	// exposes the agent externally; pair it with TLS and a Token.
	Host string
}

// New starts an HTTP server on the specified port serving the CollectorService
// and returns immediately; it does not block. Call Shutdown on the returned
// Server to stop it.
//
// Without TLS, it supports HTTP/2 over cleartext and HTTP/1.1 (necessary for browsers).
// With CertFile and KeyFile set, it supports HTTP/2 over TLS (minimum TLS 1.2).
// With HTTP3 also set, it additionally serves HTTP/3 over QUIC in parallel, with Alt-Svc
// headers on TLS responses so clients can upgrade automatically.
func New(port uint16, cfg NewConfig) (*Server, error) {
	if cfg.HTTP3 && (cfg.CertFile == "" || cfg.KeyFile == "") {
		return nil, errors.New("HTTP3 requires CertFile and KeyFile to be configured")
	}

	mux := http.NewServeMux()
	path, handler := protoconnect.NewCollectorServiceHandler(rpc.CollectorServiceHandler{}, handlerOpts(cfg.Token, cfg.CompressMinBytes)...)
	mux.Handle(path, handler)

	host := cfg.Host
	if host == "" {
		host = "localhost"
	}
	addr := net.JoinHostPort(host, strconv.Itoa(int(port)))

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
