---
type: api-reference
title: HTTP Middleware
description: The middleware chain (request ID, recovery, logging, metrics) applied to every request in both backend and frontend, including ordering and correlation.
tags: [middleware, http, observability, request-id]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-09T18:31:04.450Z
sources:
  - id: openwiki-source-e689a4a46f2ebef989178800
    resource: repo://internal/http/router.go
  - id: openwiki-source-3ae86643bd52d54c1a51eeef
    resource: repo://internal/web/server.go
generated: { by: "pi", at: "2026-11-09T18:31:04.450Z" }
---

# HTTP Middleware

## Responsibility

`internal/http/middleware/middleware.go` provides four cross-cutting concerns applied
to every request in both the control plane backend and the frontend gateway:
request identity, panic recovery, structured access logging, and Prometheus
metrics. Ordering is defined in `internal/http/router.go` (backend) and
`internal/web/server.go` (frontend).

## Chain order (outermost → innermost)

1. `RequestID` — assigns/propagates `X-Request-ID`.
2. `Recovery` — catches panics, logs them, returns 500.
3. `Logging` — emits one slog line per request with method, path, status, duration, request id, remote.
4. `Metrics` — increments `flaggo_http_requests_total` and observes
   `flaggo_http_request_duration_seconds`.
5. mux (handlers).

## Request ID propagation

- `RequestID` reads `X-Request-ID` from the incoming request; if absent it
  generates 16 random hex bytes.
- The id is stored in the request context (`requestIDKey`) and echoed on the
  response header.
- `RequestIDFromContext` exposes it to handlers and downstream calls.
- The frontend gateway sets `X-Request-ID` on the proxied request to the
  backend, so a dashboard action correlates across both services. The backend's
  echoed copy is stripped by `ModifyResponse` to avoid duplicate headers.

## Recovery

`Recovery` wraps the handler in a deferred recover. On panic it logs the panic
value, path, and request id, then writes `500 Internal Server Error`. This
prevents a single handler panic from crashing the process.

## Logging

`Logging` wraps the response writer to capture the status code, records start
time, and emits one `slog.Info("request", ...)` line after the handler returns,
including duration in milliseconds.

## Metrics

`Metrics` records per-request counters and histograms using
`observability.Metrics` (`flaggo_http_requests_total`,
`flaggo_http_request_duration_seconds`), labeled by method, path, and status.
The `/metrics` scrape endpoint is registered separately in the router.

## Configuration

None. Middleware is always applied; metrics require a non-nil `*observability.Metrics`.

## Representative tests

`internal/web/server_test.go` verifies gateway `/healthz` and `/readyz`
behavior; middleware ordering is exercised indirectly by handler tests.
See `openwiki/testing/test-strategy.md`.
