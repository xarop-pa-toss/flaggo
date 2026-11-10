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
- **HARD RULE: never create Dockerfiles, .dockerignore, compose files, CI workflows, or deploy manifests.** The operator writes the entire deployment layer themselves. This is not negotiable and not overridable by "but the task said start the stack" — ask the operator instead. If a task appears to need one, stop and confirm.
- Keep the app a good operability example: graceful shutdown, env config, health checks, metrics.
- Tests live next to code (*_test.go).
- Commits follow Conventional Commits (https://www.conventionalcommits.org): `<type>(<scope>): <description>` (e.g. `feat:`, `fix:`, `docs:`, `chore:`), with a blank line before the body and footers (`Fixes #123`, `BREAKING CHANGE:`).

## Build / Run

```sh
# Development, no database required:
STORE=memory go run ./cmd/flaggo                 # API  :8080
BACKEND_URL=http://127.0.0.1:8080 \
  go run ./cmd/flaggo-web                        # UI   :8080 (use PORT to change)
CONTROL_PLANE_URL=http://127.0.0.1:8080 \
  go run ./examples/client                       # client

# With Postgres instead:
DATABASE_URL=postgres://... go run ./cmd/flaggo
```

Or run all three locally in one shot:

```sh
./start.sh          # background start, logs in /tmp/opencode
./start.sh stop
```

## Repo Layout (target)

```
cmd/flaggo/          # Control Plane Backend Web API (entrypoint)
cmd/flaggo-web/      # Control Plane Frontend: dashboard + /api gateway
internal/
  config/            # env-based config
  http/              # API handlers, router, middleware
  web/               # frontend/gateway: embedded UI + reverse proxy
  store/             # persistence (postgres.go, memory.go)
  model/             # domain types
  eval/              # flag evaluation (rollout %, kill switch)
  observability/     # prometheus metrics
examples/
  client/            # polling data-plane client
PROJECT-ROADMAP.md
```

Not present on purpose: `Dockerfile`, `.dockerignore`, `compose.yaml`, `.github/workflows/`. Operator-owned.

<!-- OPENWIKI:START -->

## OpenWiki

This repository has a generated `openwiki/` evidence index. It is optional just-in-time context, not required startup reading.

- Do not enumerate, preload, or search wikis at task start. Use retrieval when the user asks for it, when unfamiliar architecture or dependency behavior materially affects the task, or when source inspection leaves an important uncertainty. Stop once the question is grounded.
- When those conditions apply and OpenWiki retrieval tools are available, use `openwiki_search` for just-in-time context and `openwiki_read` for the relevant complete sections. If search returns `workspace_required`, ask which listed workspace to use and retry with its ID.
- Use `openwiki_list_workspaces` or `openwiki_list_wikis` when workspace membership itself needs to be discovered.
- If the retrieval tools are unavailable, read `openwiki/quickstart.md` and follow its links to the relevant pages.
- Treat source code and tests as authoritative. A brief's unknowns and review items are verification gaps, not automatic requirements.
- Prefer the narrowest quiet validation that proves the changed behavior. Preserve complete failure output.

Use the repository's documented maintenance process to refresh generated pages after source changes. Check its configuration and completed runs before relying on automatic updates. Do not hand-edit generated OpenWiki pages unless explicitly asked; prefer updating source code/docs and letting OpenWiki regenerate.

<!-- OPENWIKI:END -->
