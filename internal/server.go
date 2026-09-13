// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"context"
	"errors"
	"runtime/metrics"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/TB004/pprof-viz-collector/internal/proto"
	"github.com/TB004/pprof-viz-collector/internal/proto/protoconnect"
	"google.golang.org/protobuf/types/known/emptypb"
)

// metadataResponse builds the MetadataResponse once and caches it: bucket
// boundaries are fixed for the process lifetime, so there's no need to read
// all runtime metrics on every call.
var metadataResponse = sync.OnceValue(func() *proto.MetadataResponse {
	histogramIndices := []Metric{
		MetricGCHeapAllocsBySize,
		MetricGCHeapFreesBySize,
		MetricSchedLatencies,
		MetricSchedPausesStoppingGC,
		MetricSchedPausesStoppingOther,
		MetricSchedPausesTotalGC,
		MetricSchedPausesTotalOther,
	}

	samples := make([]metrics.Sample, len(histogramIndices))
	for i, idx := range histogramIndices {
		samples[i].Name = AllMetricNames[idx]
	}
	metrics.Read(samples)

	buckets := make(map[Metric][]float32, len(samples))
	for i, idx := range histogramIndices {
		if h := AsHistogram(samples[i].Value); h != nil {
			buckets[Metric(idx)] = h.Buckets
		}
	}

	return &proto.MetadataResponse{
		Gc: &proto.GCMetadata{
			BytesAllocsBySizeBuckets: buckets[MetricGCHeapAllocsBySize],
			BytesFreesBySizeBuckets:  buckets[MetricGCHeapFreesBySize],
		},
		Scheduler: &proto.SchedulerMetadata{
			LatenciesSecondsBuckets:       buckets[MetricSchedLatencies],
			StoppingGcSecondsBuckets:      buckets[MetricSchedPausesStoppingGC],
			StoppingOtherSecondsBuckets:   buckets[MetricSchedPausesStoppingOther],
			TotalGcPauseSecondsBuckets:    buckets[MetricSchedPausesTotalGC],
			TotalOtherPauseSecondsBuckets: buckets[MetricSchedPausesTotalOther],
		},
	}
})

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
	return metadataResponse(), nil
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
	// are differential.
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
