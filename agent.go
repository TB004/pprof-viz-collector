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

// Helper functions for safe dereferencing
func vUint(v *uint64) uint64 {
	if v == nil {
		return 0
	}
	return *v
}

func vFloat(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func vUint32(v *uint32) uint32 {
	if v == nil {
		return 0
	}
	return *v
}

func vHist(h *proto.Histogram) *models.Histogram {
	if h == nil {
		return nil
	}
	return &models.Histogram{
		Counts:  h.Counts,
		Buckets: h.Buckets,
	}
}

// toModelHistogram helper to convert proto.Histogram to models.Histogram
func toModelHistogram(h *proto.Histogram) *models.Histogram {
	if h == nil {
		return nil
	}
	return &models.Histogram{
		Counts:  h.Counts,
		Buckets: h.Buckets,
	}
}

// SnapshotCgo retrieves a snapshot of cgo-related metrics.
func SnapshotCgo() models.CgoResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect(false, true)

	return models.CgoResponse{
		CgoCalls: vUint(s.Cgo.CgoCalls),
	}
}

// SnapshotGC retrieves a snapshot of GC-related metrics.
func SnapshotGC() *models.GCResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect(true, true)

	return &models.GCResponse{
		BytesAllocsBySize:        vHist(s.Gc.BytesAllocsBySize),
		BytesFreesBySize:         vHist(s.Gc.BytesFreesBySize),
		ExecutedCleanups:         vUint(s.Gc.ExecutedCleanups),
		QueuedCleanups:           vUint(s.Gc.QueuedCleanups),
		AutomaticCycles:          vUint(s.Gc.AutomaticCycles),
		ForcedCycles:             vUint(s.Gc.ForcedCycles),
		TotalCycles:              vUint(s.Gc.TotalCycles),
		ExecutedFinalizers:       vUint(s.Gc.ExecutedFinalizers),
		QueuedFinalizers:         vUint(s.Gc.QueuedFinalizers),
		MemoryLimitBytes:         vUint(s.Gc.MemoryLimitBytes),
		BytesAllocs:              vUint(s.Gc.BytesAllocs),
		ObjectsAllocs:            vUint(s.Gc.ObjectsAllocs),
		BytesFrees:               vUint(s.Gc.BytesFrees),
		ObjectsFrees:             vUint(s.Gc.ObjectsFrees),
		GoalHeapSizeBytes:        vUint(s.Gc.GoalHeapSizeBytes),
		LiveOccupiedBytes:        vUint(s.Gc.LiveOccupiedBytes),
		ObjectsOccupied:          vUint(s.Gc.ObjectsOccupied),
		TinyObjectsAllocs:        vUint(s.Gc.TinyObjectsAllocs),
		LimiterLastEnabledCycle:  vUint(s.Gc.LimiterLastEnabledCycle),
		GlobalScannedBytes:       vUint(s.Gc.GlobalScannedBytes),
		HeapScannedBytes:         vUint(s.Gc.HeapScannedBytes),
		StackScannedBytes:        vUint(s.Gc.StackScannedBytes),
		TotalScannedBytes:        vUint(s.Gc.TotalScannedBytes),
		StackStartingSize:        vUint(s.Gc.StackStartingSize),
		TargetHeapSizePercentage: vUint32(s.Gc.TargetHeapSizePercentage),
	}
}

// SnapshotCPU retrieves a snapshot of CPU-related metrics.
func SnapshotCPU() *models.CPUResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect(false, true)

	return &models.CPUResponse{
		GcAssistSeconds:           vFloat(s.Cpu.GcAssistSeconds),
		GcDedicatedSeconds:        vFloat(s.Cpu.GcDedicatedSeconds),
		GcIdleSeconds:             vFloat(s.Cpu.GcIdleSeconds),
		GcPauseSeconds:            vFloat(s.Cpu.GcPauseSeconds),
		GcTotalSeconds:            vFloat(s.Cpu.GcTotalSeconds),
		IdleSeconds:               vFloat(s.Cpu.IdleSeconds),
		TotalSeconds:              vFloat(s.Cpu.TotalSeconds),
		UserSeconds:               vFloat(s.Cpu.UserSeconds),
		ScavengeAssistSeconds:     vFloat(s.Cpu.ScavengeAssistSeconds),
		ScavengeBackgroundSeconds: vFloat(s.Cpu.ScavengeBackgroundSeconds),
		ScavengeTotalSeconds:      vFloat(s.Cpu.ScavengeTotalSeconds),
	}
}

// SnapshotGoDebug retrieves a snapshot of godebug-related metrics.
func SnapshotGoDebug() *models.GoDebugResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect(false, true)

	return &models.GoDebugResponse{
		CacheHashEvents:                   vUint(s.GoDebug.CacheHashEvents),
		CacheTestEvents:                   vUint(s.GoDebug.CacheTestEvents),
		CacheVerifyEvents:                 vUint(s.GoDebug.CacheVerifyEvents),
		Http2ClientEvents:                 vUint(s.GoDebug.Http2ClientEvents),
		Http2ServerEvents:                 vUint(s.GoDebug.Http2ServerEvents),
		Tls10ServerEvents:                 vUint(s.GoDebug.Tls10ServerEvents),
		Tls3DesEvents:                     vUint(s.GoDebug.Tls3DesEvents),
		TlsMaxRsaSizeEvents:               vUint(s.GoDebug.TlsMaxRsaSizeEvents),
		TlsRsaKexEvents:                   vUint(s.GoDebug.TlsRsaKexEvents),
		TlsSha1Events:                     vUint(s.GoDebug.TlsSha1Events),
		TlsUnsafeEkmEvents:                vUint(s.GoDebug.TlsUnsafeEkmEvents),
		HttpCookieMaxNumEvents:            vUint(s.GoDebug.HttpCookieMaxNumEvents),
		HttpLaxContentLengthEvents:        vUint(s.GoDebug.HttpLaxContentLengthEvents),
		HttpMuxGo121Events:                vUint(s.GoDebug.HttpMuxGo_121Events),
		HttpServeContentKeepHeadersEvents: vUint(s.GoDebug.HttpServeContentKeepHeadersEvents),
		AllowMultipleVcsEvents:            vUint(s.GoDebug.AllowMultipleVcsEvents),
		AsyncTimerChanEvents:              vUint(s.GoDebug.AsyncTimerChanEvents),
		ContainerMaxProcsEvents:           vUint(s.GoDebug.ContainerMaxProcsEvents),
		CryptoCustomRandEvents:            vUint(s.GoDebug.CryptoCustomRandEvents),
		EmbedFollowSymlinkEvents:          vUint(s.GoDebug.EmbedFollowSymlinkEvents),
		ExecErrDotEvents:                  vUint(s.GoDebug.ExecErrDotEvents),
		TestJsonBuildTextEvents:           vUint(s.GoDebug.TestJsonBuildTextEvents),
		TypeAliasEvents:                   vUint(s.GoDebug.TypeAliasEvents),
		HtmlMetaContentUrlEscapedEvents:   vUint(s.GoDebug.HtmlMetaContentUrlEscapedEvents),
		InstallGoRootEvents:               vUint(s.GoDebug.InstallGoRootEvents),
		MultipartMaxHeadersEvents:         vUint(s.GoDebug.MultipartMaxHeadersEvents),
		MultipartMaxPartsEvents:           vUint(s.GoDebug.MultipartMaxPartsEvents),
		MultipathTcpEvents:                vUint(s.GoDebug.MultipathTcpEvents),
		NetEdns0Events:                    vUint(s.GoDebug.NetEdns0Events),
		PanicNilEvents:                    vUint(s.GoDebug.PanicNilEvents),
		RandAutoSeedEvents:                vUint(s.GoDebug.RandAutoSeedEvents),
		RandSeedNopEvents:                 vUint(s.GoDebug.RandSeedNopEvents),
		Rsa1024MinEvents:                  vUint(s.GoDebug.Rsa1024MinEvents),
		TarInsecurePathEvents:             vUint(s.GoDebug.TarInsecurePathEvents),
		UpdateMaxProcsEvents:              vUint(s.GoDebug.UpdateMaxProcsEvents),
		UrlMaxQueryParamsEvents:           vUint(s.GoDebug.UrlMaxQueryParamsEvents),
		UrlStrictColonsEvents:             vUint(s.GoDebug.UrlStrictColonsEvents),
		WinReadLinkVolumeEvents:           vUint(s.GoDebug.WinReadLinkVolumeEvents),
		WinSymlinkEvents:                  vUint(s.GoDebug.WinSymlinkEvents),
		X509KeypairLeafEvents:             vUint(s.GoDebug.X509KeypairLeafEvents),
		X509NegativeSerialEvents:          vUint(s.GoDebug.X509NegativeSerialEvents),
		X509RsaCrtEvents:                  vUint(s.GoDebug.X509RsaCrtEvents),
		X509Sha256SkidEvents:              vUint(s.GoDebug.X509Sha256SkidEvents),
		X509UseFallbackRootsEvents:        vUint(s.GoDebug.X509UseFallbackRootsEvents),
		X509UsePoliciesEvents:             vUint(s.GoDebug.X509UsePoliciesEvents),
		ZipInsecurePathEvents:             vUint(s.GoDebug.ZipInsecurePathEvents),
	}
}

// SnapshotMemory retrieves a snapshot of memory-related metrics.
func SnapshotMemory() *models.MemoryResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect(false, true)

	return &models.MemoryResponse{
		HeapFreeBytes:        vUint(s.Memory.HeapFreeBytes),
		HeapObjectBytes:      vUint(s.Memory.HeapObjectBytes),
		HeapReleasedBytes:    vUint(s.Memory.HeapReleasedBytes),
		HeapStackBytes:       vUint(s.Memory.HeapStackBytes),
		HeapUnusedBytes:      vUint(s.Memory.HeapUnusedBytes),
		McacheFreeBytes:      vUint(s.Memory.McacheFreeBytes),
		McacheInuseBytes:     vUint(s.Memory.McacheInuseBytes),
		MspanFreeBytes:       vUint(s.Memory.MspanFreeBytes),
		MspanInuseBytes:      vUint(s.Memory.MspanInuseBytes),
		OtherMetadataBytes:   vUint(s.Memory.OtherMetadataBytes),
		OsStackBytes:         vUint(s.Memory.OsStackBytes),
		OtherBytes:           vUint(s.Memory.OtherBytes),
		ProfilingBucketBytes: vUint(s.Memory.ProfilingBucketBytes),
		TotalBytes:           vUint(s.Memory.TotalBytes),
	}
}

// SnapshotScheduler retrieves a snapshot of scheduler-related metrics.
func SnapshotScheduler() *models.SchedulerResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect(true, true)

	return &models.SchedulerResponse{
		LatenciesSeconds:       toModelHistogram(s.Scheduler.LatenciesSeconds),
		StoppingGcSeconds:      toModelHistogram(s.Scheduler.StoppingGcSeconds),
		StoppingOtherSeconds:   toModelHistogram(s.Scheduler.StoppingOtherSeconds),
		TotalGcPauseSeconds:    toModelHistogram(s.Scheduler.TotalGcPauseSeconds),
		TotalOtherPauseSeconds: toModelHistogram(s.Scheduler.TotalOtherPauseSeconds),
		GoroutinesCreated:      vUint(s.Scheduler.GoroutinesCreated),
		GomaxprocsThreads:      vUint32(s.Scheduler.GomaxprocsThreads),
		NotInGoGoroutines:      vUint32(s.Scheduler.NotInGoGoroutines),
		RunnableGoroutines:     vUint32(s.Scheduler.RunnableGoroutines),
		RunningGoroutines:      vUint32(s.Scheduler.RunningGoroutines),
		WaitingGoroutines:      vUint32(s.Scheduler.WaitingGoroutines),
		LiveGoroutines:         vUint32(s.Scheduler.LiveGoroutines),
		TotalLiveThreads:       vUint32(s.Scheduler.TotalLiveThreads),
	}
}

// SnapshotSync retrieves a snapshot of sync-related metrics.
func SnapshotSync() models.SyncResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect(false, true)

	return models.SyncResponse{
		TotalMutexWaitDurationSeconds: vFloat(s.Sync.TotalMutexWaitDurationSeconds),
	}
}

// SnapshotAll retrieves a single point-in-time capture of all supported Go runtime metrics.
func SnapshotAll() models.AllResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect(true, true)

	return models.AllResponse{
		Cgo: models.CgoResponse{
			CgoCalls: vUint(s.Cgo.CgoCalls),
		},
		Cpu: &models.CPUResponse{
			GcAssistSeconds:           vFloat(s.Cpu.GcAssistSeconds),
			GcDedicatedSeconds:        vFloat(s.Cpu.GcDedicatedSeconds),
			GcIdleSeconds:             vFloat(s.Cpu.GcIdleSeconds),
			GcPauseSeconds:            vFloat(s.Cpu.GcPauseSeconds),
			GcTotalSeconds:            vFloat(s.Cpu.GcTotalSeconds),
			IdleSeconds:               vFloat(s.Cpu.IdleSeconds),
			TotalSeconds:              vFloat(s.Cpu.TotalSeconds),
			UserSeconds:               vFloat(s.Cpu.UserSeconds),
			ScavengeAssistSeconds:     vFloat(s.Cpu.ScavengeAssistSeconds),
			ScavengeBackgroundSeconds: vFloat(s.Cpu.ScavengeBackgroundSeconds),
			ScavengeTotalSeconds:      vFloat(s.Cpu.ScavengeTotalSeconds),
		},
		Gc: &models.GCResponse{
			BytesAllocsBySize:        toModelHistogram(s.Gc.BytesAllocsBySize),
			BytesFreesBySize:         toModelHistogram(s.Gc.BytesFreesBySize),
			ExecutedCleanups:         vUint(s.Gc.ExecutedCleanups),
			QueuedCleanups:           vUint(s.Gc.QueuedCleanups),
			AutomaticCycles:          vUint(s.Gc.AutomaticCycles),
			ForcedCycles:             vUint(s.Gc.ForcedCycles),
			TotalCycles:              vUint(s.Gc.TotalCycles),
			ExecutedFinalizers:       vUint(s.Gc.ExecutedFinalizers),
			QueuedFinalizers:         vUint(s.Gc.QueuedFinalizers),
			MemoryLimitBytes:         vUint(s.Gc.MemoryLimitBytes),
			BytesAllocs:              vUint(s.Gc.BytesAllocs),
			ObjectsAllocs:            vUint(s.Gc.ObjectsAllocs),
			BytesFrees:               vUint(s.Gc.BytesFrees),
			ObjectsFrees:             vUint(s.Gc.ObjectsFrees),
			GoalHeapSizeBytes:        vUint(s.Gc.GoalHeapSizeBytes),
			LiveOccupiedBytes:        vUint(s.Gc.LiveOccupiedBytes),
			ObjectsOccupied:          vUint(s.Gc.ObjectsOccupied),
			TinyObjectsAllocs:        vUint(s.Gc.TinyObjectsAllocs),
			LimiterLastEnabledCycle:  vUint(s.Gc.LimiterLastEnabledCycle),
			GlobalScannedBytes:       vUint(s.Gc.GlobalScannedBytes),
			HeapScannedBytes:         vUint(s.Gc.HeapScannedBytes),
			StackScannedBytes:        vUint(s.Gc.StackScannedBytes),
			TotalScannedBytes:        vUint(s.Gc.TotalScannedBytes),
			StackStartingSize:        vUint(s.Gc.StackStartingSize),
			TargetHeapSizePercentage: vUint32(s.Gc.TargetHeapSizePercentage),
		},
		GoDebug: &models.GoDebugResponse{
			CacheHashEvents:                   vUint(s.GoDebug.CacheHashEvents),
			CacheTestEvents:                   vUint(s.GoDebug.CacheTestEvents),
			CacheVerifyEvents:                 vUint(s.GoDebug.CacheVerifyEvents),
			Http2ClientEvents:                 vUint(s.GoDebug.Http2ClientEvents),
			Http2ServerEvents:                 vUint(s.GoDebug.Http2ServerEvents),
			Tls10ServerEvents:                 vUint(s.GoDebug.Tls10ServerEvents),
			Tls3DesEvents:                     vUint(s.GoDebug.Tls3DesEvents),
			TlsMaxRsaSizeEvents:               vUint(s.GoDebug.TlsMaxRsaSizeEvents),
			TlsRsaKexEvents:                   vUint(s.GoDebug.TlsRsaKexEvents),
			TlsSha1Events:                     vUint(s.GoDebug.TlsSha1Events),
			TlsUnsafeEkmEvents:                vUint(s.GoDebug.TlsUnsafeEkmEvents),
			HttpCookieMaxNumEvents:            vUint(s.GoDebug.HttpCookieMaxNumEvents),
			HttpLaxContentLengthEvents:        vUint(s.GoDebug.HttpLaxContentLengthEvents),
			HttpMuxGo121Events:                vUint(s.GoDebug.HttpMuxGo_121Events),
			HttpServeContentKeepHeadersEvents: vUint(s.GoDebug.HttpServeContentKeepHeadersEvents),
			AllowMultipleVcsEvents:            vUint(s.GoDebug.AllowMultipleVcsEvents),
			AsyncTimerChanEvents:              vUint(s.GoDebug.AsyncTimerChanEvents),
			ContainerMaxProcsEvents:           vUint(s.GoDebug.ContainerMaxProcsEvents),
			CryptoCustomRandEvents:            vUint(s.GoDebug.CryptoCustomRandEvents),
			EmbedFollowSymlinkEvents:          vUint(s.GoDebug.EmbedFollowSymlinkEvents),
			ExecErrDotEvents:                  vUint(s.GoDebug.ExecErrDotEvents),
			TestJsonBuildTextEvents:           vUint(s.GoDebug.TestJsonBuildTextEvents),
			TypeAliasEvents:                   vUint(s.GoDebug.TypeAliasEvents),
			HtmlMetaContentUrlEscapedEvents:   vUint(s.GoDebug.HtmlMetaContentUrlEscapedEvents),
			InstallGoRootEvents:               vUint(s.GoDebug.InstallGoRootEvents),
			MultipartMaxHeadersEvents:         vUint(s.GoDebug.MultipartMaxHeadersEvents),
			MultipartMaxPartsEvents:           vUint(s.GoDebug.MultipartMaxPartsEvents),
			MultipathTcpEvents:                vUint(s.GoDebug.MultipathTcpEvents),
			NetEdns0Events:                    vUint(s.GoDebug.NetEdns0Events),
			PanicNilEvents:                    vUint(s.GoDebug.PanicNilEvents),
			RandAutoSeedEvents:                vUint(s.GoDebug.RandAutoSeedEvents),
			RandSeedNopEvents:                 vUint(s.GoDebug.RandSeedNopEvents),
			Rsa1024MinEvents:                  vUint(s.GoDebug.Rsa1024MinEvents),
			TarInsecurePathEvents:             vUint(s.GoDebug.TarInsecurePathEvents),
			UpdateMaxProcsEvents:              vUint(s.GoDebug.UpdateMaxProcsEvents),
			UrlMaxQueryParamsEvents:           vUint(s.GoDebug.UrlMaxQueryParamsEvents),
			UrlStrictColonsEvents:             vUint(s.GoDebug.UrlStrictColonsEvents),
			WinReadLinkVolumeEvents:           vUint(s.GoDebug.WinReadLinkVolumeEvents),
			WinSymlinkEvents:                  vUint(s.GoDebug.WinSymlinkEvents),
			X509KeypairLeafEvents:             vUint(s.GoDebug.X509KeypairLeafEvents),
			X509NegativeSerialEvents:          vUint(s.GoDebug.X509NegativeSerialEvents),
			X509RsaCrtEvents:                  vUint(s.GoDebug.X509RsaCrtEvents),
			X509Sha256SkidEvents:              vUint(s.GoDebug.X509Sha256SkidEvents),
			X509UseFallbackRootsEvents:        vUint(s.GoDebug.X509UseFallbackRootsEvents),
			X509UsePoliciesEvents:             vUint(s.GoDebug.X509UsePoliciesEvents),
			ZipInsecurePathEvents:             vUint(s.GoDebug.ZipInsecurePathEvents),
		},
		Memory: &models.MemoryResponse{
			HeapFreeBytes:        vUint(s.Memory.HeapFreeBytes),
			HeapObjectBytes:      vUint(s.Memory.HeapObjectBytes),
			HeapReleasedBytes:    vUint(s.Memory.HeapReleasedBytes),
			HeapStackBytes:       vUint(s.Memory.HeapStackBytes),
			HeapUnusedBytes:      vUint(s.Memory.HeapUnusedBytes),
			McacheFreeBytes:      vUint(s.Memory.McacheFreeBytes),
			McacheInuseBytes:     vUint(s.Memory.McacheInuseBytes),
			MspanFreeBytes:       vUint(s.Memory.MspanFreeBytes),
			MspanInuseBytes:      vUint(s.Memory.MspanInuseBytes),
			OtherMetadataBytes:   vUint(s.Memory.OtherMetadataBytes),
			OsStackBytes:         vUint(s.Memory.OsStackBytes),
			OtherBytes:           vUint(s.Memory.OtherBytes),
			ProfilingBucketBytes: vUint(s.Memory.ProfilingBucketBytes),
			TotalBytes:           vUint(s.Memory.TotalBytes),
		},
		Scheduler: &models.SchedulerResponse{
			LatenciesSeconds:       toModelHistogram(s.Scheduler.LatenciesSeconds),
			StoppingGcSeconds:      toModelHistogram(s.Scheduler.StoppingGcSeconds),
			StoppingOtherSeconds:   toModelHistogram(s.Scheduler.StoppingOtherSeconds),
			TotalGcPauseSeconds:    toModelHistogram(s.Scheduler.TotalGcPauseSeconds),
			TotalOtherPauseSeconds: toModelHistogram(s.Scheduler.TotalOtherPauseSeconds),
			GoroutinesCreated:      vUint(s.Scheduler.GoroutinesCreated),
			GomaxprocsThreads:      vUint32(s.Scheduler.GomaxprocsThreads),
			NotInGoGoroutines:      vUint32(s.Scheduler.NotInGoGoroutines),
			RunnableGoroutines:     vUint32(s.Scheduler.RunnableGoroutines),
			RunningGoroutines:      vUint32(s.Scheduler.RunningGoroutines),
			WaitingGoroutines:      vUint32(s.Scheduler.WaitingGoroutines),
			LiveGoroutines:         vUint32(s.Scheduler.LiveGoroutines),
			TotalLiveThreads:       vUint32(s.Scheduler.TotalLiveThreads),
		},
		Sync: models.SyncResponse{
			TotalMutexWaitDurationSeconds: vFloat(s.Sync.TotalMutexWaitDurationSeconds),
		},
	}
}
