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
	t.Parallel()
	mux := http.NewServeMux()
	Bind(mux)

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
	time.Sleep(100 * time.Millisecond)

	// ConnectRPC with JSON expects an empty object "{}" for google.protobuf.Empty.
	body := strings.NewReader("{}")
	req, err := http.NewRequest("POST", "http://localhost:50000/proto.AgentService/SnapshotAll", body)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Actually perform the request
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	assert.NotEmpty(t, resp.Body)
	defer assert.NoError(t, resp.Body.Close())
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSnapshotCgo(t *testing.T) {
	t.Parallel()
	resp := SnapshotCgo()
	assert.NotNil(t, resp)
}

func TestSnapshotGC(t *testing.T) {
	t.Parallel()
	resp := SnapshotGC()
	assert.NotNil(t, resp)
}

func TestSnapshotCPU(t *testing.T) {
	t.Parallel()
	resp := SnapshotCPU()
	assert.NotNil(t, resp)
}

func TestSnapshotGoDebug(t *testing.T) {
	t.Parallel()
	resp := SnapshotGoDebug()
	assert.NotNil(t, resp)
}

func TestSnapshotMemory(t *testing.T) {
	t.Parallel()
	resp := SnapshotMemory()
	assert.NotNil(t, resp)
}

func TestSnapshotScheduler(t *testing.T) {
	t.Parallel()
	resp := SnapshotScheduler()
	assert.NotNil(t, resp)
}

func TestSnapshotSync(t *testing.T) {
	t.Parallel()
	resp := SnapshotSync()
	assert.NotNil(t, resp)
}

func TestSnapshotAll(t *testing.T) {
	t.Parallel()
	resp := SnapshotAll()
	assert.NotNil(t, resp)
}
