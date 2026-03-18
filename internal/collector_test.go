// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSampleCollector(t *testing.T) {
	t.Parallel()

	c := NewSampleCollector()
	assert.NotNil(t, c)
	assert.Len(t, c.samples, len(AllMetricNames))

	for i, name := range AllMetricNames {
		assert.Equal(t, name, c.samples[i].Name)
	}
}

func TestSampleCollector_ReadAndCollect(t *testing.T) {
	t.Parallel()

	c := NewSampleCollector()
	c.Read()

	t.Run("IncludeBuckets", func(t *testing.T) {
		res := c.Collect(true)
		assert.NotNil(t, res)

		// Check all categories are non-nil
		assert.NotNil(t, res.Cgo)
		assert.NotNil(t, res.Cpu)
		assert.NotNil(t, res.Gc)
		assert.NotNil(t, res.GoDebug)
		assert.NotNil(t, res.Memory)
		assert.NotNil(t, res.Scheduler)
		assert.NotNil(t, res.Sync)

		// Check for some expected non-nil histograms
		if res.Gc.BytesAllocsBySize != nil {
			assert.NotEmpty(t, res.Gc.BytesAllocsBySize.Buckets)
		}
		if res.Scheduler.LatenciesSeconds != nil {
			assert.NotEmpty(t, res.Scheduler.LatenciesSeconds.Buckets)
		}
	})

	t.Run("ExcludeBuckets", func(t *testing.T) {
		res := c.Collect(false)
		assert.NotNil(t, res)

		// Histograms should still have counts but no buckets
		if res.Gc.BytesAllocsBySize != nil {
			assert.Nil(t, res.Gc.BytesAllocsBySize.Buckets)
			assert.NotNil(t, res.Gc.BytesAllocsBySize.Counts)
		}
		if res.Scheduler.LatenciesSeconds != nil {
			assert.Nil(t, res.Scheduler.LatenciesSeconds.Buckets)
			assert.NotNil(t, res.Scheduler.LatenciesSeconds.Counts)
		}
	})
}
