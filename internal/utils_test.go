// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"math"
	"runtime/metrics"
	"testing"
	"unsafe"

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
			t.Parallel()
			var v metrics.Value
			setMetricValue(&v, tt.kind, tt.scalar, nil)

			val := AsFloat(v)
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
			t.Parallel()
			var v metrics.Value
			setMetricValue(&v, tc.kind, tc.scalar, nil)

			val := AsUint(v)
			assert.Equal(t, tc.expected, val)
		})
	}
}

func TestAsHistogram(t *testing.T) {
	t.Parallel()

	t.Run("ValidHistogram", func(t *testing.T) {
		t.Parallel()
		collector := NewSampleCollector()
		collector.Read()

		// Find a histogram sample
		var histValue metrics.Value
		found := false
		for _, s := range collector.samples {
			if s.Value.Kind() == metrics.KindFloat64Histogram {
				histValue = s.Value
				found = true
				break
			}
		}

		if !found {
			t.Skip("No histogram metrics available in this environment")
		}

		histg := AsHistogram(histValue)
		assert.NotNil(t, histg)
		assert.NotEmpty(t, histg.Buckets)
		assert.NotEmpty(t, histg.Counts)
	})

	t.Run("InvalidHistogram", func(t *testing.T) {
		t.Parallel()
		var v metrics.Value
		setMetricValue(&v, int(metrics.KindUint64), 123, nil)

		h := AsHistogram(v)
		assert.Nil(t, h)
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
			t.Parallel()
			out := SafeUint32(tc.input)
			assert.Equal(t, tc.expect, out)
		})
	}
}
