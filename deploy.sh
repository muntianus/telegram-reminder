#!/usr/bin/env bash
set -euo pipefail
trap 'echo "❌ Error on line $LINENO" >&2' ERR

LOG_LINES=${LOG_LINES:-100}

show_env() {
  docker exec telegram-reminder env | grep -E "(OPENAI|TELEGRAM)" || true
}

die() { echo "❌ $1" >&2; exit 1; }

[[ -f .env ]] || die ".env not found"
grep -q '^[[:space:]]' .env && die "Strip leading spaces in .env"

docker compose pull --quiet
docker compose up -d --force-recreate
sleep 2
show_env

# wait up to 30s for healthy container
for _ in {1..15}; do
  status=$(docker inspect -f '{{.State.Health.Status}}' telegram-reminder 2>/dev/null || echo none)
  if [[ "$status" == "healthy" ]]; then
    echo "🎉 Deploy successful"
    exit 0
  fi
  if [[ "$status" == "unhealthy" ]]; then
    echo "🔍 Container unhealthy, last $LOG_LINES lines:"
    docker logs --timestamps --tail "$LOG_LINES" telegram-reminder || true
    exit 1
  fi
  sleep 2
done

echo "⌛ Timed out waiting for container health"
docker logs --timestamps --tail "$LOG_LINES" telegram-reminder || true
exit 1
