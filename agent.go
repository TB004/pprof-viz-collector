// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package pprofvizagent

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"github.com/TB004/pprof-viz-agent/internal"
	"github.com/TB004/pprof-viz-agent/internal/proto"
	"github.com/TB004/pprof-viz-agent/internal/proto/protoconnect"
	"github.com/TB004/pprof-viz-agent/pkg/models"
	"github.com/quic-go/quic-go/http3"
	"github.com/rs/cors"
)

// Config holds configuration for the agent.
type Config struct {
	// Token, if non-empty, is required on every RPC call via Authorization: Bearer <token>.
	// Applies to both Bind and New.
	Token string
	// AllowedOrigins is the list of origins permitted by CORS (e.g. ["https://app.example.com"]).
	// Use ["*"] to allow all origins. An empty slice disables CORS middleware entirely.
	// Applies to both Bind and New.
	AllowedOrigins []string
	// CertFile is the path to the TLS certificate file. Enables TLS when set. Only used by New.
	CertFile string
	// KeyFile is the path to the TLS private key file. Must be set together with CertFile. Only used by New.
	KeyFile string
	// HTTP3 also serves HTTP/3 over QUIC in parallel with the TLS/TCP server.
	// Requires CertFile and KeyFile. Only used by New.
	HTTP3 bool
}

// newCORSHandler wraps a handler with CORS middleware configured for ConnectRPC.
// It allows the standard ConnectRPC headers plus Authorization for token auth.
func newCORSHandler(origins []string, h http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins: origins,
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: append(connectcors.AllowedHeaders(), "Authorization"),
		ExposedHeaders: connectcors.ExposedHeaders(),
	}).Handler(h)
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
// cfg.CertFile, cfg.KeyFile, and cfg.HTTP3 are ignored; those are server-level concerns managed externally.
func Bind(mux *http.ServeMux, cfg Config) {
	path, handler := protoconnect.NewAgentServiceHandler(internal.AgentServiceHandler{}, handlerOpts(cfg)...)
	if len(cfg.AllowedOrigins) > 0 {
		handler = newCORSHandler(cfg.AllowedOrigins, handler)
	}
	mux.Handle(path, handler)
}

// Server timeouts applied to every TCP server New creates. WriteTimeout is
// deliberately omitted: StreamAll is a long-lived server stream and a write
// deadline would terminate it. ReadHeaderTimeout and IdleTimeout guard against
// slowloris-style connection exhaustion.
const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	idleTimeout       = 120 * time.Second
)

// newHTTPServer builds an http.Server with hardened timeouts shared by all
// New entry points.
func newHTTPServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		IdleTimeout:       idleTimeout,
	}
}

// New creates and starts an HTTP server on the specified port serving the AgentService.
// Without TLS, it supports HTTP/2 over cleartext.
// With CertFile and KeyFile set, it supports HTTP/2 over TLS (minimum TLS 1.2).
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

	// Wrap mux with CORS middleware if origins are configured.
	var h http.Handler = mux
	if len(cfg.AllowedOrigins) > 0 {
		h = newCORSHandler(cfg.AllowedOrigins, mux)
	}

	// Cleartext HTTP/2 (h2c) when no certificate is configured.
	if cfg.CertFile == "" {
		protos := new(http.Protocols)
		protos.SetHTTP1(false)
		protos.SetHTTP2(false)
		protos.SetUnencryptedHTTP2(true)
		s := newHTTPServer(addr, h)
		s.Protocols = protos
		return s.ListenAndServe()
	}

	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}

	// TLS without HTTP/3: a single HTTP/2-over-TLS server.
	if !cfg.HTTP3 {
		s := newHTTPServer(addr, h)
		s.TLSConfig = tlsConfig
		return s.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
	}

	// TLS with HTTP/3: serve HTTP/2 over TLS/TCP and HTTP/3 over QUIC in parallel,
	// advertising QUIC via Alt-Svc headers on the TLS/TCP responses. Whichever
	// server fails first wins; the other is closed before returning.
	h3 := &http3.Server{
		Addr:      addr,
		Handler:   h,
		TLSConfig: tlsConfig.Clone(),
	}
	tcp := newHTTPServer(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// #nosec G104 -- best-effort Alt-Svc advertisement
		_ = h3.SetQUICHeaders(w.Header())
		h.ServeHTTP(w, r)
	}))
	tcp.TLSConfig = tlsConfig.Clone()

	errc := make(chan error, 2)
	go func() { errc <- tcp.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile) }()
	go func() { errc <- h3.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile) }()

	// Whichever server fails first wins; close the other and surface any
	// shutdown errors alongside the original. errors.Join drops nils.
	err := <-errc
	return errors.Join(err, tcp.Close(), h3.Close())
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

func toModelHistogram(h *proto.Histogram) *models.Histogram {
	if h == nil {
		return nil
	}
	return &models.Histogram{
		Counts:  h.Counts,
		Buckets: h.Buckets,
	}
}

// The following toXxx helpers map a single category of a proto.AllResponse into
// its models counterpart. They are shared by the per-category Snapshot* funcs
// and SnapshotAll so the mapping lives in exactly one place.

func toCgoResponse(c *proto.CgoResponse) models.CgoResponse {
	return models.CgoResponse{
		CgoCalls: vUint(c.CgoCalls),
	}
}

func toCPUResponse(c *proto.CPUResponse) *models.CPUResponse {
	return &models.CPUResponse{
		GcAssistSeconds:           vFloat(c.GcAssistSeconds),
		GcDedicatedSeconds:        vFloat(c.GcDedicatedSeconds),
		GcIdleSeconds:             vFloat(c.GcIdleSeconds),
		GcPauseSeconds:            vFloat(c.GcPauseSeconds),
		GcTotalSeconds:            vFloat(c.GcTotalSeconds),
		IdleSeconds:               vFloat(c.IdleSeconds),
		TotalSeconds:              vFloat(c.TotalSeconds),
		UserSeconds:               vFloat(c.UserSeconds),
		ScavengeAssistSeconds:     vFloat(c.ScavengeAssistSeconds),
		ScavengeBackgroundSeconds: vFloat(c.ScavengeBackgroundSeconds),
		ScavengeTotalSeconds:      vFloat(c.ScavengeTotalSeconds),
	}
}

func toGCResponse(g *proto.GCResponse) *models.GCResponse {
	return &models.GCResponse{
		BytesAllocsBySize:        toModelHistogram(g.BytesAllocsBySize),
		BytesFreesBySize:         toModelHistogram(g.BytesFreesBySize),
		ExecutedCleanups:         vUint(g.ExecutedCleanups),
		QueuedCleanups:           vUint(g.QueuedCleanups),
		AutomaticCycles:          vUint(g.AutomaticCycles),
		ForcedCycles:             vUint(g.ForcedCycles),
		TotalCycles:              vUint(g.TotalCycles),
		ExecutedFinalizers:       vUint(g.ExecutedFinalizers),
		QueuedFinalizers:         vUint(g.QueuedFinalizers),
		MemoryLimitBytes:         vUint(g.MemoryLimitBytes),
		BytesAllocs:              vUint(g.BytesAllocs),
		ObjectsAllocs:            vUint(g.ObjectsAllocs),
		BytesFrees:               vUint(g.BytesFrees),
		ObjectsFrees:             vUint(g.ObjectsFrees),
		GoalHeapSizeBytes:        vUint(g.GoalHeapSizeBytes),
		LiveOccupiedBytes:        vUint(g.LiveOccupiedBytes),
		ObjectsOccupied:          vUint(g.ObjectsOccupied),
		TinyObjectsAllocs:        vUint(g.TinyObjectsAllocs),
		LimiterLastEnabledCycle:  vUint(g.LimiterLastEnabledCycle),
		GlobalScannedBytes:       vUint(g.GlobalScannedBytes),
		HeapScannedBytes:         vUint(g.HeapScannedBytes),
		StackScannedBytes:        vUint(g.StackScannedBytes),
		TotalScannedBytes:        vUint(g.TotalScannedBytes),
		StackStartingSize:        vUint(g.StackStartingSize),
		TargetHeapSizePercentage: vUint32(g.TargetHeapSizePercentage),
	}
}

func toGoDebugResponse(d *proto.GoDebugResponse) *models.GoDebugResponse {
	return &models.GoDebugResponse{
		CacheHashEvents:                   vUint(d.CacheHashEvents),
		CacheTestEvents:                   vUint(d.CacheTestEvents),
		CacheVerifyEvents:                 vUint(d.CacheVerifyEvents),
		Http2ClientEvents:                 vUint(d.Http2ClientEvents),
		Http2ServerEvents:                 vUint(d.Http2ServerEvents),
		Tls10ServerEvents:                 vUint(d.Tls10ServerEvents),
		Tls3DesEvents:                     vUint(d.Tls3DesEvents),
		TlsMaxRsaSizeEvents:               vUint(d.TlsMaxRsaSizeEvents),
		TlsRsaKexEvents:                   vUint(d.TlsRsaKexEvents),
		TlsSha1Events:                     vUint(d.TlsSha1Events),
		TlsUnsafeEkmEvents:                vUint(d.TlsUnsafeEkmEvents),
		HttpCookieMaxNumEvents:            vUint(d.HttpCookieMaxNumEvents),
		HttpLaxContentLengthEvents:        vUint(d.HttpLaxContentLengthEvents),
		HttpMuxGo121Events:                vUint(d.HttpMuxGo_121Events),
		HttpServeContentKeepHeadersEvents: vUint(d.HttpServeContentKeepHeadersEvents),
		AllowMultipleVcsEvents:            vUint(d.AllowMultipleVcsEvents),
		AsyncTimerChanEvents:              vUint(d.AsyncTimerChanEvents),
		ContainerMaxProcsEvents:           vUint(d.ContainerMaxProcsEvents),
		CryptoCustomRandEvents:            vUint(d.CryptoCustomRandEvents),
		EmbedFollowSymlinkEvents:          vUint(d.EmbedFollowSymlinkEvents),
		ExecErrDotEvents:                  vUint(d.ExecErrDotEvents),
		TestJsonBuildTextEvents:           vUint(d.TestJsonBuildTextEvents),
		TypeAliasEvents:                   vUint(d.TypeAliasEvents),
		HtmlMetaContentUrlEscapedEvents:   vUint(d.HtmlMetaContentUrlEscapedEvents),
		InstallGoRootEvents:               vUint(d.InstallGoRootEvents),
		MultipartMaxHeadersEvents:         vUint(d.MultipartMaxHeadersEvents),
		MultipartMaxPartsEvents:           vUint(d.MultipartMaxPartsEvents),
		MultipathTcpEvents:                vUint(d.MultipathTcpEvents),
		NetEdns0Events:                    vUint(d.NetEdns0Events),
		PanicNilEvents:                    vUint(d.PanicNilEvents),
		RandAutoSeedEvents:                vUint(d.RandAutoSeedEvents),
		RandSeedNopEvents:                 vUint(d.RandSeedNopEvents),
		Rsa1024MinEvents:                  vUint(d.Rsa1024MinEvents),
		TarInsecurePathEvents:             vUint(d.TarInsecurePathEvents),
		UpdateMaxProcsEvents:              vUint(d.UpdateMaxProcsEvents),
		UrlMaxQueryParamsEvents:           vUint(d.UrlMaxQueryParamsEvents),
		UrlStrictColonsEvents:             vUint(d.UrlStrictColonsEvents),
		WinReadLinkVolumeEvents:           vUint(d.WinReadLinkVolumeEvents),
		WinSymlinkEvents:                  vUint(d.WinSymlinkEvents),
		X509KeypairLeafEvents:             vUint(d.X509KeypairLeafEvents),
		X509NegativeSerialEvents:          vUint(d.X509NegativeSerialEvents),
		X509RsaCrtEvents:                  vUint(d.X509RsaCrtEvents),
		X509Sha256SkidEvents:              vUint(d.X509Sha256SkidEvents),
		X509UseFallbackRootsEvents:        vUint(d.X509UseFallbackRootsEvents),
		X509UsePoliciesEvents:             vUint(d.X509UsePoliciesEvents),
		ZipInsecurePathEvents:             vUint(d.ZipInsecurePathEvents),
	}
}

func toMemoryResponse(m *proto.MemoryResponse) *models.MemoryResponse {
	return &models.MemoryResponse{
		HeapFreeBytes:        vUint(m.HeapFreeBytes),
		HeapObjectBytes:      vUint(m.HeapObjectBytes),
		HeapReleasedBytes:    vUint(m.HeapReleasedBytes),
		HeapStackBytes:       vUint(m.HeapStackBytes),
		HeapUnusedBytes:      vUint(m.HeapUnusedBytes),
		McacheFreeBytes:      vUint(m.McacheFreeBytes),
		McacheInuseBytes:     vUint(m.McacheInuseBytes),
		MspanFreeBytes:       vUint(m.MspanFreeBytes),
		MspanInuseBytes:      vUint(m.MspanInuseBytes),
		OtherMetadataBytes:   vUint(m.OtherMetadataBytes),
		OsStackBytes:         vUint(m.OsStackBytes),
		OtherBytes:           vUint(m.OtherBytes),
		ProfilingBucketBytes: vUint(m.ProfilingBucketBytes),
		TotalBytes:           vUint(m.TotalBytes),
	}
}

func toSchedulerResponse(s *proto.SchedulerResponse) *models.SchedulerResponse {
	return &models.SchedulerResponse{
		LatenciesSeconds:       toModelHistogram(s.LatenciesSeconds),
		StoppingGcSeconds:      toModelHistogram(s.StoppingGcSeconds),
		StoppingOtherSeconds:   toModelHistogram(s.StoppingOtherSeconds),
		TotalGcPauseSeconds:    toModelHistogram(s.TotalGcPauseSeconds),
		TotalOtherPauseSeconds: toModelHistogram(s.TotalOtherPauseSeconds),
		GoroutinesCreated:      vUint(s.GoroutinesCreated),
		GomaxprocsThreads:      vUint32(s.GomaxprocsThreads),
		NotInGoGoroutines:      vUint32(s.NotInGoGoroutines),
		RunnableGoroutines:     vUint32(s.RunnableGoroutines),
		RunningGoroutines:      vUint32(s.RunningGoroutines),
		WaitingGoroutines:      vUint32(s.WaitingGoroutines),
		LiveGoroutines:         vUint32(s.LiveGoroutines),
		TotalLiveThreads:       vUint32(s.TotalLiveThreads),
	}
}

func toSyncResponse(s *proto.SyncResponse) models.SyncResponse {
	return models.SyncResponse{
		TotalMutexWaitDurationSeconds: vFloat(s.TotalMutexWaitDurationSeconds),
	}
}

// snapshot reads a single point-in-time capture of all runtime metrics.
func snapshot() *proto.AllResponse {
	return internal.Snapshot()
}

// SnapshotCgo retrieves a snapshot of cgo-related metrics.
func SnapshotCgo() models.CgoResponse {
	return toCgoResponse(snapshot().Cgo)
}

// SnapshotGC retrieves a snapshot of GC-related metrics.
func SnapshotGC() *models.GCResponse {
	return toGCResponse(snapshot().Gc)
}

// SnapshotCPU retrieves a snapshot of CPU-related metrics.
func SnapshotCPU() *models.CPUResponse {
	return toCPUResponse(snapshot().Cpu)
}

// SnapshotGoDebug retrieves a snapshot of godebug-related metrics.
func SnapshotGoDebug() *models.GoDebugResponse {
	return toGoDebugResponse(snapshot().GoDebug)
}

// SnapshotMemory retrieves a snapshot of memory-related metrics.
func SnapshotMemory() *models.MemoryResponse {
	return toMemoryResponse(snapshot().Memory)
}

// SnapshotScheduler retrieves a snapshot of scheduler-related metrics.
func SnapshotScheduler() *models.SchedulerResponse {
	return toSchedulerResponse(snapshot().Scheduler)
}

// SnapshotSync retrieves a snapshot of sync-related metrics.
func SnapshotSync() models.SyncResponse {
	return toSyncResponse(snapshot().Sync)
}

// SnapshotAll retrieves a single point-in-time capture of all supported Go runtime metrics.
func SnapshotAll() models.AllResponse {
	s := snapshot()
	return models.AllResponse{
		Cgo:       toCgoResponse(s.Cgo),
		Cpu:       toCPUResponse(s.Cpu),
		Gc:        toGCResponse(s.Gc),
		GoDebug:   toGoDebugResponse(s.GoDebug),
		Memory:    toMemoryResponse(s.Memory),
		Scheduler: toSchedulerResponse(s.Scheduler),
		Sync:      toSyncResponse(s.Sync),
	}
}
