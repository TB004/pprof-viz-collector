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
	// connect handles pooling internally when using WithCompression
	w, err := zstd.NewWriter(nil)
	if err != nil {
		panic(err)
	}
	return w
}

func NewZstdDecompressor() connect.Decompressor {
	r, err := zstd.NewReader(nil)
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

// Close does nothing here but needed to fullfil connect.Decompressor interface
func (d *zstdDecompressor) Close() error {
	return nil
}
