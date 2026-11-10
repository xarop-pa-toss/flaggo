---
type: build-containerization
title: Containerization
description: How Flaggo is packaged into container images and run as a multi-service stack with Docker Compose.
tags: [docker, containerization, compose, distroless]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-09T18:31:04.450Z
sources:
  - id: openwiki-source-e201e686a785f09b6d899f0b
    resource: repo://compose.yaml
  - id: openwiki-source-617551d27beb40063c3c4624
    resource: repo://Dockerfile.backend
  - id: openwiki-source-96c6a14d9629409d26255bf2
    resource: repo://Dockerfile.client
  - id: openwiki-source-cd191afb9f44a1a29c78d716
    resource: repo://Dockerfile.web
generated: { by: "pi", at: "2026-11-09T18:31:04.450Z" }
---

# Containerization

## Responsibility

Package each of the three binaries into a minimal, secure container image and
orchestrate them with Docker Compose on a private network.

## Dockerfiles

All three (`Dockerfile.backend`, `Dockerfile.web`, `Dockerfile.client`) use the
same multi-stage pattern:

1. **Build stage** (`golang:1.27.2-alpine`): `CGO_ENABLED=0 GOOS=linux`,
   `go build -ldflags="-s -w"` into `/app/binary`.
2. **Runtime stage** (`gcr.io/distroless/static-debian12:nonroot`): copy the
   binary, run as non-root UID 65532, `ENTRYPOINT ["/binary"]`. Backend and
   web expose 8080; client is outbound only.

Layer caching: `go.mod`/`go.sum` are copied and `go mod download` runs before
the source `COPY . .`.

## .dockerignore

Excludes `.git`, build outputs, secrets (`.env`, keys), docs, IDE files, and
the Dockerfiles themselves, keeping the build context small and protecting
layer cache from irrelevant changes.

## Docker Compose

`compose.yaml` defines four services on a shared private network:

- `postgres` (postgres:16-alpine) — persistent store.
- `backend` — control plane API, `STORE=postgres`, depends on postgres.
- `frontend` — dashboard + gateway, only service with a host-published port
  (8080), depends on backend.
- `client` — data plane example, depends on backend.

Only the frontend maps a port to the host, so all backend traffic stays on the
private network.

## Build commands

```sh
docker build -f Dockerfile.backend -t flaggo-backend:local .
docker compose up -d
```

## Related

- `openwiki/architecture/overview.md`
- `openwiki/build/ci-pipeline.md`
