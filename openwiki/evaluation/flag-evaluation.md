---
type: evaluation-logic
title: Flag Evaluation
description: In-memory flag evaluation algorithm including kill switch, enabled state, and deterministic rollout hashing by user key.
tags: [evaluation, rollout, kill-switch, hashing]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-09T18:31:04.450Z
sources:
  - id: openwiki-source-9ef53414007a0bac3914ddd2
    resource: repo://internal/eval/eval.go
generated: { by: "pi", at: "2026-11-09T18:31:04.450Z" }
---

# Flag Evaluation

## Responsibility

`internal/eval/eval.go` implements `Evaluate(flag, userKey) bool`, the single
function that decides whether a flag is on for a given user. It runs in the
data plane (client), never on the server.

## Algorithm

1. If `KillSwitch` is true → false (overrides everything).
2. If `!Enabled` → false.
3. If `RolloutPercent >= 100` → true.
4. If `RolloutPercent <= 0` → false.
5. Hash `name:userKey` with FNV-32a, take `bucket = hash % 100`.
6. Return `bucket < RolloutPercent`.

## Properties

- **Deterministic**: same flag + user key → same result across processes.
- **Kill switch wins**: overrides enabled and rollout.
- **Gradual rollout**: increasing `RolloutPercent` only adds users (monotonic
  in percentage for a fixed hash).
- **No server state**: evaluation is pure and side-effect free.

## Tests

`internal/eval/eval_test.go` covers kill switch, disabled, 0/100 percent, and
determinism. See `openwiki/testing/test-strategy.md`.

## Related

- `openwiki/domain/flag-model.md`
- `openwiki/client/data-plane-client.md`
