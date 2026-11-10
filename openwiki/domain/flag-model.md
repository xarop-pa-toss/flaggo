---
type: domain-model
title: Flag Domain Model
description: The Flag entity, its invariants, validation rules, and domain errors shared across API, store, and evaluation.
tags: [domain, model, flag, validation]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-09T18:31:04.450Z
sources:
  - id: openwiki-source-5839211a1058172e61dce73f
    resource: repo://internal/http/flags.go
  - id: openwiki-source-66c38061448b4ba09326772c
    resource: repo://internal/model/errors.go
  - id: openwiki-source-93bd1d711c8f52a4c0f6dca3
    resource: repo://internal/model/flag.go
generated: { by: "pi", at: "2026-11-09T18:31:04.450Z" }
---

# Flag Domain Model

## Responsibility

`internal/model/flag.go` defines the core `Flag` entity and its validation.
`internal/model/errors.go` defines sentinel errors used by handlers and stores.

## Flag struct

Fields (JSON-tagged): `name`, `description`, `enabled`, `rollout_percent`,
`kill_switch`, `created_at`, `updated_at`. See `model.Flag`.

## Invariants (Validate)

- `Name` must be non-empty (`ErrInvalidName`).
- `RolloutPercent` must be in `[0, 100]` (`ErrInvalidRollout`).
- Timestamps are set by the store on create/update, not by clients.

## Sentinel errors

- `ErrNotFound` — flag does not exist.
- `ErrAlreadyExists` — duplicate name on create.
- `ErrInvalidName`, `ErrInvalidRollout` — validation failures.

## Usage

- Handlers call `f.Validate()` before persistence and map sentinel errors to
  HTTP status codes (404, 409, 400).
- Stores call `Validate()` again defensively on create/update.
- Evaluation (`internal/eval`) reads fields but does not mutate.

## Related

- `openwiki/evaluation/flag-evaluation.md`
- `openwiki/api/control-plane-api.md`
- `openwiki/storage/storage-layer.md`
