# Flaggo

A small, self-hosted Feature Flag platform.

Flaggo decouples **deployment** from **release**: a Control Plane exposes an API to create flags, set rollout percentages, and trigger instant kill switches. Client applications (the Data Plane) poll the API and evaluate flags in memory — so you can change behaviour without restarting or redeploying anything.

Three parts:

| Part | What it is | Where |
|------|-----------|-------|
| **Backend API** | Flag CRUD, rollout %, kill switch | `cmd/flaggo` |
| **Frontend** | Dashboard UI + `/api` proxy to the backend | `cmd/flaggo-web` |
| **Client** | Data Plane: polls flags, evaluates them, prints changes | `examples/client` |

---

## Run it

No database needed for development — the API has an in-memory store.

**Terminal 1 — backend API (port 8080):**
```sh
STORE=memory go run ./cmd/flaggo
```

**Terminal 2 — dashboard (port 8090):**
```sh
PORT=8090 BACKEND_URL=http://127.0.0.1:8080 go run ./cmd/flaggo-web
```

**Terminal 3 — data-plane client:**
```sh
CONTROL_PLANE_URL=http://127.0.0.1:8080 go run ./examples/client
```

Open **http://localhost:8090**, create a flag called `new-checkout`, tick *Enabled*, set rollout to `100`.

Terminal 3 prints:

```
[state]  new-checkout     ON   (enabled=true rollout=100% kill=false)
```

Now flip it in the dashboard — untick *Enabled*, or hit the kill switch — and within one poll interval:

```
[CHANGE] new-checkout     ON -> OFF  (enabled=false rollout=100% kill=false)
```

That is the whole product: the client never restarted.

### Using Postgres instead

```sh
export DATABASE_URL=postgres://user:pass@localhost:5432/flaggo?sslmode=disable
go run ./cmd/flaggo     # STORE defaults to postgres
```

The `flags` table is created automatically on startup.

---

## API

| Method | Path                        | Description                    |
|--------|-----------------------------|--------------------------------|
| GET    | `/healthz`                  | Liveness probe                 |
| GET    | `/readyz`                   | Readiness probe                |
| GET    | `/metrics`                  | Prometheus metrics             |
| POST   | `/flags`                    | Create a flag                  |
| GET    | `/flags`                    | List flags                     |
| GET    | `/flags/{name}`             | Get a flag                     |
| PUT    | `/flags/{name}`             | Update a flag                  |
| DELETE | `/flags/{name}`             | Delete a flag                  |
| POST   | `/flags/{name}/kill-switch` | Toggle / set kill switch       |
| PUT    | `/flags/{name}/rollout`     | Set rollout percentage         |

Try it with curl:

```sh
# create
curl -X POST localhost:8080/flags -H 'Content-Type: application/json' \
  -d '{"name":"new-checkout","description":"new checkout flow","enabled":true,"rollout_percent":100}'

# list
curl localhost:8080/flags

# turn it off instantly
curl -X POST localhost:8080/flags/new-checkout/kill-switch \
  -H 'Content-Type: application/json' -d '{"kill_switch":true}'

# 25% rollout
curl -X PUT localhost:8080/flags/new-checkout/rollout \
  -H 'Content-Type: application/json' -d '{"rollout_percent":25}'
```

Evaluation rules (see `internal/eval`):

1. `kill_switch` on → flag is OFF, always.
2. `enabled` false → OFF.
3. `rollout_percent` 0 → OFF, 100 → ON.
4. Otherwise a deterministic hash of `flag name + user key` decides, so the same user always gets the same answer.

---

## Configuration

### Backend API (`cmd/flaggo`)

| Variable           | Default    | Description                              |
|--------------------|------------|------------------------------------------|
| `PORT`             | `8080`     | HTTP listen port                         |
| `LOG_LEVEL`        | `info`     | debug / info / warn / error              |
| `STORE`            | `postgres` | `postgres` or `memory`                   |
| `DATABASE_URL`     | (empty)    | Postgres DSN, required when `STORE=postgres` |
| `SHUTDOWN_TIMEOUT` | `10s`      | Graceful shutdown timeout                |
| `READ_TIMEOUT`     | `15s`      | HTTP read timeout                        |
| `WRITE_TIMEOUT`    | `15s`      | HTTP write timeout                       |

### Frontend (`cmd/flaggo-web`)

| Variable      | Default           | Description                     |
|---------------|-------------------|---------------------------------|
| `PORT`        | `8080`            | HTTP listen port                |
| `BACKEND_URL` | `http://api:8080` | Backend API base URL            |
| `LOG_LEVEL`   | `info`            | debug / info / warn / error     |

### Client (`examples/client`)

| Variable            | Default                 | Description                     |
|---------------------|-------------------------|---------------------------------|
| `CONTROL_PLANE_URL` | `http://localhost:8080` | Backend API base URL            |
| `POLL_INTERVAL`     | `2s`                    | Polling interval                |
| `USER_KEY`          | `demo-user`             | User key used for rollout       |
| `FLAG_NAME`         | (empty)                 | Only watch this flag            |

---

## Commands

Both binaries take a subcommand: `serve` (default), `healthcheck` (exit 0/1, for probes), `version`.

## Build & test

```sh
go build ./...
go vet ./...
go test ./...
```

## License

TBD
