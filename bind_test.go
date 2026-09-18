// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package pprofvizcollector

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBind(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	Bind(mux, BindConfig{})

	req := httptest.NewRequest("POST", "/proto.CollectorService/SnapshotAll", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// A 404 would mean Bind failed; any other status means the handler is registered.
	assert.NotEqual(t, http.StatusNotFound, rec.Code)
}

func TestBind_TokenAuth(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	Bind(mux, BindConfig{Token: "secret"})

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
	Bind(mux, BindConfig{AllowedOrigins: []string{"https://app.example.com"}})

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
		Bind(mux2, BindConfig{})
		req := httptest.NewRequest("OPTIONS", "/proto.CollectorService/Ping", nil)
		req.Header.Set("Origin", "https://app.example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		rec := httptest.NewRecorder()
		mux2.ServeHTTP(rec, req)
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})
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
	Bind(mux, BindConfig{})

	oversized := strings.Repeat("a", maxRequestBytes*2)
	req := httptest.NewRequest("POST", "/proto.CollectorService/StreamAll", strings.NewReader(`{"intervalSeconds":`+oversized+`}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusOK, rec.Code)
}

func TestBind_CompressMinBytes(t *testing.T) {
	t.Parallel()

	ping := func(t *testing.T, mux *http.ServeMux) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest("POST", "/proto.CollectorService/Ping", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "zstd")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	t.Run("DefaultThreshold_TinyResponseUncompressed", func(t *testing.T) {
		t.Parallel()
		mux := http.NewServeMux()
		Bind(mux, BindConfig{})
		rec := ping(t, mux)
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Content-Encoding"),
			"Ping's tiny response should stay under the default 256-byte compression threshold")
	})

	t.Run("LowThreshold_TinyResponseCompressed", func(t *testing.T) {
		t.Parallel()
		mux := http.NewServeMux()
		Bind(mux, BindConfig{CompressMinBytes: 1})
		rec := ping(t, mux)
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "zstd", rec.Header().Get("Content-Encoding"),
			"a 1-byte threshold should compress even Ping's tiny response")
	})
}
