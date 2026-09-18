// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenInterceptor_Check(t *testing.T) {
	ti := NewTokenInterceptor("secret").(tokenInterceptor)

	header := func(v string) http.Header {
		h := http.Header{}
		if v != "" {
			h.Set("Authorization", v)
		}
		return h
	}

	tests := []struct {
		name    string
		auth    string
		wantErr bool
	}{
		{"valid token", "Bearer secret", false},
		{"case-insensitive scheme per RFC 7235", "bearer secret", false},
		{"mixed-case scheme", "BEARER secret", false},
		{"missing header", "", true},
		{"wrong token", "Bearer wrong", true},
		{"wrong token, different length", "Bearer nope", true},
		{"missing scheme", "secret", true},
		{"empty token", "Bearer ", true},
		{"trailing whitespace", "Bearer secret ", true},
		{"prefix of scheme only", "Bear", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ti.check(header(tt.auth))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
