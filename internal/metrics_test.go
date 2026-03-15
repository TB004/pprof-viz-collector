// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetrieveCPUSamples(t *testing.T) {
	samples := RetrieveCPUSamples()
	assert.NotEmpty(t, samples)
	assert.Len(t, samples, 11) // We expect exactly 11 CPU samples
}

func TestRetrieveCgoSamples(t *testing.T) {
	samples := RetrieveCgoSamples()
	assert.NotEmpty(t, samples)
	assert.Len(t, samples, 1) // We expect exactly 1 cgo sample
}

func TestRetrieveGCSamples(t *testing.T) {
	samples := RetrieveGCSamples()
	assert.NotEmpty(t, samples)
	assert.Len(t, samples, 26) // We expect exactly 26 GC samples
}

func TestRetrieveGoDebugSamples(t *testing.T) {
	samples := RetrieveGoDebugSamples()
	// GoDebug metrics might be empty if none are enabled/tracked, but the map should exist
	assert.NotNil(t, samples)
	assert.Len(t, samples, 45) // We expect exactly 45 GoDebug samples
}

func TestRetrieveMemorySamples(t *testing.T) {
	samples := RetrieveMemorySamples()
	assert.NotEmpty(t, samples)
	assert.Len(t, samples, 14) // We expect exactly 14 memory samples
}

func TestRetrieveSchedulerSamples(t *testing.T) {
	samples := RetrieveSchedulerSamples()
	assert.NotEmpty(t, samples)
	assert.Len(t, samples, 13) // We expect exactly 13 scheduler samples
}

func TestRetrieveSyncSamples(t *testing.T) {
	samples := RetrieveSyncSamples()
	assert.NotEmpty(t, samples)
	assert.Len(t, samples, 1) // We expect exactly 1 sync sample
}
