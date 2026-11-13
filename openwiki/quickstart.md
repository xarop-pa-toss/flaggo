---
type: "Guide"
title: "Quickstart Guide"
description: "Quickstart guide and navigation entrypoint for developers exploring the Flaggo repository."
tags: ["quickstart", "onboarding", "navigation", "development"]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-13T15:53:50.497Z
sources:
  - id: openwiki-source-8037e2358a2c4f9b2c722a11
    resource: repo://AGENTS.md
  - id: openwiki-source-7bd911fdd3026b7b031a01e3
    resource: repo://go.mod
  - id: openwiki-source-23775c3de52f3ab95a13cb8b
    resource: repo://README.md
generated: { by: "openwiki/0.9.0", at: "2026-11-13T15:53:50.497Z" }
---

# Quickstart Guide

## What is Flaggo?

Flaggo is a small, self-hosted Feature Flag platform. It decouples **deployment** from **release**: a Control Plane exposes an API to create flags, set rollout percentages, and trigger instant kill switches. Client applications (the Data Plane) poll the API and evaluate flags in memory — so you can change behaviour without restarting or redeploying anything.

The repository contains three primary components:
- **Backend API (`cmd/flaggo`)**: Flag CRUD, rollout percentages, and kill switches.
- **Frontend (`cmd/flaggo-web`)**: Dashboard UI plus an `/api` proxy to the backend.
- **Client (`examples/client`)**: Data Plane example that polls flags, evaluates them, and prints changes.

---

## Navigation Map

Use the following documentation hierarchy to explore the OpenWiki knowledge base:

| Domain / Topic | Wiki Page | Description |
|----------------|-----------|-------------|
| **Architecture** | [Architecture Overview](architecture/overview.md) | High-level architectural overview of the system, components, and data flows. |
| **Configuration** | [Environment Configuration](configuration/environment.md) | Comprehensive reference for environment variables and configuration settings across backend, frontend, and client. |

---

## Local Development & Running

No external database is required for local development because the backend supports an in-memory store (`STORE=memory`).

### Option 1: Using the Non-Docker Script

Build and start the backend, dashboard, and client in the background using the helper script:

```sh
./start-non-docker.sh          # builds + starts backend, dashboard, client in the background
./start-non-docker.sh stop     # stops all three
```
Logs are stored under `/tmp/opencode/`. Open the dashboard at **http://localhost:8090**.

### Option 2: Running in Three Terminals

You can run each component in a separate terminal:

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

---

## Docker Compose Setup

To run the complete containerized stack (including PostgreSQL persistence, backend, frontend dashboard, and client simulation):

```sh
docker compose up --build
```

The frontend dashboard is accessible on port `8080` of the host machine.

---

## Building and Testing

To build binaries, run code vetting, and execute tests across the module:

```sh
go build ./...
go vet ./...
go test ./...
```
