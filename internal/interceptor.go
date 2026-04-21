// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
)

type tokenInterceptor struct{ token string }

// NewTokenInterceptor returns a ConnectRPC interceptor that enforces bearer token
// authentication on every RPC. Requests must carry Authorization: Bearer <token>;
// missing or mismatched tokens are rejected with CodeUnauthenticated.
func NewTokenInterceptor(token string) connect.Interceptor {
	return tokenInterceptor{token: token}
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
	if h.Get("Authorization") != "Bearer "+ti.token {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("invalid or missing token"))
	}
	return nil
}
