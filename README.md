# Flaggo

A lightweight, self-hosted Feature Flag & Runtime Configuration Platform. Built to demonstrate containerization and modern CI/CD patterns.

Flaggo decouples **deployment** from **release**: a centralized Control Plane (this service) exposes an API to create flags, set rollout percentages, and trigger instant kill switches. Client applications (Data Plane) poll the API and evaluate flags in memory — no restarts, no redeployments.

> Dockerfiles and deployment manifests are intentionally **not** included — the operator owns deployment. The app is structured as a good operability example: structured JSON logs, Prometheus metrics, health/readiness probes, graceful shutdown, 12-factor config.

## Configuration (environment variables)

| Variable            | Default     | Description                                  |
|---------------------|-------------|----------------------------------------------|
| `PORT`              | `8080`      | HTTP listen port                             |
| `LOG_LEVEL`         | `info`      | debug / info / warn / error                  |
| `DATABASE_URL`      | (empty)     | Postgres connection string; if empty, degraded mode (readyz not ready, no flag routes) |
| `SHUTDOWN_TIMEOUT`  | `10s`       | Graceful shutdown timeout                    |
| `READ_TIMEOUT`      | `15s`      | HTTP read timeout                            |
| `WRITE_TIMEOUT`     | `15s`      | HTTP write timeout                           |

Client example (`examples/client`):

| Variable             | Default                  | Description                      |
|----------------------|--------------------------|----------------------------------|
| `CONTROL_PLANE_URL`  | `http://localhost:8080`  | Control plane base URL           |
| `POLL_INTERVAL`      | `2s`                     | Polling interval                 |
| `USER_KEY`           | `demo-user`              | User key for rollout evaluation  |
| `FLAG_NAME`          | `new-ui`                 | Flag to evaluate                 |

## Run

```sh
# 1. Start Postgres (operator's responsibility), then:
export DATABASE_URL=postgres://user:pass@localhost:5432/flaggo?sslmode=disable
go run ./cmd/flaggo

# 2. In another terminal, run the data-plane example client:
CONTROL_PLANE_URL=http://localhost:8080 go run ./examples/client
```

## API

| Method | Path                              | Description                          |
|--------|-----------------------------------|--------------------------------------|
| GET    | `/healthz`                        | Liveness probe                       |
| GET    | `/readyz`                         | Readiness probe (DB connectivity)    |
| GET    | `/metrics`                        | Prometheus metrics                   |
| POST   | `/flags`                          | Create a flag                        |
| GET    | `/flags`                          | List flags                           |
| GET    | `/flags/{name}`                   | Get a flag                           |
| PUT    | `/flags/{name}`                   | Update a flag                        |
| DELETE | `/flags/{name}`                   | Delete a flag                        |
| POST   | `/flags/{name}/kill-switch`       | Toggle / set kill switch             |
| PUT    | `/flags/{name}/rollout`           | Set rollout percentage               |

### Example walkthrough

```sh
# create
curl -X POST localhost:8080/flags -d '{"name":"new-ui","enabled":true,"rollout_percent":25}'
# toggle kill switch
curl -X POST localhost:8080/flags/new-ui/kill-switch
# set rollout
curl -X PUT localhost:8080/flags/new-ui/rollout -d '{"rollout_percent":50}'
```

## Commands

| Command | Description |
|---------|-------------|
| `flaggo serve` | Run the control plane (default when no command given) |
| `flaggo healthcheck` | Probe `/healthz`; exit 0 = healthy, 1 = unhealthy. Designed for container `HEALTHCHECK` in shell-less base images |
| `flaggo version` | Print the build version (set via `-ldflags "-X main.version=..."`) |

`flaggo healthcheck` targets `http://127.0.0.1:$PORT/healthz` by default; override with `HEALTHCHECK_URL`.

## Build & test

```sh
go build ./...
go vet ./...
go test ./...
```

## License

TBD
