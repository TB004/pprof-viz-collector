// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"math"
	"runtime/metrics"
)

type Histogram struct {
	Counts  []uint64  `json:"counts"`
	Buckets []float32 `json:"buckets"`
}

func AsHistogram(v metrics.Value) *Histogram {
	if v.Kind() != metrics.KindFloat64Histogram {
		return nil
	}
	h := v.Float64Histogram()

	buckets := make([]float32, len(h.Buckets))
	for i, b := range h.Buckets {
		buckets[i] = float32(b)
	}

	return &Histogram{
		Counts:  h.Counts,
		Buckets: buckets,
	}
}

// histCounts returns just the counts of a histogram metric value.
func histCounts(v metrics.Value) []uint64 {
	if v.Kind() != metrics.KindFloat64Histogram {
		return nil
	}
	return v.Float64Histogram().Counts
}

func AsFloat(v metrics.Value) float64 {
	switch v.Kind() {
	case metrics.KindFloat64:
		return v.Float64()
	case metrics.KindUint64:
		return float64(v.Uint64())
	default:
		return 0
	}
}

func AsUint(v metrics.Value) uint64 {
	switch v.Kind() {
	case metrics.KindUint64:
		return v.Uint64()
	case metrics.KindFloat64:
		return uint64(v.Float64())
	default:
		return 0
	}
}

func SafeUint32(v uint64) uint32 {
	if v > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(v)
}
