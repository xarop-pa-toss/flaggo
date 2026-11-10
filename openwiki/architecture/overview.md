---
type: architecture-overview
title: Architecture Overview
description: Flaggo is a Go feature-flag control plane with a management dashboard gateway and a polling data-plane client, shipped as three binaries from one module.
tags: [architecture, overview, components, monorepo]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-10T11:19:27.436Z
sources:
  - id: openwiki-source-8037e2358a2c4f9b2c722a11
    resource: repo://AGENTS.md
  - id: openwiki-source-4cae579004db10634e547e3a
    resource: repo://cmd/flaggo-web/main.go
  - id: openwiki-source-693bc5f0b4443bd18cfa76b4
    resource: repo://cmd/flaggo/main.go
  - id: openwiki-source-e201e686a785f09b6d899f0b
    resource: repo://compose.yaml
  - id: openwiki-source-25b53892bc5b00be5d8f0155
    resource: repo://examples/client/main.go
  - id: openwiki-source-7bd911fdd3026b7b031a01e3
    resource: repo://go.mod
generated: { by: "pi", at: "2026-11-09T18:31:04.450Z" }
---

# Architecture Overview

## Responsibility

Flaggo decouples deployment from release: a centralized control plane stores and
serves feature flags; client applications (data plane) poll it and evaluate
flags in memory, so toggling a flag takes effect without a restart.

## Components

- **Control Plane Backend** (`cmd/flaggo`): REST API, persistence (Postgres or
  in-memory), health/readiness, Prometheus metrics, graceful shutdown.
- **Control Plane Frontend** (`cmd/flaggo-web`): embedded dashboard plus an API
  gateway that reverse-proxies `/api/*` to the backend. Only this service
  publishes a host port.
- **Data Plane Client** (`examples/client`): polling example that fetches flags
  and evaluates them per user key, printing state changes.
- **Postgres** (optional): persistent store; `STORE=memory` removes the
  dependency for local dev and tests.

## Monorepo layout

Single Go module (`go.mod`, module `github.com/xarop-pa-toss/flaggo`). Shared
packages under `internal/`: `config`, `http`, `web`, `store`, `model`, `eval`,
`observability`. Three build targets: `./cmd/flaggo`, `./cmd/flaggo-web`,
`./examples/client`.

## Runtime topology (compose)

`compose.yaml` defines four services on a private network: `postgres`,
`backend`, `frontend`, `client`. Only `frontend` maps host port 8080. The
frontend depends on the backend; the backend depends on postgres. See
`openwiki/build/containerization.md`.

## Key design points

- One container image per binary, built from the repo root as build context.
- Same image runs in every environment; behavior changes via environment
  variables (12-factor).
- Storage is pluggable behind the `store.Store` interface, enabling degraded
  mode when Postgres is unavailable.
- Evaluation happens in the client, not the server, so flag changes propagate
  within one poll interval.

## Related

- `openwiki/configuration/environment.md`
- `openwiki/architecture/request-flow.md`
- `openwiki/build/containerization.md`
