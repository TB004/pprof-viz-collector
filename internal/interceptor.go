// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect"
)

const bearerPrefix = "Bearer "

type tokenInterceptor struct {
	// wantHash is the SHA-256 digest of the expected token, precomputed once
	// so check doesn't rebuild and re-hash it on every RPC.
	wantHash [sha256.Size]byte
}

// NewTokenInterceptor returns a ConnectRPC interceptor that enforces bearer token
// authentication on every RPC.
func NewTokenInterceptor(token string) connect.Interceptor {
	return tokenInterceptor{wantHash: sha256.Sum256([]byte(token))}
}

func (ti tokenInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if err := ti.check(req.Header()); err != nil {
			return nil, err
		}
		return next(ctx, req)
	}
}

func (ti tokenInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (ti tokenInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		if err := ti.check(conn.RequestHeader()); err != nil {
			return err
		}
		return next(ctx, conn)
	}
}

func (ti tokenInterceptor) check(h http.Header) error {
	got := h.Get("Authorization")
	if len(got) < len(bearerPrefix) || !strings.EqualFold(got[:len(bearerPrefix)], bearerPrefix) {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("invalid or missing token"))
	}

	gotHash := sha256.Sum256([]byte(got[len(bearerPrefix):]))
	if subtle.ConstantTimeCompare(gotHash[:], ti.wantHash[:]) != 1 {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("invalid or missing token"))
	}
	return nil
}
