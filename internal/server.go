// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"context"
	"errors"
	"runtime/metrics"
	"time"

	"connectrpc.com/connect"
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
	return Snapshot(), nil
}

// GetMetadata returns the static bucket boundaries for all histogram metrics.
// Clients should call this once before streaming; StreamAll never sends bucket boundaries.
func (h CollectorServiceHandler) GetMetadata(context.Context, *emptypb.Empty) (*proto.MetadataResponse, error) {
	collector := NewSampleCollector()
	collector.Read()

	// Helper function to get buckets out of a histogram value
	getBuckets := func(v metrics.Value) []float32 {
		h := AsHistogram(v)
		if h == nil {
			return nil
		}
		return h.Buckets
	}

	v := func(name string) metrics.Value {
		return collector.samples[GetIndex(name)].Value
	}

	resp := &proto.MetadataResponse{
		Gc: &proto.GCMetadata{
			BytesAllocsBySizeBuckets: getBuckets(v(MetricGCHeapAllocsBySize)),
			BytesFreesBySizeBuckets:  getBuckets(v(MetricGCHeapFreesBySize)),
		},
		Scheduler: &proto.SchedulerMetadata{
			LatenciesSecondsBuckets:       getBuckets(v(MetricSchedLatencies)),
			StoppingGcSecondsBuckets:      getBuckets(v(MetricSchedPausesStoppingGC)),
			StoppingOtherSecondsBuckets:   getBuckets(v(MetricSchedPausesStoppingOther)),
			TotalGcPauseSecondsBuckets:    getBuckets(v(MetricSchedPausesTotalGC)),
			TotalOtherPauseSecondsBuckets: getBuckets(v(MetricSchedPausesTotalOther)),
		},
	}

	return resp, nil
}

// StreamAll streams runtime metrics at a client-provided interval in a client-provided number of cycles or indefinitely.
func (h CollectorServiceHandler) StreamAll(ctx context.Context, req *proto.StreamConfigRequest, stream *connect.ServerStream[proto.AllResponse]) error {
	interval := req.GetIntervalSeconds()
	if interval == 0 {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("interval must be at least 1 second"))
	}

	numCycles := int(req.GetNumberOfCycles())

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	collector := NewSampleCollector()

	send := func() error {
		collector.Read()
		return stream.Send(collector.Collect())
	}

	// Send a full initial frame immediately so clients aren't blind for the
	// first interval. The first Read on a fresh collector diffs against an empty
	// baseline, so this frame carries the full current state; subsequent frames
	// are differential. This counts as the first cycle.
	if err := send(); err != nil {
		return err
	}

	for i := 1; numCycles == 0 || i < numCycles; i++ {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := send(); err != nil {
				return err
			}
		}
	}

	return nil
}
