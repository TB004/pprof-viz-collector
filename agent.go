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
	"github.com/TB004/pprof-viz-agent/internal/proto"
	"github.com/TB004/pprof-viz-agent/internal/proto/protoconnect"
	"github.com/TB004/pprof-viz-agent/pkg/models"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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

	// Wrap the mux with h2c to support HTTP/2 over cleartext.
	h2cHandler := h2c.NewHandler(mux, &http2.Server{})

	s := http.Server{
		Addr:        fmt.Sprintf("localhost:%d", port),
		Handler:     h2cHandler,
		Protocols:   proto,
		TLSConfig:   tls,
		ReadTimeout: 10 * time.Second,
	}
	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

// SnapshotCgo retrieves a snapshot of cgo-related metrics.
func SnapshotCgo() models.CgoResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect(false)

	return models.CgoResponse{
		CgoCalls: s.Cgo.CgoCalls,
	}
}

// SnapshotGC retrieves a snapshot of GC-related metrics.
func SnapshotGC() *models.GCResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect(true)

	// Helper to convert proto.Histogram to models.Histogram
	toModelHistogram := func(h *proto.Histogram) *models.Histogram {
		if h == nil {
			return nil
		}
		return &models.Histogram{
			Counts:  h.Counts,
			Buckets: h.Buckets,
		}
	}

	return &models.GCResponse{
		BytesAllocsBySize:        toModelHistogram(s.Gc.BytesAllocsBySize),
		BytesFreesBySize:         toModelHistogram(s.Gc.BytesFreesBySize),
		ExecutedCleanups:         s.Gc.ExecutedCleanups,
		QueuedCleanups:           s.Gc.QueuedCleanups,
		AutomaticCycles:          s.Gc.AutomaticCycles,
		ForcedCycles:             s.Gc.ForcedCycles,
		TotalCycles:              s.Gc.TotalCycles,
		ExecutedFinalizers:       s.Gc.ExecutedFinalizers,
		QueuedFinalizers:         s.Gc.QueuedFinalizers,
		MemoryLimitBytes:         s.Gc.MemoryLimitBytes,
		BytesAllocs:              s.Gc.BytesAllocs,
		ObjectsAllocs:            s.Gc.ObjectsAllocs,
		BytesFrees:               s.Gc.BytesFrees,
		ObjectsFrees:             s.Gc.ObjectsFrees,
		GoalHeapSizeBytes:        s.Gc.GoalHeapSizeBytes,
		LiveOccupiedBytes:        s.Gc.LiveOccupiedBytes,
		ObjectsOccupied:          s.Gc.ObjectsOccupied,
		TinyObjectsAllocs:        s.Gc.TinyObjectsAllocs,
		LimiterLastEnabledCycle:  s.Gc.LimiterLastEnabledCycle,
		GlobalScannedBytes:       s.Gc.GlobalScannedBytes,
		HeapScannedBytes:         s.Gc.HeapScannedBytes,
		StackScannedBytes:        s.Gc.StackScannedBytes,
		TotalScannedBytes:        s.Gc.TotalScannedBytes,
		StackStartingSize:        s.Gc.StackStartingSize,
		TargetHeapSizePercentage: s.Gc.TargetHeapSizePercentage,
	}
}

// SnapshotCPU retrieves a snapshot of CPU-related metrics.
func SnapshotCPU() *models.CPUResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect(false)

	return &models.CPUResponse{
		GcAssistSeconds:           s.Cpu.GcAssistSeconds,
		GcDedicatedSeconds:        s.Cpu.GcDedicatedSeconds,
		GcIdleSeconds:             s.Cpu.GcIdleSeconds,
		GcPauseSeconds:            s.Cpu.GcPauseSeconds,
		GcTotalSeconds:            s.Cpu.GcTotalSeconds,
		IdleSeconds:               s.Cpu.IdleSeconds,
		TotalSeconds:              s.Cpu.TotalSeconds,
		UserSeconds:               s.Cpu.UserSeconds,
		ScavengeAssistSeconds:     s.Cpu.ScavengeAssistSeconds,
		ScavengeBackgroundSeconds: s.Cpu.ScavengeBackgroundSeconds,
		ScavengeTotalSeconds:      s.Cpu.ScavengeTotalSeconds,
	}
}

// SnapshotGoDebug retrieves a snapshot of godebug-related metrics.
func SnapshotGoDebug() *models.GoDebugResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect(false)

	return &models.GoDebugResponse{
		CacheHashEvents:                   s.GoDebug.CacheHashEvents,
		CacheTestEvents:                   s.GoDebug.CacheTestEvents,
		CacheVerifyEvents:                 s.GoDebug.CacheVerifyEvents,
		Http2ClientEvents:                 s.GoDebug.Http2ClientEvents,
		Http2ServerEvents:                 s.GoDebug.Http2ServerEvents,
		Tls10ServerEvents:                 s.GoDebug.Tls10ServerEvents,
		Tls3DesEvents:                     s.GoDebug.Tls3DesEvents,
		TlsMaxRsaSizeEvents:               s.GoDebug.TlsMaxRsaSizeEvents,
		TlsRsaKexEvents:                   s.GoDebug.TlsRsaKexEvents,
		TlsSha1Events:                     s.GoDebug.TlsSha1Events,
		TlsUnsafeEkmEvents:                s.GoDebug.TlsUnsafeEkmEvents,
		HttpCookieMaxNumEvents:            s.GoDebug.HttpCookieMaxNumEvents,
		HttpLaxContentLengthEvents:        s.GoDebug.HttpLaxContentLengthEvents,
		HttpMuxGo121Events:                s.GoDebug.HttpMuxGo_121Events,
		HttpServeContentKeepHeadersEvents: s.GoDebug.HttpServeContentKeepHeadersEvents,
		AllowMultipleVcsEvents:            s.GoDebug.AllowMultipleVcsEvents,
		AsyncTimerChanEvents:              s.GoDebug.AsyncTimerChanEvents,
		ContainerMaxProcsEvents:           s.GoDebug.ContainerMaxProcsEvents,
		CryptoCustomRandEvents:            s.GoDebug.CryptoCustomRandEvents,
		EmbedFollowSymlinkEvents:          s.GoDebug.EmbedFollowSymlinkEvents,
		ExecErrDotEvents:                  s.GoDebug.ExecErrDotEvents,
		TestJsonBuildTextEvents:           s.GoDebug.TestJsonBuildTextEvents,
		TypeAliasEvents:                   s.GoDebug.TypeAliasEvents,
		HtmlMetaContentUrlEscapedEvents:   s.GoDebug.HtmlMetaContentUrlEscapedEvents,
		InstallGoRootEvents:               s.GoDebug.InstallGoRootEvents,
		MultipartMaxHeadersEvents:         s.GoDebug.MultipartMaxHeadersEvents,
		MultipartMaxPartsEvents:           s.GoDebug.MultipartMaxPartsEvents,
		MultipathTcpEvents:                s.GoDebug.MultipathTcpEvents,
		NetEdns0Events:                    s.GoDebug.NetEdns0Events,
		PanicNilEvents:                    s.GoDebug.PanicNilEvents,
		RandAutoSeedEvents:                s.GoDebug.RandAutoSeedEvents,
		RandSeedNopEvents:                 s.GoDebug.RandSeedNopEvents,
		Rsa1024MinEvents:                  s.GoDebug.Rsa1024MinEvents,
		TarInsecurePathEvents:             s.GoDebug.TarInsecurePathEvents,
		UpdateMaxProcsEvents:              s.GoDebug.UpdateMaxProcsEvents,
		UrlMaxQueryParamsEvents:           s.GoDebug.UrlMaxQueryParamsEvents,
		UrlStrictColonsEvents:             s.GoDebug.UrlStrictColonsEvents,
		WinReadLinkVolumeEvents:           s.GoDebug.WinReadLinkVolumeEvents,
		WinSymlinkEvents:                  s.GoDebug.WinSymlinkEvents,
		X509KeyparLeafEvents:              s.GoDebug.X509KeyparLeafEvents,
		X509NegativSerialEvents:           s.GoDebug.X509NegativSerialEvents,
		X509RsaCrtEvents:                  s.GoDebug.X509RsaCrtEvents,
		X509Sha256SkidEvents:              s.GoDebug.X509Sha256SkidEvents,
		X509UseFallbackRootsEvents:        s.GoDebug.X509UseFallbackRootsEvents,
		X509UsePoliciesEvents:             s.GoDebug.X509UsePoliciesEvents,
		ZipInsecurePathEvents:             s.GoDebug.ZipInsecurePathEvents,
	}
}

// SnapshotMemory retrieves a snapshot of memory-related metrics.
func SnapshotMemory() *models.MemoryResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect(false)

	return &models.MemoryResponse{
		HeapFreeBytes:        s.Memory.HeapFreeBytes,
		HeapObjectBytes:      s.Memory.HeapObjectBytes,
		HeapReleasedBytes:    s.Memory.HeapReleasedBytes,
		HeapStackBytes:       s.Memory.HeapStackBytes,
		HeapUnusedBytes:      s.Memory.HeapUnusedBytes,
		McacheFreeBytes:      s.Memory.McacheFreeBytes,
		McacheInuseBytes:     s.Memory.McacheInuseBytes,
		MspanFreeBytes:       s.Memory.MspanFreeBytes,
		MspanInuseBytes:      s.Memory.MspanInuseBytes,
		OtherMetadataBytes:   s.Memory.OtherMetadataBytes,
		OsStackBytes:         s.Memory.OsStackBytes,
		OtherBytes:           s.Memory.OtherBytes,
		ProfilingBucketBytes: s.Memory.ProfilingBucketBytes,
		TotalBytes:           s.Memory.TotalBytes,
	}
}

// SnapshotSchedular retrieves a snapshot of scheduler-related metrics.
func SnapshotSchedular() *models.SchedulerResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect(true)

	// Helper to convert proto.Histogram to models.Histogram
	toModelHistogram := func(h *proto.Histogram) *models.Histogram {
		if h == nil {
			return nil
		}
		return &models.Histogram{
			Counts:  h.Counts,
			Buckets: h.Buckets,
		}
	}

	return &models.SchedulerResponse{
		LatenciesSeconds:       toModelHistogram(s.Scheduler.LatenciesSeconds),
		StoppingGcSeconds:      toModelHistogram(s.Scheduler.StoppingGcSeconds),
		StoppingOtherSeconds:   toModelHistogram(s.Scheduler.StoppingOtherSeconds),
		TotalGcPauseSeconds:    toModelHistogram(s.Scheduler.TotalGcPauseSeconds),
		TotalOtherPauseSeconds: toModelHistogram(s.Scheduler.TotalOtherPauseSeconds),
		GoroutinesCreated:      s.Scheduler.GoroutinesCreated,
		GomaxprocsThreads:      s.Scheduler.GomaxprocsThreads,
		NotInGoGoroutines:      s.Scheduler.NotInGoGoroutines,
		RunnableGoroutines:     s.Scheduler.RunnableGoroutines,
		RunningGoroutines:      s.Scheduler.RunningGoroutines,
		WaitingGoroutines:      s.Scheduler.WaitingGoroutines,
		LiveGoroutines:         s.Scheduler.LiveGoroutines,
		TotalLiveThreads:       s.Scheduler.TotalLiveThreads,
	}
}

// SnapshotSync retrieves a snapshot of sync-related metrics.
func SnapshotSync() models.SyncResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect(false)

	return models.SyncResponse{
		TotalMutexWaitDurationSeconds: s.Sync.TotalMutexWaitDurationSeconds,
	}
}

// SnapshotAll retrieves a single point-in-time capture of all supported Go runtime metrics.
func SnapshotAll() models.AllResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect(true)

	// Helper to convert proto.Histogram to models.Histogram
	toModelHistogram := func(h *proto.Histogram) *models.Histogram {
		if h == nil {
			return nil
		}
		return &models.Histogram{
			Counts:  h.Counts,
			Buckets: h.Buckets,
		}
	}

	return models.AllResponse{
		Cgo: models.CgoResponse{
			CgoCalls: s.Cgo.CgoCalls,
		},
		Cpu: &models.CPUResponse{
			GcAssistSeconds:           s.Cpu.GcAssistSeconds,
			GcDedicatedSeconds:        s.Cpu.GcDedicatedSeconds,
			GcIdleSeconds:             s.Cpu.GcIdleSeconds,
			GcPauseSeconds:            s.Cpu.GcPauseSeconds,
			GcTotalSeconds:            s.Cpu.GcTotalSeconds,
			IdleSeconds:               s.Cpu.IdleSeconds,
			TotalSeconds:              s.Cpu.TotalSeconds,
			UserSeconds:               s.Cpu.UserSeconds,
			ScavengeAssistSeconds:     s.Cpu.ScavengeAssistSeconds,
			ScavengeBackgroundSeconds: s.Cpu.ScavengeBackgroundSeconds,
			ScavengeTotalSeconds:      s.Cpu.ScavengeTotalSeconds,
		},
		Gc: &models.GCResponse{
			BytesAllocsBySize:        toModelHistogram(s.Gc.BytesAllocsBySize),
			BytesFreesBySize:         toModelHistogram(s.Gc.BytesFreesBySize),
			ExecutedCleanups:         s.Gc.ExecutedCleanups,
			QueuedCleanups:           s.Gc.QueuedCleanups,
			AutomaticCycles:          s.Gc.AutomaticCycles,
			ForcedCycles:             s.Gc.ForcedCycles,
			TotalCycles:              s.Gc.TotalCycles,
			ExecutedFinalizers:       s.Gc.ExecutedFinalizers,
			QueuedFinalizers:         s.Gc.QueuedFinalizers,
			MemoryLimitBytes:         s.Gc.MemoryLimitBytes,
			BytesAllocs:              s.Gc.BytesAllocs,
			ObjectsAllocs:            s.Gc.ObjectsAllocs,
			BytesFrees:               s.Gc.BytesFrees,
			ObjectsFrees:             s.Gc.ObjectsFrees,
			GoalHeapSizeBytes:        s.Gc.GoalHeapSizeBytes,
			LiveOccupiedBytes:        s.Gc.LiveOccupiedBytes,
			ObjectsOccupied:          s.Gc.ObjectsOccupied,
			TinyObjectsAllocs:        s.Gc.TinyObjectsAllocs,
			LimiterLastEnabledCycle:  s.Gc.LimiterLastEnabledCycle,
			GlobalScannedBytes:       s.Gc.GlobalScannedBytes,
			HeapScannedBytes:         s.Gc.HeapScannedBytes,
			StackScannedBytes:        s.Gc.StackScannedBytes,
			TotalScannedBytes:        s.Gc.TotalScannedBytes,
			StackStartingSize:        s.Gc.StackStartingSize,
			TargetHeapSizePercentage: s.Gc.TargetHeapSizePercentage,
		},
		GoDebug: &models.GoDebugResponse{
			CacheHashEvents:                   s.GoDebug.CacheHashEvents,
			CacheTestEvents:                   s.GoDebug.CacheTestEvents,
			CacheVerifyEvents:                 s.GoDebug.CacheVerifyEvents,
			Http2ClientEvents:                 s.GoDebug.Http2ClientEvents,
			Http2ServerEvents:                 s.GoDebug.Http2ServerEvents,
			Tls10ServerEvents:                 s.GoDebug.Tls10ServerEvents,
			Tls3DesEvents:                     s.GoDebug.Tls3DesEvents,
			TlsMaxRsaSizeEvents:               s.GoDebug.TlsMaxRsaSizeEvents,
			TlsRsaKexEvents:                   s.GoDebug.TlsRsaKexEvents,
			TlsSha1Events:                     s.GoDebug.TlsSha1Events,
			TlsUnsafeEkmEvents:                s.GoDebug.TlsUnsafeEkmEvents,
			HttpCookieMaxNumEvents:            s.GoDebug.HttpCookieMaxNumEvents,
			HttpLaxContentLengthEvents:        s.GoDebug.HttpLaxContentLengthEvents,
			HttpMuxGo121Events:                s.GoDebug.HttpMuxGo_121Events,
			HttpServeContentKeepHeadersEvents: s.GoDebug.HttpServeContentKeepHeadersEvents,
			AllowMultipleVcsEvents:            s.GoDebug.AllowMultipleVcsEvents,
			AsyncTimerChanEvents:              s.GoDebug.AsyncTimerChanEvents,
			ContainerMaxProcsEvents:           s.GoDebug.ContainerMaxProcsEvents,
			CryptoCustomRandEvents:            s.GoDebug.CryptoCustomRandEvents,
			EmbedFollowSymlinkEvents:          s.GoDebug.EmbedFollowSymlinkEvents,
			ExecErrDotEvents:                  s.GoDebug.ExecErrDotEvents,
			TestJsonBuildTextEvents:           s.GoDebug.TestJsonBuildTextEvents,
			TypeAliasEvents:                   s.GoDebug.TypeAliasEvents,
			HtmlMetaContentUrlEscapedEvents:   s.GoDebug.HtmlMetaContentUrlEscapedEvents,
			InstallGoRootEvents:               s.GoDebug.InstallGoRootEvents,
			MultipartMaxHeadersEvents:         s.GoDebug.MultipartMaxHeadersEvents,
			MultipartMaxPartsEvents:           s.GoDebug.MultipartMaxPartsEvents,
			MultipathTcpEvents:                s.GoDebug.MultipathTcpEvents,
			NetEdns0Events:                    s.GoDebug.NetEdns0Events,
			PanicNilEvents:                    s.GoDebug.PanicNilEvents,
			RandAutoSeedEvents:                s.GoDebug.RandAutoSeedEvents,
			RandSeedNopEvents:                 s.GoDebug.RandSeedNopEvents,
			Rsa1024MinEvents:                  s.GoDebug.Rsa1024MinEvents,
			TarInsecurePathEvents:             s.GoDebug.TarInsecurePathEvents,
			UpdateMaxProcsEvents:              s.GoDebug.UpdateMaxProcsEvents,
			UrlMaxQueryParamsEvents:           s.GoDebug.UrlMaxQueryParamsEvents,
			UrlStrictColonsEvents:             s.GoDebug.UrlStrictColonsEvents,
			WinReadLinkVolumeEvents:           s.GoDebug.WinReadLinkVolumeEvents,
			WinSymlinkEvents:                  s.GoDebug.WinSymlinkEvents,
			X509KeyparLeafEvents:              s.GoDebug.X509KeyparLeafEvents,
			X509NegativSerialEvents:           s.GoDebug.X509NegativSerialEvents,
			X509RsaCrtEvents:                  s.GoDebug.X509RsaCrtEvents,
			X509Sha256SkidEvents:              s.GoDebug.X509Sha256SkidEvents,
			X509UseFallbackRootsEvents:        s.GoDebug.X509UseFallbackRootsEvents,
			X509UsePoliciesEvents:             s.GoDebug.X509UsePoliciesEvents,
			ZipInsecurePathEvents:             s.GoDebug.ZipInsecurePathEvents,
		},
		Memory: &models.MemoryResponse{
			HeapFreeBytes:        s.Memory.HeapFreeBytes,
			HeapObjectBytes:      s.Memory.HeapObjectBytes,
			HeapReleasedBytes:    s.Memory.HeapReleasedBytes,
			HeapStackBytes:       s.Memory.HeapStackBytes,
			HeapUnusedBytes:      s.Memory.HeapUnusedBytes,
			McacheFreeBytes:      s.Memory.McacheFreeBytes,
			McacheInuseBytes:     s.Memory.McacheInuseBytes,
			MspanFreeBytes:       s.Memory.MspanFreeBytes,
			MspanInuseBytes:      s.Memory.MspanInuseBytes,
			OtherMetadataBytes:   s.Memory.OtherMetadataBytes,
			OsStackBytes:         s.Memory.OsStackBytes,
			OtherBytes:           s.Memory.OtherBytes,
			ProfilingBucketBytes: s.Memory.ProfilingBucketBytes,
			TotalBytes:           s.Memory.TotalBytes,
		},
		Scheduler: &models.SchedulerResponse{
			LatenciesSeconds:       toModelHistogram(s.Scheduler.LatenciesSeconds),
			StoppingGcSeconds:      toModelHistogram(s.Scheduler.StoppingGcSeconds),
			StoppingOtherSeconds:   toModelHistogram(s.Scheduler.StoppingOtherSeconds),
			TotalGcPauseSeconds:    toModelHistogram(s.Scheduler.TotalGcPauseSeconds),
			TotalOtherPauseSeconds: toModelHistogram(s.Scheduler.TotalOtherPauseSeconds),
			GoroutinesCreated:      s.Scheduler.GoroutinesCreated,
			GomaxprocsThreads:      s.Scheduler.GomaxprocsThreads,
			NotInGoGoroutines:      s.Scheduler.NotInGoGoroutines,
			RunnableGoroutines:     s.Scheduler.RunnableGoroutines,
			RunningGoroutines:      s.Scheduler.RunningGoroutines,
			WaitingGoroutines:      s.Scheduler.WaitingGoroutines,
			LiveGoroutines:         s.Scheduler.LiveGoroutines,
			TotalLiveThreads:       s.Scheduler.TotalLiveThreads,
		},
		Sync: models.SyncResponse{
			TotalMutexWaitDurationSeconds: s.Sync.TotalMutexWaitDurationSeconds,
		},
	}
}
