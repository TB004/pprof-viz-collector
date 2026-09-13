// SPDX-License-Identifier: MIT
// Copyright (C) 2026 Tizian Bahr

package internal

import (
	"io"

	"connectrpc.com/connect"
	"github.com/klauspost/compress/zstd"
)

const CompressionZstd = "zstd"

func NewZstdCompressor() connect.Compressor {
	// NewWriter only returns an error for invalid options; the options here are
	// always valid, so the error is unreachable.
	w, err := zstd.NewWriter(nil,
		zstd.WithEncoderConcurrency(1),
		zstd.WithEncoderLevel(zstd.SpeedFastest),
	)
	if err != nil {
		panic(err)
	}
	return w
}

func NewZstdDecompressor() connect.Decompressor {
	// Same as above, the error is unreachable.
	r, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(1))
	if err != nil {
		panic(err)
	}
	return &zstdDecompressor{r}
}

type zstdDecompressor struct {
	*zstd.Decoder
}

// Reset resets the decoder with a new reader.
func (d *zstdDecompressor) Reset(r io.Reader) error {
	return d.Decoder.Reset(r)
}

// Close does nothing here but needed to fulfil connect.Decompressor interface.
// Decoder.Close() is intentionally not called: with concurrency forced to 1
// (see NewZstdDecompressor) it stops no goroutines, and connect reuses this
// decompressor across the connection's lifetime via Reset, so closing it here
// would make it unusable for the next message.
func (d *zstdDecompressor) Close() error {
	return nil
}
