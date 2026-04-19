// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"context"
	"errors"
	"runtime/metrics"
	"time"

	"connectrpc.com/connect"
	"github.com/TB004/pprof-viz-agent/internal/proto"
	"github.com/TB004/pprof-viz-agent/internal/proto/protoconnect"
	"google.golang.org/protobuf/types/known/emptypb"
)

// AgentServiceHandler implements the AgentService ConnectRPC service.
type AgentServiceHandler struct {
	protoconnect.UnimplementedAgentServiceHandler
}

// Ping is a no-op health check that returns OK if the agent is reachable.
func (h AgentServiceHandler) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// SnapshotAll returns a single point-in-time capture of all supported Go runtime metrics.
func (h AgentServiceHandler) SnapshotAll(context.Context, *emptypb.Empty) (*proto.AllResponse, error) {
	collector := NewSampleCollector()
	collector.Read()
	return collector.Collect(), nil
}

// GetMetadata returns the static bucket boundaries for all histogram metrics.
// Clients should call this once before streaming; StreamAll never sends bucket boundaries.
func (h AgentServiceHandler) GetMetadata(context.Context, *emptypb.Empty) (*proto.MetadataResponse, error) {
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
func (h AgentServiceHandler) StreamAll(ctx context.Context, req *proto.StreamConfigRequest, stream *connect.ServerStream[proto.AllResponse]) error {
	interval := req.GetIntervalSeconds()
	if interval == 0 {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("interval must be at least 1 second"))
	}

	numCycles := int(req.GetNumberOfCycles())

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	collector := NewSampleCollector()

	for i := 0; numCycles == 0 || i < numCycles; i++ {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			collector.Read()
			all := collector.Collect()
			if err := stream.Send(all); err != nil {
				return err
			}
		}
	}

	return nil
}
