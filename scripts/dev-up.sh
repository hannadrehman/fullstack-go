#!/usr/bin/env bash
# Start the full stack for THIS worktree without clashing with other worktrees.
#  - compose project name is derived from the worktree dir (own containers + own DB volume)
#  - host ports are picked free and persisted in .env (gitignored), reused on restart
#  - the DB volume is per-project, so a new worktree gets a fresh DB and the server's
#    migrate() applies the full schema on startup
# Usage: scripts/dev-up.sh [up|down|status]   (down -v wipes this worktree's data)
set -euo pipefail
cd "$(dirname "$0")/.."

cmd="${1:-up}"
name="scraps-$(basename "$PWD" | tr 'A-Z' 'a-z' | tr -c 'a-z0-9\n_-' '-' | cut -c1-40)"
export COMPOSE_PROJECT_NAME="$name"
ENV_FILE=.env

used_ports() {
  { ss -ltnH 2>/dev/null | awk '{print $4}' | sed 's/.*://'
    docker ps --format '{{.Ports}}' | grep -oE ':[0-9]+->' | tr -d ':>-'
  } | sort -u
}

port_free() { ! used_ports | grep -qx "$1"; }

pick_port() { # pick_port <preferred>
  local p="$1"
  if port_free "$p" && [[ " ${CHOSEN:-} " != *" $p "* ]]; then echo "$p"; return; fi
  while :; do
    p=$((20000 + RANDOM % 20000))
    if port_free "$p" && [[ " ${CHOSEN:-} " != *" $p "* ]]; then echo "$p"; return; fi
  done
}

running() { [[ -n "$(docker compose ps -q 2>/dev/null)" ]]; }

others() {
  docker ps --filter "label=com.docker.compose.project" \
    --format '{{.Label "com.docker.compose.project"}} {{.Names}} {{.Ports}}' \
    | grep '^scraps-' | grep -v "^$name " || true
}

case "$cmd" in
  down) shift; docker compose down "$@"; exit ;;
  status)
    docker compose ps; echo "--- other worktrees ---"; others; exit ;;
  up) ;;
  *) echo "usage: $0 [up|down|status]"; exit 1 ;;
esac

if [[ -n "$(others)" ]]; then
  echo "Note: the project is also running from other worktrees:"; others | sed 's/^/  /'
fi

# Load saved ports; re-pick only if not already ours and now taken.
[[ -f $ENV_FILE ]] && set -a && . "./$ENV_FILE" && set +a
CHOSEN=""
resolve() { # resolve VAR preferred
  local var="$1" cur="${!1:-}"
  if [[ -n "$cur" ]] && { running || port_free "$cur"; }; then :; else cur="$(pick_port "$2")"; fi
  CHOSEN="$CHOSEN $cur"; printf -v "$var" '%s' "$cur"
}
resolve CLIENT_PORT 3000
resolve SERVER_PORT 8080
resolve DB_PORT 5432

cat > "$ENV_FILE" <<EOT
COMPOSE_PROJECT_NAME=$name
CLIENT_PORT=$CLIENT_PORT
SERVER_PORT=$SERVER_PORT
DB_PORT=$DB_PORT
EOT
export CLIENT_PORT SERVER_PORT DB_PORT

docker compose up --build -d

echo "Waiting for API..."
for _ in $(seq 1 60); do
  { curl -fs "http://localhost:$SERVER_PORT/api/health" || wget -qO- "http://localhost:$SERVER_PORT/api/health"; } >/dev/null 2>&1 && break
  sleep 1
done
echo "Project:  $name"
echo "Frontend: http://localhost:$CLIENT_PORT"
echo "API:      http://localhost:$SERVER_PORT"
echo "Postgres: localhost:$DB_PORT"
