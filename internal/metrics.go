// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"fmt"
	"runtime/metrics"
)

//go:generate go run ./gen

func ReadAllSamples(samples []metrics.Sample) {
	if len(samples) != len(AllMetricNames) {
		panic(fmt.Sprintf("ReadAllSamples: got %d samples, want %d", len(samples), len(AllMetricNames)))
	}
	metrics.Read(samples)
}
