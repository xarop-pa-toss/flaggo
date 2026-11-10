---
type: client-example
title: Data Plane Client
description: Polling client example that fetches flags and evaluates them in memory, printing state changes without restart.
tags: [client, data-plane, polling, evaluation]
verified:
  - by: openwiki/0.9.0
    at: 2026-11-09T18:31:04.450Z
sources:
  - id: openwiki-source-25b53892bc5b00be5d8f0155
    resource: repo://examples/client/main.go
  - id: openwiki-source-9ef53414007a0bac3914ddd2
    resource: repo://internal/eval/eval.go
generated: { by: "pi", at: "2026-11-09T18:31:04.450Z" }
---

# Data Plane Client

## Responsibility

`examples/client/main.go` is the reference data plane: it polls the control
plane, evaluates every flag in memory for one user key, and prints a CHANGE
line when a flag's effective value flips. It demonstrates the decoupling of
deploy from release.

## Flow

1. Read `CONTROL_PLANE_URL`, `POLL_INTERVAL`, `USER_KEY`, optional
   `FLAG_NAME` from the environment.
2. Loop: fetch flags (`GET /flags` or `GET /flags/{name}`), sort by name.
3. For each flag call `eval.Evaluate(flag, userKey)`.
4. Print `[state]` on first sight and `[CHANGE] old -> new` when the
   effective value differs from the previous poll.
5. Sleep `POLL_INTERVAL`, repeat. Errors are logged and retried.

## Evaluation

Evaluation runs locally in the client via `internal/eval` (kill switch, enabled,
deterministic rollout by user key). The server never evaluates; it only stores
and serves flag definitions. See `openwiki/evaluation/flag-evaluation.md`.

## Configuration

- `CONTROL_PLANE_URL` (default http://localhost:8080)
- `POLL_INTERVAL` (default 2s)
- `USER_KEY` (default demo-user)
- `FLAG_NAME` (optional filter)

## Related

- `openwiki/evaluation/flag-evaluation.md`
- `openwiki/configuration/environment.md`
