// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package pprofvizcollector

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// shutdown gracefully stops srv within a short timeout and fails the test if
// that doesn't succeed. Intended for t.Cleanup.
func shutdown(t *testing.T, srv *Server) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	assert.NoError(t, srv.Shutdown(ctx))
}

func TestBind(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	Bind(mux, Config{})

	req := httptest.NewRequest("POST", "/proto.CollectorService/SnapshotAll", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// A 404 would mean Bind failed; any other status means the handler is registered.
	assert.NotEqual(t, http.StatusNotFound, rec.Code)
}

func TestBind_TokenAuth(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	Bind(mux, Config{Token: "secret"})

	t.Run("MissingToken", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest("POST", "/proto.CollectorService/Ping", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("WrongToken", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest("POST", "/proto.CollectorService/Ping", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer wrong")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("CorrectToken", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest("POST", "/proto.CollectorService/Ping", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer secret")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestBind_CORS(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	Bind(mux, Config{AllowedOrigins: []string{"https://app.example.com"}})

	t.Run("PreflightAllowedOrigin", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest("OPTIONS", "/proto.CollectorService/Ping", nil)
		req.Header.Set("Origin", "https://app.example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		assert.Equal(t, "https://app.example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("PreflightDisallowedOrigin", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest("OPTIONS", "/proto.CollectorService/Ping", nil)
		req.Header.Set("Origin", "https://evil.example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("NoCORSWithoutOrigins", func(t *testing.T) {
		t.Parallel()
		mux2 := http.NewServeMux()
		Bind(mux2, Config{})
		req := httptest.NewRequest("OPTIONS", "/proto.CollectorService/Ping", nil)
		req.Header.Set("Origin", "https://app.example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		rec := httptest.NewRecorder()
		mux2.ServeHTTP(rec, req)
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestNew(t *testing.T) {
	t.Parallel()

	server, err := New(0, Config{})
	require.NoError(t, err)
	t.Cleanup(func() { shutdown(t, server) })
	assert.NotEmpty(t, server.Addr())

	protos := new(http.Protocols)
	protos.SetUnencryptedHTTP2(true)
	client := &http.Client{Transport: &http.Transport{Protocols: protos}}

	body := strings.NewReader("{}")
	req, err := http.NewRequest("POST", "http://"+server.Addr()+"/proto.CollectorService/SnapshotAll", body)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer assert.NoError(t, resp.Body.Close())
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestNew_PortZeroPicksFreePort(t *testing.T) {
	t.Parallel()

	a, err := New(0, Config{})
	require.NoError(t, err)
	t.Cleanup(func() { shutdown(t, a) })

	b, err := New(0, Config{})
	require.NoError(t, err)
	t.Cleanup(func() { shutdown(t, b) })

	assert.NotEmpty(t, a.Addr())
	assert.NotEmpty(t, b.Addr())
	assert.NotEqual(t, a.Addr(), b.Addr(), "two servers on port 0 should get distinct ports")
}

func TestNew_Shutdown(t *testing.T) {
	t.Parallel()

	server, err := New(0, Config{})
	require.NoError(t, err)
	addr := server.Addr()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, server.Shutdown(ctx))

	// Idempotent: a second call must not error or hang.
	assert.NoError(t, server.Shutdown(ctx))

	// The listener must actually be closed.
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if conn != nil {
		_ = conn.Close()
	}
	assert.Error(t, err)
}

func TestNew_HTTP3(t *testing.T) {
	t.Parallel()

	server, err := New(0, Config{
		CertFile: "example/cert.pem",
		KeyFile:  "example/key.pem",
		HTTP3:    true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { shutdown(t, server) })
	require.NotEmpty(t, server.Addr())

	conn, err := tls.Dial("tcp", server.Addr(), &tls.Config{InsecureSkipVerify: true}) //nolint:gosec // test dial, not a real client
	assert.NoError(t, err)
	if conn != nil {
		_ = conn.Close()
	}
}

func TestRecoverPanic(t *testing.T) {
	t.Parallel()

	err := recoverPanic(context.Background(), connect.Spec{}, nil, "boom")
	var connectErr *connect.Error
	assert.True(t, errors.As(err, &connectErr))
	assert.Equal(t, connect.CodeInternal, connectErr.Code())
	assert.Contains(t, connectErr.Error(), "boom")
}

func TestBind_ReadMaxBytes(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	Bind(mux, Config{})

	oversized := strings.Repeat("a", maxRequestBytes*2)
	req := httptest.NewRequest("POST", "/proto.CollectorService/StreamAll", strings.NewReader(`{"intervalSeconds":`+oversized+`}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusOK, rec.Code)
}

func TestNew_HTTP3RequiresCertFiles(t *testing.T) {
	t.Parallel()
	server, err := New(0, Config{HTTP3: true})
	assert.Nil(t, server)
	assert.EqualError(t, err, "HTTP3 requires CertFile and KeyFile to be configured")
}
