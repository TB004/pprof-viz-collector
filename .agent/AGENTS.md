# Gemini Project Context: pprof-viz-agent

This document serves as the foundational mandate for all AI interactions within the `pprof-viz-agent` repository. It outlines the project's purpose, architecture, and engineering standards to ensure consistency and technical integrity.

## 🚀 Project Overview
`pprof-viz-agent` is a lightweight, high-fidelity observability agent for Go applications. It exposes comprehensive Go runtime metrics (CPU, Memory, GC, Scheduler) via a performant ConnectRPC API, optimized for real-time streaming and visualization.

### Tech Stack
- **Language**: Go 1.26.1+
- **Protocol**: ConnectRPC (Protobuf over HTTP/2)
- **Compression**: `zstd` (with a 256-byte threshold)
- **Task Runner**: `just`

## 🏗️ Architecture & Directory Structure
The project follows a refined Go library structure focused on a minimal public API surface and optimized internal logic.

- **`/agent.go`**: The primary entry point. Contains the `Bind()` (for existing servers) and `New()` (for standalone execution) functions.
- **`/internal/`**: Private implementation details.
    - **`server.go`**: Implements the ConnectRPC service handlers.
    - **`collector.go`**: Manages the `SampleCollector` for efficient, single-pass metric collection. Supports differential updates to minimize bandwidth.
    - **`metrics.go` / `paths.go`**: Logic for interfacing with `runtime/metrics` and pre-calculating metric indices.
    - **`utils.go`**: Low-level conversion helpers for `metrics.Value`.
    - **`compression.go`**: Custom `zstd` compressor/decompressor for ConnectRPC.
    - **`proto/`**: Generated Go code from Protobuf definitions.
- **`/pkg/models/`**: Shared, public data structures (e.g., `Histogram`).
- **`/proto/`**: Source `.proto` files defining the service contract.

## 🛠️ Development & Commands
All maintenance tasks should be performed using the `Justfile`.

| Task | Command | Description |
| :--- | :--- | :--- |
| **Test** | `just test` | Runs unit tests with the race detector. |
| **Coverage** | `just cover` | Generates and opens an HTML coverage report. |
| **Lint** | `just lint` | Runs `golangci-lint`, `go vet`, and `modernize`. |
| **Security** | `just vuln` | Runs `staticcheck`, `govulncheck`, and `gosec`. |
| **Generate** | `buf generate` | Regenerates Protobuf code (if `buf` is installed). |

## 📏 Engineering Standards

### Coding Style
- **Idiomatic Go**: Adhere strictly to standard Go formatting and naming conventions.
- **Memory Optimization**: 
    - Order struct fields by size (descending) to minimize padding.
    - Use pre-allocated slices and `SampleCollector` to minimize heap allocations during collection.
- **Performance**: 
    - Replaced map-based metric lookups with a pre-allocated `[]metrics.Sample` slice.
    - Use `GetIndex(name)` for O(1) index retrieval, initialized via `init()`.
    - Use Go 1.26's `new(value)` for efficient, single-line pointer creation for `optional` fields.
- **Documentation**: All exported functions and types must have descriptive comments.

### Testing Practices
- **Parallelism**: Use `t.Parallel()` for all safe tests.
- **Subtest Pattern**: When using `t.Run`, always call `t.Parallel()` **inside** the subtest function to ensure concurrent execution.
- **Network Tests**: Use `httptest.NewServer` inside individual tests to avoid port conflicts.

### Protobuf & Network
- **Wire Efficiency**: Prioritize tags 1-15 for frequently updated fields.
- **Type Selection**: Use `uint32` for gauges (counts) and `double` for cumulative time counters to maintain precision.
- **Differential Updates**: 
    - Use `optional` scalar fields to support differential updates (only sending changed values) in streaming RPCs.
    - Treat initial stream messages as a delta from the default (zero) state to omit 0-value fields.
- **Sparse Histograms**: Use `map<uint32, uint64> updated_counts` to transmit only buckets that have changed since the last sample, including in the initial synchronization.
- **Compression**: The agent uses a 256-byte threshold for `zstd` compression to balance CPU usage and bandwidth.

## 🤖 AI Interaction Guidelines
- **Verify Logic**: Contributions can be AI-assisted, but all code must be verified for correctness. Never trust AI output blindly.
- **No "Vibe-Coding"**: Every change must have a clear technical rationale grounded in the Go runtime's behavior or ConnectRPC's requirements.
- **Surgical Edits**: When modifying logic, prioritize minimal, high-impact changes that maintain the existing architectural patterns.
