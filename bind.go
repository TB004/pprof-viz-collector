// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package pprofvizcollector

import (
	"net/http"

	"github.com/TB004/pprof-viz-collector/internal/proto/protoconnect"
	"github.com/TB004/pprof-viz-collector/internal/rpc"
)

// BindConfig configures Bind.
type BindConfig struct {
	// Token, if non-empty, is required on every RPC call via Authorization: Bearer <token>.
	Token string
	// AllowedOrigins is the list of origins permitted by CORS (e.g. ["https://app.example.com"]).
	// Use ["*"] to allow all origins. An empty slice disables CORS middleware entirely.
	AllowedOrigins []string
	// CompressMinBytes is the minimum response size, in bytes, below which
	// responses are sent uncompressed. Defaults to 256 when zero. Lower it on
	// bandwidth-constrained links, where compressing even small frames is
	// worth the CPU cost; raise it if compression's CPU cost outweighs the
	// bandwidth savings for your typical message sizes.
	CompressMinBytes int
}

// Bind registers the CollectorService handler to the provided http.ServeMux.
// Server-level concerns like TLS and HTTP/3 are managed externally by whatever
// serves mux; use New instead if you want this package to run that server.
func Bind(mux *http.ServeMux, cfg BindConfig) {
	path, handler := protoconnect.NewCollectorServiceHandler(rpc.CollectorServiceHandler{}, handlerOpts(cfg.Token, cfg.CompressMinBytes)...)
	if len(cfg.AllowedOrigins) > 0 {
		handler = newCORSHandler(cfg.AllowedOrigins, handler)
	}
	mux.Handle(path, handler)
}
