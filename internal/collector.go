// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"runtime/metrics"

	"github.com/TB004/pprof-viz-agent/internal/proto"
	"github.com/TB004/pprof-viz-agent/pkg/models"
)

// SampleCollector manages a pre-allocated slice of metrics.Sample for efficient collection.
type SampleCollector struct {
	samples []metrics.Sample
}

// NewSampleCollector creates a new SampleCollector with pre-allocated samples.
func NewSampleCollector() *SampleCollector {
	samples := make([]metrics.Sample, len(AllMetricNames))
	for i, name := range AllMetricNames {
		samples[i].Name = name
	}
	return &SampleCollector{samples: samples}
}

// Read updates the internal samples with current runtime metrics.
func (c *SampleCollector) Read() {
	ReadAllSamples(c.samples)
}

// Collect maps the current samples into a proto response.
func (c *SampleCollector) Collect(includeBuckets bool) *proto.AllResponse {
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

	v := func(name string) metrics.Value {
		return c.samples[GetIndex(name)].Value
	}

	// Direct mapping into the proto response
	return &proto.AllResponse{
		Cgo: &proto.CgoResponse{
			CgoCalls: AsUint(v(MetricCgoGoToCCalls)),
		},
		Cpu: &proto.CPUResponse{
			GcAssistSeconds:           AsFloat(v(MetricCPUClassesGCMarkAssist)),
			GcDedicatedSeconds:        AsFloat(v(MetricCPUClassesGCMarkDedicated)),
			GcIdleSeconds:             AsFloat(v(MetricCPUClassesGCMarkIdle)),
			GcPauseSeconds:            AsFloat(v(MetricCPUClassesGCPause)),
			GcTotalSeconds:            AsFloat(v(MetricCPUClassesGCTotal)),
			IdleSeconds:               AsFloat(v(MetricCPUClassesIdle)),
			TotalSeconds:              AsFloat(v(MetricCPUClassesTotal)),
			UserSeconds:               AsFloat(v(MetricCPUClassesUser)),
			ScavengeAssistSeconds:     AsFloat(v(MetricCPUClassesScavengeAssist)),
			ScavengeBackgroundSeconds: AsFloat(v(MetricCPUClassesScavengeBackground)),
			ScavengeTotalSeconds:      AsFloat(v(MetricCPUClassesScavengeTotal)),
		},
		Gc: &proto.GCResponse{
			BytesAllocsBySize:        toProtoHistogram(AsHistogram(v(MetricGCHeapAllocsBySize)), includeBuckets),
			BytesFreesBySize:         toProtoHistogram(AsHistogram(v(MetricGCHeapFreesBySize)), includeBuckets),
			ExecutedCleanups:         AsUint(v(MetricGCCleanupsExecuted)),
			QueuedCleanups:           AsUint(v(MetricGCCleanupsQueued)),
			AutomaticCycles:          AsUint(v(MetricGCCyclesAutomatic)),
			ForcedCycles:             AsUint(v(MetricGCCyclesForced)),
			TotalCycles:              AsUint(v(MetricGCCyclesTotal)),
			ExecutedFinalizers:       AsUint(v(MetricGCFinalizersExecuted)),
			QueuedFinalizers:         AsUint(v(MetricGCFinalizersQueued)),
			MemoryLimitBytes:         AsUint(v(MetricGCGOMemLimit)),
			BytesAllocs:              AsUint(v(MetricGCHeapAllocs)),
			ObjectsAllocs:            AsUint(v(MetricGCHeapAllocsObjects)),
			BytesFrees:               AsUint(v(MetricGCHeapFrees)),
			ObjectsFrees:             AsUint(v(MetricGCHeapFreesObjects)),
			GoalHeapSizeBytes:        AsUint(v(MetricGCHeapGoal)),
			LiveOccupiedBytes:        AsUint(v(MetricGCHeapLive)),
			ObjectsOccupied:          AsUint(v(MetricGCHeapObjects)),
			TinyObjectsAllocs:        AsUint(v(MetricGCHeapTinyAllocs)),
			LimiterLastEnabledCycle:  AsUint(v(MetricGCLimiterLastEnabled)),
			GlobalScannedBytes:       AsUint(v(MetricGCScanGlobals)),
			HeapScannedBytes:         AsUint(v(MetricGCScanHeap)),
			StackScannedBytes:        AsUint(v(MetricGCScanStack)),
			TotalScannedBytes:        AsUint(v(MetricGCScanTotal)),
			StackStartingSize:        AsUint(v(MetricGCStackStartingSize)),
			TargetHeapSizePercentage: SafeUint32(AsUint(v(MetricGCGOGC))),
		},
		GoDebug: &proto.GoDebugResponse{
			AllowMultipleVcsEvents:            AsUint(v(MetricGoDebugAllowMultipleVCs)),
			AsyncTimerChanEvents:              AsUint(v(MetricGoDebugAsyncTimerChan)),
			ContainerMaxProcsEvents:           AsUint(v(MetricGoDebugContainerMaxProcs)),
			CryptoCustomRandEvents:            AsUint(v(MetricGoDebugCryptoCustomRand)),
			EmbedFollowSymlinkEvents:          AsUint(v(MetricGoDebugEmbedFollowSymlinks)),
			ExecErrDotEvents:                  AsUint(v(MetricGoDebugExecErrDot)),
			CacheHashEvents:                   AsUint(v(MetricGoDebugGoCacheHash)),
			CacheTestEvents:                   AsUint(v(MetricGoDebugGoCacheTest)),
			CacheVerifyEvents:                 AsUint(v(MetricGoDebugGoCacheVerify)),
			TestJsonBuildTextEvents:           AsUint(v(MetricGoDebugGoTypesAlias)),
			TypeAliasEvents:                   AsUint(v(MetricGoDebugGoTypesAlias)),
			HtmlMetaContentUrlEscapedEvents:   AsUint(v(MetricGoDebugHTMLMetaContentURLEscape)),
			Http2ClientEvents:                 AsUint(v(MetricGoDebugHTTP2Client)),
			Http2ServerEvents:                 AsUint(v(MetricGoDebugHTTP2Server)),
			HttpCookieMaxNumEvents:            AsUint(v(MetricGoDebugHTTPCookieMaxNum)),
			HttpLaxContentLengthEvents:        AsUint(v(MetricGoDebugHTTPLaxContentLength)),
			HttpMuxGo_121Events:               AsUint(v(MetricGoDebugHTTPMuxGo121)),
			HttpServeContentKeepHeadersEvents: AsUint(v(MetricGoDebugHTTPServeContentKeepHeaders)),
			InstallGoRootEvents:               AsUint(v(MetricGoDebugInstallGoroot)),
			MultipartMaxHeadersEvents:         AsUint(v(MetricGoDebugMultipartMaxHeaders)),
			MultipartMaxPartsEvents:           AsUint(v(MetricGoDebugMultipartMaxParts)),
			MultipathTcpEvents:                AsUint(v(MetricGoDebugMultipathTCP)),
			NetEdns0Events:                    AsUint(v(MetricGoDebugNetEdns0)),
			PanicNilEvents:                    AsUint(v(MetricGoDebugPanicNil)),
			RandAutoSeedEvents:                AsUint(v(MetricGoDebugRandAutoSeed)),
			RandSeedNopEvents:                 AsUint(v(MetricGoDebugRandSeedNop)),
			Rsa1024MinEvents:                  AsUint(v(MetricGoDebugRSA1024Min)),
			TarInsecurePathEvents:             AsUint(v(MetricGoDebugTarInsecurePath)),
			Tls10ServerEvents:                 AsUint(v(MetricGoDebugTLS10Server)),
			Tls3DesEvents:                     AsUint(v(MetricGoDebugTLS3DES)),
			TlsMaxRsaSizeEvents:               AsUint(v(MetricGoDebugTLSMaxRSASize)),
			TlsRsaKexEvents:                   AsUint(v(MetricGoDebugTLSRSAKex)),
			TlsSha1Events:                     AsUint(v(MetricGoDebugTLSSHA1)),
			TlsUnsafeEkmEvents:                AsUint(v(MetricGoDebugTLSUnsafeEKM)),
			UpdateMaxProcsEvents:              AsUint(v(MetricGoDebugUpdateMaxProcs)),
			UrlMaxQueryParamsEvents:           AsUint(v(MetricGoDebugURLMaxQueryParams)),
			UrlStrictColonsEvents:             AsUint(v(MetricGoDebugURLStrictColons)),
			WinReadLinkVolumeEvents:           AsUint(v(MetricGoDebugWinReadLinkVolume)),
			WinSymlinkEvents:                  AsUint(v(MetricGoDebugWinSymlink)),
			X509KeyparLeafEvents:              AsUint(v(MetricGoDebugX509KeyPairLeaf)),
			X509NegativSerialEvents:           AsUint(v(MetricGoDebugX509NegativeSerial)),
			X509RsaCrtEvents:                  AsUint(v(MetricGoDebugX509RSACRT)),
			X509Sha256SkidEvents:              AsUint(v(MetricGoDebugX509SKID256)),
			X509UseFallbackRootsEvents:        AsUint(v(MetricGoDebugX509UseFallbackRoots)),
			X509UsePoliciesEvents:             AsUint(v(MetricGoDebugX509UsePolicies)),
			ZipInsecurePathEvents:             AsUint(v(MetricGoDebugZipInsecurePath)),
		},
		Memory: &proto.MemoryResponse{
			HeapFreeBytes:        AsUint(v(MetricMemoryClassesHeapFree)),
			HeapObjectBytes:      AsUint(v(MetricMemoryClassesHeapObjects)),
			HeapReleasedBytes:    AsUint(v(MetricMemoryClassesHeapReleased)),
			HeapStackBytes:       AsUint(v(MetricMemoryClassesHeapStacks)),
			HeapUnusedBytes:      AsUint(v(MetricMemoryClassesHeapUnused)),
			McacheFreeBytes:      AsUint(v(MetricMemoryClassesMetadataMCacheFree)),
			McacheInuseBytes:     AsUint(v(MetricMemoryClassesMetadataMCacheInUse)),
			MspanFreeBytes:       AsUint(v(MetricMemoryClassesMetadataMSpanFree)),
			MspanInuseBytes:      AsUint(v(MetricMemoryClassesMetadataMSpanInUse)),
			OtherMetadataBytes:   AsUint(v(MetricMemoryClassesMetadataOther)),
			OsStackBytes:         AsUint(v(MetricMemoryClassesOSStacks)),
			OtherBytes:           AsUint(v(MetricMemoryClassesOther)),
			ProfilingBucketBytes: AsUint(v(MetricMemoryClassesProfilingBuckets)),
			TotalBytes:           AsUint(v(MetricMemoryClassesTotal)),
		},
		Scheduler: &proto.SchedulerResponse{
			LatenciesSeconds:       toProtoHistogram(AsHistogram(v(MetricSchedLatencies)), includeBuckets),
			StoppingGcSeconds:      toProtoHistogram(AsHistogram(v(MetricSchedPausesStoppingGC)), includeBuckets),
			StoppingOtherSeconds:   toProtoHistogram(AsHistogram(v(MetricSchedPausesStoppingOther)), includeBuckets),
			TotalGcPauseSeconds:    toProtoHistogram(AsHistogram(v(MetricSchedPausesTotalGC)), includeBuckets),
			TotalOtherPauseSeconds: toProtoHistogram(AsHistogram(v(MetricSchedPausesTotalOther)), includeBuckets),
			GoroutinesCreated:      AsUint(v(MetricSchedGoroutinesCreated)),
			GomaxprocsThreads:      SafeUint32(AsUint(v(MetricSchedGOMAXPROCS))),
			NotInGoGoroutines:      SafeUint32(AsUint(v(MetricSchedGoroutinesNotInGo))),
			RunnableGoroutines:     SafeUint32(AsUint(v(MetricSchedGoroutinesRunnable))),
			RunningGoroutines:      SafeUint32(AsUint(v(MetricSchedGoroutinesRunning))),
			WaitingGoroutines:      SafeUint32(AsUint(v(MetricSchedGoroutinesWaiting))),
			LiveGoroutines:         SafeUint32(AsUint(v(MetricSchedGoroutines))),
			TotalLiveThreads:       SafeUint32(AsUint(v(MetricSchedThreadsTotal))),
		},
		Sync: &proto.SyncResponse{
			TotalMutexWaitDurationSeconds: AsFloat(v(MetricSyncMutexWaitTotal)),
		},
	}
}
