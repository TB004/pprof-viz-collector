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

	t.Run("IncludeBuckets_ForceFull", func(t *testing.T) {
		res := c.Collect(true, true)
		assert.NotNil(t, res)

		// Check all categories are non-nil
		assert.NotNil(t, res.Cgo)
		assert.NotNil(t, res.Cpu)
		assert.NotNil(t, res.Gc)
		assert.NotNil(t, res.GoDebug)
		assert.NotNil(t, res.Memory)
		assert.NotNil(t, res.Scheduler)
		assert.NotNil(t, res.Sync)

		// In force full mode, fields should be populated if non-zero.
		// Some metrics might be 0 in tests (like CPU time), so we check those that
		// are guaranteed to be non-zero like TotalBytes.
		assert.NotNil(t, res.Memory.TotalBytes)

		// Check for some expected non-nil histograms
		// In the first message (forceFull), we use UpdatedCounts even for full snapshots.
		if res.Gc.BytesAllocsBySize != nil {
			assert.NotEmpty(t, res.Gc.BytesAllocsBySize.Buckets)
			assert.NotEmpty(t, res.Gc.BytesAllocsBySize.UpdatedCounts)
		}
		if res.Scheduler.LatenciesSeconds != nil {
			assert.NotEmpty(t, res.Scheduler.LatenciesSeconds.Buckets)
			assert.NotEmpty(t, res.Scheduler.LatenciesSeconds.UpdatedCounts)
		}
	})

	t.Run("ExcludeBuckets_ForceFull", func(t *testing.T) {
		res := c.Collect(false, true)
		assert.NotNil(t, res)

		// Histograms should still have updated counts but no buckets
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
		// Create a manual collector
		c2 := NewSampleCollector()
		
		res := c2.Collect(true, false)
		// Since samples and lastSamples are both zero, they are equal.
		// So result fields should be nil.
		assert.Nil(t, res.Memory.TotalBytes)
		assert.Nil(t, res.Cpu.TotalSeconds)
		assert.Nil(t, res.Gc.BytesAllocsBySize)
	})
}
