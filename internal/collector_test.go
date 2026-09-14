// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSampleCollector_ReadAndCollect(t *testing.T) {
	c := NewSampleCollector()

	// Initial read to populate samples
	c.Read()

	t.Run("ForceFull", func(t *testing.T) {
		res := c.Collect()
		assert.NotNil(t, res)

		assert.NotNil(t, res.Cgo)
		assert.NotNil(t, res.Cpu)
		assert.NotNil(t, res.Gc)
		assert.NotNil(t, res.GoDebug)
		assert.NotNil(t, res.Memory)
		assert.NotNil(t, res.Scheduler)
		assert.NotNil(t, res.Sync)

		assert.NotNil(t, res.Memory.TotalBytes)

		if res.Gc.BytesAllocsBySize != nil {
			assert.Nil(t, res.Gc.BytesAllocsBySize.Buckets)
			assert.NotEmpty(t, res.Gc.BytesAllocsBySize.UpdatedCounts)
		}
		if res.Scheduler.LatenciesSeconds != nil {
			assert.Nil(t, res.Scheduler.LatenciesSeconds.Buckets)
			assert.NotEmpty(t, res.Scheduler.LatenciesSeconds.UpdatedCounts)
		}
	})

	t.Run("DiffCollection_NoChange", func(t *testing.T) {
		c2 := NewSampleCollector()

		res := c2.Collect()
		// Since samples and lastSamples are both zero, they are equal.
		// So result fields should be nil.
		assert.Nil(t, res.Memory.TotalBytes)
		assert.Nil(t, res.Cpu.TotalSeconds)
		assert.Nil(t, res.Gc.BytesAllocsBySize)
	})
}

// TestResetBaseline verifies that resetBaseline restores full-state collection
// semantics on a collector that has already accumulated real values, which is
// what makes pooled reuse for one-shot snapshots correct.
func TestResetBaseline(t *testing.T) {
	c := NewSampleCollector()

	c.Read()
	_ = c.Collect() // first read returns full state
	c.Read()
	diff := c.Collect()
	// Total memory rarely changes between two back-to-back reads, so the diff
	// typically omits it; the point is we are no longer in the full-state regime.
	_ = diff

	// Simulate what the pool does before handing the collector back out.
	c.resetBaseline()
	c.Read()
	res := c.Collect()

	// A running program always has non-zero total memory, so a full snapshot
	// must include it again.
	assert.NotNil(t, res.Memory.TotalBytes, "resetBaseline should restore full-state collection")
}

// TestSnapshotFullStateAcrossCalls verifies the pooled Snapshot helper returns a
// full snapshot on every call, even when a recycled collector is reused.
func TestSnapshotFullStateAcrossCalls(t *testing.T) {
	for i := range 5 {
		res := Snapshot()
		assert.NotNil(t, res)
		assert.NotNilf(t, res.Memory.TotalBytes, "Snapshot must return full state on call %d", i)
	}
}

// BenchmarkCollect measures the per-tick cost of Read+Collect: exactly what
// StreamAll's send() does on every interval.
func BenchmarkCollect(b *testing.B) {
	c := NewSampleCollector()
	c.Read()

	b.ReportAllocs()
	for b.Loop() {
		c.Read()
		_ = c.Collect()
	}
}

// BenchmarkSnapshot measures the pooled one-shot path used by SnapshotAll.
func BenchmarkSnapshot(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = Snapshot()
	}
}
