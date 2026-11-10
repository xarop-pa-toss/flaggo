---
type: "Reference"
title: "Environment Configuration"
openwiki_generated: true
verified:
  - by: openwiki/0.9.0
    at: 2026-11-09T18:31:04.450Z
sources:
  - id: openwiki-source-693bc5f0b4443bd18cfa76b4
    resource: repo://cmd/flaggo/main.go
  - id: openwiki-source-a8910515ddd14810ad43f5c1
    resource: repo://internal/config/config.go
generated: { by: "pi", at: "2026-11-09T18:31:04.450Z" }
---


# Environment Configuration

## Responsibility

`internal/config/config.go` loads all backend runtime configuration from
environment variables with sensible defaults. The client (`examples/client`)
reads its own subset directly. No config files are used (12-factor).

## Backend variables

| Var | Default | Meaning |
|---|---|---|
| `PORT` | 8080 | Listen port |
| `LOG_LEVEL` | info | debug/info/warn/error |
| `STORE` | postgres | `postgres` or `memory` |
| `DATABASE_URL` | (empty) | Postgres connection string (required when STORE=postgres) |
| `SHUTDOWN_TIMEOUT` | 10s | Graceful shutdown budget |
| `READ_TIMEOUT` | 15s | HTTP read timeout |
| `WRITE_TIMEOUT` | 15s | HTTP write timeout |

Validation: `PORT` must be non-empty; `STORE` must be `postgres` or `memory`.
`DATABASE_URL` is required only in postgres mode.

## Client variables

| Var | Default | Meaning |
|---|---|---|
| `CONTROL_PLANE_URL` | http://localhost:8080 | Backend base URL |
| `POLL_INTERVAL` | 2s | Poll cadence |
| `USER_KEY` | demo-user | User key for rollout hashing |
| `FLAG_NAME` | (empty) | Optional single-flag filter |

## Flow

`config.Load()` is called once in `runServer()` (`cmd/flaggo/main.go`). The
resulting `Config` drives store selection, logger level, server timeouts, and
router wiring. The frontend (`cmd/flaggo-web`) uses `BACKEND_URL`.

## Degraded mode

If `STORE=postgres` but `DATABASE_URL` is empty or connection/migration fails,
the backend logs the error, sets readiness false, and starts without flag
routes. `/healthz` stays ok; `/readyz` returns 503.

## Local template

`.env.example` documents the variables; copy to `.env` for local dev (gitignored).

## Related

- `openwiki/runtime/lifecycle.md`
- `openwiki/architecture/overview.md`
