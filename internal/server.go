// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"context"
	"errors"
	"runtime/metrics"
	"time"

	"connectrpc.com/connect"
	"github.com/TB004/pprof-viz-agent/internal/proto"
	"github.com/TB004/pprof-viz-agent/internal/proto/protoconnect"
	"github.com/TB004/pprof-viz-agent/pkg/models"
	"google.golang.org/protobuf/types/known/emptypb"
)

// AgentServiceHandler implements the AgentService ConnectRPC service.
type AgentServiceHandler struct {
	protoconnect.UnimplementedAgentServiceHandler
}

// SnapshotAll returns a single point-in-time capture of all supported Go runtime metrics.
func (h AgentServiceHandler) SnapshotAll(context.Context, *emptypb.Empty) (*proto.AllResponse, error) {
	return collectSamples(true), nil
}

// GetMetadata returns the static bucket boundaries for all histogram metrics.
// Clients should call this once and then use StreamAll with exclude_histogram_buckets=true.
func (h AgentServiceHandler) GetMetadata(context.Context, *emptypb.Empty) (*proto.MetadataResponse, error) {
	// Helper function to get buckets out of a histogram sample
	getBuckets := func(s metrics.Sample) []float32 {
		h := AsHistogram(s)
		if h == nil {
			return nil
		}
		return h.Buckets
	}

	gcVals := RetrieveGCSamples()
	schedVals := RetrieveSchedulerSamples()

	resp := &proto.MetadataResponse{
		Gc: &proto.GCMetadata{
			BytesAllocsBySizeBuckets: getBuckets(gcVals[MetricGCHeapAllocsBySize]),
			BytesFreesBySizeBuckets:  getBuckets(gcVals[MetricGCHeapFreesBySize]),
		},
		Scheduler: &proto.SchedulerMetadata{
			LatenciesSecondsBuckets:       getBuckets(schedVals[MetricSchedLatencies]),
			StoppingGcSecondsBuckets:      getBuckets(schedVals[MetricSchedPausesStoppingGC]),
			StoppingOtherSecondsBuckets:   getBuckets(schedVals[MetricSchedPausesStoppingOther]),
			TotalGcPauseSecondsBuckets:    getBuckets(schedVals[MetricSchedPausesTotalGC]),
			TotalOtherPauseSecondsBuckets: getBuckets(schedVals[MetricSchedPausesTotalOther]),
		},
	}

	return resp, nil
}

// StreamAll streams runtime metrics at a client-provided interval in a client-provided number of cycles or indefinitely.
func (h AgentServiceHandler) StreamAll(ctx context.Context, req *proto.StreamConfigRequest, stream *connect.ServerStream[proto.AllResponse]) error {
	interval := req.GetIntervalSeconds()
	if interval == 0 {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("interval must be at least 1 second"))
	}

	numCycles := int(req.GetNumberOfCycles())
	includeBuckets := !req.GetExcludeHistogramBuckets()

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for i := 0; numCycles == 0 || i < numCycles; i++ {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			all := collectSamples(includeBuckets)
			if err := stream.Send(all); err != nil {
				return err
			}
		}
	}

	return nil
}

// collectSamples retrieves and maps all current runtime metrics into a proto response.
func collectSamples(includeBuckets bool) *proto.AllResponse {
	// Helper function to convert *models.Histogram to *proto.Histogram
	toProtoHistogram := func(h *models.Histogram, includeBuckets bool) *proto.Histogram {
		if h == nil {
			return nil
		}
		buckets := h.Buckets
		if !includeBuckets {
			buckets = nil
		}
		return &proto.Histogram{
			Counts:  h.Counts,
			Buckets: buckets,
		}
	}

	// cgo
	cgoVals := RetrieveCgoSamples()
	cgo := &proto.CgoResponse{
		CgoCalls: AsUint(cgoVals[MetricCgoGoToCCalls]),
	}

	// cpu
	cpuVals := RetrieveCPUSamples()
	cpu := &proto.CPUResponse{
		GcAssistSeconds:           AsFloat(cpuVals[MetricCPUClassesGCMarkAssist]),
		GcDedicatedSeconds:        AsFloat(cpuVals[MetricCPUClassesGCMarkDedicated]),
		GcIdleSeconds:             AsFloat(cpuVals[MetricCPUClassesGCMarkIdle]),
		GcPauseSeconds:            AsFloat(cpuVals[MetricCPUClassesGCPause]),
		GcTotalSeconds:            AsFloat(cpuVals[MetricCPUClassesGCTotal]),
		IdleSeconds:               AsFloat(cpuVals[MetricCPUClassesIdle]),
		TotalSeconds:              AsFloat(cpuVals[MetricCPUClassesTotal]),
		UserSeconds:               AsFloat(cpuVals[MetricCPUClassesUser]),
		ScavengeAssistSeconds:     AsFloat(cpuVals[MetricCPUClassesScavengeAssist]),
		ScavengeBackgroundSeconds: AsFloat(cpuVals[MetricCPUClassesScavengeBackground]),
		ScavengeTotalSeconds:      AsFloat(cpuVals[MetricCPUClassesScavengeTotal]),
	}

	// gc
	gcVals := RetrieveGCSamples()
	gc := &proto.GCResponse{
		BytesAllocsBySize:        toProtoHistogram(AsHistogram(gcVals[MetricGCHeapAllocsBySize]), includeBuckets),
		BytesFreesBySize:         toProtoHistogram(AsHistogram(gcVals[MetricGCHeapFreesBySize]), includeBuckets),
		ExecutedCleanups:         AsUint(gcVals[MetricGCCleanupsExecuted]),
		QueuedCleanups:           AsUint(gcVals[MetricGCCleanupsQueued]),
		AutomaticCycles:          AsUint(gcVals[MetricGCCyclesAutomatic]),
		ForcedCycles:             AsUint(gcVals[MetricGCCyclesForced]),
		TotalCycles:              AsUint(gcVals[MetricGCCyclesTotal]),
		ExecutedFinalizers:       AsUint(gcVals[MetricGCFinalizersExecuted]),
		QueuedFinalizers:         AsUint(gcVals[MetricGCFinalizersQueued]),
		MemoryLimitBytes:         AsUint(gcVals[MetricGCGOMemLimit]),
		BytesAllocs:              AsUint(gcVals[MetricGCHeapAllocs]),
		ObjectsAllocs:            AsUint(gcVals[MetricGCHeapAllocsObjects]),
		BytesFrees:               AsUint(gcVals[MetricGCHeapFrees]),
		ObjectsFrees:             AsUint(gcVals[MetricGCHeapFreesObjects]),
		GoalHeapSizeBytes:        AsUint(gcVals[MetricGCHeapGoal]),
		LiveOccupiedBytes:        AsUint(gcVals[MetricGCHeapLive]),
		ObjectsOccupied:          AsUint(gcVals[MetricGCHeapObjects]),
		TinyObjectsAllocs:        AsUint(gcVals[MetricGCHeapTinyAllocs]),
		LimiterLastEnabledCycle:  AsUint(gcVals[MetricGCLimiterLastEnabled]),
		GlobalScannedBytes:       AsUint(gcVals[MetricGCScanGlobals]),
		HeapScannedBytes:         AsUint(gcVals[MetricGCScanHeap]),
		StackScannedBytes:        AsUint(gcVals[MetricGCScanStack]),
		TotalScannedBytes:        AsUint(gcVals[MetricGCScanTotal]),
		StackStartingSize:        AsUint(gcVals[MetricGCStackStartingSize]),
		TargetHeapSizePercentage: SafeUint32(AsUint(gcVals[MetricGCGOGC])),
	}

	// godebug
	gdVals := RetrieveGoDebugSamples()
	gd := &proto.GoDebugResponse{
		CacheHashEvents:                   AsUint(gdVals[MetricGoDebugGoCacheHash]),
		CacheTestEvents:                   AsUint(gdVals[MetricGoDebugGoCacheTest]),
		CacheVerifyEvents:                 AsUint(gdVals[MetricGoDebugGoCacheVerify]),
		Http2ClientEvents:                 AsUint(gdVals[MetricGoDebugHTTP2Client]),
		Http2ServerEvents:                 AsUint(gdVals[MetricGoDebugHTTP2Server]),
		Tls10ServerEvents:                 AsUint(gdVals[MetricGoDebugTLS10Server]),
		Tls3DesEvents:                     AsUint(gdVals[MetricGoDebugTLS3DES]),
		TlsMaxRsaSizeEvents:               AsUint(gdVals[MetricGoDebugTLSMaxRSASize]),
		TlsRsaKexEvents:                   AsUint(gdVals[MetricGoDebugTLSRSAKex]),
		TlsSha1Events:                     AsUint(gdVals[MetricGoDebugTLSSHA1]),
		TlsUnsafeEkmEvents:                AsUint(gdVals[MetricGoDebugTLSUnsafeEKM]),
		HttpCookieMaxNumEvents:            AsUint(gdVals[MetricGoDebugHTTPCookieMaxNum]),
		HttpLaxContentLengthEvents:        AsUint(gdVals[MetricGoDebugHTTPLaxContentLength]),
		HttpMuxGo_121Events:               AsUint(gdVals[MetricGoDebugHTTPMuxGo121]),
		HttpServeContentKeepHeadersEvents: AsUint(gdVals[MetricGoDebugHTTPServeContentKeepHeaders]),
		AllowMultipleVcsEvents:            AsUint(gdVals[MetricGoDebugAllowMultipleVCs]),
		AsyncTimerChanEvents:              AsUint(gdVals[MetricGoDebugAsyncTimerChan]),
		ContainerMaxProcsEvents:           AsUint(gdVals[MetricGoDebugContainerMaxProcs]),
		CryptoCustomRandEvents:            AsUint(gdVals[MetricGoDebugCryptoCustomRand]),
		EmbedFollowSymlinkEvents:          AsUint(gdVals[MetricGoDebugEmbedFollowSymlinks]),
		ExecErrDotEvents:                  AsUint(gdVals[MetricGoDebugExecErrDot]),
		TestJsonBuildTextEvents:           AsUint(gdVals[MetricGoDebugGoTypesAlias]),
		TypeAliasEvents:                   AsUint(gdVals[MetricGoDebugGoTypesAlias]),
		HtmlMetaContentUrlEscapedEvents:   AsUint(gdVals[MetricGoDebugHTMLMetaContentURLEscape]),
		InstallGoRootEvents:               AsUint(gdVals[MetricGoDebugInstallGoroot]),
		MultipartMaxHeadersEvents:         AsUint(gdVals[MetricGoDebugMultipartMaxHeaders]),
		MultipartMaxPartsEvents:           AsUint(gdVals[MetricGoDebugMultipartMaxParts]),
		MultipathTcpEvents:                AsUint(gdVals[MetricGoDebugMultipathTCP]),
		NetEdns0Events:                    AsUint(gdVals[MetricGoDebugNetEdns0]),
		PanicNilEvents:                    AsUint(gdVals[MetricGoDebugPanicNil]),
		RandAutoSeedEvents:                AsUint(gdVals[MetricGoDebugRandAutoSeed]),
		RandSeedNopEvents:                 AsUint(gdVals[MetricGoDebugRandSeedNop]),
		Rsa1024MinEvents:                  AsUint(gdVals[MetricGoDebugRSA1024Min]),
		TarInsecurePathEvents:             AsUint(gdVals[MetricGoDebugTarInsecurePath]),
		UpdateMaxProcsEvents:              AsUint(gdVals[MetricGoDebugUpdateMaxProcs]),
		UrlMaxQueryParamsEvents:           AsUint(gdVals[MetricGoDebugURLMaxQueryParams]),
		UrlStrictColonsEvents:             AsUint(gdVals[MetricGoDebugURLStrictColons]),
		WinReadLinkVolumeEvents:           AsUint(gdVals[MetricGoDebugWinReadLinkVolume]),
		WinSymlinkEvents:                  AsUint(gdVals[MetricGoDebugWinSymlink]),
		X509KeyparLeafEvents:              AsUint(gdVals[MetricGoDebugX509KeyPairLeaf]),
		X509NegativSerialEvents:           AsUint(gdVals[MetricGoDebugX509NegativeSerial]),
		X509RsaCrtEvents:                  AsUint(gdVals[MetricGoDebugX509RSACRT]),
		X509Sha256SkidEvents:              AsUint(gdVals[MetricGoDebugX509SKID256]),
		X509UseFallbackRootsEvents:        AsUint(gdVals[MetricGoDebugX509UseFallbackRoots]),
		X509UsePoliciesEvents:             AsUint(gdVals[MetricGoDebugX509UsePolicies]),
		ZipInsecurePathEvents:             AsUint(gdVals[MetricGoDebugZipInsecurePath]),
	}

	// memory
	memVals := RetrieveMemorySamples()
	memory := &proto.MemoryResponse{
		HeapFreeBytes:        AsUint(memVals[MetricMemoryClassesHeapFree]),
		HeapObjectBytes:      AsUint(memVals[MetricMemoryClassesHeapObjects]),
		HeapReleasedBytes:    AsUint(memVals[MetricMemoryClassesHeapReleased]),
		HeapStackBytes:       AsUint(memVals[MetricMemoryClassesHeapStacks]),
		HeapUnusedBytes:      AsUint(memVals[MetricMemoryClassesHeapUnused]),
		McacheFreeBytes:      AsUint(memVals[MetricMemoryClassesMetadataMCacheFree]),
		McacheInuseBytes:     AsUint(memVals[MetricMemoryClassesMetadataMCacheInUse]),
		MspanFreeBytes:       AsUint(memVals[MetricMemoryClassesMetadataMSpanFree]),
		MspanInuseBytes:      AsUint(memVals[MetricMemoryClassesMetadataMSpanInUse]),
		OtherMetadataBytes:   AsUint(memVals[MetricMemoryClassesMetadataOther]),
		OsStackBytes:         AsUint(memVals[MetricMemoryClassesOSStacks]),
		OtherBytes:           AsUint(memVals[MetricMemoryClassesOther]),
		ProfilingBucketBytes: AsUint(memVals[MetricMemoryClassesProfilingBuckets]),
		TotalBytes:           AsUint(memVals[MetricMemoryClassesTotal]),
	}

	// scheduler
	schedVals := RetrieveSchedulerSamples()
	scheduler := &proto.SchedulerResponse{
		LatenciesSeconds:       toProtoHistogram(AsHistogram(schedVals[MetricSchedLatencies]), includeBuckets),
		StoppingGcSeconds:      toProtoHistogram(AsHistogram(schedVals[MetricSchedPausesStoppingGC]), includeBuckets),
		StoppingOtherSeconds:   toProtoHistogram(AsHistogram(schedVals[MetricSchedPausesStoppingOther]), includeBuckets),
		TotalGcPauseSeconds:    toProtoHistogram(AsHistogram(schedVals[MetricSchedPausesTotalGC]), includeBuckets),
		TotalOtherPauseSeconds: toProtoHistogram(AsHistogram(schedVals[MetricSchedPausesTotalOther]), includeBuckets),
		GoroutinesCreated:      AsUint(schedVals[MetricSchedGoroutinesCreated]),
		GomaxprocsThreads:      SafeUint32(AsUint(schedVals[MetricSchedGOMAXPROCS])),
		NotInGoGoroutines:      SafeUint32(AsUint(schedVals[MetricSchedGoroutinesNotInGo])),
		RunnableGoroutines:     SafeUint32(AsUint(schedVals[MetricSchedGoroutinesRunnable])),
		RunningGoroutines:      SafeUint32(AsUint(schedVals[MetricSchedGoroutinesRunning])),
		WaitingGoroutines:      SafeUint32(AsUint(schedVals[MetricSchedGoroutinesWaiting])),
		LiveGoroutines:         SafeUint32(AsUint(schedVals[MetricSchedGoroutines])),
		TotalLiveThreads:       SafeUint32(AsUint(schedVals[MetricSchedThreadsTotal])),
	}

	// sync
	syncVals := RetrieveSyncSamples()
	sync := &proto.SyncResponse{
		TotalMutexWaitDurationSeconds: AsFloat(syncVals[MetricSyncMutexWaitTotal]),
	}

	return &proto.AllResponse{
		Cgo:       cgo,
		Cpu:       cpu,
		Gc:        gc,
		GoDebug:   gd,
		Memory:    memory,
		Scheduler: scheduler,
		Sync:      sync,
	}
}
