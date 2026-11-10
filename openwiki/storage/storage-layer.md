---
type: storage-layer
title: Storage Layer
description: Store interface and its in-memory and Postgres implementations, migrations, pooling, and degraded-mode behavior.
tags: [storage, postgres, memory, persistence, store]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-09T18:31:04.450Z
sources:
  - id: openwiki-source-f628907702addc1ee81ddcb3
    resource: repo://internal/store/memory.go
  - id: openwiki-source-3998dcb9b7abce50f21ad523
    resource: repo://internal/store/postgres.go
  - id: openwiki-source-4a81fcd95533ed8ba5a77739
    resource: repo://internal/store/store.go
generated: { by: "pi", at: "2026-11-09T18:31:04.450Z" }
---

# Storage Layer

## Responsibility

`internal/store` defines the `Store` interface and two implementations:
`MemoryStore` (dev/tests) and `PostgresStore` (production). Handlers depend on
the interface, so storage is swappable and degraded mode is possible.

## Store interface

Methods: `Migrate`, `Create`, `Get`, `List`, `Update`, `Delete`, `Close`.
Errors use sentinel values from `internal/model` (`ErrNotFound`,
`ErrAlreadyExists`).

## MemoryStore

In-memory map guarded by `sync.RWMutex`. `Migrate` and `Close` are no-ops.
Used for `STORE=memory` and for tests. Not persistent across restarts.

## PostgresStore

- `NewPostgres` opens a `pgxpool.Pool` and pings it.
- `Migrate` creates the `flags` table if missing.
- CRUD via parameterized SQL; unique violations map to `ErrAlreadyExists`;
  `pgx.ErrNoRows` maps to `ErrNotFound`.
- `Close` closes the pool.

## Degraded mode

If connection or migration fails, `main.go` leaves `st` nil, sets readiness
false, and omits flag routes. The process still serves health and metrics.

## Related

- `openwiki/api/control-plane-api.md`
- `openwiki/operations/health-and-readiness.md`
- `openwiki/runtime/lifecycle.md`
- `openwiki/testing/test-strategy.md`
