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
	t.Parallel()
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
	// Snapshot should include buckets
	assert.NotEmpty(t, resp.Gc.BytesAllocsBySize.Buckets)
}

func TestAgentServiceHandler_GetMetadata(t *testing.T) {
	t.Parallel()
	// Setup Connect server
	mux := http.NewServeMux()
	path, handler := protoconnect.NewAgentServiceHandler(AgentServiceHandler{})
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	client := protoconnect.NewAgentServiceClient(server.Client(), server.URL)
	resp, err := client.GetMetadata(context.Background(), &emptypb.Empty{})
	assert.NoError(t, err)

	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Gc)
	assert.NotNil(t, resp.Scheduler)
	// Metadata should include buckets
	assert.NotEmpty(t, resp.Gc.BytesAllocsBySizeBuckets)
	assert.NotEmpty(t, resp.Scheduler.LatenciesSecondsBuckets)
}

func TestAgentServiceHandler_StreamAll(t *testing.T) {
	t.Parallel()

	t.Run("InvalidInterval", func(t *testing.T) {
		t.Parallel()
		mux := http.NewServeMux()
		path, handler := protoconnect.NewAgentServiceHandler(AgentServiceHandler{})
		mux.Handle(path, handler)
		server := httptest.NewServer(mux)
		defer server.Close()

		client := protoconnect.NewAgentServiceClient(server.Client(), server.URL)
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
		t.Parallel()
		mux := http.NewServeMux()
		path, handler := protoconnect.NewAgentServiceHandler(AgentServiceHandler{})
		mux.Handle(path, handler)
		server := httptest.NewServer(mux)
		defer server.Close()

		client := protoconnect.NewAgentServiceClient(server.Client(), server.URL)

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
		// By default (ExcludeHistogramBuckets=false), buckets should be present
		assert.NotEmpty(t, msg.Gc.BytesAllocsBySize.Buckets)

		// We expect stream to close after 1 cycle
		assert.False(t, stream.Receive(), "expected stream to close after 1 cycle")
		assert.NoError(t, stream.Err())
	})

	t.Run("StreamWithoutBuckets", func(t *testing.T) {
		t.Parallel()
		mux := http.NewServeMux()
		path, handler := protoconnect.NewAgentServiceHandler(AgentServiceHandler{})
		mux.Handle(path, handler)
		server := httptest.NewServer(mux)
		defer server.Close()

		client := protoconnect.NewAgentServiceClient(server.Client(), server.URL)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		stream, err := client.StreamAll(ctx, &proto.StreamConfigRequest{
			IntervalSeconds:         1,
			NumberOfCycles:          1,
			ExcludeHistogramBuckets: true,
		})
		assert.NoError(t, err)

		assert.True(t, stream.Receive())
		msg := stream.Msg()
		assert.NotNil(t, msg)
		// Buckets should be empty
		assert.Empty(t, msg.Gc.BytesAllocsBySize.Buckets)
		// Counts should still be present
		assert.NotEmpty(t, msg.Gc.BytesAllocsBySize.Counts)

		assert.False(t, stream.Receive())
		assert.NoError(t, stream.Err())
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		mux := http.NewServeMux()
		path, handler := protoconnect.NewAgentServiceHandler(AgentServiceHandler{})
		mux.Handle(path, handler)
		server := httptest.NewServer(mux)
		defer server.Close()

		client := protoconnect.NewAgentServiceClient(server.Client(), server.URL)

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
	t.Parallel()
	collector := NewSampleCollector()
	collector.Read()
	resp := collector.Collect(true)
	assert.NotNil(t, resp)
	// Basic validation that we got some data populated
	assert.NotNil(t, resp.Cpu)
	assert.NotNil(t, resp.Gc)
}
