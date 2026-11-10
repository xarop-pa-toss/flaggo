---
type: api-reference
title: Control Plane HTTP API
description: REST endpoints of the Flaggo control plane for creating, reading, updating, deleting, and toggling feature flags, plus kill switch and rollout controls.
tags: [api, http, flags, control-plane]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-09T18:31:04.450Z
sources:
  - id: openwiki-source-693bc5f0b4443bd18cfa76b4
    resource: repo://cmd/flaggo/main.go
  - id: openwiki-source-5839211a1058172e61dce73f
    resource: repo://internal/http/flags.go
generated: { by: "pi", at: "2026-11-09T18:31:04.450Z" }
---

# Control Plane HTTP API

## Responsibility

The control plane exposes a small REST API over feature flags. It is the only
write path for flag state: the dashboard (frontend gateway) and any external
automation talk to these endpoints, and the data plane (polling clients) reads
them. All handlers live in `internal/http/flags.go` (`FlagsHandler`) and are
wired into the mux by `Register`.

## Entrypoints

- `cmd/flaggo/main.go` builds a `FlagsHandler` only when storage is usable
  (`case "memory"` or successful Postgres `Migrate`); otherwise the process
  runs in degraded mode with no flag routes.
- Routes are registered in `internal/http/router.go` via
  `d.Flags.Register(mux)` only when `d.Flags != nil`.

## Endpoints

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/flags` | Create a flag |
| `GET` | `/flags` | List all flags (sorted by name) |
| `GET` | `/flags/{name}` | Get one flag |
| `PUT` | `/flags/{name}` | Replace a flag |
| `DELETE` | `/flags/{name}` | Delete a flag |
| `POST` | `/flags/{name}/kill-switch` | Toggle/force kill switch |
| `PUT` | `/flags/{name}/rollout` | Set rollout percentage |

Health and metrics endpoints (`/healthz`, `/readyz`, `/metrics`) are served
separately; see `openwiki/operations/health-and-readiness.md`.

## Request flow

1. Request enters the mux in `internal/http/router.go`.
2. Middleware chain (outermost first): `RequestID` → `Recovery` → `Logging` →
   `Metrics` → mux. See `openwiki/api/middleware.md`.
3. Handler decodes JSON with `DisallowUnknownFields`, validates the model,
   calls the `Store`, and re-reads the persisted row before responding.

## Contracts and invariants

- `model.Flag` requires a non-empty `Name` and `RolloutPercent` in `[0,100]`
  (`internal/model/flag.go` `Validate`).
- Create rejects duplicates with `409 Conflict` (`model.ErrAlreadyExists`).
- Get/Update/Delete/Rollout on a missing flag return `404 Not Found`.
- Update forces the path name onto the body (`f.Name = name`), so the URL is
  authoritative.
- Kill switch accepts an optional `{"kill_switch": bool}`; when omitted it
  toggles the current value.
- Rollout validates the new percentage before persisting.

## Failure behavior

- Malformed or unknown JSON fields → `400 Bad Request`.
- Validation failure → `400 Bad Request`.
- Store errors other than not-found/already-exists → `500 Internal Server
  Error` with a generic message; details go to the slog logger with the request
  id.
- In degraded mode (no store), flag routes are simply not registered, so
  requests hit the default mux and return `404`.

## Configuration

No endpoint-specific configuration. The base path is fixed; the frontend
gateway rewrites `/api/<path>` to `/<path>` (see
`openwiki/frontend/gateway.md`).

## Representative tests

`internal/http/flags_test.go` exercises create/list/get/update/delete and
kill-switch/rollout against the in-memory store. See
`openwiki/testing/test-strategy.md`.
