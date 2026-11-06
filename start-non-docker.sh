#!/usr/bin/env sh
#
# start.sh — run the Flaggo stack in the background.
#
#   ./start.sh          start backend + frontend + client, then return to the shell
#   ./start.sh stop     stop everything
#
# The fourth app, the database, is the API's in-memory store here
# (STORE=memory), so it needs no process of its own. Point DATABASE_URL at a
# real Postgres and set STORE=postgres to use one instead.
#
# Binaries are built into $RUN_DIR and run directly (not via `go run`) so the
# recorded PIDs are the actual processes and `stop` is reliable.

set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$ROOT"

RUN_DIR="${TMPDIR:-/tmp}/opencode"
mkdir -p "$RUN_DIR"

BIN_API="$RUN_DIR/flaggo"
BIN_WEB="$RUN_DIR/flaggo-web"
BIN_CLI="$RUN_DIR/flaggo-client"
PID_FILE="$RUN_DIR/flaggo.pids"

PORT_API="${PORT_API:-8080}"
PORT_WEB="${PORT_WEB:-8090}"
POLL_INTERVAL="${POLL_INTERVAL:-1s}"

# wait_ready <url> — block until <url> answers 2xx, or ~10s elapse.
# Keeps the client from logging a spurious "connection refused" on boot.
wait_ready() {
  if ! command -v curl >/dev/null 2>&1; then
    sleep 1
    return 0
  fi
  i=0
  while [ "$i" -lt 50 ]; do
    if curl -fsS -o /dev/null --max-time 1 "$1" 2>/dev/null; then
      return 0
    fi
    sleep 0.2
    i=$((i + 1))
  done
  echo "warning: $1 did not become ready in 10s" >&2
  return 1
}

start_all() {
  echo "building binaries into $RUN_DIR ..."
  go build -o "$BIN_API" ./cmd/flaggo
  go build -o "$BIN_WEB" ./cmd/flaggo-web
  go build -o "$BIN_CLI" ./examples/client

  : > "$PID_FILE"

  echo "starting ..."

  # 1. Control Plane Backend API (+ in-memory database)
  STORE=memory PORT="$PORT_API" "$BIN_API" serve \
    >"$RUN_DIR/flaggo-api.log" 2>&1 &
  echo "api $!" >> "$PID_FILE"
  wait_ready "http://127.0.0.1:$PORT_API/readyz" || true

  # 2. Control Plane Frontend + /api gateway
  PORT="$PORT_WEB" BACKEND_URL="http://127.0.0.1:$PORT_API" "$BIN_WEB" serve \
    >"$RUN_DIR/flaggo-web.log" 2>&1 &
  echo "web $!" >> "$PID_FILE"
  wait_ready "http://127.0.0.1:$PORT_WEB/healthz" || true

  # 3. Data Plane client
  CONTROL_PLANE_URL="http://127.0.0.1:$PORT_API" POLL_INTERVAL="$POLL_INTERVAL" "$BIN_CLI" \
    >"$RUN_DIR/flaggo-client.log" 2>&1 &
  echo "client $!" >> "$PID_FILE"

  sleep 0.5

  printf '\n'
  printf '  frontend   http://localhost:%s\n' "$PORT_WEB"
  printf '  api        http://127.0.0.1:%s\n' "$PORT_API"
  printf '  client     %s/flaggo-client.log\n' "$RUN_DIR"
  printf '  pids       %s\n' "$PID_FILE"
  printf '\n  stop with: ./start.sh stop\n\n'
}

stop_all() {
  if [ ! -f "$PID_FILE" ]; then
    echo "nothing recorded in $PID_FILE"
    return 0
  fi
  # SIGTERM first so the servers exercise graceful shutdown.
  while read -r name pid; do
    [ -n "${pid:-}" ] || continue
    if kill "$pid" 2>/dev/null; then
      echo "stopped $name (pid $pid)"
    fi
  done < "$PID_FILE"

  # Give them up to 5s to exit; SIGKILL anything still standing.
  i=0
  while [ "$i" -lt 25 ]; do
    alive=0
    while read -r name pid; do
      [ -n "${pid:-}" ] || continue
      if kill -0 "$pid" 2>/dev/null; then alive=1; fi
    done < "$PID_FILE"
    if [ "$alive" -eq 0 ]; then break; fi
    sleep 0.2
    i=$((i + 1))
  done

  while read -r name pid; do
    [ -n "${pid:-}" ] || continue
    if kill -0 "$pid" 2>/dev/null; then
      if kill -9 "$pid" 2>/dev/null; then
        echo "killed $name (pid $pid) after timeout"
      fi
    fi
  done < "$PID_FILE"

  rm -f "$PID_FILE"
}

case "${1:-start}" in
  start) start_all ;;
  stop)  stop_all ;;
  *)
    echo "usage: $0 [start|stop]" >&2
    exit 2
    ;;
esac
