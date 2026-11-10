---
type: frontend-gateway
title: Frontend Gateway
description: Embedded dashboard and reverse proxy that fronts the control plane backend, with readiness probing and single-binary shipping.
tags: [frontend, gateway, proxy, dashboard]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-09T18:31:04.450Z
sources:
  - id: openwiki-source-3ae86643bd52d54c1a51eeef
    resource: repo://internal/web/server.go
generated: { by: "pi", at: "2026-11-09T18:31:04.450Z" }
---

# Frontend Gateway

## Responsibility

`internal/web` implements the control plane frontend: an embedded dashboard
(`ui/` assets) plus an API gateway that reverse-proxies `/api/*` to the
backend. Only this service publishes a host port, so backend traffic stays on
the private network.

## Handler construction

`NewHandler(Deps)` requires a non-empty `BackendURL`. It builds a mux with:

- `/api/` → reverse proxy to backend (path rewritten).
- `/static/` → embedded static assets.
- `/healthz` → liveness (always ok).
- `/readyz` → readiness (probes backend `/readyz`).
- `/metrics` → Prometheus scrape.
- `/` → dashboard index.

Middleware chain matches the backend (request id, recovery, logging, metrics).

## Reverse proxy

`newProxy` uses `httputil.NewSingleHostReverseProxy`. The director strips `/api`
from the path and propagates `X-Request-ID`. `ErrorHandler` returns 502 on
backend failure; `ModifyResponse` removes the backend's echoed request id.

## Readiness probing

`probe(target, "/readyz", timeout)` performs an HTTP GET with a bounded timeout.
`/readyz` on the gateway returns 200 only when the backend is ready, 503
otherwise.

## Shipping

The dashboard is embedded via `//go:embed ui`, so the frontend ships as a single
static binary with no node_modules or asset pipeline.

## Related

- `openwiki/architecture/request-flow.md`
- `openwiki/operations/health-and-readiness.md`
