# AGENTS.md — Flaggo

## Project Purpose

Flaggo is a lightweight, self-hosted Feature Flag & Runtime Configuration Platform built to demonstrate containerization and modern CI/CD patterns. It decouples deployment from release: a centralized Control Plane (management dashboard and API) where engineers toggle features, set user rollout percentages, and trigger instant kill switches in real time. Client applications (Data Plane) consume these rules via lightweight polling, evaluating feature logic dynamically in memory with zero downtime and zero restarts.

**Project roadmap:** PROJECT-ROADMAP.md — the living plan that gates scope creep and guides controlled evolution.

## Intent (one sentence)

Build a small, well-operable Go feature-flag control plane backed by Postgres, with a polling data-plane client example, health/metrics/logging for DevOps training — without Dockerfiles or deployment manifests (operator owns those).

## Stack

- Language: Go 1.27
- Storage: Postgres (via pgx)
- HTTP: net/http (stdlib) + chi or stdlib mux
- Config: environment variables (12-factor)
- Observability: structured JSON logging (slog), Prometheus metrics, /healthz + /readyz
- Module: github.com/xarop-pa-toss/flaggo

## Conventions

- `go build ./...`, `go test ./...`, `go vet ./...` must pass.
- No Dockerfiles, docker-compose, or deployment manifests — operator owns deployment.
- Keep the app a good operability example: graceful shutdown, env config, health checks, metrics.
- Tests live next to code (*_test.go).

## Build / Run

```sh
# Requires Postgres. Set DATABASE_URL.
go run ./cmd/flaggo
# or
DATABASE_URL=postgres://... go run ./cmd/flaggo
```

## Repo Layout (target)

```
cmd/flaggo/          # main entrypoint (control plane)
internal/
  config/            # env-based config
  http/              # handlers, router, middleware
  store/             # Postgres persistence (flags, rules)
  model/             # domain types
  eval/              # flag evaluation (rollout %, kill switch)
examples/
  client/            # polling data-plane example
PROJECT-ROADMAP.md
```
