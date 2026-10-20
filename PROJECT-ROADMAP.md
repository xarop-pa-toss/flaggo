# PROJECT-ROADMAP — Flaggo

## 1. Intent statement (verbatim)

Flaggo is a lightweight, self-hosted Feature Flag & Runtime Configuration Platform built to demonstrate containerization and modern CI/CD patterns. It decouples deployment from release: a centralized Control Plane (management dashboard and API) where engineers toggle features, set user rollout percentages, and trigger instant kill switches in real time. Client applications (Data Plane) consume these rules via lightweight polling, evaluating feature logic dynamically in memory with zero downtime and zero restarts.

Operator intent (MVP): Control Plane REST API only, backed by Postgres, polling data plane, with strong operability (health, metrics, logging, graceful shutdown, env config). Operator owns Dockerfiles and deployment. App is an example for a DevOps training roadmap — management/operation matters more than app sophistication.

## 2. Current state

- Repo created: github.com/xarop-pa-toss/flaggo (public), Go 1.27 module.
- Initial scaffold: main.go (hello), .gitignore, README.md, go.mod.
- No domain code yet. No DB schema. No HTTP server. No tests.
- Operator has NOT asked for Dockerfiles/deployment manifests — out of scope (operator owns).

## 3. Planned phases

### Phase 1 — Operable HTTP skeleton (no DB yet)
Goal: a running Go HTTP server with /healthz, /readyz, /metrics, structured JSON logging, graceful shutdown, env-based config.
Success: `go run ./cmd/flaggo` starts; curl /healthz returns 200; SIGTERM shuts down cleanly; config via env (PORT, LOG_LEVEL).

### Phase 2 — Postgres store + schema
Goal: connect to Postgres via DATABASE_URL; migrate a `flags` table (name, enabled, rollout_percent, kill_switch, updated_at); store layer with CRUD.
Success: integration test against a real Postgres (or testcontainers) passes; migrate runs on startup.

### Phase 3 — Control Plane REST API
Goal: endpoints to create/update/delete/get/list flags; toggle enabled; set rollout %; trigger kill switch. JSON in/out. Validation. Error responses.
Success: curl walkthrough creates a flag, toggles it, sets 25% rollout, kill-switches it; all endpoints tested.

### Phase 4 — Flag evaluation engine
Goal: deterministic in-memory evaluation: kill switch overrides; rollout % by user key (consistent hashing); enabled/disabled.
Success: unit tests cover boundary cases (0%, 100%, kill switch, unknown flag).

### Phase 5 — Polling data-plane example client
Goal: `examples/client` polls /flags (or /flags/{name}) and evaluates flags locally; prints state; demonstrates zero-restart toggle.
Success: run control plane + client; toggle a flag via API; client reflects change within poll interval without restart.

### Phase 6 — Polish & operability hardening
Goal: request IDs, panic recovery, timeouts, graceful degradation when DB down (readyz flips), README with run instructions, example env file.
Success: README documents env vars, endpoints, and how to run; `go test ./...` green.

## 4. Scope-creep defenses

### In scope (v1)
- Control Plane REST API for flags (CRUD, toggle, rollout %, kill switch).
- Postgres persistence.
- Polling data-plane example client.
- Operability: healthz/readyz, metrics, structured logs, graceful shutdown, env config.
- Tests (unit + integration where feasible).

### Out of scope (v1) — do NOT build unless operator asks
- Web UI / dashboard (operator may add later).
- Authentication/authorization (note: add API key auth in v2).
- SSE or webhooks (polling only for v1).
- Multi-tenancy / organizations.
- SDKs beyond the example Go client (Node, Python, etc.).
- Dockerfiles, docker-compose, Helm, Kubernetes manifests (operator owns deployment).
- Feature flag targeting rules beyond rollout % (user segments, attributes).

### Adding scope
A new feature/phase is added ONLY if the operator explicitly requests it OR the gap widens (e.g., a discovered requirement makes v1 unusable). "While I'm here" ideas go to the "Note for later" list, not into code.

### Note for later (deferred ideas)
- API key auth (v2).
- Web UI dashboard (v2).
- SSE push for lower latency (v2).
- Webhooks for integrations (v2).
- Rich targeting rules (user attributes, segments) (v2).
- OpenAPI spec generation (v2).
- Rate limiting / quotas (v2).

## 5. Controlled-evolution path

- Trigger new phase: previous phase's success criteria all met + operator confirms.
- Trigger re-orientation: any drift signal (see below), or operator says "that's not what I wanted."
- Trigger phase transition vs re-orientation: phase transition = scope unchanged, moving to next planned phase. Re-orientation = scope/question changed, stop and re-read intent.
- "Note for later" reactivation criteria: operator explicitly asks, OR the item blocks a v1 success criterion.

## 6. Drift-check checkpoints

- Before starting each phase: re-read intent + this roadmap; confirm phase is next.
- Every ~10 tool actions: run the 6 drift signals (can I state intent? am I making code correct or realizing intent? would operator recognize this? asked operator recently? optimizing outside gap? output growing while intent stalls?).
- After each phase: verify success criteria; update "Current state" section.
- If a subagent's output doesn't match intent: reject and re-dispatch with clearer instructions.

## 7. Failure modes to watch (devops example specifics)

- Don't build a fancy app — operability is the point. Resist feature creep on flag features.
- Don't create Dockerfiles/deployment — operator owns. If asked, point to roadmap out-of-scope.
- Don't skip health/metrics "to save time" — they ARE the deliverable for a DevOps training example.
- Don't over-engineer storage — Postgres with a single `flags` table is enough for v1.
