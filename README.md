# pprof-viz-agent

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Real-time, high-fidelity Go runtime telemetry delivered via ConnectRPC.

## Overview

`pprof-viz-agent` is a lightweight observability agent designed to be embedded in Go applications. It exposes the comprehensive `runtime/metrics` suite—providing much deeper insight into GC behavior, scheduler latencies, and memory classes than traditional pprof—via a high-performance ConnectRPC API.

While standard `pprof` is indispensable for deep-dive analysis and stack-trace inspection, `pprof-viz-agent` is built for **continuous, high-resolution monitoring**. It enables external visualization tools to stream the internal state of the Go runtime with minimal overhead and maximum detail.

## Key Features

- **High Fidelity**: Exposes the full range of `runtime/metrics` (CPU, Memory, GC, Scheduler, etc.).
- **Modern Protocol**: Built on ConnectRPC for type-safe, performant streaming over HTTP/2 and HTTP/3.
- **Low Overhead**: Optimized for zero-alloc collection loops using a single-pass `runtime/metrics.Read` approach and O(1) metric indexing.
- **Efficient**: Uses `zstd` compression and intelligent field alignment to minimize network and memory footprint.
- **Easy Integration**: Minimal-config setup for both existing web services and standalone CLI tools.
- **Secure**: Optional bearer token authentication to restrict access to authorized clients.

## Installation

```bash
go get github.com/TB004/pprof-viz-agent
```

## Bandwidth Optimization

`pprof-viz-agent` employs several sophisticated strategies to minimize network usage during long-running streams:

1.  **Differential Updates**: The agent tracks the state of all metrics and only transmits values that have changed since the last sample. Unchanged fields are omitted from the wire entirely, drastically reducing payload size for stable applications.
2.  **Zero-Value Omission**: The very first message of a stream is treated as a delta from the default state. Any metric that is zero (like many `goDebug` events) is omitted from the initial synchronization, ensuring the baseline is as small as possible.
3.  **Sparse Histogram Updates**: Instead of sending full arrays of histogram counts, the agent uses a map-based "sparse update" approach. Only the specific buckets that have changed are transmitted. This is particularly effective for large histograms where only a few buckets (e.g., specific latency ranges) are active.
4.  **Metadata Separation**: Histogram bucket boundaries are static and bandwidth-heavy. The agent enforces a strict "Metadata + Data" pattern:
    *   Call `GetMetadata` once to retrieve the static bucket definitions.
    *   `StreamAll` always streams only the sparse changed counts — bucket boundaries are never re-sent on the wire.

This combination ensures that the agent consumes negligible bandwidth when the runtime state is stable, while still providing high-fidelity data during activity spikes.

## Usage

The library provides two primary entry points depending on your application's architecture. Both accept a `Config` struct.

### Configuration

```go
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
```

### 1. Using `Bind()` (Existing HTTP Server)

If your project already runs an HTTP server (e.g., a REST API or gRPC-web service), use `Bind` to register the agent's handlers to your existing `http.ServeMux`.

```go
package main

import (
    "net/http"

    pprofvizagent "github.com/TB004/pprof-viz-agent"
)

func main() {
    mux := http.NewServeMux()

    // Bind the agent to your existing mux, optionally with a token
    pprofvizagent.Bind(mux, pprofvizagent.Config{
        Token: "your-secret-token",
    })

    // Register your other handlers
    mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    http.ListenAndServe(":8080", mux)
}
```

### 2. Using `New()` (Standalone/CLI Tools)

If your project is a long-running process without an existing HTTP server (e.g., a background worker or CLI tool), use `New` to start a dedicated telemetry server.

**Cleartext (H2C):**
```go
go func() {
    if err := pprofvizagent.New(9999, pprofvizagent.Config{}); err != nil {
        log.Fatalf("failed to start agent: %v", err)
    }
}()
```

**TLS with token auth:**
```go
go func() {
    if err := pprofvizagent.New(9999, pprofvizagent.Config{
        Token:    "your-secret-token",
        CertFile: "/path/to/cert.pem",
        KeyFile:  "/path/to/key.pem",
    }); err != nil {
        log.Fatalf("failed to start agent: %v", err)
    }
}()
```

**TLS + HTTP/3:**
```go
go func() {
    if err := pprofvizagent.New(9999, pprofvizagent.Config{
        Token:    "your-secret-token",
        CertFile: "/path/to/cert.pem",
        KeyFile:  "/path/to/key.pem",
        HTTP3:    true,
    }); err != nil {
        log.Fatalf("failed to start agent: %v", err)
    }
}()
```

### Authentication

When `Config.Token` is set, every RPC call must include the token as a bearer credential:

```
Authorization: Bearer your-secret-token
```

Requests without the header or with a wrong token are rejected with HTTP 401 Unauthenticated. This applies to all RPCs including `Ping` and `GetMetadata`.

## API

| RPC | Type | Description |
| :--- | :--- | :--- |
| `Ping` | Unary | Health check — returns OK if the agent is reachable. |
| `SnapshotAll` | Unary | One-shot capture of all non-zero runtime metrics. |
| `StreamAll` | Server-streaming | Streams differential metric updates at a client-specified interval. |
| `GetMetadata` | Unary | Returns static histogram bucket boundaries. Call once before streaming. |

## Pprof vs. Pprof-Viz

| Feature | pprof | pprof-viz-agent |
| :--- | :---: | :---: |
| **Primary Goal** | Deep profiling / Bottleneck analysis | Continuous Runtime Monitoring |
| **Data Source** | Sampling & Stack Traces | Native `runtime/metrics` |
| **Information Depth** | Specific function calls | System-wide runtime internals |
| **Ease of Use** | Requires manual captures/analysis | Automated, streaming-first |
| **Stack Traces** | Yes | No |

## Visualization

This agent is designed to work seamlessly with the upcoming **Pprof-Viz Web App**, a free-to-use visualization platform that provides real-time dashboards for your Go applications. Stay tuned for the release!

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
