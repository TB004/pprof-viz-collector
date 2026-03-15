// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/TB004/pprof-viz-agent/internal/proto"
	"github.com/TB004/pprof-viz-agent/internal/proto/protoconnect"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestAgentServiceHandler_SnapshotAll(t *testing.T) {
	// Setup Connect server
	mux := http.NewServeMux()
	path, handler := protoconnect.NewAgentServiceHandler(AgentServiceHandler{})
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	client := protoconnect.NewAgentServiceClient(server.Client(), server.URL)
	resp, err := client.SnapshotAll(context.Background(), &emptypb.Empty{})
	assert.NoError(t, err)

	// Basic validation that we got some data populated
	assert.NotNil(t, resp.Cpu)
	assert.NotNil(t, resp.Memory)
	assert.NotNil(t, resp.Gc)
}

func TestAgentServiceHandler_StreamAll(t *testing.T) {
	// Setup Connect server
	mux := http.NewServeMux()
	path, handler := protoconnect.NewAgentServiceHandler(AgentServiceHandler{})
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	client := protoconnect.NewAgentServiceClient(server.Client(), server.URL)

	t.Parallel()
	t.Run("InvalidInterval", func(t *testing.T) {
		stream, err := client.StreamAll(context.Background(), &proto.StreamConfigRequest{
			IntervalSeconds: 0, // Invalid
			NumberOfCycles:  1,
		})
		assert.NoError(t, err) // Stream creation itself usually succeeds

		// Receive should fail
		hasNext := stream.Receive()
		assert.False(t, hasNext)
		err = stream.Err()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "interval must be at least 1 second")
	})

	t.Run("StreamMultipleCycles", func(t *testing.T) {
		// Use a short interval for testing
		// Since the server implementation enforces "seconds" (interval * time.Second),
		// we have to wait at least 1 second per cycle.
		// For a fast test, we might only test 1 cycle with 1 second.
		// Or we trust the logic.

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		stream, err := client.StreamAll(ctx, &proto.StreamConfigRequest{
			IntervalSeconds: 1,
			NumberOfCycles:  1,
		})
		assert.NoError(t, err)

		// We expect 1 message
		assert.True(t, stream.Receive(), "expected first message")
		msg := stream.Msg()
		assert.NotNil(t, msg)
		assert.NotNil(t, msg.Cpu)

		// We expect stream to close after 1 cycle
		assert.False(t, stream.Receive(), "expected stream to close after 1 cycle")
		assert.NoError(t, stream.Err())
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		// Test that it stops when context is cancelled
		ctx, cancel := context.WithCancel(context.Background())

		stream, err := client.StreamAll(ctx, &proto.StreamConfigRequest{
			IntervalSeconds: 1,
			NumberOfCycles:  10, // Many cycles
		})
		assert.NoError(t, err)

		// Cancel immediately
		cancel()

		// Attempt to receive - might get one or error immediately depending on race
		// But it shouldn't hang forever
		for stream.Receive() {
			// Drain if any
		}

		// Should be done
		assert.Error(t, stream.Err())
		assert.Contains(t, stream.Err().Error(), "canceled")
	})
}

func TestCollectSamples(t *testing.T) {
	resp := collectSamples()
	assert.NotNil(t, resp)
	// Basic validation that we got some data populated
	assert.NotNil(t, resp.Cpu)
	assert.NotNil(t, resp.Gc)
}
