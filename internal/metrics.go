// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"fmt"
	"runtime/metrics"
)

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
	MetricGCScanGlobals,
	MetricGCScanHeap,
	MetricGCScanStack,
	MetricGCScanTotal,
	MetricGCStackStartingSize,
	MetricGoDebugAllowMultipleVCs,
	MetricGoDebugContainerMaxProcs,
	MetricGoDebugCryptoCustomRand,
	MetricGoDebugEmbedFollowSymlinks,
	MetricGoDebugExecErrDot,
	MetricGoDebugFIPS140EMS,
	MetricGoDebugGoCacheHash,
	MetricGoDebugGoCacheTest,
	MetricGoDebugGoCacheVerify,
	MetricGoDebugGoTestJSONBuildText,
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
	MetricGoDebugTLSMaxRSASize,
	MetricGoDebugTLSSHA1,
	MetricGoDebugUpdateMaxProcs,
	MetricGoDebugURLMaxQueryParams,
	MetricGoDebugURLStrictColons,
	MetricGoDebugWinReadLinkVolume,
	MetricGoDebugWinSymlink,
	MetricGoDebugX509NegativeSerial,
	MetricGoDebugX509RSACRT,
	MetricGoDebugX509SKID256,
	MetricGoDebugX509SSLCertOverridePlatform,
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

func GetIndex(name string) int {
	return metricIndices[name]
}

func ReadAllSamples(samples []metrics.Sample) {
	if len(samples) != len(AllMetricNames) {
		panic(fmt.Sprintf("ReadAllSamples: got %d samples, want %d", len(samples), len(AllMetricNames)))
	}
	metrics.Read(samples)
}
