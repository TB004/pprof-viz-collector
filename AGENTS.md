# AI Project Context: pprof-viz-collector

This document serves as the foundational mandate for all AI interactions within the `pprof-viz-collector` repository. It outlines the project's purpose, architecture, and engineering standards to ensure consistency and technical integrity.

## 🚀 Project Overview
`pprof-viz-collector` is a lightweight, high-fidelity observability agent for Go applications. It exposes comprehensive Go runtime metrics (CPU, Memory, GC, Scheduler) via a performant ConnectRPC API, optimized for real-time streaming and visualization.

### Tech Stack
- **Language**: Go 1.27+
- **Protocol**: ConnectRPC (Protobuf over HTTP/2 and HTTP/3)
- **Compression**: `zstd` (with a 256-byte threshold)
- **HTTP/3**: `github.com/quic-go/quic-go/http3`
- **CORS**: `connectrpc.com/cors` (header constants) + `github.com/rs/cors` (middleware)
- **Task Runner**: `just`

## 🏗️ Architecture & Directory Structure
The project follows a refined Go library structure focused on a minimal public API surface and optimized internal logic.

- **`/lib.go`**: The primary entry point. Contains the `Config` struct, `Bind()` (for existing servers), and `New()` (for standalone execution) functions, plus public `Snapshot*()` helpers.
- **`/internal/`**: Private implementation details.
    - **`server.go`**: Implements the ConnectRPC service handlers (`Ping`, `SnapshotAll`, `StreamAll`, `GetMetadata`).
    - **`collector.go`**: Manages the `SampleCollector` for efficient, single-pass metric collection with double-buffering and differential updates.
    - **`interceptor.go`**: `NewTokenInterceptor(token)` — a ConnectRPC interceptor enforcing `Authorization: Bearer <token>` on all RPCs.
    - **`metrics.go` / `paths.go`**: Logic for interfacing with `runtime/metrics` and pre-calculating metric indices.
    - **`utils.go`**: Low-level conversion helpers for `metrics.Value`.
    - **`compression.go`**: Custom `zstd` compressor/decompressor for ConnectRPC.
    - **`proto/`**: Generated Go code from Protobuf definitions. Never edit by hand — use `just gen`.
- **`/pkg/models/`**: Shared, public data structures (e.g., `Histogram`).
- **`/proto/`**: Source `.proto` files defining the service contract.

## 🔌 Public API

### `Config` struct
```go
type Config struct {
    Token          string   // bearer token for auth; empty = no auth
    AllowedOrigins []string // CORS origins; nil/empty disables CORS
    CertFile       string   // TLS cert path; enables TLS when set (New only)
    KeyFile        string   // TLS key path; paired with CertFile (New only)
    HTTP3          bool     // serve HTTP/3 over QUIC in parallel; requires CertFile+KeyFile (New only)
    Host           string   // bind interface; empty = "localhost" (New only)
}
```

### Entry points
- **`Bind(mux *http.ServeMux, cfg Config)`**: Registers the agent onto an existing mux. `cfg.Token` and `cfg.AllowedOrigins` apply; TLS/HTTP3 are server-level concerns managed externally.
- **`New(port int, cfg Config) error`**: Starts a dedicated server bound to `cfg.Host` (default `localhost`).
  - No TLS → cleartext HTTP/2 (h2c) / HTTP/1.1 (browsers cannot talk h2c).
  - `CertFile`+`KeyFile` set → TLS + HTTP/2 (minimum TLS 1.2).
  - `HTTP3: true` → HTTP/2 over TLS/TCP and HTTP/3 over QUIC are served in parallel on two separate servers; the TCP handler injects `Alt-Svc` headers via `http3.Server.SetQUICHeaders` so clients can upgrade. The package-level `http3.ListenAndServeTLS` helper is deliberately **not** used because it builds a default `http.Server` with no timeouts and no TLS `MinVersion`.
  - Every TCP server is built by `newHTTPServer` with hardened timeouts (`ReadHeaderTimeout`, `ReadTimeout`, `IdleTimeout`). `WriteTimeout` is intentionally omitted so the long-lived `StreamAll` stream is not terminated. All TLS configs set `MinVersion: tls.VersionTLS12`.

### RPC methods
| RPC | Type | Description |
| :--- | :--- | :--- |
| `Ping` | Unary | No-op health check. Returns OK if the agent is reachable. |
| `SnapshotAll` | Unary | One-shot capture of all non-zero runtime metrics. |
| `StreamAll(StreamConfigRequest)` | Server-streaming | Differential metric updates at a client-specified interval; emits an immediate full baseline frame first. |
| `GetMetadata` | Unary | Static histogram bucket boundaries. Call once before streaming. |

## 🛠️ Development & Commands
All maintenance tasks should be performed using the `Justfile`.

| Task | Command | Description |
| :--- | :--- | :--- |
| **Test** | `just test` | Runs unit tests with the race detector. |
| **Coverage** | `just cover` | Generates and opens an HTML coverage report. |
| **Lint** | `just lint` | Runs `golangci-lint`, `go vet`, and `modernize`. |
| **Security** | `just vuln` | Runs `staticcheck`, `govulncheck`, and `gosec`. |
| **Generate** | `just gen` | Regenerates Protobuf/ConnectRPC code via `buf generate`. |

## 📏 Engineering Standards

### Coding Style
- **Idiomatic Go**: Adhere strictly to standard Go formatting and naming conventions.
- **Unused parameters**: Omit parameter names entirely when unused — do not use `_`. Match the style of existing handler signatures (e.g., `func (h CollectorServiceHandler) Ping(context.Context, *emptypb.Empty)`).
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
- **Test helpers**: Use `newTestClient(t *testing.T)` (defined in `internal/server_test.go`) to create isolated ConnectRPC test servers. It uses `t.Cleanup(server.Close)` — never `defer server.Close()` — so the server lifetime outlives parallel subtests.
- **`defer` vs `t.Cleanup`**: When a parent test spawns parallel subtests, `defer` fires at parent function return (before subtests complete). Always use `t.Cleanup` for resources shared with parallel subtests.
- **Metric coverage guard**: `TestMetricCoverage` (in `internal/metrics_test.go`) asserts every `runtime/metrics.All()` entry is either collected (in `AllMetricNames`) or listed in the documented `intentionallyOmitted` map. When adding or dropping a collected metric, update one of the two — a new Go version's added metrics will fail this test until they are wired up or explicitly omitted with a reason.

### Collector design
- **`SampleCollector.Collect()`** (no parameters): Always returns differential updates. On a fresh collector (first call after `NewSampleCollector()`), the `KindBad` branch in `hasChanged` correctly treats all non-zero values as "changed from zero", making an explicit `forceFull` flag unnecessary.
- **Histogram buckets are never included in `Collect()`**: Bucket boundaries are static and sent once via `GetMetadata`. `StreamAll` and `SnapshotAll` always omit `Buckets`; only `UpdatedCounts` (sparse delta map) is transmitted.
- **One-shot snapshots use a pooled collector**: `internal.Snapshot()` recycles `SampleCollector`s through a `sync.Pool` and calls `resetBaseline()` (zeroing the buffer that the next `Read` swaps in as the diff baseline) before each read, so repeated `SnapshotAll`/`Snapshot*()` calls reproduce fresh-collector full-state semantics without re-allocating the sample slices. `StreamAll` keeps its own persistent collector across cycles for genuine differential diffing — never route it through the pool.
- **`StreamAll` sends an immediate first frame**: The full baseline frame is sent at connect time (t=0), counting as the first cycle, rather than after one interval. Frame semantics are unchanged (the first frame was always full); only the timing moved earlier.
- **`ReadAllSamples` panics on a wrong-sized buffer**: The sample slice must have exactly `len(AllMetricNames)` entries. A mismatch is a programming error within the package and panics rather than silently emitting zeroed metrics.

### Authentication
- Token auth is implemented as a ConnectRPC interceptor in `internal/interceptor.go` via `NewTokenInterceptor(token string) connect.Interceptor`.
- Checks `Authorization: Bearer <token>` on every RPC (unary and streaming) using a constant-time comparison (`crypto/subtle.ConstantTimeCompare`) to avoid leaking the token through response timing.
- Returns `CodeUnauthenticated` (HTTP 401) on mismatch.
- **Auth belongs in HTTP headers, not proto messages.** Never add token fields to proto message definitions.

### CORS
- CORS is applied via `newCORSHandler(origins []string, h http.Handler) http.Handler` in `agent.go`.
- Uses `connectrpc.com/cors` for the ConnectRPC-specific header sets and `github.com/rs/cors` for the middleware.
- `AllowedHeaders` = `connectcors.AllowedHeaders()` + `"Authorization"` (required for token auth).
- In `Bind`, CORS wraps the ConnectRPC handler before it is registered on the external mux — scoped to the agent's path only.
- In `New`, CORS wraps the entire mux before it is passed to any server branch (h2c, TLS, or HTTP3).
- An empty `AllowedOrigins` slice skips the middleware entirely — no CORS headers are ever set.

### Protobuf & Network
- **Wire Efficiency**: Prioritize tags 1-15 for frequently updated fields.
- **Type Selection**: Use `uint32` for gauges (counts) and `double` for cumulative time counters to maintain precision.
- **Differential Updates**:
    - Use `optional` scalar fields to support differential updates (only sending changed values) in streaming RPCs.
    - Treat initial stream messages as a delta from the default (zero) state to omit 0-value fields.
- **Sparse Histograms**: Use `map<uint32, uint64> updated_counts` to transmit only buckets that have changed since the last sample.
- **No `Buckets` in streaming**: `StreamConfigRequest` has no `exclude_histogram_buckets` field — exclusion is unconditional. Clients call `GetMetadata` once for bucket boundaries.
- **Compression**: The agent uses a 256-byte threshold for `zstd` compression to balance CPU usage and bandwidth.
- **Code generation**: All `.pb.go` and `.connect.go` files are generated by `just gen` (`buf generate`). Never edit them by hand.

## 🤖 AI Interaction Guidelines
- **Verify Logic**: Contributions can be AI-assisted, but all code must be verified for correctness. Never trust AI output blindly.
- **No "Vibe-Coding"**: Every change must have a clear technical rationale grounded in the Go runtime's behavior or ConnectRPC's requirements.
- **Surgical Edits**: When modifying logic, prioritize minimal, high-impact changes that maintain the existing architectural patterns.
