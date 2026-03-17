// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package pprofvizagent

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/TB004/pprof-viz-agent/internal"
	"github.com/TB004/pprof-viz-agent/internal/proto/protoconnect"
	"github.com/TB004/pprof-viz-agent/pkg/models"
)

// Bind registers the AgentService handler to the provided http.ServeMux.
//
// Useful, if you already have a HTTP server.
func Bind(mux *http.ServeMux) {
	path, handler := protoconnect.NewAgentServiceHandler(
		internal.AgentServiceHandler{},
		connect.WithCompression(internal.CompressionZstd, internal.NewZstdDecompressor, internal.NewZstdCompressor),
		connect.WithCompressMinBytes(256),
	)
	mux.Handle(path, handler)
}

// New creates and starts a new HTTP server on the specified port, serving the AgentService.
func New(port int, proto *http.Protocols, tls *tls.Config) error {
	mux := http.NewServeMux()

	path, handler := protoconnect.NewAgentServiceHandler(
		internal.AgentServiceHandler{},
		connect.WithCompression(internal.CompressionZstd, internal.NewZstdDecompressor, internal.NewZstdCompressor),
		connect.WithCompressMinBytes(256),
	)
	mux.Handle(path, handler)

	s := http.Server{
		Addr:        fmt.Sprintf("localhost:%d", port),
		Handler:     mux,
		Protocols:   proto,
		TLSConfig:   tls,
		ReadTimeout: 3 * time.Millisecond,
	}
	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

// SnapshotCgo retrieves a snapshot of cgo-related metrics.
func SnapshotCgo() models.CgoResponse {
	vals := internal.RetrieveCgoSamples()

	return models.CgoResponse{
		CgoCalls: internal.AsUint(vals[internal.MetricCgoGoToCCalls]),
	}
}

// SnapshotGC retrieves a snapshot of GC-related metrics.
func SnapshotGC() *models.GCResponse {
	gcVals := internal.RetrieveGCSamples()
	return &models.GCResponse{
		BytesAllocsBySize:        internal.AsHistogram(gcVals[internal.MetricGCHeapAllocsBySize]),
		BytesFreesBySize:         internal.AsHistogram(gcVals[internal.MetricGCHeapFreesBySize]),
		ExecutedCleanups:         internal.AsUint(gcVals[internal.MetricGCCleanupsExecuted]),
		QueuedCleanups:           internal.AsUint(gcVals[internal.MetricGCCleanupsQueued]),
		AutomaticCycles:          internal.AsUint(gcVals[internal.MetricGCCyclesAutomatic]),
		ForcedCycles:             internal.AsUint(gcVals[internal.MetricGCCyclesForced]),
		TotalCycles:              internal.AsUint(gcVals[internal.MetricGCCyclesTotal]),
		ExecutedFinalizers:       internal.AsUint(gcVals[internal.MetricGCFinalizersExecuted]),
		QueuedFinalizers:         internal.AsUint(gcVals[internal.MetricGCFinalizersQueued]),
		MemoryLimitBytes:         internal.AsUint(gcVals[internal.MetricGCGOMemLimit]),
		BytesAllocs:              internal.AsUint(gcVals[internal.MetricGCHeapAllocs]),
		ObjectsAllocs:            internal.AsUint(gcVals[internal.MetricGCHeapAllocsObjects]),
		BytesFrees:               internal.AsUint(gcVals[internal.MetricGCHeapFrees]),
		ObjectsFrees:             internal.AsUint(gcVals[internal.MetricGCHeapFreesObjects]),
		GoalHeapSizeBytes:        internal.AsUint(gcVals[internal.MetricGCHeapGoal]),
		LiveOccupiedBytes:        internal.AsUint(gcVals[internal.MetricGCHeapLive]),
		ObjectsOccupied:          internal.AsUint(gcVals[internal.MetricGCHeapObjects]),
		TinyObjectsAllocs:        internal.AsUint(gcVals[internal.MetricGCHeapTinyAllocs]),
		LimiterLastEnabledCycle:  internal.AsUint(gcVals[internal.MetricGCLimiterLastEnabled]),
		GlobalScannedBytes:       internal.AsUint(gcVals[internal.MetricGCScanGlobals]),
		HeapScannedBytes:         internal.AsUint(gcVals[internal.MetricGCScanHeap]),
		StackScannedBytes:        internal.AsUint(gcVals[internal.MetricGCScanStack]),
		TotalScannedBytes:        internal.AsUint(gcVals[internal.MetricGCScanTotal]),
		StackStartingSize:        internal.AsUint(gcVals[internal.MetricGCScanStack]),
		TargetHeapSizePercentage: internal.SafeUint32(internal.AsUint(gcVals[internal.MetricGCGOGC])),
	}
}

// SnapshotCPU retrieves a snapshot of CPU-related metrics.
func SnapshotCPU() *models.CPUResponse {
	cpuVals := internal.RetrieveCPUSamples()
	return &models.CPUResponse{
		GcAssistSeconds:           internal.AsFloat(cpuVals[internal.MetricCPUClassesGCMarkAssist]),
		GcDedicatedSeconds:        internal.AsFloat(cpuVals[internal.MetricCPUClassesGCMarkDedicated]),
		GcIdleSeconds:             internal.AsFloat(cpuVals[internal.MetricCPUClassesGCMarkIdle]),
		GcPauseSeconds:            internal.AsFloat(cpuVals[internal.MetricCPUClassesGCPause]),
		GcTotalSeconds:            internal.AsFloat(cpuVals[internal.MetricCPUClassesGCTotal]),
		IdleSeconds:               internal.AsFloat(cpuVals[internal.MetricCPUClassesIdle]),
		TotalSeconds:              internal.AsFloat(cpuVals[internal.MetricCPUClassesTotal]),
		UserSeconds:               internal.AsFloat(cpuVals[internal.MetricCPUClassesUser]),
		ScavengeAssistSeconds:     internal.AsFloat(cpuVals[internal.MetricCPUClassesScavengeAssist]),
		ScavengeBackgroundSeconds: internal.AsFloat(cpuVals[internal.MetricCPUClassesScavengeBackground]),
		ScavengeTotalSeconds:      internal.AsFloat(cpuVals[internal.MetricCPUClassesScavengeTotal]),
	}
}

// SnapshotGoDebug retrieves a snapshot of godebug-related metrics.
func SnapshotGoDebug() *models.GoDebugResponse {
	gdVals := internal.RetrieveGoDebugSamples()
	return &models.GoDebugResponse{
		CacheHashEvents:                   internal.AsUint(gdVals[internal.MetricGoDebugGoCacheHash]),
		CacheTestEvents:                   internal.AsUint(gdVals[internal.MetricGoDebugGoCacheTest]),
		CacheVerifyEvents:                 internal.AsUint(gdVals[internal.MetricGoDebugGoCacheVerify]),
		Http2ClientEvents:                 internal.AsUint(gdVals[internal.MetricGoDebugHTTP2Client]),
		Http2ServerEvents:                 internal.AsUint(gdVals[internal.MetricGoDebugHTTP2Server]),
		Tls10ServerEvents:                 internal.AsUint(gdVals[internal.MetricGoDebugTLS10Server]),
		Tls3DesEvents:                     internal.AsUint(gdVals[internal.MetricGoDebugTLS3DES]),
		TlsMaxRsaSizeEvents:               internal.AsUint(gdVals[internal.MetricGoDebugTLSMaxRSASize]),
		TlsRsaKexEvents:                   internal.AsUint(gdVals[internal.MetricGoDebugTLSRSAKex]),
		TlsSha1Events:                     internal.AsUint(gdVals[internal.MetricGoDebugTLSSHA1]),
		TlsUnsafeEkmEvents:                internal.AsUint(gdVals[internal.MetricGoDebugTLSUnsafeEKM]),
		HttpCookieMaxNumEvents:            internal.AsUint(gdVals[internal.MetricGoDebugHTTPCookieMaxNum]),
		HttpLaxContentLengthEvents:        internal.AsUint(gdVals[internal.MetricGoDebugHTTPLaxContentLength]),
		HttpMuxGo121Events:                internal.AsUint(gdVals[internal.MetricGoDebugHTTPMuxGo121]),
		HttpServeContentKeepHeadersEvents: internal.AsUint(gdVals[internal.MetricGoDebugHTTPServeContentKeepHeaders]),
		AllowMultipleVcsEvents:            internal.AsUint(gdVals[internal.MetricGoDebugAllowMultipleVCs]),
		AsyncTimerChanEvents:              internal.AsUint(gdVals[internal.MetricGoDebugAsyncTimerChan]),
		ContainerMaxProcsEvents:           internal.AsUint(gdVals[internal.MetricGoDebugContainerMaxProcs]),
		CryptoCustomRandEvents:            internal.AsUint(gdVals[internal.MetricGoDebugCryptoCustomRand]),
		EmbedFollowSymlinkEvents:          internal.AsUint(gdVals[internal.MetricGoDebugEmbedFollowSymlinks]),
		ExecErrDotEvents:                  internal.AsUint(gdVals[internal.MetricGoDebugExecErrDot]),
		TestJsonBuildTextEvents:           internal.AsUint(gdVals[internal.MetricGoDebugGoTypesAlias]),
		TypeAliasEvents:                   internal.AsUint(gdVals[internal.MetricGoDebugGoTypesAlias]),
		HtmlMetaContentUrlEscapedEvents:   internal.AsUint(gdVals[internal.MetricGoDebugHTMLMetaContentURLEscape]),
		InstallGoRootEvents:               internal.AsUint(gdVals[internal.MetricGoDebugInstallGoroot]),
		MultipartMaxHeadersEvents:         internal.AsUint(gdVals[internal.MetricGoDebugMultipartMaxHeaders]),
		MultipartMaxPartsEvents:           internal.AsUint(gdVals[internal.MetricGoDebugMultipartMaxParts]),
		MultipathTcpEvents:                internal.AsUint(gdVals[internal.MetricGoDebugMultipathTCP]),
		NetEdns0Events:                    internal.AsUint(gdVals[internal.MetricGoDebugNetEdns0]),
		PanicNilEvents:                    internal.AsUint(gdVals[internal.MetricGoDebugPanicNil]),
		RandAutoSeedEvents:                internal.AsUint(gdVals[internal.MetricGoDebugRandAutoSeed]),
		RandSeedNopEvents:                 internal.AsUint(gdVals[internal.MetricGoDebugRandSeedNop]),
		Rsa1024MinEvents:                  internal.AsUint(gdVals[internal.MetricGoDebugRSA1024Min]),
		TarInsecurePathEvents:             internal.AsUint(gdVals[internal.MetricGoDebugTarInsecurePath]),
		UpdateMaxProcsEvents:              internal.AsUint(gdVals[internal.MetricGoDebugUpdateMaxProcs]),
		UrlMaxQueryParamsEvents:           internal.AsUint(gdVals[internal.MetricGoDebugURLMaxQueryParams]),
		UrlStrictColonsEvents:             internal.AsUint(gdVals[internal.MetricGoDebugURLStrictColons]),
		WinReadLinkVolumeEvents:           internal.AsUint(gdVals[internal.MetricGoDebugWinReadLinkVolume]),
		WinSymlinkEvents:                  internal.AsUint(gdVals[internal.MetricGoDebugWinSymlink]),
		X509KeyparLeafEvents:              internal.AsUint(gdVals[internal.MetricGoDebugX509KeyPairLeaf]),
		X509NegativSerialEvents:           internal.AsUint(gdVals[internal.MetricGoDebugX509NegativeSerial]),
		X509RsaCrtEvents:                  internal.AsUint(gdVals[internal.MetricGoDebugX509RSACRT]),
		X509Sha256SkidEvents:              internal.AsUint(gdVals[internal.MetricGoDebugX509SKID256]),
		X509UseFallbackRootsEvents:        internal.AsUint(gdVals[internal.MetricGoDebugX509UseFallbackRoots]),
		X509UsePoliciesEvents:             internal.AsUint(gdVals[internal.MetricGoDebugX509UsePolicies]),
		ZipInsecurePathEvents:             internal.AsUint(gdVals[internal.MetricGoDebugZipInsecurePath]),
	}
}

// SnapshotMemory retrieves a snapshot of memory-related metrics.
func SnapshotMemory() *models.MemoryResponse {
	memVals := internal.RetrieveMemorySamples()
	return &models.MemoryResponse{
		HeapFreeBytes:        internal.AsUint(memVals[internal.MetricMemoryClassesHeapFree]),
		HeapObjectBytes:      internal.AsUint(memVals[internal.MetricMemoryClassesHeapObjects]),
		HeapReleasedBytes:    internal.AsUint(memVals[internal.MetricMemoryClassesHeapReleased]),
		HeapStackBytes:       internal.AsUint(memVals[internal.MetricMemoryClassesHeapStacks]),
		HeapUnusedBytes:      internal.AsUint(memVals[internal.MetricMemoryClassesHeapUnused]),
		McacheFreeBytes:      internal.AsUint(memVals[internal.MetricMemoryClassesMetadataMCacheFree]),
		McacheInuseBytes:     internal.AsUint(memVals[internal.MetricMemoryClassesMetadataMCacheInUse]),
		MspanFreeBytes:       internal.AsUint(memVals[internal.MetricMemoryClassesMetadataMSpanFree]),
		MspanInuseBytes:      internal.AsUint(memVals[internal.MetricMemoryClassesMetadataMSpanInUse]),
		OtherMetadataBytes:   internal.AsUint(memVals[internal.MetricMemoryClassesMetadataOther]),
		OsStackBytes:         internal.AsUint(memVals[internal.MetricMemoryClassesOSStacks]),
		OtherBytes:           internal.AsUint(memVals[internal.MetricMemoryClassesOther]),
		ProfilingBucketBytes: internal.AsUint(memVals[internal.MetricMemoryClassesProfilingBuckets]),
		TotalBytes:           internal.AsUint(memVals[internal.MetricMemoryClassesTotal]),
	}
}

// SnapshotSchedular retrieves a snapshot of scheduler-related metrics.
func SnapshotSchedular() *models.SchedulerResponse {
	schedVals := internal.RetrieveSchedulerSamples()
	return &models.SchedulerResponse{
		LatenciesSeconds:       internal.AsHistogram(schedVals[internal.MetricSchedLatencies]),
		StoppingGcSeconds:      internal.AsHistogram(schedVals[internal.MetricSchedPausesStoppingGC]),
		StoppingOtherSeconds:   internal.AsHistogram(schedVals[internal.MetricSchedPausesStoppingOther]),
		TotalGcPauseSeconds:    internal.AsHistogram(schedVals[internal.MetricSchedPausesTotalGC]),
		TotalOtherPauseSeconds: internal.AsHistogram(schedVals[internal.MetricSchedPausesTotalOther]),
		GoroutinesCreated:      internal.AsUint(schedVals[internal.MetricSchedGoroutinesCreated]),
		GomaxprocsThreads:      internal.SafeUint32(internal.AsUint(schedVals[internal.MetricSchedGOMAXPROCS])),
		NotInGoGoroutines:      internal.SafeUint32(internal.AsUint(schedVals[internal.MetricSchedGoroutinesNotInGo])),
		RunnableGoroutines:     internal.SafeUint32(internal.AsUint(schedVals[internal.MetricSchedGoroutinesRunnable])),
		RunningGoroutines:      internal.SafeUint32(internal.AsUint(schedVals[internal.MetricSchedGoroutinesRunning])),
		WaitingGoroutines:      internal.SafeUint32(internal.AsUint(schedVals[internal.MetricSchedGoroutinesWaiting])),
		LiveGoroutines:         internal.SafeUint32(internal.AsUint(schedVals[internal.MetricSchedGoroutines])),
		TotalLiveThreads:       internal.SafeUint32(internal.AsUint(schedVals[internal.MetricSchedThreadsTotal])),
	}
}

// SnapshotSync retrieves a snapshot of sync-related metrics.
func SnapshotSync() models.SyncResponse {
	syncVals := internal.RetrieveSyncSamples()
	return models.SyncResponse{
		TotalMutexWaitDurationSeconds: internal.AsFloat(syncVals[internal.MetricSyncMutexWaitTotal]),
	}
}

// SnapshotAll retrieves a single point-in-time capture of all supported Go runtime metrics.
func SnapshotAll() models.AllResponse {
	cgoResp := SnapshotCgo()
	cpuResp := SnapshotCPU()
	gcResp := SnapshotGC()
	gdResp := SnapshotGoDebug()
	memResp := SnapshotMemory()
	schedResp := SnapshotSchedular()
	syncResp := SnapshotSync()

	return models.AllResponse{
		Cgo:       cgoResp,
		Cpu:       cpuResp,
		Gc:        gcResp,
		GoDebug:   gdResp,
		Memory:    memResp,
		Scheduler: schedResp,
		Sync:      syncResp,
	}
}
