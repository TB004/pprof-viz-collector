// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/TB004/pprof-viz-collector/internal/proto"
	"github.com/TB004/pprof-viz-collector/internal/proto/protoconnect"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"
)

func newTestClient(t *testing.T) protoconnect.CollectorServiceClient {
	t.Helper()
	mux := http.NewServeMux()
	path, handler := protoconnect.NewCollectorServiceHandler(CollectorServiceHandler{})
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return protoconnect.NewCollectorServiceClient(server.Client(), server.URL)
}

func TestAgentServiceHandler_Ping(t *testing.T) {
	t.Parallel()
	_, err := newTestClient(t).Ping(context.Background(), &emptypb.Empty{})
	assert.NoError(t, err)
}

func TestAgentServiceHandler_SnapshotAll(t *testing.T) {
	t.Parallel()
	resp, err := newTestClient(t).SnapshotAll(context.Background(), &emptypb.Empty{})
	assert.NoError(t, err)

	assert.NotNil(t, resp.Cpu)
	assert.NotNil(t, resp.Memory)
	assert.NotNil(t, resp.Gc)
	assert.NotEmpty(t, resp.Gc.BytesAllocsBySize.UpdatedCounts)
}

func TestAgentServiceHandler_GetMetadata(t *testing.T) {
	t.Parallel()
	resp, err := newTestClient(t).GetMetadata(context.Background(), &emptypb.Empty{})
	assert.NoError(t, err)

	assert.NotNil(t, resp.Gc)
	assert.NotNil(t, resp.Scheduler)
	assert.NotEmpty(t, resp.Gc.BytesAllocsBySizeBuckets)
	assert.NotEmpty(t, resp.Scheduler.LatenciesSecondsBuckets)
}

func TestAgentServiceHandler_StreamAll(t *testing.T) {
	t.Parallel()

	t.Run("InvalidInterval", func(t *testing.T) {
		t.Parallel()
		client := newTestClient(t)
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
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		stream, err := newTestClient(t).StreamAll(ctx, &proto.StreamConfigRequest{
			IntervalSeconds: 1,
			NumberOfCycles:  1,
		})
		assert.NoError(t, err)

		// We expect 1 message
		assert.True(t, stream.Receive(), "expected first message")
		msg := stream.Msg()
		assert.NotNil(t, msg)
		assert.NotNil(t, msg.Cpu)
		assert.Empty(t, msg.Gc.BytesAllocsBySize.Buckets)
		assert.NotEmpty(t, msg.Gc.BytesAllocsBySize.UpdatedCounts)

		// We expect stream to close after 1 cycle
		assert.False(t, stream.Receive(), "expected stream to close after 1 cycle")
		assert.NoError(t, stream.Err())
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())

		stream, err := newTestClient(t).StreamAll(ctx, &proto.StreamConfigRequest{
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
	resp := collector.Collect()
	assert.NotNil(t, resp)
	// Basic validation that we got some data populated
	assert.NotNil(t, resp.Cpu)
	assert.NotNil(t, resp.Gc)
}
