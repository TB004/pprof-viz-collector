// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package main

// kind is the runtime/metrics value shape a metric is read as, and therefore
// which vXxx helper in collector.go's Collect reads it and which Go type its
// proto field must be.
type kind int

const (
	kindUint64 kind = iota
	kindUint32
	kindFloat64
	kindHistogram
)

// field is one row of the single source of truth for every runtime/metrics
// value this collector exposes.
//
//   - name is the Go identifier suffix; the generator emits it as MetricName.
//   - path is the runtime/metrics path, e.g. "/gc/heap/allocs:bytes".
//   - group is the proto.AllResponse field this metric belongs under, e.g.
//     "Gc" for proto.AllResponse.Gc (of type *proto.GCResponse).
//   - protoField is the field name on that group's proto message, e.g.
//     "BytesAllocs" for proto.GCResponse.BytesAllocs.
type field struct {
	name       string
	path       string
	kind       kind
	group      string
	protoField string
}

var groupOrder = []struct {
	group   string
	message string
}{
	{"Cgo", "CgoResponse"},
	{"Cpu", "CPUResponse"},
	{"Gc", "GCResponse"},
	{"GoDebug", "GoDebugResponse"},
	{"Memory", "MemoryResponse"},
	{"Scheduler", "SchedulerResponse"},
	{"Sync", "SyncResponse"},
}

var metrics = []field{
	{"CgoGoToCCalls", "/cgo/go-to-c-calls:calls", kindUint64, "Cgo", "CgoCalls"},

	{"CPUClassesGCMarkAssist", "/cpu/classes/gc/mark/assist:cpu-seconds", kindFloat64, "Cpu", "GcAssistSeconds"},
	{"CPUClassesGCMarkDedicated", "/cpu/classes/gc/mark/dedicated:cpu-seconds", kindFloat64, "Cpu", "GcDedicatedSeconds"},
	{"CPUClassesGCMarkIdle", "/cpu/classes/gc/mark/idle:cpu-seconds", kindFloat64, "Cpu", "GcIdleSeconds"},
	{"CPUClassesGCPause", "/cpu/classes/gc/pause:cpu-seconds", kindFloat64, "Cpu", "GcPauseSeconds"},
	{"CPUClassesGCTotal", "/cpu/classes/gc/total:cpu-seconds", kindFloat64, "Cpu", "GcTotalSeconds"},
	{"CPUClassesIdle", "/cpu/classes/idle:cpu-seconds", kindFloat64, "Cpu", "IdleSeconds"},
	{"CPUClassesTotal", "/cpu/classes/total:cpu-seconds", kindFloat64, "Cpu", "TotalSeconds"},
	{"CPUClassesUser", "/cpu/classes/user:cpu-seconds", kindFloat64, "Cpu", "UserSeconds"},
	{"CPUClassesScavengeAssist", "/cpu/classes/scavenge/assist:cpu-seconds", kindFloat64, "Cpu", "ScavengeAssistSeconds"},
	{"CPUClassesScavengeBackground", "/cpu/classes/scavenge/background:cpu-seconds", kindFloat64, "Cpu", "ScavengeBackgroundSeconds"},
	{"CPUClassesScavengeTotal", "/cpu/classes/scavenge/total:cpu-seconds", kindFloat64, "Cpu", "ScavengeTotalSeconds"},

	{"GCHeapAllocsBySize", "/gc/heap/allocs-by-size:bytes", kindHistogram, "Gc", "BytesAllocsBySize"},
	{"GCHeapFreesBySize", "/gc/heap/frees-by-size:bytes", kindHistogram, "Gc", "BytesFreesBySize"},
	{"GCCleanupsExecuted", "/gc/cleanups/executed:cleanups", kindUint64, "Gc", "ExecutedCleanups"},
	{"GCCleanupsQueued", "/gc/cleanups/queued:cleanups", kindUint64, "Gc", "QueuedCleanups"},
	{"GCCyclesAutomatic", "/gc/cycles/automatic:gc-cycles", kindUint64, "Gc", "AutomaticCycles"},
	{"GCCyclesForced", "/gc/cycles/forced:gc-cycles", kindUint64, "Gc", "ForcedCycles"},
	{"GCCyclesTotal", "/gc/cycles/total:gc-cycles", kindUint64, "Gc", "TotalCycles"},
	{"GCFinalizersExecuted", "/gc/finalizers/executed:finalizers", kindUint64, "Gc", "ExecutedFinalizers"},
	{"GCFinalizersQueued", "/gc/finalizers/queued:finalizers", kindUint64, "Gc", "QueuedFinalizers"},
	{"GCGOMemLimit", "/gc/gomemlimit:bytes", kindUint64, "Gc", "MemoryLimitBytes"},
	{"GCHeapAllocs", "/gc/heap/allocs:bytes", kindUint64, "Gc", "BytesAllocs"},
	{"GCHeapAllocsObjects", "/gc/heap/allocs:objects", kindUint64, "Gc", "ObjectsAllocs"},
	{"GCHeapFrees", "/gc/heap/frees:bytes", kindUint64, "Gc", "BytesFrees"},
	{"GCHeapFreesObjects", "/gc/heap/frees:objects", kindUint64, "Gc", "ObjectsFrees"},
	{"GCHeapGoal", "/gc/heap/goal:bytes", kindUint64, "Gc", "GoalHeapSizeBytes"},
	{"GCHeapLive", "/gc/heap/live:bytes", kindUint64, "Gc", "LiveOccupiedBytes"},
	{"GCHeapObjects", "/gc/heap/objects:objects", kindUint64, "Gc", "ObjectsOccupied"},
	{"GCHeapTinyAllocs", "/gc/heap/tiny/allocs:objects", kindUint64, "Gc", "TinyObjectsAllocs"},
	{"GCLimiterLastEnabled", "/gc/limiter/last-enabled:gc-cycle", kindUint64, "Gc", "LimiterLastEnabledCycle"},
	{"GCScanGlobals", "/gc/scan/globals:bytes", kindUint64, "Gc", "GlobalScannedBytes"},
	{"GCScanHeap", "/gc/scan/heap:bytes", kindUint64, "Gc", "HeapScannedBytes"},
	{"GCScanStack", "/gc/scan/stack:bytes", kindUint64, "Gc", "StackScannedBytes"},
	{"GCScanTotal", "/gc/scan/total:bytes", kindUint64, "Gc", "TotalScannedBytes"},
	{"GCStackStartingSize", "/gc/stack/starting-size:bytes", kindUint64, "Gc", "StackStartingSize"},
	{"GCGOGC", "/gc/gogc:percent", kindUint32, "Gc", "TargetHeapSizePercentage"},

	{"GoDebugAllowMultipleVCs", "/godebug/non-default-behavior/allowmultiplevcs:events", kindUint64, "GoDebug", "AllowMultipleVcsEvents"},
	{"GoDebugContainerMaxProcs", "/godebug/non-default-behavior/containermaxprocs:events", kindUint64, "GoDebug", "ContainerMaxProcsEvents"},
	{"GoDebugCryptoCustomRand", "/godebug/non-default-behavior/cryptocustomrand:events", kindUint64, "GoDebug", "CryptoCustomRandEvents"},
	{"GoDebugEmbedFollowSymlinks", "/godebug/non-default-behavior/embedfollowsymlinks:events", kindUint64, "GoDebug", "EmbedFollowSymlinkEvents"},
	{"GoDebugExecErrDot", "/godebug/non-default-behavior/execerrdot:events", kindUint64, "GoDebug", "ExecErrDotEvents"},
	{"GoDebugFIPS140EMS", "/godebug/non-default-behavior/fips140ems:events", kindUint64, "GoDebug", "Fips140EmsEvents"},
	{"GoDebugGoCacheHash", "/godebug/non-default-behavior/gocachehash:events", kindUint64, "GoDebug", "CacheHashEvents"},
	{"GoDebugGoCacheTest", "/godebug/non-default-behavior/gocachetest:events", kindUint64, "GoDebug", "CacheTestEvents"},
	{"GoDebugGoCacheVerify", "/godebug/non-default-behavior/gocacheverify:events", kindUint64, "GoDebug", "CacheVerifyEvents"},
	{"GoDebugGoTestJSONBuildText", "/godebug/non-default-behavior/gotestjsonbuildtext:events", kindUint64, "GoDebug", "TestJsonBuildTextEvents"},
	{"GoDebugHTMLMetaContentURLEscape", "/godebug/non-default-behavior/htmlmetacontenturlescape:events", kindUint64, "GoDebug", "HtmlMetaContentUrlEscapedEvents"},
	{"GoDebugHTTP2Client", "/godebug/non-default-behavior/http2client:events", kindUint64, "GoDebug", "Http2ClientEvents"},
	{"GoDebugHTTP2Server", "/godebug/non-default-behavior/http2server:events", kindUint64, "GoDebug", "Http2ServerEvents"},
	{"GoDebugHTTPCookieMaxNum", "/godebug/non-default-behavior/httpcookiemaxnum:events", kindUint64, "GoDebug", "HttpCookieMaxNumEvents"},
	{"GoDebugHTTPLaxContentLength", "/godebug/non-default-behavior/httplaxcontentlength:events", kindUint64, "GoDebug", "HttpLaxContentLengthEvents"},
	{"GoDebugHTTPMuxGo121", "/godebug/non-default-behavior/httpmuxgo121:events", kindUint64, "GoDebug", "HttpMuxGo_121Events"},
	{"GoDebugHTTPServeContentKeepHeaders", "/godebug/non-default-behavior/httpservecontentkeepheaders:events", kindUint64, "GoDebug", "HttpServeContentKeepHeadersEvents"},
	{"GoDebugInstallGoroot", "/godebug/non-default-behavior/installgoroot:events", kindUint64, "GoDebug", "InstallGoRootEvents"},
	{"GoDebugMultipartMaxHeaders", "/godebug/non-default-behavior/multipartmaxheaders:events", kindUint64, "GoDebug", "MultipartMaxHeadersEvents"},
	{"GoDebugMultipartMaxParts", "/godebug/non-default-behavior/multipartmaxparts:events", kindUint64, "GoDebug", "MultipartMaxPartsEvents"},
	{"GoDebugMultipathTCP", "/godebug/non-default-behavior/multipathtcp:events", kindUint64, "GoDebug", "MultipathTcpEvents"},
	{"GoDebugNetEdns0", "/godebug/non-default-behavior/netedns0:events", kindUint64, "GoDebug", "NetEdns0Events"},
	{"GoDebugPanicNil", "/godebug/non-default-behavior/panicnil:events", kindUint64, "GoDebug", "PanicNilEvents"},
	{"GoDebugRandAutoSeed", "/godebug/non-default-behavior/randautoseed:events", kindUint64, "GoDebug", "RandAutoSeedEvents"},
	{"GoDebugRandSeedNop", "/godebug/non-default-behavior/randseednop:events", kindUint64, "GoDebug", "RandSeedNopEvents"},
	{"GoDebugRSA1024Min", "/godebug/non-default-behavior/rsa1024min:events", kindUint64, "GoDebug", "Rsa1024MinEvents"},
	{"GoDebugTarInsecurePath", "/godebug/non-default-behavior/tarinsecurepath:events", kindUint64, "GoDebug", "TarInsecurePathEvents"},
	{"GoDebugTLSMaxRSASize", "/godebug/non-default-behavior/tlsmaxrsasize:events", kindUint64, "GoDebug", "TlsMaxRsaSizeEvents"},
	{"GoDebugTLSSHA1", "/godebug/non-default-behavior/tlssha1:events", kindUint64, "GoDebug", "TlsSha1Events"},
	{"GoDebugUpdateMaxProcs", "/godebug/non-default-behavior/updatemaxprocs:events", kindUint64, "GoDebug", "UpdateMaxProcsEvents"},
	{"GoDebugURLMaxQueryParams", "/godebug/non-default-behavior/urlmaxqueryparams:events", kindUint64, "GoDebug", "UrlMaxQueryParamsEvents"},
	{"GoDebugURLStrictColons", "/godebug/non-default-behavior/urlstrictcolons:events", kindUint64, "GoDebug", "UrlStrictColonsEvents"},
	{"GoDebugWinReadLinkVolume", "/godebug/non-default-behavior/winreadlinkvolume:events", kindUint64, "GoDebug", "WinReadLinkVolumeEvents"},
	{"GoDebugWinSymlink", "/godebug/non-default-behavior/winsymlink:events", kindUint64, "GoDebug", "WinSymlinkEvents"},
	{"GoDebugX509NegativeSerial", "/godebug/non-default-behavior/x509negativeserial:events", kindUint64, "GoDebug", "X509NegativeSerialEvents"},
	{"GoDebugX509RSACRT", "/godebug/non-default-behavior/x509rsacrt:events", kindUint64, "GoDebug", "X509RsaCrtEvents"},
	{"GoDebugX509SKID256", "/godebug/non-default-behavior/x509sha256skid:events", kindUint64, "GoDebug", "X509Sha256SkidEvents"},
	{"GoDebugX509SSLCertOverridePlatform", "/godebug/non-default-behavior/x509sslcertoverrideplatform:events", kindUint64, "GoDebug", "X509SslCertOverridePlatformEvents"},
	{"GoDebugX509UseFallbackRoots", "/godebug/non-default-behavior/x509usefallbackroots:events", kindUint64, "GoDebug", "X509UseFallbackRootsEvents"},
	{"GoDebugX509UsePolicies", "/godebug/non-default-behavior/x509usepolicies:events", kindUint64, "GoDebug", "X509UsePoliciesEvents"},
	{"GoDebugZipInsecurePath", "/godebug/non-default-behavior/zipinsecurepath:events", kindUint64, "GoDebug", "ZipInsecurePathEvents"},

	{"MemoryClassesHeapFree", "/memory/classes/heap/free:bytes", kindUint64, "Memory", "HeapFreeBytes"},
	{"MemoryClassesHeapObjects", "/memory/classes/heap/objects:bytes", kindUint64, "Memory", "HeapObjectBytes"},
	{"MemoryClassesHeapReleased", "/memory/classes/heap/released:bytes", kindUint64, "Memory", "HeapReleasedBytes"},
	{"MemoryClassesHeapStacks", "/memory/classes/heap/stacks:bytes", kindUint64, "Memory", "HeapStackBytes"},
	{"MemoryClassesHeapUnused", "/memory/classes/heap/unused:bytes", kindUint64, "Memory", "HeapUnusedBytes"},
	{"MemoryClassesMetadataMCacheFree", "/memory/classes/metadata/mcache/free:bytes", kindUint64, "Memory", "McacheFreeBytes"},
	{"MemoryClassesMetadataMCacheInUse", "/memory/classes/metadata/mcache/inuse:bytes", kindUint64, "Memory", "McacheInuseBytes"},
	{"MemoryClassesMetadataMSpanFree", "/memory/classes/metadata/mspan/free:bytes", kindUint64, "Memory", "MspanFreeBytes"},
	{"MemoryClassesMetadataMSpanInUse", "/memory/classes/metadata/mspan/inuse:bytes", kindUint64, "Memory", "MspanInuseBytes"},
	{"MemoryClassesMetadataOther", "/memory/classes/metadata/other:bytes", kindUint64, "Memory", "OtherMetadataBytes"},
	{"MemoryClassesOSStacks", "/memory/classes/os-stacks:bytes", kindUint64, "Memory", "OsStackBytes"},
	{"MemoryClassesOther", "/memory/classes/other:bytes", kindUint64, "Memory", "OtherBytes"},
	{"MemoryClassesProfilingBuckets", "/memory/classes/profiling/buckets:bytes", kindUint64, "Memory", "ProfilingBucketBytes"},
	{"MemoryClassesTotal", "/memory/classes/total:bytes", kindUint64, "Memory", "TotalBytes"},

	{"SchedLatencies", "/sched/latencies:seconds", kindHistogram, "Scheduler", "LatenciesSeconds"},
	{"SchedPausesStoppingGC", "/sched/pauses/stopping/gc:seconds", kindHistogram, "Scheduler", "StoppingGcSeconds"},
	{"SchedPausesStoppingOther", "/sched/pauses/stopping/other:seconds", kindHistogram, "Scheduler", "StoppingOtherSeconds"},
	{"SchedPausesTotalGC", "/sched/pauses/total/gc:seconds", kindHistogram, "Scheduler", "TotalGcPauseSeconds"},
	{"SchedPausesTotalOther", "/sched/pauses/total/other:seconds", kindHistogram, "Scheduler", "TotalOtherPauseSeconds"},
	{"SchedGoroutinesCreated", "/sched/goroutines-created:goroutines", kindUint64, "Scheduler", "GoroutinesCreated"},
	{"SchedGOMAXPROCS", "/sched/gomaxprocs:threads", kindUint32, "Scheduler", "GomaxprocsThreads"},
	{"SchedGoroutinesNotInGo", "/sched/goroutines/not-in-go:goroutines", kindUint32, "Scheduler", "NotInGoGoroutines"},
	{"SchedGoroutinesRunnable", "/sched/goroutines/runnable:goroutines", kindUint32, "Scheduler", "RunnableGoroutines"},
	{"SchedGoroutinesRunning", "/sched/goroutines/running:goroutines", kindUint32, "Scheduler", "RunningGoroutines"},
	{"SchedGoroutinesWaiting", "/sched/goroutines/waiting:goroutines", kindUint32, "Scheduler", "WaitingGoroutines"},
	{"SchedGoroutines", "/sched/goroutines:goroutines", kindUint32, "Scheduler", "LiveGoroutines"},
	{"SchedThreadsTotal", "/sched/threads/total:threads", kindUint32, "Scheduler", "TotalLiveThreads"},

	{"SyncMutexWaitTotal", "/sync/mutex/wait/total:seconds", kindFloat64, "Sync", "TotalMutexWaitDurationSeconds"},
}
