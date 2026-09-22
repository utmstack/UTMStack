#!/usr/bin/env bash
# Simulate the CI backend build (.github/workflows/reusable-golang.yml) locally,
# then build the PROD image (backend/Dockerfile, used unchanged) so that
# dev-backend/compose.yml can run it.
#
#   1. Compile the backend to backend/backend — the exact location and -o name
#      the CI step "Build Binary" produces, which backend/Dockerfile COPYs in.
#      The CI runner is native linux/amd64; on macOS we cross-compile with
#      GOOS/GOARCH + CGO_ENABLED=0 (the Postgres driver is pure Go) so the static
#      binary runs in the ubuntu:24.04 prod image.
#   2. docker build with context = repo root and the prod Dockerfile, exactly
#      like reusable-golang.yml's "Build and Push the Image" step
#      (v11 pipeline passes build_context "." + dockerfile ./backend/Dockerfile).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${IMAGE:-utmstack-backend:dev}"

# Optional ldflags — the pipeline injects billing PublicKey/EncryptSalt here.
# Not needed for the migration test; license verification falls back to
# community mode when empty. Override with: FLAGS="-X '...'" ./build.sh
FLAGS="${FLAGS:-}"

echo "==> [1/2] Compiling backend -> backend/backend (linux/amd64)"
cd "$ROOT/backend"
if [ -n "$FLAGS" ]; then
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o backend -v -ldflags "$FLAGS" .
else
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o backend -v .
fi

echo "==> [2/2] Building image $IMAGE from backend/Dockerfile (context = repo root)"
cd "$ROOT"
docker build -f backend/Dockerfile -t "$IMAGE" .

echo "==> Done. Now:  cd dev-backend && docker compose up"
