// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package pprofvizcollector

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
	go func() {
		_ = New(50000, Config{})
	}()

	time.Sleep(100 * time.Millisecond)

	protos := new(http.Protocols)
	protos.SetUnencryptedHTTP2(true)
	client := &http.Client{Transport: &http.Transport{Protocols: protos}}

	body := strings.NewReader("{}")
	req, err := http.NewRequest("POST", "http://localhost:50000/proto.CollectorService/SnapshotAll", body)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer assert.NoError(t, resp.Body.Close())
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestNew_HTTP3RequiresCertFiles(t *testing.T) {
	t.Parallel()
	err := New(50001, Config{HTTP3: true})
	assert.EqualError(t, err, "HTTP3 requires CertFile and KeyFile to be configured")
}
