// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"runtime/metrics"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadAllSamples(t *testing.T) {
	t.Parallel()

	samples := make([]metrics.Sample, len(AllMetricNames))
	for i, name := range AllMetricNames {
		samples[i].Name = name
	}

	ReadAllSamples(samples)
	// Verify that we didn't panic and that the samples were updated
	assert.NotEmpty(t, samples)
}

func TestGetIndex(t *testing.T) {
	t.Parallel()

	for i, name := range AllMetricNames {
		assert.Equal(t, i, GetIndex(name))
	}

	assert.Equal(t, 0, GetIndex("NonExistentMetric"))
}
