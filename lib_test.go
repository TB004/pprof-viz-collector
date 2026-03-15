// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package pprofvizagent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBind(t *testing.T) {
	mux := http.NewServeMux()
	Bind(mux)

	// Verify the route is registered
	// The path for AgentService is /proto.AgentService/
	req := httptest.NewRequest("POST", "/proto.AgentService/SnapshotAll", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	// We expect 415 Unsupported Media Type or similar because we didn't send proper headers/body,
	// but getting a 404 would mean Bind failed.
	assert.NotEqual(t, http.StatusNotFound, rec.Code, "handler was not bound correctly")
	assert.NotNil(t, rec.Body)
}

func TestNew(t *testing.T) {
	// Start server in background.
	// Note: New blocks until the server stops or fails.
	go func() {
		_ = New(50000, nil, nil)
	}()

	// Give the server a moment to start up and bind to the port.
	time.Sleep(3 * time.Millisecond)

	// ConnectRPC with JSON expects an empty object "{}" for google.protobuf.Empty.
	body := strings.NewReader("{}")
	req, err := http.NewRequest("POST", "http://localhost:50000/proto.AgentService/SnapshotAll", body)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Actually perform the request
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	assert.NotEmpty(t, resp.Body)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSnapshotCgo(t *testing.T) {
	resp := SnapshotCgo()
	assert.NotNil(t, resp)
}

func TestSnapshotGC(t *testing.T) {
	resp := SnapshotGC()
	assert.NotNil(t, resp)
}

func TestSnapshotCPU(t *testing.T) {
	resp := SnapshotCPU()
	assert.NotNil(t, resp)
}

func TestSnapshotGoDebug(t *testing.T) {
	resp := SnapshotGoDebug()
	assert.NotNil(t, resp)
}

func TestSnapshotMemory(t *testing.T) {
	resp := SnapshotMemory()
	assert.NotNil(t, resp)
}

func TestSnapshotSchedular(t *testing.T) {
	resp := SnapshotSchedular()
	assert.NotNil(t, resp)
}

func TestSnapshotSync(t *testing.T) {
	resp := SnapshotSync()
	assert.NotNil(t, resp)
}

func TestSnapshotAll(t *testing.T) {
	resp := SnapshotAll()
	assert.NotNil(t, resp)
}
