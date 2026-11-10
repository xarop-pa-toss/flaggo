---
type: build-pipeline
title: CI Pipeline
description: GitHub Actions workflow that lints, tests, and compiles Flaggo on every push and pull request to main.
tags: [ci, github-actions, build, testing]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-09T18:31:04.450Z
sources:
  - id: openwiki-source-164e2da859b5277df81c7d94
    resource: repo://.github/workflows/ci.yml
generated: { by: "pi", at: "2026-11-09T18:31:04.450Z" }
---

# CI Pipeline

## Responsibility

`.github/workflows/ci.yml` ("Flaggo CI Pipeline") automates quality checks on
every `push` and `pull_request` to `main`: static analysis (`go vet`), unit
tests, and compilation of all three binaries.

## Trigger and environment

- Triggers: `push`/`pull_request` on `main`.
- Runner: `ubuntu-latest`.
- Go version: `1.27.2` (aligned with `go.mod` and the Dockerfiles).
- `STORE=memory` is set for tests so no Postgres is required.

## Steps

1. Checkout code (`actions/checkout`).
2. Setup Go with module cache enabled.
3. `go vet ./...`
4. `go test -v ./...`
5. `go build` for `./cmd/flaggo`, `./cmd/flaggo-web`, `./examples/client`
   (output discarded, compile check only).

## Gaps / planned work

- No container image build or publish yet (planned: `docker/build-push-action`
  to GHCR, SHA-tagged, multi-platform).
- No vulnerability scanning (`govulncheck`, Trivy) yet.
- Actions are pinned by mutable tag (`@v4`/`@v5`), not by SHA.
- No top-level `permissions:` or `timeout-minutes` yet.
See `openwiki/build/containerization.md` for the image build plan.

## Related

- `openwiki/build/containerization.md`
- `openwiki/testing/test-strategy.md`
