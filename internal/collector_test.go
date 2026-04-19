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
