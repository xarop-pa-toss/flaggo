---
type: observability
title: Metrics and Logging
description: Prometheus metrics, structured JSON logging via slog, and request correlation across the backend and frontend.
tags: [observability, prometheus, slog, logging, metrics]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-09T18:31:04.450Z
sources:
  - id: openwiki-source-693bc5f0b4443bd18cfa76b4
    resource: repo://cmd/flaggo/main.go
  - id: openwiki-source-993789787071e9ec6a9b9ca6
    resource: repo://internal/http/middleware/middleware.go
  - id: openwiki-source-162f50532496def982cf3987
    resource: repo://internal/observability/metrics.go
  - id: openwiki-source-3ae86643bd52d54c1a51eeef
    resource: repo://internal/web/server.go
generated: { by: "pi", at: "2026-11-09T18:31:04.450Z" }
---

# Metrics and Logging

## Responsibility

`internal/observability/metrics.go` defines the Prometheus metrics;
`internal/http/middleware/middleware.go` (`Logging`, `Metrics`) emits access logs
and records them. Both backend and frontend use the same patterns.

## Metrics

`observability.Metrics` holds:
- `flaggo_http_requests_total` (CounterVec) — labels method, path, status.
- `flaggo_http_request_duration_seconds` (HistogramVec) — labels method, path.

Registered on the default registerer in production (`NewDefaultMetrics`);
tests can use a fresh registerer. Scrape endpoint: `GET /metrics`
(`promhttp.Handler()`).

## Logging

- Backend and frontend use `slog.NewJSONHandler` to stdout, level from
  `LOG_LEVEL` (debug/info/warn/error).
- Each request produces one `slog.Info("request", ...)` line: method, path,
  status, duration_ms, request_id, remote.
- Panics are logged via `Recovery` with panic value, path, request id.

## Correlation

`X-Request-ID` is generated at the outermost middleware and echoed on the
response. The frontend propagates it to the backend, so one id spans gateway
access log, backend access log, and any error logs.

## Related

- `openwiki/api/middleware.md`
- `openwiki/architecture/request-flow.md`
