// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"math"
	"runtime/metrics"

	"github.com/TB004/pprof-viz-agent/internal/proto"
	"github.com/TB004/pprof-viz-agent/pkg"
)

// ReadSamples reads the given metric names and returns a map from name to Sample.
func ReadSamples(names []string) map[string]metrics.Sample {
	samples := make([]metrics.Sample, len(names))
	for i, n := range names {
		samples[i].Name = n
	}
	metrics.Read(samples)
	m := make(map[string]metrics.Sample, len(samples))
	for _, s := range samples {
		m[s.Name] = s
	}
	return m
}

// AsHistogram returns the sample as a pkg.Histogram when possible.
func AsHistogram(s metrics.Sample) *pkg.Histogram {
	if s.Value.Kind() != metrics.KindFloat64Histogram {
		return nil
	}
	h := s.Value.Float64Histogram()

	buckets := make([]float32, len(h.Buckets))
	for i, b := range h.Buckets {
		buckets[i] = float32(b)
	}

	return &pkg.Histogram{
		Counts:  h.Counts,
		Buckets: buckets,
	}
}

// AsFloat returns the sample as float64 when possible.
// For KindFloat64 returns the float64 value. For KindUint64 returns float64(uint64).
func AsFloat(s metrics.Sample) float64 {
	switch s.Value.Kind() {
	case metrics.KindFloat64:
		return s.Value.Float64()
	case metrics.KindUint64:
		return float64(s.Value.Uint64())
	default:
		return 0
	}
}

// AsUint returns the sample as uint64 when possible.
// For KindUint64 returns uint64. For KindFloat64 returns uint64(value) after truncation.
func AsUint(s metrics.Sample) uint64 {
	switch s.Value.Kind() {
	case metrics.KindUint64:
		return s.Value.Uint64()
	case metrics.KindFloat64:
		return uint64(s.Value.Float64())
	default:
		return 0
	}
}

func ToProtoHistogram(h *pkg.Histogram) *proto.Histogram {
	if h == nil {
		return nil
	}
	return &proto.Histogram{
		Counts:  h.Counts,
		Buckets: h.Buckets,
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
