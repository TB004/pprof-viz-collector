// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import "runtime/metrics"

// AllMetricNames is a comprehensive list of all supported Go runtime metrics.
var AllMetricNames = []string{
	MetricCgoGoToCCalls,
	MetricCPUClassesGCMarkAssist,
	MetricCPUClassesGCMarkDedicated,
	MetricCPUClassesGCMarkIdle,
	MetricCPUClassesGCPause,
	MetricCPUClassesGCTotal,
	MetricCPUClassesIdle,
	MetricCPUClassesTotal,
	MetricCPUClassesUser,
	MetricCPUClassesScavengeAssist,
	MetricCPUClassesScavengeBackground,
	MetricCPUClassesScavengeTotal,
	MetricGCCleanupsExecuted,
	MetricGCCleanupsQueued,
	MetricGCCyclesAutomatic,
	MetricGCCyclesForced,
	MetricGCCyclesTotal,
	MetricGCFinalizersExecuted,
	MetricGCFinalizersQueued,
	MetricGCGOGC,
	MetricGCGOMemLimit,
	MetricGCHeapAllocsBySize,
	MetricGCHeapAllocs,
	MetricGCHeapAllocsObjects,
	MetricGCHeapFreesBySize,
	MetricGCHeapFrees,
	MetricGCHeapFreesObjects,
	MetricGCHeapGoal,
	MetricGCHeapLive,
	MetricGCHeapObjects,
	MetricGCHeapTinyAllocs,
	MetricGCLimiterLastEnabled,
	MetricGCPausesSeconds,
	MetricGCScanGlobals,
	MetricGCScanHeap,
	MetricGCScanStack,
	MetricGCScanTotal,
	MetricGCStackStartingSize,
	MetricGoDebugAllowMultipleVCs,
	MetricGoDebugAsyncTimerChan,
	MetricGoDebugContainerMaxProcs,
	MetricGoDebugCryptoCustomRand,
	MetricGoDebugEmbedFollowSymlinks,
	MetricGoDebugExecErrDot,
	MetricGoDebugGoCacheHash,
	MetricGoDebugGoCacheTest,
	MetricGoDebugGoCacheVerify,
	MetricGoDebugGoTestJSONBuildText,
	MetricGoDebugGoTypesAlias,
	MetricGoDebugHTMLMetaContentURLEscape,
	MetricGoDebugHTTP2Client,
	MetricGoDebugHTTP2Server,
	MetricGoDebugHTTPCookieMaxNum,
	MetricGoDebugHTTPLaxContentLength,
	MetricGoDebugHTTPMuxGo121,
	MetricGoDebugHTTPServeContentKeepHeaders,
	MetricGoDebugInstallGoroot,
	MetricGoDebugMultipartMaxHeaders,
	MetricGoDebugMultipartMaxParts,
	MetricGoDebugMultipathTCP,
	MetricGoDebugNetEdns0,
	MetricGoDebugPanicNil,
	MetricGoDebugRandAutoSeed,
	MetricGoDebugRandSeedNop,
	MetricGoDebugRSA1024Min,
	MetricGoDebugTarInsecurePath,
	MetricGoDebugTLS10Server,
	MetricGoDebugTLS3DES,
	MetricGoDebugTLSMaxRSASize,
	MetricGoDebugTLSRSAKex,
	MetricGoDebugTLSSHA1,
	MetricGoDebugTLSUnsafeEKM,
	MetricGoDebugUpdateMaxProcs,
	MetricGoDebugURLMaxQueryParams,
	MetricGoDebugURLStrictColons,
	MetricGoDebugWinReadLinkVolume,
	MetricGoDebugWinSymlink,
	MetricGoDebugX509KeyPairLeaf,
	MetricGoDebugX509NegativeSerial,
	MetricGoDebugX509RSACRT,
	MetricGoDebugX509SKID256,
	MetricGoDebugX509UseFallbackRoots,
	MetricGoDebugX509UsePolicies,
	MetricGoDebugZipInsecurePath,
	MetricMemoryClassesHeapFree,
	MetricMemoryClassesHeapObjects,
	MetricMemoryClassesHeapReleased,
	MetricMemoryClassesHeapStacks,
	MetricMemoryClassesHeapUnused,
	MetricMemoryClassesMetadataMCacheFree,
	MetricMemoryClassesMetadataMCacheInUse,
	MetricMemoryClassesMetadataMSpanFree,
	MetricMemoryClassesMetadataMSpanInUse,
	MetricMemoryClassesMetadataOther,
	MetricMemoryClassesOSStacks,
	MetricMemoryClassesOther,
	MetricMemoryClassesProfilingBuckets,
	MetricMemoryClassesTotal,
	MetricSchedGOMAXPROCS,
	MetricSchedGoroutinesCreated,
	MetricSchedGoroutinesNotInGo,
	MetricSchedGoroutinesRunnable,
	MetricSchedGoroutinesRunning,
	MetricSchedGoroutinesWaiting,
	MetricSchedGoroutines,
	MetricSchedLatencies,
	MetricSchedPausesStoppingGC,
	MetricSchedPausesStoppingOther,
	MetricSchedPausesTotalGC,
	MetricSchedPausesTotalOther,
	MetricSchedThreadsTotal,
	MetricSyncMutexWaitTotal,
}

var metricIndices map[string]int

func init() {
	metricIndices = make(map[string]int, len(AllMetricNames))
	for i, name := range AllMetricNames {
		metricIndices[name] = i
	}
}

// GetIndex returns the index of the given metric name in the pre-allocated slice.
func GetIndex(name string) int {
	return metricIndices[name]
}

// ReadAllSamples reads all supported Go runtime metrics in a single pass.
func ReadAllSamples(samples []metrics.Sample) {
	if len(samples) != len(AllMetricNames) {
		return
	}
	metrics.Read(samples)
}
