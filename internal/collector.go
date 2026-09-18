// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"runtime/metrics"
	"sync"

	"github.com/TB004/pprof-viz-collector/internal/proto"
)

// SampleCollector manages a pre-allocated slice of metrics.Sample for efficient collection.
type SampleCollector struct {
	samples     []metrics.Sample
	lastSamples []metrics.Sample
}

// NewSampleCollector creates a new SampleCollector with pre-allocated samples.
func NewSampleCollector() *SampleCollector {
	samples := make([]metrics.Sample, len(AllMetricNames))
	lastSamples := make([]metrics.Sample, len(AllMetricNames))
	for i, name := range AllMetricNames {
		samples[i].Name = name
		lastSamples[i].Name = name
	}
	return &SampleCollector{
		samples:     samples,
		lastSamples: lastSamples,
	}
}

// Read updates the internal samples with current runtime metrics.
func (c *SampleCollector) Read() {
	c.samples, c.lastSamples = c.lastSamples, c.samples
	ReadAllSamples(c.samples)
}

// resetBaseline clears the current sample values so that the next Read+Collect
// reports the full current state rather than a diff.
//
// Histogram metrics are cleared by zeroing their Counts in place rather than
// replacing the whole Value. This matters because runtime/metrics only
// reuses an existing *Float64Histogram (avoiding a fresh allocation of the
// histogram and its counts slice) when the Value it's given already has
// KindFloat64Histogram and a non-nil pointer (see float64HistOrInit in
// runtime/metrics.go). Overwriting the Value with its zero value would drop
// that pointer, so a pooled collector would force the runtime to reallocate
// every histogram metric on its very next Read.
func (c *SampleCollector) resetBaseline() {
	for i := range c.samples {
		val := c.samples[i].Value
		if val.Kind() == metrics.KindFloat64Histogram {
			if h := val.Float64Histogram(); h != nil {
				for j := range h.Counts {
					h.Counts[j] = 0
				}
			}
			continue
		}
		c.samples[i].Value = metrics.Value{}
	}
}

// collectorPool recycles SampleCollectors for one-shot snapshots so repeated
// Snapshot calls don't each allocate two full sample slices.
var collectorPool = sync.Pool{
	New: func() any { return NewSampleCollector() },
}

// Snapshot performs a single full collection of all runtime metrics using a
// pooled collector.
func Snapshot() *proto.AllResponse {
	c := collectorPool.Get().(*SampleCollector)
	defer collectorPool.Put(c)

	c.resetBaseline()
	c.Read()
	return c.Collect()
}

// Collect maps the current samples into a proto response, returning only fields that changed
// since the last Read. On a fresh collector, all non-zero fields are returned.
// Histogram bucket boundaries are never included; use GetMetadata to retrieve them once.
func (c *SampleCollector) Collect() *proto.AllResponse {
	return buildAllResponse(c)
}

// hasChanged reports whether the metric at idx differs between the current
// and previous Read. If the previous value is uninitialized (KindBad), it is
// treated as a delta from the default (0) value.
func (c *SampleCollector) hasChanged(idx Metric) bool {
	cur := c.samples[idx].Value
	last := c.lastSamples[idx].Value

	if last.Kind() == metrics.KindBad {
		switch cur.Kind() {
		case metrics.KindUint64:
			return cur.Uint64() != 0
		case metrics.KindFloat64:
			return cur.Float64() != 0
		case metrics.KindFloat64Histogram:
			h := cur.Float64Histogram()
			if h == nil {
				return false
			}
			for _, count := range h.Counts {
				if count != 0 {
					return true
				}
			}
			return false
		case metrics.KindBad:
			return false // Both are bad
		default:
			return true
		}
	}

	if cur.Kind() != last.Kind() {
		return true
	}

	switch cur.Kind() {
	case metrics.KindUint64:
		return cur.Uint64() != last.Uint64()
	case metrics.KindFloat64:
		return cur.Float64() != last.Float64()
	case metrics.KindBad:
		return false
	case metrics.KindFloat64Histogram:
		hCur := cur.Float64Histogram()
		hLast := last.Float64Histogram()
		if hCur == nil || hLast == nil {
			return hCur != hLast
		}
		if len(hCur.Counts) != len(hLast.Counts) {
			return true
		}
		for i, v := range hCur.Counts {
			if v != hLast.Counts[i] {
				return true
			}
		}
		return false
	default:
		return true
	}
}

// vUint returns the current uint64 value of idx, or nil if it hasn't changed
// since the last Read.
func (c *SampleCollector) vUint(idx Metric) *uint64 {
	if !c.hasChanged(idx) {
		return nil
	}
	return new(AsUint(c.samples[idx].Value))
}

// vFloat returns the current float64 value of idx, or nil if it hasn't
// changed since the last Read.
func (c *SampleCollector) vFloat(idx Metric) *float64 {
	if !c.hasChanged(idx) {
		return nil
	}
	return new(AsFloat(c.samples[idx].Value))
}

// vUint32 returns the current value of idx narrowed to uint32, or nil if it
// hasn't changed since the last Read.
func (c *SampleCollector) vUint32(idx Metric) *uint32 {
	if !c.hasChanged(idx) {
		return nil
	}
	return new(SafeUint32(AsUint(c.samples[idx].Value)))
}

// vHist returns the histogram buckets of idx that changed since the last
// Read, or nil if none did. Unlike hasChanged+vUint's two-step pattern, this
// makes a single pass over the counts and only allocates the result map once
// a difference is actually found.
func (c *SampleCollector) vHist(idx Metric) *proto.Histogram {
	counts := histCounts(c.samples[idx].Value)
	if counts == nil {
		return nil
	}

	var lastCounts []uint64
	if last := c.lastSamples[idx].Value; last.Kind() == metrics.KindFloat64Histogram {
		if hLast := last.Float64Histogram(); hLast != nil {
			lastCounts = hLast.Counts
		}
	}

	var updated map[uint32]uint64
	for i, v := range counts {
		var lv uint64
		if i < len(lastCounts) {
			lv = lastCounts[i]
		}
		if v != lv {
			if updated == nil {
				updated = make(map[uint32]uint64, len(counts)-i)
			}
			updated[uint32(i)] = v
		}
	}

	if len(updated) == 0 {
		return nil
	}

	return &proto.Histogram{
		UpdatedCounts: updated,
	}
}
