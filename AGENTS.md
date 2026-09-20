# AI Project Context: pprof-viz-collector

Engineering conventions for working in this repository. See `README.md` for what the
library does, its public API, and usage examples — that is not repeated here.

## Directory Structure

- **Package root (`pprofvizcollector`)**: the public API surface, split by concern rather than
  crammed into one file. Each exported entry point owns its config struct in the same file.
    - **`bind.go`**: `BindConfig` and `Bind()`.
    - **`new.go`**: `NewConfig` and `New()` — the h2c/TLS/HTTP3 listener branching. `New`'s `port`
      parameter is `uint16`, not `int`: the type itself rules out negative or out-of-range values
      instead of leaving that to a runtime check.
    - **`handler.go`**: the ConnectRPC handler-option/CORS wiring shared by `Bind` and `New`
      (`newCORSHandler`, `handlerOpts`, `recoverPanic`, `maxRequestBytes`, `defaultCompressMinBytes`).
    - **`server.go`**: the `Server` type returned by `New()` (`Addr`, `Shutdown`) and the shared
      `newHTTPServer` helper.
- **`/internal/`**: Private implementation, split into subpackages so a symbol's import path says
  where it comes from instead of everything being flatly `internal`.
    - **`internal/rpc/`**: the ConnectRPC transport layer.
        - **`server.go`**: `CollectorServiceHandler` (`Ping`, `SnapshotAll`, `StreamAll`,
          `GetMetadata`) — thin glue that delegates metric work to `internal/collect`.
        - **`interceptor.go`**: `NewTokenInterceptor(token)`, the bearer-token interceptor.
        - **`compression.go`**: custom `zstd` compressor/decompressor for ConnectRPC (256-byte
          threshold).
    - **`internal/collect/`**: runtime-metrics collection, independent of ConnectRPC.
        - **`collector.go`**: `Collector` — single-pass collection with double-buffering and
          differential updates; `Snapshot()` for pooled one-shot reads.
        - **`metadata.go`**: `Metadata()`, the cached histogram bucket-boundary response.
        - **`metrics.go`**: `runtime/metrics` interfacing (`readAllSamples`).
        - **`metrics_gen.go` / `collect_gen.go`**: generated from `gen/table.go` — never edit by
          hand, run `just gen`.
        - **`utils.go`**: conversion helpers for `metrics.Value`.
        - **`gen/`**: the code generator (`table.go` is the single source of truth for every
          collected metric).
    - **`internal/proto/`**: Generated Go code shared by both subpackages. Never edit by hand — use
      `just gen`.
- **`/proto/`**: Source `.proto` files defining the service contract.

## Commands

All maintenance tasks run through the `Justfile`.

| Task | Command | Description |
| :--- | :--- | :--- |
| **Test** | `just test` | Unit tests with the race detector. |
| **Coverage** | `just cover` | HTML coverage report. |
| **Lint** | `just lint` | `golangci-lint`, `go vet`, `modernize`. |
| **Security** | `just vuln` | `staticcheck`, `govulncheck`, `gosec`. |
| **Generate** | `just gen` | Regenerates Protobuf/ConnectRPC code via `buf generate`. |

## Coding Style

- **Idiomatic Go**: standard formatting and naming conventions.
- **Unused parameters**: omit parameter names entirely — do not use `_`. Match existing handler
  signatures (e.g. `func (h CollectorServiceHandler) Ping(context.Context, *emptypb.Empty)`).
- **Memory**: order struct fields by size (descending) to minimize padding; use pre-allocated
  slices and `Collector` to avoid heap allocations during collection.
- **Metric lookup**: each `Metric` constant is a direct index into the pre-allocated
  `[]metrics.Sample` slice — an O(1) array access, never a map.
- **Pointers**: use `new(value)` for single-line pointer creation for `optional` fields.
- **Documentation**: all exported functions and types need descriptive comments.

## Testing Practices

- **Parallelism**: use `t.Parallel()` for all safe tests. With `t.Run`, call `t.Parallel()`
  *inside* the subtest function.
- **Test helpers**: use `newTestClient(t *testing.T)` (in `internal/rpc/server_test.go`) for
  isolated ConnectRPC test servers.
- **`defer` vs `t.Cleanup`**: when a parent test spawns parallel subtests, `defer` fires at parent
  return, before the subtests finish. Always use `t.Cleanup` for resources shared with parallel
  subtests — this is why `newTestClient` uses `t.Cleanup(server.Close)`.
- **Metric coverage guard**: `TestMetricCoverage` (`internal/collect/metrics_test.go`) asserts
  every `runtime/metrics.All()` entry is either collected (in `AllMetricNames`) or listed in the
  documented `intentionallyOmitted` map. A new Go version's added metrics will fail this test
  until they are wired up or explicitly omitted with a reason.

## Collector Design

- **`Collector.Collect()`** always returns differential updates. On a fresh collector the
  `KindBad` branch in `hasChanged` treats all non-zero values as "changed from zero", which is why
  no explicit `forceFull` flag exists.
- **Histogram buckets are never included in `Collect()`**. Bucket boundaries are static and sent
  once via `GetMetadata`; only `UpdatedCounts` (sparse delta map) goes on the wire.
- **One-shot snapshots use a pooled collector**: `collect.Snapshot()` recycles `Collector`s
  through a `sync.Pool` and calls `resetBaseline()` before each read, reproducing fresh-collector
  full-state semantics without re-allocating sample slices. `StreamAll` keeps its own persistent
  collector for genuine differential diffing — never route it through the pool.
- **`readAllSamples` panics on a wrong-sized buffer**: the slice must have exactly
  `len(AllMetricNames)` entries. A mismatch is a programming error inside the package, and panicking
  beats silently emitting zeroed metrics.

## Server Construction

- Every TCP server is built by `newHTTPServer` with hardened timeouts. `WriteTimeout` is
  deliberately omitted so long-lived `StreamAll` streams are not terminated. All TLS configs set
  `MinVersion: tls.VersionTLS12`.
- With `HTTP3: true`, TCP and QUIC run as two separate servers; the TCP handler injects `Alt-Svc`
  via `http3.Server.SetQUICHeaders`. The package-level `http3.ListenAndServeTLS` helper is
  deliberately **not** used — it builds a default `http.Server` with no timeouts and no TLS
  `MinVersion`.
- CORS placement differs by entry point: `Bind` wraps the ConnectRPC handler before registering it
  on the external mux (scoped to the agent's path only); `New` wraps the entire mux before it
  reaches any server branch (h2c, TLS, or HTTP3). `AllowedHeaders` is
  `connectcors.AllowedHeaders()` plus `"Authorization"`.

## Protobuf & Wire Format

- **Auth belongs in HTTP headers, not proto messages.** Never add token fields to proto definitions.
- **Wire efficiency**: reserve tags 1-15 for frequently updated fields.
- **Type selection**: `uint32` for gauges (counts), `double` for cumulative time counters.
- **Differential updates**: use `optional` scalar fields so unchanged values can be omitted; treat
  initial stream messages as a delta from the zero state.
- **Sparse histograms**: `map<uint32, uint64> updated_counts` carries only changed buckets.
- **No bucket exclusion knob**: `StreamConfigRequest` has no `exclude_histogram_buckets` field —
  exclusion is unconditional.
- **Code generation**: all `.pb.go` and `.connect.go` files come from `just gen`. Never hand-edit.

## AI Interaction Guidelines

- **Verify logic**: AI-assisted contributions are fine, but all code must be verified for
  correctness. Never trust AI output blindly.
- **No "vibe-coding"**: every change needs a technical rationale grounded in Go runtime behavior or
  ConnectRPC requirements.
- **Surgical edits**: prefer minimal, high-impact changes that preserve existing architectural patterns.
