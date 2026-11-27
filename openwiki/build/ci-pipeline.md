---
type: build-pipeline
title: CI Pipeline
description: GitHub Actions workflow that lints, tests, and compiles Flaggo binaries on every push and pull request to main.
tags: [ci, github-actions, build, testing]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-27T17:49:20.614Z
sources:
  - id: openwiki-source-164e2da859b5277df81c7d94
    resource: repo://.github/workflows/ci.yml
generated: { by: "openwiki/0.9.0", at: "2026-11-27T17:49:20.614Z" }
---

# CI Pipeline

## Responsibility

`.github/workflows/ci.yml` ("Flaggo CI Pipeline") automates quality checks on
every `push` and `pull_request` to `main` (as well as manual `workflow_dispatch`),
ignoring documentation and wiki-only updates: static analysis (`go vet`), unit
tests, and compilation of all three binaries.

## Trigger and environment

- Triggers: `push`/`pull_request` on `main` (excluding Markdown/docs/wiki paths), plus `workflow_dispatch`.
- Runner: `ubuntu-latest`.
- Go version: `1.27.2` (configured via `actions/setup-go@v5` with built-in module caching).
- `STORE=memory` is set for tests so no Postgres instance is required.

## Steps

1. Checkout code (`actions/checkout@v4`).
2. Setup Go 1.27.2 with dependency caching enabled (`actions/setup-go@v5`).
3. `go vet ./...` (static analysis linter).
4. `go test -v ./...` with `STORE=memory`.
5. `go build -v -o /dev/null` for `./cmd/flaggo`, `./cmd/flaggo-web`, and `./examples/client` (compilation verification).

## Gaps / planned work

- No container image build or publish yet (planned: `docker/build-push-action`
  to GHCR, SHA-tagged, multi-platform).
- No vulnerability scanning (`govulncheck`, Trivy) yet.
- Actions are pinned by mutable tag (`@v4`/`@v5`), not by SHA.
- No top-level `permissions:` or `timeout-minutes` yet.
See [Containerization](containerization.md) for the image build plan.

## Related

- [Containerization](containerization.md)
- Test Strategy](../testing/test-strategy.md)
