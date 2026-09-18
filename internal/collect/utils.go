// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package collect

import (
	"math"
	"runtime/metrics"
	"slices"
)

type histogram struct {
	Counts  []uint64  `json:"counts"`
	Buckets []float32 `json:"buckets"`
}

// asHistogram converts v into a histogram that owns its data, safe to retain
// past the call. This is unlike v.Float64Histogram() itself, whose Counts
// slice is the runtime's live buffer: metrics.Read overwrites it in place on
// every subsequent read of the same metrics.Sample, so holding onto that
// slice directly (rather than a copy) is a use-after-mutate trap for any
// caller that keeps the result around.
func asHistogram(v metrics.Value) *histogram {
	if v.Kind() != metrics.KindFloat64Histogram {
		return nil
	}
	h := v.Float64Histogram()

	buckets := make([]float32, len(h.Buckets))
	for i, b := range h.Buckets {
		buckets[i] = float32(b)
	}

	return &histogram{
		Counts:  slices.Clone(h.Counts),
		Buckets: buckets,
	}
}

// histCounts returns just the counts of a histogram metric value, aliasing
// the runtime's live buffer. Only safe for the immediate, synchronous diff in
// collect; unlike asHistogram, it does not copy, so the result must not be
// retained past the current read.
func histCounts(v metrics.Value) []uint64 {
	if v.Kind() != metrics.KindFloat64Histogram {
		return nil
	}
	return v.Float64Histogram().Counts
}

func asFloat(v metrics.Value) float64 {
	switch v.Kind() {
	case metrics.KindFloat64:
		return v.Float64()
	case metrics.KindUint64:
		return float64(v.Uint64())
	default:
		return 0
	}
}

func asUint(v metrics.Value) uint64 {
	switch v.Kind() {
	case metrics.KindUint64:
		return v.Uint64()
	case metrics.KindFloat64:
		return uint64(v.Float64())
	default:
		return 0
	}
}

func safeUint32(v uint64) uint32 {
	if v > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(v)
}
