#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

COMPOSE_FILE="docker-compose.prod.yml"

echo "==> Pulling latest image..."
podman compose -f "$COMPOSE_FILE" pull

echo "==> Restarting container..."
podman compose -f "$COMPOSE_FILE" up -d --remove-orphans

echo "==> Health check..."
sleep 2
for i in {1..10}; do
  if curl -sf http://localhost:7300/health >/dev/null 2>&1; then
    echo "Healthy"
    break
  fi
  if [ "$i" -eq 10 ]; then
    echo "Health check failed" >&2
    podman compose -f "$COMPOSE_FILE" logs --tail=50
    exit 1
  fi
  sleep 2
done

echo "==> Pruning dangling images..."
podman image prune -f

echo "==> Done. Current image:"
podman images ghcr.io/samaita/address-quality --format '{{.ID}}  {{.Tag}}  {{.CreatedSince}}'
