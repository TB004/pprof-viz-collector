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
	group      string
	protoField string
	kind       kind
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
	{"CgoGoToCCalls", "/cgo/go-to-c-calls:calls", "Cgo", "CgoCalls", kindUint64},

	{"CPUClassesGCMarkAssist", "/cpu/classes/gc/mark/assist:cpu-seconds", "Cpu", "GcAssistSeconds", kindFloat64},
	{"CPUClassesGCMarkDedicated", "/cpu/classes/gc/mark/dedicated:cpu-seconds", "Cpu", "GcDedicatedSeconds", kindFloat64},
	{"CPUClassesGCMarkIdle", "/cpu/classes/gc/mark/idle:cpu-seconds", "Cpu", "GcIdleSeconds", kindFloat64},
	{"CPUClassesGCPause", "/cpu/classes/gc/pause:cpu-seconds", "Cpu", "GcPauseSeconds", kindFloat64},
	{"CPUClassesGCTotal", "/cpu/classes/gc/total:cpu-seconds", "Cpu", "GcTotalSeconds", kindFloat64},
	{"CPUClassesIdle", "/cpu/classes/idle:cpu-seconds", "Cpu", "IdleSeconds", kindFloat64},
	{"CPUClassesTotal", "/cpu/classes/total:cpu-seconds", "Cpu", "TotalSeconds", kindFloat64},
	{"CPUClassesUser", "/cpu/classes/user:cpu-seconds", "Cpu", "UserSeconds", kindFloat64},
	{"CPUClassesScavengeAssist", "/cpu/classes/scavenge/assist:cpu-seconds", "Cpu", "ScavengeAssistSeconds", kindFloat64},
	{"CPUClassesScavengeBackground", "/cpu/classes/scavenge/background:cpu-seconds", "Cpu", "ScavengeBackgroundSeconds", kindFloat64},
	{"CPUClassesScavengeTotal", "/cpu/classes/scavenge/total:cpu-seconds", "Cpu", "ScavengeTotalSeconds", kindFloat64},

	{"GCHeapAllocsBySize", "/gc/heap/allocs-by-size:bytes", "Gc", "BytesAllocsBySize", kindHistogram},
	{"GCHeapFreesBySize", "/gc/heap/frees-by-size:bytes", "Gc", "BytesFreesBySize", kindHistogram},
	{"GCCleanupsExecuted", "/gc/cleanups/executed:cleanups", "Gc", "ExecutedCleanups", kindUint64},
	{"GCCleanupsQueued", "/gc/cleanups/queued:cleanups", "Gc", "QueuedCleanups", kindUint64},
	{"GCCyclesAutomatic", "/gc/cycles/automatic:gc-cycles", "Gc", "AutomaticCycles", kindUint64},
	{"GCCyclesForced", "/gc/cycles/forced:gc-cycles", "Gc", "ForcedCycles", kindUint64},
	{"GCCyclesTotal", "/gc/cycles/total:gc-cycles", "Gc", "TotalCycles", kindUint64},
	{"GCFinalizersExecuted", "/gc/finalizers/executed:finalizers", "Gc", "ExecutedFinalizers", kindUint64},
	{"GCFinalizersQueued", "/gc/finalizers/queued:finalizers", "Gc", "QueuedFinalizers", kindUint64},
	{"GCGOMemLimit", "/gc/gomemlimit:bytes", "Gc", "MemoryLimitBytes", kindUint64},
	{"GCHeapAllocs", "/gc/heap/allocs:bytes", "Gc", "BytesAllocs", kindUint64},
	{"GCHeapAllocsObjects", "/gc/heap/allocs:objects", "Gc", "ObjectsAllocs", kindUint64},
	{"GCHeapFrees", "/gc/heap/frees:bytes", "Gc", "BytesFrees", kindUint64},
	{"GCHeapFreesObjects", "/gc/heap/frees:objects", "Gc", "ObjectsFrees", kindUint64},
	{"GCHeapGoal", "/gc/heap/goal:bytes", "Gc", "GoalHeapSizeBytes", kindUint64},
	{"GCHeapLive", "/gc/heap/live:bytes", "Gc", "LiveOccupiedBytes", kindUint64},
	{"GCHeapObjects", "/gc/heap/objects:objects", "Gc", "ObjectsOccupied", kindUint64},
	{"GCHeapTinyAllocs", "/gc/heap/tiny/allocs:objects", "Gc", "TinyObjectsAllocs", kindUint64},
	{"GCLimiterLastEnabled", "/gc/limiter/last-enabled:gc-cycle", "Gc", "LimiterLastEnabledCycle", kindUint64},
	{"GCScanGlobals", "/gc/scan/globals:bytes", "Gc", "GlobalScannedBytes", kindUint64},
	{"GCScanHeap", "/gc/scan/heap:bytes", "Gc", "HeapScannedBytes", kindUint64},
	{"GCScanStack", "/gc/scan/stack:bytes", "Gc", "StackScannedBytes", kindUint64},
	{"GCScanTotal", "/gc/scan/total:bytes", "Gc", "TotalScannedBytes", kindUint64},
	{"GCStackStartingSize", "/gc/stack/starting-size:bytes", "Gc", "StackStartingSize", kindUint64},
	{"GCGOGC", "/gc/gogc:percent", "Gc", "TargetHeapSizePercentage", kindUint32},

	{"GoDebugAllowMultipleVCs", "/godebug/non-default-behavior/allowmultiplevcs:events", "GoDebug", "AllowMultipleVcsEvents", kindUint64},
	{"GoDebugContainerMaxProcs", "/godebug/non-default-behavior/containermaxprocs:events", "GoDebug", "ContainerMaxProcsEvents", kindUint64},
	{"GoDebugCryptoCustomRand", "/godebug/non-default-behavior/cryptocustomrand:events", "GoDebug", "CryptoCustomRandEvents", kindUint64},
	{"GoDebugEmbedFollowSymlinks", "/godebug/non-default-behavior/embedfollowsymlinks:events", "GoDebug", "EmbedFollowSymlinkEvents", kindUint64},
	{"GoDebugExecErrDot", "/godebug/non-default-behavior/execerrdot:events", "GoDebug", "ExecErrDotEvents", kindUint64},
	{"GoDebugFIPS140EMS", "/godebug/non-default-behavior/fips140ems:events", "GoDebug", "Fips140EmsEvents", kindUint64},
	{"GoDebugGoCacheHash", "/godebug/non-default-behavior/gocachehash:events", "GoDebug", "CacheHashEvents", kindUint64},
	{"GoDebugGoCacheTest", "/godebug/non-default-behavior/gocachetest:events", "GoDebug", "CacheTestEvents", kindUint64},
	{"GoDebugGoCacheVerify", "/godebug/non-default-behavior/gocacheverify:events", "GoDebug", "CacheVerifyEvents", kindUint64},
	{"GoDebugGoTestJSONBuildText", "/godebug/non-default-behavior/gotestjsonbuildtext:events", "GoDebug", "TestJsonBuildTextEvents", kindUint64},
	{"GoDebugHTMLMetaContentURLEscape", "/godebug/non-default-behavior/htmlmetacontenturlescape:events", "GoDebug", "HtmlMetaContentUrlEscapedEvents", kindUint64},
	{"GoDebugHTTP2Client", "/godebug/non-default-behavior/http2client:events", "GoDebug", "Http2ClientEvents", kindUint64},
	{"GoDebugHTTP2Server", "/godebug/non-default-behavior/http2server:events", "GoDebug", "Http2ServerEvents", kindUint64},
	{"GoDebugHTTPCookieMaxNum", "/godebug/non-default-behavior/httpcookiemaxnum:events", "GoDebug", "HttpCookieMaxNumEvents", kindUint64},
	{"GoDebugHTTPLaxContentLength", "/godebug/non-default-behavior/httplaxcontentlength:events", "GoDebug", "HttpLaxContentLengthEvents", kindUint64},
	{"GoDebugHTTPMuxGo121", "/godebug/non-default-behavior/httpmuxgo121:events", "GoDebug", "HttpMuxGo_121Events", kindUint64},
	{"GoDebugHTTPServeContentKeepHeaders", "/godebug/non-default-behavior/httpservecontentkeepheaders:events", "GoDebug", "HttpServeContentKeepHeadersEvents", kindUint64},
	{"GoDebugHTTPServeContentMaxRanges", "/godebug/non-default-behavior/httpservecontentmaxranges:events", "GoDebug", "HttpServeContentMaxRangesEvents", kindUint64},
	{"GoDebugInstallGoroot", "/godebug/non-default-behavior/installgoroot:events", "GoDebug", "InstallGoRootEvents", kindUint64},
	{"GoDebugMultipartMaxHeaders", "/godebug/non-default-behavior/multipartmaxheaders:events", "GoDebug", "MultipartMaxHeadersEvents", kindUint64},
	{"GoDebugMultipartMaxParts", "/godebug/non-default-behavior/multipartmaxparts:events", "GoDebug", "MultipartMaxPartsEvents", kindUint64},
	{"GoDebugMultipathTCP", "/godebug/non-default-behavior/multipathtcp:events", "GoDebug", "MultipathTcpEvents", kindUint64},
	{"GoDebugNetEdns0", "/godebug/non-default-behavior/netedns0:events", "GoDebug", "NetEdns0Events", kindUint64},
	{"GoDebugPanicNil", "/godebug/non-default-behavior/panicnil:events", "GoDebug", "PanicNilEvents", kindUint64},
	{"GoDebugRandAutoSeed", "/godebug/non-default-behavior/randautoseed:events", "GoDebug", "RandAutoSeedEvents", kindUint64},
	{"GoDebugRandSeedNop", "/godebug/non-default-behavior/randseednop:events", "GoDebug", "RandSeedNopEvents", kindUint64},
	{"GoDebugRSA1024Min", "/godebug/non-default-behavior/rsa1024min:events", "GoDebug", "Rsa1024MinEvents", kindUint64},
	{"GoDebugTarInsecurePath", "/godebug/non-default-behavior/tarinsecurepath:events", "GoDebug", "TarInsecurePathEvents", kindUint64},
	{"GoDebugTLSMaxRSASize", "/godebug/non-default-behavior/tlsmaxrsasize:events", "GoDebug", "TlsMaxRsaSizeEvents", kindUint64},
	{"GoDebugTLSSHA1", "/godebug/non-default-behavior/tlssha1:events", "GoDebug", "TlsSha1Events", kindUint64},
	{"GoDebugUpdateMaxProcs", "/godebug/non-default-behavior/updatemaxprocs:events", "GoDebug", "UpdateMaxProcsEvents", kindUint64},
	{"GoDebugURLMaxQueryParams", "/godebug/non-default-behavior/urlmaxqueryparams:events", "GoDebug", "UrlMaxQueryParamsEvents", kindUint64},
	{"GoDebugURLStrictColons", "/godebug/non-default-behavior/urlstrictcolons:events", "GoDebug", "UrlStrictColonsEvents", kindUint64},
	{"GoDebugWinReadLinkVolume", "/godebug/non-default-behavior/winreadlinkvolume:events", "GoDebug", "WinReadLinkVolumeEvents", kindUint64},
	{"GoDebugWinSymlink", "/godebug/non-default-behavior/winsymlink:events", "GoDebug", "WinSymlinkEvents", kindUint64},
	{"GoDebugX509NegativeSerial", "/godebug/non-default-behavior/x509negativeserial:events", "GoDebug", "X509NegativeSerialEvents", kindUint64},
	{"GoDebugX509RSACRT", "/godebug/non-default-behavior/x509rsacrt:events", "GoDebug", "X509RsaCrtEvents", kindUint64},
	{"GoDebugX509SKID256", "/godebug/non-default-behavior/x509sha256skid:events", "GoDebug", "X509Sha256SkidEvents", kindUint64},
	{"GoDebugX509SSLCertOverridePlatform", "/godebug/non-default-behavior/x509sslcertoverrideplatform:events", "GoDebug", "X509SslCertOverridePlatformEvents", kindUint64},
	{"GoDebugX509UseFallbackRoots", "/godebug/non-default-behavior/x509usefallbackroots:events", "GoDebug", "X509UseFallbackRootsEvents", kindUint64},
	{"GoDebugX509UsePolicies", "/godebug/non-default-behavior/x509usepolicies:events", "GoDebug", "X509UsePoliciesEvents", kindUint64},
	{"GoDebugZipInsecurePath", "/godebug/non-default-behavior/zipinsecurepath:events", "GoDebug", "ZipInsecurePathEvents", kindUint64},

	{"MemoryClassesHeapFree", "/memory/classes/heap/free:bytes", "Memory", "HeapFreeBytes", kindUint64},
	{"MemoryClassesHeapObjects", "/memory/classes/heap/objects:bytes", "Memory", "HeapObjectBytes", kindUint64},
	{"MemoryClassesHeapReleased", "/memory/classes/heap/released:bytes", "Memory", "HeapReleasedBytes", kindUint64},
	{"MemoryClassesHeapStacks", "/memory/classes/heap/stacks:bytes", "Memory", "HeapStackBytes", kindUint64},
	{"MemoryClassesHeapUnused", "/memory/classes/heap/unused:bytes", "Memory", "HeapUnusedBytes", kindUint64},
	{"MemoryClassesMetadataMCacheFree", "/memory/classes/metadata/mcache/free:bytes", "Memory", "McacheFreeBytes", kindUint64},
	{"MemoryClassesMetadataMCacheInUse", "/memory/classes/metadata/mcache/inuse:bytes", "Memory", "McacheInuseBytes", kindUint64},
	{"MemoryClassesMetadataMSpanFree", "/memory/classes/metadata/mspan/free:bytes", "Memory", "MspanFreeBytes", kindUint64},
	{"MemoryClassesMetadataMSpanInUse", "/memory/classes/metadata/mspan/inuse:bytes", "Memory", "MspanInuseBytes", kindUint64},
	{"MemoryClassesMetadataOther", "/memory/classes/metadata/other:bytes", "Memory", "OtherMetadataBytes", kindUint64},
	{"MemoryClassesOSStacks", "/memory/classes/os-stacks:bytes", "Memory", "OsStackBytes", kindUint64},
	{"MemoryClassesOther", "/memory/classes/other:bytes", "Memory", "OtherBytes", kindUint64},
	{"MemoryClassesProfilingBuckets", "/memory/classes/profiling/buckets:bytes", "Memory", "ProfilingBucketBytes", kindUint64},
	{"MemoryClassesTotal", "/memory/classes/total:bytes", "Memory", "TotalBytes", kindUint64},

	{"SchedLatencies", "/sched/latencies:seconds", "Scheduler", "LatenciesSeconds", kindHistogram},
	{"SchedPausesStoppingGC", "/sched/pauses/stopping/gc:seconds", "Scheduler", "StoppingGcSeconds", kindHistogram},
	{"SchedPausesStoppingOther", "/sched/pauses/stopping/other:seconds", "Scheduler", "StoppingOtherSeconds", kindHistogram},
	{"SchedPausesTotalGC", "/sched/pauses/total/gc:seconds", "Scheduler", "TotalGcPauseSeconds", kindHistogram},
	{"SchedPausesTotalOther", "/sched/pauses/total/other:seconds", "Scheduler", "TotalOtherPauseSeconds", kindHistogram},
	{"SchedGoroutinesCreated", "/sched/goroutines-created:goroutines", "Scheduler", "GoroutinesCreated", kindUint64},
	{"SchedGOMAXPROCS", "/sched/gomaxprocs:threads", "Scheduler", "GomaxprocsThreads", kindUint32},
	{"SchedGoroutinesNotInGo", "/sched/goroutines/not-in-go:goroutines", "Scheduler", "NotInGoGoroutines", kindUint32},
	{"SchedGoroutinesRunnable", "/sched/goroutines/runnable:goroutines", "Scheduler", "RunnableGoroutines", kindUint32},
	{"SchedGoroutinesRunning", "/sched/goroutines/running:goroutines", "Scheduler", "RunningGoroutines", kindUint32},
	{"SchedGoroutinesWaiting", "/sched/goroutines/waiting:goroutines", "Scheduler", "WaitingGoroutines", kindUint32},
	{"SchedGoroutines", "/sched/goroutines:goroutines", "Scheduler", "LiveGoroutines", kindUint32},
	{"SchedThreadsTotal", "/sched/threads/total:threads", "Scheduler", "TotalLiveThreads", kindUint32},

	{"SyncMutexWaitTotal", "/sync/mutex/wait/total:seconds", "Sync", "TotalMutexWaitDurationSeconds", kindFloat64},
}
