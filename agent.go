// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package pprofvizagent

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/TB004/pprof-viz-agent/internal"
	"github.com/TB004/pprof-viz-agent/internal/proto"
	"github.com/TB004/pprof-viz-agent/internal/proto/protoconnect"
	"github.com/TB004/pprof-viz-agent/pkg/models"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Config holds configuration for the agent.
type Config struct {
	// Token, if non-empty, is required on every RPC call via Authorization: Bearer <token>.
	// Applies to both Bind and New.
	Token string
	// CertFile is the path to the TLS certificate file. Enables TLS when set. Only used by New.
	CertFile string
	// KeyFile is the path to the TLS private key file. Must be set together with CertFile. Only used by New.
	KeyFile string
	// HTTP3 also serves HTTP/3 over QUIC in parallel with the TLS/TCP server.
	// Requires CertFile and KeyFile. Only used by New.
	HTTP3 bool
}

// handlerOpts builds the shared ConnectRPC handler options from a Config.
func handlerOpts(cfg Config) []connect.HandlerOption {
	opts := []connect.HandlerOption{
		connect.WithCompression(internal.CompressionZstd, internal.NewZstdDecompressor, internal.NewZstdCompressor),
		connect.WithCompressMinBytes(256),
	}
	if cfg.Token != "" {
		opts = append(opts, connect.WithInterceptors(internal.NewTokenInterceptor(cfg.Token)))
	}
	return opts
}

// Bind registers the AgentService handler to the provided http.ServeMux.
// cfg.TLS and cfg.HTTP3 are ignored; those are server-level concerns managed externally.
func Bind(mux *http.ServeMux, cfg Config) {
	path, handler := protoconnect.NewAgentServiceHandler(internal.AgentServiceHandler{}, handlerOpts(cfg)...)
	mux.Handle(path, handler)
}

// New creates and starts an HTTP server on the specified port serving the AgentService.
// Without TLS, it supports HTTP/1.1 and HTTP/2 over cleartext (h2c).
// With CertFile and KeyFile set, it supports HTTP/1.1 and HTTP/2 over TLS.
// With HTTP3 also set, it additionally serves HTTP/3 over QUIC in parallel, with Alt-Svc
// headers on TLS responses so clients can upgrade automatically.
func New(port int, cfg Config) error {
	if cfg.HTTP3 && (cfg.CertFile == "" || cfg.KeyFile == "") {
		return errors.New("HTTP3 requires CertFile and KeyFile to be configured")
	}

	mux := http.NewServeMux()
	path, handler := protoconnect.NewAgentServiceHandler(internal.AgentServiceHandler{}, handlerOpts(cfg)...)
	mux.Handle(path, handler)

	addr := fmt.Sprintf("localhost:%d", port)

	if cfg.CertFile != "" {
		if cfg.HTTP3 {
			return http3.ListenAndServeTLS(addr, cfg.CertFile, cfg.KeyFile, mux)
		}
		s := &http.Server{
			Addr:        addr,
			Handler:     mux,
			ReadTimeout: 10 * time.Second,
		}
		return s.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
	}

	s := &http.Server{
		Addr:        addr,
		Handler:     h2c.NewHandler(mux, &http2.Server{}),
		ReadTimeout: 10 * time.Second,
	}
	return s.ListenAndServe()
}

// Helper functions for safe dereferencing of optional proto fields.
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
	s := c.Collect()

	return models.CgoResponse{
		CgoCalls: vUint(s.Cgo.CgoCalls),
	}
}

// SnapshotGC retrieves a snapshot of GC-related metrics.
func SnapshotGC() *models.GCResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect()

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
	s := c.Collect()

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
	s := c.Collect()

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
	s := c.Collect()

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
	s := c.Collect()

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
	s := c.Collect()

	return models.SyncResponse{
		TotalMutexWaitDurationSeconds: vFloat(s.Sync.TotalMutexWaitDurationSeconds),
	}
}

// SnapshotAll retrieves a single point-in-time capture of all supported Go runtime metrics.
func SnapshotAll() models.AllResponse {
	c := internal.NewSampleCollector()
	c.Read()
	s := c.Collect()

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
