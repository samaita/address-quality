#!/usr/bin/env bash
set -euo pipefail

SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="$SELF_DIR/deploy/docker-compose.prod.yml"
CONFIG_DIR="/etc/address-quality"
ENV_FILE="$CONFIG_DIR/.env.prod"
RELEASE_FILE="$CONFIG_DIR/.release"
IMAGE="ghcr.io/samaita/address-quality"

API_IMAGE_TAG="${API_IMAGE_TAG:-latest}"

usage() {
  cat <<EOF
Usage: deploy.sh [--rollback]

Deploys the Address Quality stack (api) with podman compose.

  --rollback   Redeploy the last successfully recorded image tags.

Environment:
  API_IMAGE_TAG   image tag to deploy (default: latest)
  GHCR_USERNAME / GHCR_TOKEN     optional, needed for private packages

Prerequisites on the VPS:
  podman + a compose provider (podman-compose >= 1.0.4 or docker-compose)
  curl
  $ENV_FILE  (backend env, copied from deploy/.env.prod.example)
  $CONFIG_DIR owned by the user running this script
  $CONFIG_DIR/db populated manually with address.db + location.db
EOF
}

[ -f "$ENV_FILE" ] || { echo "FATAL: missing $ENV_FILE (see deploy/.env.prod.example)" >&2; exit 1; }
command -v podman >/dev/null || { echo "FATAL: podman not installed" >&2; exit 1; }
command -v curl >/dev/null || { echo "FATAL: curl not installed" >&2; exit 1; }
[ -w "$CONFIG_DIR" ] || { echo "FATAL: $CONFIG_DIR not writable by $(id -un) (chown it to the deploy user)" >&2; exit 1; }

case "${1:-}" in
  --rollback) MODE="rollback" ;;
  -h|--help) usage; exit 0 ;;
  "") MODE="deploy" ;;
  *) usage; exit 2 ;;
esac

# Serialize concurrent deploys (best effort)
if command -v flock >/dev/null 2>&1; then
  exec 9>"$CONFIG_DIR/.deploy.lock"
  flock 9 || { echo "Another deploy is in progress; aborting." >&2; exit 1; }
fi

# Optional auth for private GHCR packages
if [ -n "${GHCR_TOKEN:-}" ]; then
  podman login --username "${GHCR_USERNAME:-deploy}" --password "$GHCR_TOKEN" ghcr.io >/dev/null
fi

if [ "$MODE" = "rollback" ]; then
  [ -f "$RELEASE_FILE" ] || { echo "FATAL: no previous release recorded ($RELEASE_FILE)" >&2; exit 1; }
  API_IMAGE_TAG="$(awk '$1=="api" {print $2}' "$RELEASE_FILE")"
  [ -n "$API_IMAGE_TAG" ] || { echo "FATAL: corrupt release file" >&2; exit 1; }
  echo "==> Rollback to api:${API_IMAGE_TAG}"
fi

export API_IMAGE_TAG

echo "==> Pulling api:${API_IMAGE_TAG}"
podman compose -f "$COMPOSE_FILE" pull

echo "==> (Re)creating containers..."
podman compose -f "$COMPOSE_FILE" up -d --force-recreate --no-build --remove-orphans

echo "==> Waiting for API health..."
status=""
for _ in {1..30}; do
  status="$(podman inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' address-quality-api 2>/dev/null || true)"
  [ "$status" = "healthy" ] && break
  sleep 2
done
if [ "$status" != "healthy" ]; then
  echo "ERROR: API not healthy (status: ${status:-unknown})" >&2
  podman compose -f "$COMPOSE_FILE" logs --tail=50 api
  exit 1
fi

echo "==> Verifying API reverse proxy..."
curl -sf http://127.0.0.1/address-quality/health >/dev/null \
  || { echo "ERROR: nginx -> api proxy failed" >&2; exit 1; }

echo "==> Recording release..."
printf 'api %s\n' "$API_IMAGE_TAG" > "$RELEASE_FILE"

echo "==> Pruning dangling images..."
podman image prune -f >/dev/null || true

echo "==> Deployed:"
podman images "$IMAGE" --format '{{.Repository}}:{{.Tag}}  {{.ID}}  {{.CreatedSince}}'
