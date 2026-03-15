// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"math"
	"runtime/metrics"
	"testing"
	"unsafe"

	"github.com/TB004/pprof-viz-agent/pkg"
	"github.com/stretchr/testify/assert"
)

// params mimics the internal structure of metrics.Value to allow us to inject values for testing.
// This is necessary because metrics.Value fields are private.
// CAUTION: This depends on the internal layout of runtime/metrics.Value.
type metricValueStruct struct {
	kind    int // metrics.Kind is int
	scalar  uint64
	pointer unsafe.Pointer
}

func setMetricValue(v *metrics.Value, kind int, scalar uint64, ptr unsafe.Pointer) {
	dst := (*metricValueStruct)(unsafe.Pointer(v))
	dst.kind = kind
	dst.scalar = scalar
	dst.pointer = ptr
}

func TestReadSamples(t *testing.T) {
	name := []string{MetricCgoGoToCCalls}
	samples := ReadSamples(name)
	assert.NotEmpty(t, samples)
	assert.NotNil(t, samples[MetricCgoGoToCCalls])
}

func TestAsFloat(t *testing.T) {
	tests := []struct {
		name     string
		kind     int
		scalar   uint64
		expected float64
	}{
		{
			name:     "Float64",
			kind:     int(metrics.KindFloat64),
			scalar:   123456789, // This is interpreted as bits of float64
			expected: 0,         // Will be calculated below
		},
		{
			name:     "Uint64",
			kind:     int(metrics.KindUint64),
			scalar:   42,
			expected: 42.0,
		},
		{
			name:     "BadKind",
			kind:     int(metrics.KindBad),
			scalar:   100,
			expected: 0.0,
		},
	}

	// Prepare float bits for the first case
	fVal := 123.456
	tests[0].scalar = *(*uint64)(unsafe.Pointer(&fVal))
	tests[0].expected = fVal

	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s metrics.Sample
			setMetricValue(&s.Value, tt.kind, tt.scalar, nil)

			val := AsFloat(s)
			assert.InDelta(t, tt.expected, val, 0.0001)
		})
	}
}

func TestAsUint(t *testing.T) {
	tests := []struct {
		name     string
		kind     int
		scalar   uint64
		expected uint64
	}{
		{
			name:     "Uint64",
			kind:     int(metrics.KindUint64),
			scalar:   100,
			expected: 100,
		},
		{
			name:     "Float64",
			kind:     int(metrics.KindFloat64),
			scalar:   0, // Calculated below
			expected: 55,
		},
		{
			name:     "BadKind",
			kind:     int(metrics.KindBad),
			scalar:   100,
			expected: 0,
		},
	}

	// Prepare float bits
	fVal := 55.9 // Should truncate to 55
	tests[1].scalar = *(*uint64)(unsafe.Pointer(&fVal))

	t.Parallel()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var s metrics.Sample
			setMetricValue(&s.Value, tc.kind, tc.scalar, nil)

			val := AsUint(s)
			assert.Equal(t, tc.expected, val)
		})
	}
}

func TestAsHistogram(t *testing.T) {
	t.Parallel()

	t.Run("ValidHistogram", func(t *testing.T) {
		name := []string{MetricGCHeapAllocsBySize}
		samples := ReadSamples(name)

		histg := AsHistogram(samples[MetricGCHeapAllocsBySize])
		assert.NotNil(t, histg)
		assert.NotEmpty(t, histg.Buckets)
		assert.NotEmpty(t, histg.Counts)
	})

	t.Run("InvalidHistogram", func(t *testing.T) {
		// Constructing a fake Float64Histogram is complex because it involves pointers.
		// Instead, we verify that non-histograms return nil.
		var s metrics.Sample
		setMetricValue(&s.Value, int(metrics.KindUint64), 123, nil)

		h := AsHistogram(s)
		assert.Nil(t, h)
	})
}

func TestToProtoHistogram(t *testing.T) {
	t.Parallel()

	t.Run("NilInput", func(t *testing.T) {
		assert.Nil(t, ToProtoHistogram(nil))
	})

	t.Run("ValidInput", func(t *testing.T) {
		h := &pkg.Histogram{
			Counts:  []uint64{1, 2, 3},
			Buckets: []float32{0.1, 0.5, 1.0},
		}
		protoH := ToProtoHistogram(h)
		assert.NotNil(t, protoH)
		assert.Equal(t, h.Counts, protoH.Counts)
		assert.Equal(t, h.Buckets, protoH.Buckets)
	})
}

func TestSafeUint32(t *testing.T) {
	tests := []struct {
		name   string
		input  uint64
		expect uint32
	}{
		{
			name:   "Value greater than max uint32",
			input:  18446744073,
			expect: math.MaxUint32,
		},
		{
			name:   "Value lower than max uint32",
			input:  10,
			expect: 10,
		},
	}

	t.Parallel()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := SafeUint32(tc.input)
			assert.Equal(t, tc.expect, out)
		})
	}
}
