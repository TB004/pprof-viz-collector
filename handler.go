// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package pprofvizcollector

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"github.com/TB004/pprof-viz-collector/internal/rpc"
	"github.com/rs/cors"
)

func newCORSHandler(origins []string, h http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins: origins,
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: slices.Concat(connectcors.AllowedHeaders(), []string{"Authorization"}),
		ExposedHeaders: connectcors.ExposedHeaders(),
	}).Handler(h)
}

// Every RPC this service exposes takes either Empty or a two-field message, so
// this is generous headroom, not a real limit.
const maxRequestBytes = 4096

const defaultCompressMinBytes = 256

func recoverPanic(_ context.Context, _ connect.Spec, _ http.Header, r any) error {
	return connect.NewError(connect.CodeInternal, fmt.Errorf("panic: %v", r))
}

func handlerOpts(token string, compressMinBytes int) []connect.HandlerOption {
	if compressMinBytes <= 0 {
		compressMinBytes = defaultCompressMinBytes
	}
	opts := []connect.HandlerOption{
		connect.WithCompression(rpc.CompressionZstd, rpc.NewZstdDecompressor, rpc.NewZstdCompressor),
		connect.WithCompressMinBytes(compressMinBytes),
		connect.WithReadMaxBytes(maxRequestBytes),
		connect.WithRecover(recoverPanic),
	}
	if token != "" {
		opts = append(opts, connect.WithInterceptors(rpc.NewTokenInterceptor(token)))
	}
	return opts
}
