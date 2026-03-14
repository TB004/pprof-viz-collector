# pprof-viz-agent

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Real-time, high-fidelity Go runtime telemetry delivered via ConnectRPC.

## Overview

`pprof-viz-agent` is a lightweight observability agent designed to be embedded in Go applications. It exposes the comprehensive `runtime/metrics` suite—providing much deeper insight into GC behavior, scheduler latencies, and memory classes than traditional pprof—via a high-performance ConnectRPC API.

While standard `pprof` is indispensable for deep-dive analysis and stack-trace inspection, `pprof-viz-agent` is built for **continuous, high-resolution monitoring**. It enables external visualization tools to stream the internal state of the Go runtime with minimal overhead and maximum detail.

## Key Features

- **High Fidelity**: Exposes the full range of `runtime/metrics` (CPU, Memory, GC, Scheduler, etc.).
- **Modern Protocol**: Built on ConnectRPC for type-safe, performant streaming over HTTP/2.
- **Optimized**: Uses `zstd` compression and intelligent field alignment to minimize network and memory footprint.
- **Easy Integration**: Zero-config setup for both existing web services and standalone CLI tools.

## Installation

```bash
go get github.com/TB004/pprof-viz-agent
```

## Usage

The library provides two primary entry points depending on your application's architecture.

### 1. Using `Bind()` (Existing HTTP Server)

If your project already runs an HTTP server (e.g., a REST API or gRPC-web service), use `Bind` to register the agent's handlers to your existing `http.ServeMux`.

```go
package main

import (
    "net/http"
    "github.com/TB004/pprof-viz-agent"
)

func main() {
    mux := http.NewServeMux()
    
    // Bind the agent to your existing mux
    pprofvizagent.Bind(mux)
    
    // Register your other handlers
    mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    http.ListenAndServe(":8080", mux)
}
```

### 2. Using `New()` (Standalone/CLI Tools)

If your project is a long-running process without an existing HTTP server (e.g., a background worker or CLI tool), use `New` to start a dedicated telemetry server.

```go
package main

import (
    "log"
    "github.com/TB004/pprof-viz-agent"
)

func main() {
    // Start a dedicated telemetry server on port 9999
    // Pass nil for protocols and TLS to use standard HTTP/1.1 or H2C defaults
    go func() {
        if err := pprofvizagent.New(9999, nil, nil); err != nil {
            log.Fatalf("failed to start agent: %v", err)
        }
    }()

    // Your application logic here
    select {} 
}
```

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
