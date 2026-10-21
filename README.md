# Flaggo

A lightweight, self-hosted Feature Flag & Runtime Configuration Platform. Built to demonstrate containerization and modern CI/CD patterns.

Flaggo decouples **deployment** from **release**: a centralized Control Plane exposes an API to create flags, set rollout percentages, and trigger instant kill switches. Client applications (Data Plane) poll the API and evaluate flags in memory — no restarts, no redeployments.

---

## Architecture — the 4-app stack

```
                     host port 8080
                           │
                    ┌──────────────┐
                    │   frontend   │  Control Plane dashboard + API gateway
                    │  (public)    │  serves the UI, proxies /api → api:8080
                    └──────┬───────┘
                           │  internal "flaggo" bridge network
              ┌────────────┼────────────┐
              │            │            │
        ┌─────▼─────┐ ┌────▼─────┐ ┌────▼──────┐
        │    api    │ │  client  │ │    db     │
        │ (private) │ │ (private)│ │ (private) │
        │  Go API   │ │  worker  │ │ PostgreSQL│
        └─────┬─────┘ └──────────┘ └───────────┘
              │                    polls api directly
        ┌─────▼─────┐
        │    db     │
        │ PostgreSQL│
        └───────────┘
```

**Port-exposure rule (enforced by a test in `internal/deploy/`):**

| Service   | Host port | Reachable from            |
|-----------|-----------|---------------------------|
| `frontend`| **yes**   | host + internal network   |
| `api`     | no        | internal network only     |
| `client`  | no        | internal network only     |
| `db`      | no        | internal network only     |

Only the frontend binds to the host. The API, the data-plane worker, and the database are reachable exclusively over the internal bridge network, addressed by service name (`http://api:8080`, `db:5432`).

---

## Quick start (containers)

```sh
docker compose up --build      # build + start all four apps
docker compose ps             # status
docker compose logs -f client # watch the data plane poll
docker compose down           # stop (keeps the db volume)
docker compose down -v        # stop and wipe the db volume
```

Then open **http://localhost:8080** — that is the dashboard, and the only published port.

Create a flag in the UI, and watch the data-plane client pick it up within one poll interval, with no restart of anything.

### Images

One Dockerfile, three targets:

```sh
docker build --target api      -t flaggo/api:dev .
docker build --target frontend -t flaggo/frontend:dev .
docker build --target client   -t flaggo/client:dev .
```

The `db` service is the upstream `postgres:16-alpine` image — not built from this repo.

---

## Local development (no container required)

You do **not** need Postgres installed on your host to develop the app. Use the in-memory store:

```sh
# 1. Control Plane API, in-memory store
STORE=memory go run ./cmd/flaggo

# 2. Frontend + gateway (separate terminal)
BACKEND_URL=http://127.0.0.1:8080 go run ./cmd/flaggo-web

# 3. Data plane client (separate terminal)
CONTROL_PLANE_URL=http://127.0.0.1:8080 go run ./examples/client
```

Open http://localhost:8080. Flags you create live in memory and vanish on restart — that is the point of the dev mode.

With Postgres instead: set `DATABASE_URL` and leave `STORE=postgres` (the default).

---

## Configuration

### Control Plane API (`cmd/flaggo`)

| Variable           | Default   | Description                                                    |
|--------------------|-----------|----------------------------------------------------------------|
| `PORT`             | `8080`    | HTTP listen port                                               |
| `LOG_LEVEL`        | `info`    | debug / info / warn / error                                    |
| `STORE`            | `postgres`| `postgres` or `memory` (in-memory = development)              |
| `DATABASE_URL`     | (empty)   | Postgres DSN; required when `STORE=postgres`                   |
| `SHUTDOWN_TIMEOUT` | `10s`     | Graceful shutdown timeout                                      |
| `READ_TIMEOUT`     | `15s`     | HTTP read timeout                                              |
| `WRITE_TIMEOUT`    | `15s`     | HTTP write timeout                                             |

### Frontend / gateway (`cmd/flaggo-web`)

| Variable        | Default             | Description                                    |
|-----------------|---------------------|------------------------------------------------|
| `PORT`          | `8080`              | HTTP listen port (the published port)          |
| `BACKEND_URL`   | `http://api:8080`   | Control Plane API base URL                     |
| `LOG_LEVEL`     | `info`              | debug / info / warn / error                    |
| `HEALTHCHECK_URL` | `http://127.0.0.1:$PORT/healthz` | Override the probe target       |

### Data plane client (`examples/client`)

| Variable            | Default                 | Description                     |
|---------------------|-------------------------|---------------------------------|
| `CONTROL_PLANE_URL` | `http://localhost:8080` | API base URL                    |
| `POLL_INTERVAL`     | `2s`                    | Polling interval                |
| `USER_KEY`          | `demo-user`             | User key for rollout evaluation |
| `FLAG_NAME`         | `new-ui`                | Flag to evaluate                |

---

## API

All flag routes are served by the API on its internal port and proxied by the gateway under `/api`.

| Method | Path                            | Description                    |
|--------|---------------------------------|--------------------------------|
| GET    | `/healthz`                      | Liveness probe                 |
| GET    | `/readyz`                       | Readiness probe (DB reachable) |
| GET    | `/metrics`                      | Prometheus metrics             |
| POST   | `/flags`                        | Create a flag                  |
| GET    | `/flags`                        | List flags                     |
| GET    | `/flags/{name}`                 | Get a flag                     |
| PUT    | `/flags/{name}`                 | Update a flag                  |
| DELETE | `/flags/{name}`                 | Delete a flag                  |
| POST   | `/flags/{name}/kill-switch`     | Toggle / set kill switch       |
| PUT    | `/flags/{name}/rollout`         | Set rollout percentage         |

```sh
curl -X POST localhost:8080/api/flags -d '{"name":"new-checkout","enabled":true,"rollout_percent":25}'
curl -X POST localhost:8080/api/flags/new-checkout/kill-switch
curl -X PUT  localhost:8080/api/flags/new-checkout/rollout -d '{"rollout_percent":50}'
```

---

## Commands

Both binaries take a subcommand:

| Command            | Description                                                                 |
|--------------------|-----------------------------------------------------------------------------|
| `serve`            | Run the service (default when no command is given)                         |
| `healthcheck`      | Probe `/healthz`; exit 0 = healthy, 1 = unhealthy. For container `HEALTHCHECK` in shell-less base images |
| `version`          | Print the build version (set via `-ldflags "-X main.version=..."`)          |

---

## Build & test

```sh
go build ./...
go vet ./...
go test ./...
```

The test suite covers flag evaluation, the REST handlers, the frontend↔API gateway contract (including request-id propagation and 502/503 behaviour), and the deployment policies (4 services, one published port).

---

## License

TBD
