// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import "runtime/metrics"

// RetrieveCPUSamples returns CPU-related runtime/metrics samples.
func RetrieveCPUSamples() map[string]metrics.Sample {
	names := [11]string{
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
	}

	return ReadSamples(names[:])
}

// RetrieveCgoSamples returns cgo-related runtime/metrics samples.
func RetrieveCgoSamples() map[string]metrics.Sample {
	names := [1]string{
		MetricCgoGoToCCalls,
	}

	return ReadSamples(names[:])
}

// RetrieveGCSamples returns GC-related runtime/metrics samples.
func RetrieveGCSamples() map[string]metrics.Sample {
	names := [26]string{
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
	}

	return ReadSamples(names[:])
}

// RetrieveGoDebugSamples returns godebug-related runtime/metrics samples.
func RetrieveGoDebugSamples() map[string]metrics.Sample {
	names := [45]string{
		MetricGoDebugAllowMultipleVCs,
		MetricGoDebugAsyncTimerChan,
		MetricGoDebugContainerMaxProcs,
		MetricGoDebugCryptoCustomRand,
		MetricGoDebugEmbedFollowSymlinks,
		MetricGoDebugExecErrDot,
		MetricGoDebugGoCacheHash,
		MetricGoDebugGoCacheTest,
		MetricGoDebugGoCacheVerify,
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
	}

	return ReadSamples(names[:])
}

// RetrieveMemorySamples returns memory-related runtime/metrics samples.
func RetrieveMemorySamples() map[string]metrics.Sample {
	names := [14]string{
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
	}

	return ReadSamples(names[:])
}

// RetrieveSchedulerSamples returns scheduler-related runtime/metrics samples.
func RetrieveSchedulerSamples() map[string]metrics.Sample {
	names := [13]string{
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
	}

	return ReadSamples(names[:])
}

// RetrieveSyncSamples returns sync-related runtime/metrics samples.
func RetrieveSyncSamples() map[string]metrics.Sample {
	names := [1]string{
		MetricSyncMutexWaitTotal,
	}

	return ReadSamples(names[:])
}
