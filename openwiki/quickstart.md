---
type: "Reference"
title: "Flaggo Quickstart"
openwiki_generated: true
verified:
  - by: openwiki/0.9.0
    at: 2026-11-09T18:31:04.450Z
sources:
  - id: openwiki-source-8037e2358a2c4f9b2c722a11
    resource: repo://AGENTS.md
  - id: openwiki-source-7bd911fdd3026b7b031a01e3
    resource: repo://go.mod
generated: { by: "pi", at: "2026-11-09T18:31:04.450Z" }
---


# Flaggo Quickstart

## What is Flaggo?

A lightweight Go feature-flag control plane (backend + dashboard gateway) with a
polling data-plane client example. See `openwiki/architecture/overview.md`.

## Task routing

- **Understand the system** → `openwiki/architecture/overview.md`, then
  `openwiki/architecture/request-flow.md`.
- **Add or change an API endpoint** → `openwiki/api/control-plane-api.md`,
  `openwiki/api/middleware.md`, `openwiki/domain/flag-model.md`.
- **Change persistence** → `openwiki/storage/storage-layer.md`,
  `openwiki/operations/health-and-readiness.md`.
- **Change evaluation/rollout logic** → `openwiki/evaluation/flag-evaluation.md`,
  `openwiki/domain/flag-model.md`.
- **Configure environments** → `openwiki/configuration/environment.md`.
- **Run locally / containerize** → `openwiki/build/containerization.md`,
  `openwiki/runtime/lifecycle.md`.
- **CI / tests** → `openwiki/build/ci-pipeline.md`,
  `openwiki/testing/test-strategy.md`.
- **Operate / debug** → `openwiki/operations/health-and-readiness.md`,
  `openwiki/observability/metrics-and-logging.md`,
  `openwiki/frontend/gateway.md`.
- **Data plane client** → `openwiki/client/data-plane-client.md`.

## Build & run (dev)

```sh
STORE=memory go run ./cmd/flaggo
BACKEND_URL=http://127.0.0.1:8080 go run ./cmd/flaggo-web
CONTROL_PLANE_URL=http://127.0.0.1:8080 go run ./examples/client
```

Or `./start.sh`. See AGENTS.md for the full layout and conventions.
