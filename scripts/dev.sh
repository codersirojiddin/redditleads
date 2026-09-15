#!/usr/bin/env bash
# Starts everything needed for local development:
# - Postgres via docker-compose
# - the Go API (auto-migrates on boot)
# - the Go worker
# - the Vite dev server for the frontend
#
# Requires: docker, go 1.24+, node 20+. Run `cp .env.example .env` and fill
# in secrets before using this.
set -euo pipefail
cd "$(dirname "$0")/.."

if [ ! -f .env ]; then
  echo "No .env found. Run: cp .env.example .env  (then fill in secrets)"
  exit 1
fi

echo "Starting Postgres..."
docker compose up -d db

echo "Waiting for Postgres to be healthy..."
until docker compose exec -T db pg_isready -U postgres >/dev/null 2>&1; do
  sleep 1
done

if [ ! -d web/node_modules ]; then
  echo "Installing frontend dependencies..."
  (cd web && npm install)
fi

cleanup() {
  echo "Stopping..."
  kill 0
}
trap cleanup EXIT

echo "Starting API on :8080, worker, and frontend on :3000..."
(export $(grep -v '^#' .env | xargs -0 2>/dev/null || true); go run ./apps/api) &
(export $(grep -v '^#' .env | xargs -0 2>/dev/null || true); go run ./apps/worker) &
(cd web && npm run dev) &

wait
