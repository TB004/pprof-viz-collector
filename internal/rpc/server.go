// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package rpc

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/TB004/pprof-viz-collector/internal/collect"
	"github.com/TB004/pprof-viz-collector/internal/proto"
	"github.com/TB004/pprof-viz-collector/internal/proto/protoconnect"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CollectorServiceHandler struct {
	protoconnect.UnimplementedCollectorServiceHandler
}

// Ping is a no-op health check that returns OK if the collector is reachable.
func (h CollectorServiceHandler) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// SnapshotAll returns a single point-in-time capture of all supported Go runtime metrics.
func (h CollectorServiceHandler) SnapshotAll(context.Context, *emptypb.Empty) (*proto.AllResponse, error) {
	return collect.Snapshot(), nil
}

// GetMetadata returns the static bucket boundaries for all histogram metrics.
// Clients should call this once before streaming; StreamAll never sends bucket boundaries.
func (h CollectorServiceHandler) GetMetadata(context.Context, *emptypb.Empty) (*proto.MetadataResponse, error) {
	return collect.Metadata(), nil
}

// StreamAll streams runtime metrics at a client-provided interval in a client-provided number of cycles or indefinitely.
func (h CollectorServiceHandler) StreamAll(ctx context.Context, req *proto.StreamConfigRequest, stream *connect.ServerStream[proto.AllResponse]) error {
	interval := req.GetIntervalSeconds()
	if interval == 0 {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("interval must be at least 1 second"))
	}

	numCycles := int(req.GetNumberOfCycles())

	intervalDuration := time.Duration(interval) * time.Second
	ticker := time.NewTicker(intervalDuration)
	defer ticker.Stop()

	collector := collect.NewCollector()

	send := func() error {
		collector.Read()
		return stream.Send(collector.Collect())
	}

	// Send a full initial frame immediately so clients aren't blind for the
	// first interval. The first read on a fresh collector diffs against an empty
	// baseline, so this frame carries the full current state; subsequent frames
	// are differential.
	if err := send(); err != nil {
		return err
	}
	// Reset so the first scheduled tick is measured from here, not from when
	// the ticker was constructed a moment earlier.
	ticker.Reset(intervalDuration)

	for i := 1; numCycles == 0 || i < numCycles; i++ {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := send(); err != nil {
				return err
			}
			// stream.Send blocks until the client accepts the frame. Without
			// this reset, a client that stalls past one or more ticks would
			// get an immediate catch-up frame from the ticker's one buffered
			// tick as soon as it recovers. Resetting here instead measures
			// the interval between frames, not between tick attempts: the
			// next frame always waits a full, fresh interval.
			ticker.Reset(intervalDuration)
		}
	}

	return nil
}
