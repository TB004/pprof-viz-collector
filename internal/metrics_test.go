// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime/metrics"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadAllSamples(t *testing.T) {
	t.Parallel()

	samples := make([]metrics.Sample, len(AllMetricNames))
	for i, name := range AllMetricNames {
		samples[i].Name = name
	}

	ReadAllSamples(samples)

	assert.NotEmpty(t, samples)
}

// intentionallyOmitted lists runtime metrics we deliberately do not collect,
// keyed by metric name with the reason. Keep this in sync with AllMetricNames:
// TestMetricCoverage fails if a runtime metric is neither collected nor listed
// here, forcing a conscious decision whenever the Go toolchain adds a metric.
var intentionallyOmitted = map[string]string{
	"/gc/pauses:seconds": "deprecated; identical to /sched/pauses/total/gc:seconds",
}

// TestAllMetricNamesUnique guards against an accidental duplicate entry in
// AllMetricNames.
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

// TestGeneratedFilesUpToDate regenerates metrics_gen.go and collect_gen.go
// from table.go into a scratch directory and fails if that differs from what's
// checked in — i.e. someone edited table.go without running `go generate`.
func TestGeneratedFilesUpToDate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cmd := exec.Command("go", "run", "./gen", "-out", dir) //nolint:gosec // fixed args, not user input
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "go run ./gen failed: %s", out)

	for _, name := range []string{"metrics_gen.go", "collect_gen.go"} {
		want, err := os.ReadFile(filepath.Join(dir, name))
		require.NoError(t, err)
		got, err := os.ReadFile(name)
		require.NoError(t, err)
		assert.Equalf(t, string(want), string(got),
			"%s is stale relative to table.go; run `go generate ./internal/...`", name)
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
