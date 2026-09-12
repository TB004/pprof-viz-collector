// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

type Histogram struct {
	Counts  []uint64  `json:"counts"`
	Buckets []float32 `json:"buckets"`
}

type CgoResponse struct {
	CgoCalls uint64 `json:"cgo_calls"`
}

type CPUResponse struct {
	GcAssistSeconds           float64 `json:"gc_assist_seconds"`
	GcDedicatedSeconds        float64 `json:"gc_dedicated_seconds"`
	GcIdleSeconds             float64 `json:"gc_idle_seconds"`
	GcPauseSeconds            float64 `json:"gc_pause_seconds"`
	GcTotalSeconds            float64 `json:"gc_total_seconds"`
	IdleSeconds               float64 `json:"idle_seconds"`
	TotalSeconds              float64 `json:"total_seconds"`
	UserSeconds               float64 `json:"user_seconds"`
	ScavengeAssistSeconds     float64 `json:"scavenge_assist_seconds"`
	ScavengeBackgroundSeconds float64 `json:"scavenge_background_seconds"`
	ScavengeTotalSeconds      float64 `json:"scavenge_total_seconds"`
}

type GCResponse struct {
	BytesAllocsBySize        *Histogram `json:"bytes_allocs_by_size"`
	BytesFreesBySize         *Histogram `json:"bytes_frees_by_size"`
	ExecutedCleanups         uint64     `json:"executed_cleanups"`
	QueuedCleanups           uint64     `json:"queued_cleanups"`
	AutomaticCycles          uint64     `json:"automatic_cycles"`
	ForcedCycles             uint64     `json:"forced_cycles"`
	TotalCycles              uint64     `json:"total_cycles"`
	ExecutedFinalizers       uint64     `json:"executed_finalizers"`
	QueuedFinalizers         uint64     `json:"queued_finalizers"`
	MemoryLimitBytes         uint64     `json:"memory_limit_bytes"`
	BytesAllocs              uint64     `json:"bytes_allocs"`
	ObjectsAllocs            uint64     `json:"objects_allocs"`
	BytesFrees               uint64     `json:"bytes_frees"`
	ObjectsFrees             uint64     `json:"objects_frees"`
	GoalHeapSizeBytes        uint64     `json:"goal_heap_size_bytes"`
	LiveOccupiedBytes        uint64     `json:"live_occupied_bytes"`
	ObjectsOccupied          uint64     `json:"objects_occupied"`
	TinyObjectsAllocs        uint64     `json:"tiny_objects_allocs"`
	LimiterLastEnabledCycle  uint64     `json:"limiter_last_enabled_cycle"`
	GlobalScannedBytes       uint64     `json:"global_scanned_bytes"`
	HeapScannedBytes         uint64     `json:"heap_scanned_bytes"`
	StackScannedBytes        uint64     `json:"stack_scanned_bytes"`
	TotalScannedBytes        uint64     `json:"total_scanned_bytes"`
	StackStartingSize        uint64     `json:"stack_starting_size"`
	TargetHeapSizePercentage uint32     `json:"target_heap_size_percentage"`
}

type GoDebugResponse struct {
	CacheHashEvents                   uint64 `json:"cache_hash_events"`
	CacheTestEvents                   uint64 `json:"cache_test_events"`
	CacheVerifyEvents                 uint64 `json:"cache_verify_events"`
	Http2ClientEvents                 uint64 `json:"http2_client_events"`
	Http2ServerEvents                 uint64 `json:"http2_server_events"`
	TlsMaxRsaSizeEvents               uint64 `json:"tls_max_rsa_size_events"`
	TlsSha1Events                     uint64 `json:"tls_sha1_events"`
	HttpCookieMaxNumEvents            uint64 `json:"http_cookie_max_num_events"`
	HttpLaxContentLengthEvents        uint64 `json:"http_lax_content_length_events"`
	HttpMuxGo121Events                uint64 `json:"http_mux_go_121_events"`
	HttpServeContentKeepHeadersEvents uint64 `json:"http_serve_content_keep_headers_events"`
	AllowMultipleVcsEvents            uint64 `json:"allow_multiple_vcs_events"`
	ContainerMaxProcsEvents           uint64 `json:"container_max_procs_events"`
	CryptoCustomRandEvents            uint64 `json:"crypto_custom_rand_events"`
	EmbedFollowSymlinkEvents          uint64 `json:"embed_follow_symlink_events"`
	ExecErrDotEvents                  uint64 `json:"exec_err_dot_events"`
	Fips140EmsEvents                  uint64 `json:"fips140_ems_events"`
	TestJsonBuildTextEvents           uint64 `json:"test_json_build_text_events"`
	HtmlMetaContentUrlEscapedEvents   uint64 `json:"html_meta_content_url_escaped_events"`
	InstallGoRootEvents               uint64 `json:"install_go_root_events"`
	MultipartMaxHeadersEvents         uint64 `json:"multipart_max_headers_events"`
	MultipartMaxPartsEvents           uint64 `json:"multipart_max_parts_events"`
	MultipathTcpEvents                uint64 `json:"multipath_tcp_events"`
	NetEdns0Events                    uint64 `json:"net_edns0_events"`
	PanicNilEvents                    uint64 `json:"panic_nil_events"`
	RandAutoSeedEvents                uint64 `json:"rand_auto_seed_events"`
	RandSeedNopEvents                 uint64 `json:"rand_seed_nop_events"`
	Rsa1024MinEvents                  uint64 `json:"rsa1024_min_events"`
	TarInsecurePathEvents             uint64 `json:"tar_insecure_path_events"`
	UpdateMaxProcsEvents              uint64 `json:"update_max_procs_events"`
	UrlMaxQueryParamsEvents           uint64 `json:"url_max_query_params_events"`
	UrlStrictColonsEvents             uint64 `json:"url_strict_colons_events"`
	WinReadLinkVolumeEvents           uint64 `json:"win_read_link_volume_events"`
	WinSymlinkEvents                  uint64 `json:"win_symlink_events"`
	X509NegativeSerialEvents          uint64 `json:"x509_negative_serial_events"`
	X509RsaCrtEvents                  uint64 `json:"x509_rsa_crt_events"`
	X509Sha256SkidEvents              uint64 `json:"x509_sha256_skid_events"`
	X509SslCertOverridePlatformEvents uint64 `json:"x509_ssl_cert_override_platform_events"`
	X509UseFallbackRootsEvents        uint64 `json:"x509_use_fallback_roots_events"`
	X509UsePoliciesEvents             uint64 `json:"x509_use_policies_events"`
	ZipInsecurePathEvents             uint64 `json:"zip_insecure_path_events"`
}

type MemoryResponse struct {
	HeapFreeBytes        uint64 `json:"heap_free_bytes"`
	HeapObjectBytes      uint64 `json:"heap_object_bytes"`
	HeapReleasedBytes    uint64 `json:"heap_released_bytes"`
	HeapStackBytes       uint64 `json:"heap_stack_bytes"`
	HeapUnusedBytes      uint64 `json:"heap_unused_bytes"`
	McacheFreeBytes      uint64 `json:"mcache_free_bytes"`
	McacheInuseBytes     uint64 `json:"mcache_inuse_bytes"`
	MspanFreeBytes       uint64 `json:"mspan_free_bytes"`
	MspanInuseBytes      uint64 `json:"mspan_inuse_bytes"`
	OtherMetadataBytes   uint64 `json:"other_metadata_bytes"`
	OsStackBytes         uint64 `json:"os_stack_bytes"`
	OtherBytes           uint64 `json:"other_bytes"`
	ProfilingBucketBytes uint64 `json:"profiling_bucket_bytes"`
	TotalBytes           uint64 `json:"total_bytes"`
}

type SchedulerResponse struct {
	LatenciesSeconds       *Histogram `json:"latencies_seconds"`
	StoppingGcSeconds      *Histogram `json:"stopping_gc_seconds"`
	StoppingOtherSeconds   *Histogram `json:"stopping_other_seconds"`
	TotalGcPauseSeconds    *Histogram `json:"total_gc_pause_seconds"`
	TotalOtherPauseSeconds *Histogram `json:"total_other_pause_seconds"`
	GoroutinesCreated      uint64     `json:"goroutines_created"`
	GomaxprocsThreads      uint32     `json:"gomaxprocs_threads"`
	NotInGoGoroutines      uint32     `json:"not_in_go_goroutines"`
	RunnableGoroutines     uint32     `json:"runnable_goroutines"`
	RunningGoroutines      uint32     `json:"running_goroutines"`
	WaitingGoroutines      uint32     `json:"waiting_goroutines"`
	LiveGoroutines         uint32     `json:"live_goroutines"`
	TotalLiveThreads       uint32     `json:"total_live_threads"`
}

type SyncResponse struct {
	TotalMutexWaitDurationSeconds float64 `json:"total_mutex_wait_duration_seconds"`
}

type AllResponse struct {
	Cgo       CgoResponse        `json:"cgo"`
	Sync      SyncResponse       `json:"sync"`
	Cpu       *CPUResponse       `json:"cpu"`
	Gc        *GCResponse        `json:"gc"`
	GoDebug   *GoDebugResponse   `json:"go_debug"`
	Memory    *MemoryResponse    `json:"memory"`
	Scheduler *SchedulerResponse `json:"scheduler"`
}
