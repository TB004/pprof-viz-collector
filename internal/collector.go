// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"runtime/metrics"

	"github.com/TB004/pprof-viz-agent/internal/proto"
)

// SampleCollector manages a pre-allocated slice of metrics.Sample for efficient collection.
type SampleCollector struct {
	samples     []metrics.Sample
	lastSamples []metrics.Sample
}

// NewSampleCollector creates a new SampleCollector with pre-allocated samples.
func NewSampleCollector() *SampleCollector {
	samples := make([]metrics.Sample, len(AllMetricNames))
	lastSamples := make([]metrics.Sample, len(AllMetricNames))
	for i, name := range AllMetricNames {
		samples[i].Name = name
		lastSamples[i].Name = name
	}
	return &SampleCollector{
		samples:     samples,
		lastSamples: lastSamples,
	}
}

// Read updates the internal samples with current runtime metrics.
func (c *SampleCollector) Read() {
	// Swap buffers so we write to the "old" buffer, making it the "new" one.
	// The one that was "current" becomes "last".
	c.samples, c.lastSamples = c.lastSamples, c.samples
	ReadAllSamples(c.samples)
}

// Collect maps the current samples into a proto response, returning only fields that changed
// since the last Read. On a fresh collector, all non-zero fields are returned.
// Histogram bucket boundaries are never included; use GetMetadata to retrieve them once.
func (c *SampleCollector) Collect() *proto.AllResponse {
	// Helpers for diffing and value extraction
	hasChanged := func(idx int) bool {
		cur := c.samples[idx].Value
		last := c.lastSamples[idx].Value

		// If last is uninitialized (KindBad), we treat it as a delta from the default (0) value.
		if last.Kind() == metrics.KindBad {
			switch cur.Kind() {
			case metrics.KindUint64:
				return cur.Uint64() != 0
			case metrics.KindFloat64:
				return cur.Float64() != 0
			case metrics.KindFloat64Histogram:
				h := cur.Float64Histogram()
				if h == nil {
					return false
				}
				for _, count := range h.Counts {
					if count != 0 {
						return true
					}
				}
				return false
			case metrics.KindBad:
				return false // Both are bad
			default:
				return true
			}
		}

		if cur.Kind() != last.Kind() {
			return true
		}

		switch cur.Kind() {
		case metrics.KindUint64:
			return cur.Uint64() != last.Uint64()
		case metrics.KindFloat64:
			return cur.Float64() != last.Float64()
		case metrics.KindBad:
			return false
		case metrics.KindFloat64Histogram:
			hCur := cur.Float64Histogram()
			hLast := last.Float64Histogram()
			if hCur == nil || hLast == nil {
				return hCur != hLast
			}
			if len(hCur.Counts) != len(hLast.Counts) {
				return true
			}
			for i, v := range hCur.Counts {
				if v != hLast.Counts[i] {
					return true
				}
			}
			return false
		default:
			return true
		}
	}

	// Value getters that return pointer if changed, or nil if not
	vUint := func(name string) *uint64 {
		idx := GetIndex(name)
		if !hasChanged(idx) {
			return nil
		}
		return new(AsUint(c.samples[idx].Value))
	}

	vFloat := func(name string) *float64 {
		idx := GetIndex(name)
		if !hasChanged(idx) {
			return nil
		}
		return new(AsFloat(c.samples[idx].Value))
	}

	vUint32 := func(name string) *uint32 {
		idx := GetIndex(name)
		if !hasChanged(idx) {
			return nil
		}
		return new(SafeUint32(AsUint(c.samples[idx].Value)))
	}

	vHist := func(name string) *proto.Histogram {
		idx := GetIndex(name)
		if !hasChanged(idx) {
			return nil
		}

		h := AsHistogram(c.samples[idx].Value)
		if h == nil {
			return nil
		}

		var lastCounts []uint64
		if last := c.lastSamples[idx].Value; last.Kind() == metrics.KindFloat64Histogram {
			if hLast := last.Float64Histogram(); hLast != nil {
				lastCounts = hLast.Counts
			}
		}

		updated := make(map[uint32]uint64)
		for i, v := range h.Counts {
			var lv uint64
			if i < len(lastCounts) {
				lv = lastCounts[i]
			}
			if v != lv {
				updated[uint32(i)] = v
			}
		}

		if len(updated) == 0 {
			return nil
		}

		return &proto.Histogram{
			UpdatedCounts: updated,
		}
	}

	// Direct mapping into the proto response
	return &proto.AllResponse{
		Cgo: &proto.CgoResponse{
			CgoCalls: vUint(MetricCgoGoToCCalls),
		},
		Cpu: &proto.CPUResponse{
			GcAssistSeconds:           vFloat(MetricCPUClassesGCMarkAssist),
			GcDedicatedSeconds:        vFloat(MetricCPUClassesGCMarkDedicated),
			GcIdleSeconds:             vFloat(MetricCPUClassesGCMarkIdle),
			GcPauseSeconds:            vFloat(MetricCPUClassesGCPause),
			GcTotalSeconds:            vFloat(MetricCPUClassesGCTotal),
			IdleSeconds:               vFloat(MetricCPUClassesIdle),
			TotalSeconds:              vFloat(MetricCPUClassesTotal),
			UserSeconds:               vFloat(MetricCPUClassesUser),
			ScavengeAssistSeconds:     vFloat(MetricCPUClassesScavengeAssist),
			ScavengeBackgroundSeconds: vFloat(MetricCPUClassesScavengeBackground),
			ScavengeTotalSeconds:      vFloat(MetricCPUClassesScavengeTotal),
		},
		Gc: &proto.GCResponse{
			BytesAllocsBySize:        vHist(MetricGCHeapAllocsBySize),
			BytesFreesBySize:         vHist(MetricGCHeapFreesBySize),
			ExecutedCleanups:         vUint(MetricGCCleanupsExecuted),
			QueuedCleanups:           vUint(MetricGCCleanupsQueued),
			AutomaticCycles:          vUint(MetricGCCyclesAutomatic),
			ForcedCycles:             vUint(MetricGCCyclesForced),
			TotalCycles:              vUint(MetricGCCyclesTotal),
			ExecutedFinalizers:       vUint(MetricGCFinalizersExecuted),
			QueuedFinalizers:         vUint(MetricGCFinalizersQueued),
			MemoryLimitBytes:         vUint(MetricGCGOMemLimit),
			BytesAllocs:              vUint(MetricGCHeapAllocs),
			ObjectsAllocs:            vUint(MetricGCHeapAllocsObjects),
			BytesFrees:               vUint(MetricGCHeapFrees),
			ObjectsFrees:             vUint(MetricGCHeapFreesObjects),
			GoalHeapSizeBytes:        vUint(MetricGCHeapGoal),
			LiveOccupiedBytes:        vUint(MetricGCHeapLive),
			ObjectsOccupied:          vUint(MetricGCHeapObjects),
			TinyObjectsAllocs:        vUint(MetricGCHeapTinyAllocs),
			LimiterLastEnabledCycle:  vUint(MetricGCLimiterLastEnabled),
			GlobalScannedBytes:       vUint(MetricGCScanGlobals),
			HeapScannedBytes:         vUint(MetricGCScanHeap),
			StackScannedBytes:        vUint(MetricGCScanStack),
			TotalScannedBytes:        vUint(MetricGCScanTotal),
			StackStartingSize:        vUint(MetricGCStackStartingSize),
			TargetHeapSizePercentage: vUint32(MetricGCGOGC),
		},
		GoDebug: &proto.GoDebugResponse{
			AllowMultipleVcsEvents:            vUint(MetricGoDebugAllowMultipleVCs),
			AsyncTimerChanEvents:              vUint(MetricGoDebugAsyncTimerChan),
			ContainerMaxProcsEvents:           vUint(MetricGoDebugContainerMaxProcs),
			CryptoCustomRandEvents:            vUint(MetricGoDebugCryptoCustomRand),
			EmbedFollowSymlinkEvents:          vUint(MetricGoDebugEmbedFollowSymlinks),
			ExecErrDotEvents:                  vUint(MetricGoDebugExecErrDot),
			CacheHashEvents:                   vUint(MetricGoDebugGoCacheHash),
			CacheTestEvents:                   vUint(MetricGoDebugGoCacheTest),
			CacheVerifyEvents:                 vUint(MetricGoDebugGoCacheVerify),
			TestJsonBuildTextEvents:           vUint(MetricGoDebugGoTestJSONBuildText),
			TypeAliasEvents:                   vUint(MetricGoDebugGoTypesAlias),
			HtmlMetaContentUrlEscapedEvents:   vUint(MetricGoDebugHTMLMetaContentURLEscape),
			Http2ClientEvents:                 vUint(MetricGoDebugHTTP2Client),
			Http2ServerEvents:                 vUint(MetricGoDebugHTTP2Server),
			HttpCookieMaxNumEvents:            vUint(MetricGoDebugHTTPCookieMaxNum),
			HttpLaxContentLengthEvents:        vUint(MetricGoDebugHTTPLaxContentLength),
			HttpMuxGo_121Events:               vUint(MetricGoDebugHTTPMuxGo121),
			HttpServeContentKeepHeadersEvents: vUint(MetricGoDebugHTTPServeContentKeepHeaders),
			InstallGoRootEvents:               vUint(MetricGoDebugInstallGoroot),
			MultipartMaxHeadersEvents:         vUint(MetricGoDebugMultipartMaxHeaders),
			MultipartMaxPartsEvents:           vUint(MetricGoDebugMultipartMaxParts),
			MultipathTcpEvents:                vUint(MetricGoDebugMultipathTCP),
			NetEdns0Events:                    vUint(MetricGoDebugNetEdns0),
			PanicNilEvents:                    vUint(MetricGoDebugPanicNil),
			RandAutoSeedEvents:                vUint(MetricGoDebugRandAutoSeed),
			RandSeedNopEvents:                 vUint(MetricGoDebugRandSeedNop),
			Rsa1024MinEvents:                  vUint(MetricGoDebugRSA1024Min),
			TarInsecurePathEvents:             vUint(MetricGoDebugTarInsecurePath),
			Tls10ServerEvents:                 vUint(MetricGoDebugTLS10Server),
			Tls3DesEvents:                     vUint(MetricGoDebugTLS3DES),
			TlsMaxRsaSizeEvents:               vUint(MetricGoDebugTLSMaxRSASize),
			TlsRsaKexEvents:                   vUint(MetricGoDebugTLSRSAKex),
			TlsSha1Events:                     vUint(MetricGoDebugTLSSHA1),
			TlsUnsafeEkmEvents:                vUint(MetricGoDebugTLSUnsafeEKM),
			UpdateMaxProcsEvents:              vUint(MetricGoDebugUpdateMaxProcs),
			UrlMaxQueryParamsEvents:           vUint(MetricGoDebugURLMaxQueryParams),
			UrlStrictColonsEvents:             vUint(MetricGoDebugURLStrictColons),
			WinReadLinkVolumeEvents:           vUint(MetricGoDebugWinReadLinkVolume),
			WinSymlinkEvents:                  vUint(MetricGoDebugWinSymlink),
			X509KeypairLeafEvents:             vUint(MetricGoDebugX509KeyPairLeaf),
			X509NegativeSerialEvents:          vUint(MetricGoDebugX509NegativeSerial),
			X509RsaCrtEvents:                  vUint(MetricGoDebugX509RSACRT),
			X509Sha256SkidEvents:              vUint(MetricGoDebugX509SKID256),
			X509UseFallbackRootsEvents:        vUint(MetricGoDebugX509UseFallbackRoots),
			X509UsePoliciesEvents:             vUint(MetricGoDebugX509UsePolicies),
			ZipInsecurePathEvents:             vUint(MetricGoDebugZipInsecurePath),
		},
		Memory: &proto.MemoryResponse{
			HeapFreeBytes:        vUint(MetricMemoryClassesHeapFree),
			HeapObjectBytes:      vUint(MetricMemoryClassesHeapObjects),
			HeapReleasedBytes:    vUint(MetricMemoryClassesHeapReleased),
			HeapStackBytes:       vUint(MetricMemoryClassesHeapStacks),
			HeapUnusedBytes:      vUint(MetricMemoryClassesHeapUnused),
			McacheFreeBytes:      vUint(MetricMemoryClassesMetadataMCacheFree),
			McacheInuseBytes:     vUint(MetricMemoryClassesMetadataMCacheInUse),
			MspanFreeBytes:       vUint(MetricMemoryClassesMetadataMSpanFree),
			MspanInuseBytes:      vUint(MetricMemoryClassesMetadataMSpanInUse),
			OtherMetadataBytes:   vUint(MetricMemoryClassesMetadataOther),
			OsStackBytes:         vUint(MetricMemoryClassesOSStacks),
			OtherBytes:           vUint(MetricMemoryClassesOther),
			ProfilingBucketBytes: vUint(MetricMemoryClassesProfilingBuckets),
			TotalBytes:           vUint(MetricMemoryClassesTotal),
		},
		Scheduler: &proto.SchedulerResponse{
			LatenciesSeconds:       vHist(MetricSchedLatencies),
			StoppingGcSeconds:      vHist(MetricSchedPausesStoppingGC),
			StoppingOtherSeconds:   vHist(MetricSchedPausesStoppingOther),
			TotalGcPauseSeconds:    vHist(MetricSchedPausesTotalGC),
			TotalOtherPauseSeconds: vHist(MetricSchedPausesTotalOther),
			GoroutinesCreated:      vUint(MetricSchedGoroutinesCreated),
			GomaxprocsThreads:      vUint32(MetricSchedGOMAXPROCS),
			NotInGoGoroutines:      vUint32(MetricSchedGoroutinesNotInGo),
			RunnableGoroutines:     vUint32(MetricSchedGoroutinesRunnable),
			RunningGoroutines:      vUint32(MetricSchedGoroutinesRunning),
			WaitingGoroutines:      vUint32(MetricSchedGoroutinesWaiting),
			LiveGoroutines:         vUint32(MetricSchedGoroutines),
			TotalLiveThreads:       vUint32(MetricSchedThreadsTotal),
		},
		Sync: &proto.SyncResponse{
			TotalMutexWaitDurationSeconds: vFloat(MetricSyncMutexWaitTotal),
		},
	}
}
