---
type: architecture-flow
title: Request Flow
description: How an HTTP request travels from the browser through the frontend gateway reverse proxy into the backend, with middleware correlation.
tags: [architecture, http, proxy, request-flow]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-09T18:31:04.450Z
sources:
  - id: openwiki-source-3ae86643bd52d54c1a51eeef
    resource: repo://internal/web/server.go
generated: { by: "pi", at: "2026-11-09T18:31:04.450Z" }
---

# Request Flow

## Responsibility

Describe the end-to-end path of a management request (e.g. a dashboard click that
creates a flag) from the browser to the database and back, so a reader can
correlate logs and metrics across the two services.

## Path

1. Browser sends `POST /api/flags` to the frontend (host port 8080).
2. Frontend mux matches `/api/` → `newProxy` (`internal/web/server.go`).
3. Proxy director strips `/api`, rewrites path to `/flags`, and sets
   `X-Request-ID` from the gateway context.
4. Request traverses the backend middleware chain (request id, recovery,
   logging, metrics) and reaches `FlagsHandler.Create`.
5. Handler decodes/validates, calls `Store.Create`, re-reads the row, returns
   JSON.
6. Response flows back through the backend middleware (metrics recorded,
   access log emitted) to the proxy.
7. Proxy `ModifyResponse` deletes the backend's `X-Request-ID` copy; gateway's
   own id is already on the response.
8. Browser receives the JSON.

## Correlation

A single `X-Request-ID` appears in: the gateway access log, the backend access
log, the backend error logs (on failure), and both Prometheus scrape series are
labeled by path. This makes a dashboard click traceable end to end.

## Failure behavior

- Backend unreachable: proxy `ErrorHandler` returns `502 Bad Gateway` with
  `{"error":"control plane backend unreachable"}` and logs the error.
- Backend not ready: frontend `/readyz` returns 503 (probes backend `/readyz`).
- Handler panic: `Recovery` middleware returns 500 instead of crashing.

## Related

- `openwiki/api/middleware.md`
- `openwiki/frontend/gateway.md`
- `openwiki/api/control-plane-api.md`
