---
type: operations-health
title: Health and Readiness
description: Liveness and readiness endpoints, degraded mode, and gateway backend probing for Flaggo services.
tags: [health, readiness, liveness, operations, probes]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-09T18:31:04.450Z
sources:
  - id: openwiki-source-e3cb3b23134555f93bb85e79
    resource: repo://internal/http/health.go
  - id: openwiki-source-3ae86643bd52d54c1a51eeef
    resource: repo://internal/web/server.go
generated: { by: "pi", at: "2026-11-09T18:31:04.450Z" }
---

# Health and Readiness

## Responsibility

Distinguish liveness (`/healthz`, process is up) from readiness (`/readyz`,
service can serve traffic). Used by orchestrators (Kubernetes probes, compose
healthchecks) and by the frontend gateway to gate traffic to the backend.

## Backend (`internal/http/health.go`)

`HealthHandler` holds an atomic `ready` bool (default true). `/healthz` always
returns 200 `{"status":"ok"}`. `/readyz` returns 200 when ready, 503
`{"status":"not ready"}` otherwise. `SetReady` toggles it (called from
`main.go` on store connection/migration outcome and on shutdown).

## Degraded mode

When Postgres is unavailable or migration fails, the backend keeps running:
`/healthz` stays 200, `/readyz` returns 503, and flag routes are not
registered. This lets the process stay alive for diagnosis while signaling
unready to orchestrators.

## Frontend gateway

The gateway exposes its own `/healthz` (always ok) and `/readyz`, which probes
the backend's `/readyz` with a bounded timeout (`ReadyTimeout`, default 2s).
Returns 503 if the backend is not ready or unreachable.

## Lifecycle integration

On shutdown signal, `main.go` calls `health.SetReady(false)` before
`server.Shutdown`, so in-flight readiness probes fail and traffic drains.

## Related

- `openwiki/runtime/lifecycle.md`
- `openwiki/frontend/gateway.md`
- `openwiki/storage/storage-layer.md`
