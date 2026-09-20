// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package collect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMetadataCached locks in the sync.OnceValue caching in metadataOnce:
// histogram bucket boundaries are fixed for the process lifetime, so every
// call after the first must be free.
func TestMetadataCached(t *testing.T) {
	_ = Metadata() // warm the cache

	allocs := testing.AllocsPerRun(100, func() {
		_ = Metadata()
	})
	assert.Zerof(t, allocs, "Metadata should be free after the first call, got %.1f allocs/op", allocs)
}
