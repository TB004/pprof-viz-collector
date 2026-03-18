// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"math"
	"runtime/metrics"

	"github.com/TB004/pprof-viz-agent/pkg/models"
)

// AsHistogram returns the value as a models.Histogram when possible.
func AsHistogram(v metrics.Value) *models.Histogram {
	if v.Kind() != metrics.KindFloat64Histogram {
		return nil
	}
	h := v.Float64Histogram()

	buckets := make([]float32, len(h.Buckets))
	for i, b := range h.Buckets {
		buckets[i] = float32(b)
	}

	return &models.Histogram{
		Counts:  h.Counts,
		Buckets: buckets,
	}
}

// AsFloat returns the value as float64 when possible.
// For KindFloat64 returns the float64 value. For KindUint64 returns float64(uint64).
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

// AsUint returns the value as uint64 when possible.
// For KindUint64 returns uint64. For KindFloat64 returns uint64(value) after truncation.
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

// SafeUint32 clamps the uint64 value to math.MaxUint32 to prevent overflow
// when converting to uint32.
func SafeUint32(v uint64) uint32 {
	if v > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(v)
}
