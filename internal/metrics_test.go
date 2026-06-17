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

// intentionallyOmitted lists runtime metrics we deliberately do not collect,
// keyed by metric name with the reason. Keep this in sync with AllMetricNames:
// TestMetricCoverage fails if a runtime metric is neither collected nor listed
// here, forcing a conscious decision whenever the Go toolchain adds a metric.
var intentionallyOmitted = map[string]string{
	"/gc/pauses:seconds": "deprecated; identical to /sched/pauses/total/gc:seconds",
}

// TestAllMetricNamesUnique guards against an accidental duplicate entry in
// AllMetricNames, which would waste a sample slot and desync the index map.
func TestAllMetricNamesUnique(t *testing.T) {
	t.Parallel()

	seen := make(map[string]bool, len(AllMetricNames))
	for _, name := range AllMetricNames {
		assert.Falsef(t, seen[name], "duplicate metric name in AllMetricNames: %s", name)
		seen[name] = true
	}
}

// TestAllMetricNamesExist verifies every name we collect is a real metric known
// to this Go runtime, catching typos and metrics removed in a Go upgrade.
func TestAllMetricNamesExist(t *testing.T) {
	t.Parallel()

	known := make(map[string]bool)
	for _, d := range metrics.All() {
		known[d.Name] = true
	}

	for _, name := range AllMetricNames {
		assert.Truef(t, known[name], "metric %q is not provided by runtime/metrics on this Go version", name)
	}
}

// TestMetricCoverage verifies we collect every runtime metric except those in intentionallyOmitted.
func TestMetricCoverage(t *testing.T) {
	t.Parallel()

	collected := make(map[string]bool, len(AllMetricNames))
	for _, name := range AllMetricNames {
		collected[name] = true
	}

	for _, d := range metrics.All() {
		if collected[d.Name] {
			continue
		}
		_, omitted := intentionallyOmitted[d.Name]
		assert.Truef(t, omitted,
			"runtime metric %q is neither collected nor in intentionallyOmitted; add it to AllMetricNames or document the omission",
			d.Name)
	}

	// Keep the omission list honest: every omitted metric must still exist.
	known := make(map[string]bool)
	for _, d := range metrics.All() {
		known[d.Name] = true
	}
	for name := range intentionallyOmitted {
		assert.Truef(t, known[name], "intentionallyOmitted lists %q, which no longer exists in runtime/metrics", name)
		assert.Falsef(t, collected[name], "%q is in both intentionallyOmitted and AllMetricNames", name)
	}
}
