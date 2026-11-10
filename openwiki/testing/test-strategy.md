---
type: testing-strategy
title: Testing Strategy
description: Unit and handler tests for Flaggo, the in-memory store in tests, and STORE=memory usage.
tags: [testing, unit-tests, handlers, eval]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-09T18:31:04.450Z
sources:
  - id: openwiki-source-164e2da859b5277df81c7d94
    resource: repo://.github/workflows/ci.yml
  - id: openwiki-source-1272ca1baf6d0cd9ea048363
    resource: repo://internal/eval/eval_test.go
  - id: openwiki-source-932b905de7c3220be428f5c6
    resource: repo://internal/http/flags_test.go
  - id: openwiki-source-34807c38e67e1c45c91262c2
    resource: repo://internal/web/server_test.go
generated: { by: "pi", at: "2026-11-09T18:31:04.450Z" }
---

# Testing Strategy

## Responsibility

Tests live next to code (`*_test.go`) and run with `STORE=memory` so no
Postgres is needed. CI runs `go test -v ./...`.

## Test files

- `internal/eval/eval_test.go` — kill switch, disabled, 0/100 rollout, and
  determinism of `Evaluate`.
- `internal/http/flags_test.go` — CRUD, kill switch, rollout, and error paths
  against the in-memory store.
- `internal/web/server_test.go` — gateway `/healthz` and `/readyz` behavior.

## Approach

- Handler tests use `MemoryStore` (no DB), keeping tests fast and isolated.
- `httptest` recorder / server for HTTP-level assertions.
- Evaluation tests are pure function tests (no I/O).
- Tests must pass with `go test ./...` and `go vet ./...`.

## Gaps

- No integration tests against Postgres yet.
- No end-to-end test of the polling client.
- No load/soak tests.

## Related

- `openwiki/api/control-plane-api.md`
- `openwiki/evaluation/flag-evaluation.md`
- `openwiki/storage/storage-layer.md`
- `openwiki/build/ci-pipeline.md`
