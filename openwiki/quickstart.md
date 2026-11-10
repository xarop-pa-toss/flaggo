---
type: "Guide"
title: "Flaggo Quickstart & Task Routing"
description: "Developer onboarding, environment setup, and wiki task routing for the Flaggo feature flag platform."
tags: ["quickstart", "onboarding", "task-routing", "development"]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-10T11:51:15.631Z
sources:
  - id: openwiki-source-8037e2358a2c4f9b2c722a11
    resource: repo://AGENTS.md
  - id: openwiki-source-7bd911fdd3026b7b031a01e3
    resource: repo://go.mod
  - id: openwiki-source-23775c3de52f3ab95a13cb8b
    resource: repo://README.md
generated: { by: "openwiki/0.9.0", at: "2026-11-10T11:19:27.436Z" }
---

# Quickstart & Task Routing

## What is Flaggo?

Flaggo is a lightweight, self-hosted feature flag and runtime configuration platform built in Go. It decouples deployment from release by providing a centralized Control Plane (backend API and management dashboard) where engineers can toggle features, adjust rollout percentages, and trigger instant kill switches in real time. Client applications (Data Plane) poll these rules and evaluate them locally in memory for zero downtime and zero restarts.

Three binaries are built from the Go module:
- **Backend API (`cmd/flaggo`)**: Control plane backend managing flag CRUD, rollout percentages, and kill switches.
- **Frontend Gateway (`cmd/flaggo-web`)**: Management dashboard UI combined with an `/api` reverse proxy to the backend.
- **Data Plane Client (`examples/client`)**: Example client that polls the control plane, evaluates flags, and prints changes.

## Task Routing Map

Use this routing table to navigate the wiki based on your development or operational task:

| Task / Goal | Relevant Wiki Pages |
|-------------|---------------------|
| **Understand system architecture & request flows** | [Architecture Overview](architecture/overview.md), [Request Flow](architecture/request-flow.md) |
| **Add or modify API endpoints** | [Control Plane API](api/control-plane-api.md), [HTTP Middleware](api/middleware.md), [Flag Domain Model](domain/flag-model.md) |
| **Modify storage or persistence** | [Storage Layer](storage/storage-layer.md), [Health & Readiness](operations/health-and-readiness.md) |
| **Change evaluation / rollout logic** | [Flag Evaluation](evaluation/flag-evaluation.md), [Flag Domain Model](domain/flag-model.md) |
| **Configure environment variables** | [Environment Configuration](configuration/environment.md) |
| **Run locally, containerize, or manage lifecycle** | [Containerization](build/containerization.md), [Runtime Lifecycle](runtime/lifecycle.md) |
| **CI pipelines & test strategy** | [CI Pipeline](build/ci-pipeline.md), [Test Strategy](testing/test-strategy.md) |
| **Operate, monitor, or debug** | [Health & Readiness](operations/health-and-readiness.md), [Metrics & Logging](observability/metrics-and-logging.md), [Frontend Gateway](frontend/gateway.md) |
| **Data plane client development** | [Data Plane Client](client/data-plane-client.md) |

## Development Setup & Running Locally

No database is required for local development; the backend defaults to an in-memory store when `STORE=memory` is specified.

### Option A: One Command (Background Stack)

```sh
./start.sh          # builds + starts backend, dashboard, client in the background
./start.sh stop     # stops all three
```
Logs are written to `/tmp/opencode/`. Access the dashboard at **http://localhost:8090**.

### Option B: Three Terminals

**Terminal 1 — Backend API (port 8080):**
```sh
STORE=memory go run ./cmd/flaggo
```

**Terminal 2 — Frontend Dashboard (port 8090):**
```sh
PORT=8090 BACKEND_URL=http://127.0.0.1:8080 go run ./cmd/flaggo-web
```

**Terminal 3 — Data-Plane Client:**
```sh
CONTROL_PLANE_URL=http://127.0.0.1:8080 go run ./examples/client
```

### Using PostgreSQL

To run with PostgreSQL persistence instead of memory:
```sh
export DATABASE_URL=postgres://user:pass@localhost:5432/flaggo?sslmode=disable
STORE=postgres go run ./cmd/flaggo
```
The `flags` table is initialized automatically upon startup.
