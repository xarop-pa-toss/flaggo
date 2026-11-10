---
type: runtime-lifecycle
title: Runtime Lifecycle
description: Startup, degraded mode, graceful shutdown on SIGTERM, and resource cleanup for Flaggo services.
tags: [runtime, lifecycle, shutdown, graceful]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-09T18:31:04.450Z
sources:
  - id: openwiki-source-693bc5f0b4443bd18cfa76b4
    resource: repo://cmd/flaggo/main.go
generated: { by: "pi", at: "2026-11-09T18:31:04.450Z" }
---

# Runtime Lifecycle

## Responsibility

`cmd/flaggo/main.go` (`runServer`) and `cmd/flaggo-web/main.go` own process
startup, signal handling, graceful shutdown, and resource cleanup.

## Backend startup

1. Load config; panic on invalid config.
2. Create logger, metrics, health handler.
3. Select store: memory (ready), postgres (connect + migrate; degraded on
   failure), or neither (degraded).
4. Build router; if store usable, attach `FlagsHandler`.
5. Start HTTP server in a goroutine.
6. Block on SIGINT/SIGTERM.

## Degraded mode

If Postgres is unavailable, the backend logs the error, sets readiness false,
and continues serving health/metrics only. See
`openwiki/operations/health-and-readiness.md`.

## Graceful shutdown

On signal: log, set readiness false, call `server.Shutdown(ctx)` with
`ShutdownTimeout` (default 10s), then close the store. In-flight requests get
the timeout budget to complete.

## Frontend

Same pattern: start server, block on signal, graceful shutdown. Readiness
depends on backend probe.

## Related

- `openwiki/configuration/environment.md`
- `openwiki/operations/health-and-readiness.md`
- `openwiki/storage/storage-layer.md`
