// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package collect

import (
	"runtime/metrics"
	"sync"

	"github.com/TB004/pprof-viz-collector/internal/proto"
)

// metadataOnce builds the MetadataResponse once and caches it: bucket
// boundaries are fixed for the process lifetime, so there's no need to read
// all runtime metrics on every call.
var metadataOnce = sync.OnceValue(func() *proto.MetadataResponse {
	samples := make([]metrics.Sample, len(histogramMetrics))
	for i, idx := range histogramMetrics {
		samples[i].Name = AllMetricNames[idx]
	}
	metrics.Read(samples)

	buckets := make(map[Metric][]float32, len(samples))
	for i, idx := range histogramMetrics {
		if h := asHistogram(samples[i].Value); h != nil {
			buckets[idx] = h.Buckets
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

// Metadata returns the static bucket boundaries for all histogram metrics.
// Clients should call this once before streaming; StreamAll never sends bucket boundaries.
func Metadata() *proto.MetadataResponse {
	return metadataOnce()
}
