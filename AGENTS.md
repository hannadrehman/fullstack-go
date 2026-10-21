# AGENTS

Instructions for AI coding agents (and humans) working in this repo.

## What this project is

An Orkut-style scrapbook portal. Go backend + React/Vite frontend + Postgres, all orchestrated with Docker Compose. **No auth anywhere** — all APIs are intentionally public and open. Do not add authentication unless explicitly asked.

## Repo layout

- `cmd/server/` — entire Go backend lives in `package main`:
  - `main.go` — env config, routes, HTTP server
  - `db.go` — connection with retry, migrations, models, all SQL queries
  - `handler.go` — HTTP handlers, CORS, JSON errors
- `client/` — React + Vite app (`src/App.jsx` is the whole UI; `src/api.js` is the API client)
- `Dockerfile` — backend build (multi-stage, CGO disabled)
- `client/Dockerfile` — frontend build → nginx; `client/nginx.conf` proxies `/api/` to `server:8080`
- `docker-compose.yml` — `server`, `client` services (per worktree, own host ports)
- `docker-compose.db.yml` — shared Postgres (`scraps-db`, host port 5432) used by ALL worktrees

## How to run

Full stack (preferred). Postgres is **shared by all worktrees**: one container
`scraps-db` on host port 5432. Never create another db container, volume or db port.
The app stack (server + client) is per worktree and **must use free host ports**,
never the fixed 3000/8080 defaults.

```bash
# 1. Shared DB. Idempotent: a no-op if already running (e.g. from another worktree).
docker compose -f docker-compose.db.yml up -d --wait

# 2. Free host ports + unique project name for THIS worktree
free_port() { while :; do p=$((20000 + RANDOM % 20000)); (exec 3<>/dev/tcp/127.0.0.1/$p) 2>/dev/null || { echo $p; return; }; done; }  # bash-only; port is free if connect fails
export SERVER_PORT=$(free_port) CLIENT_PORT=$(free_port)
export COMPOSE_PROJECT_NAME=$(basename "$PWD" | tr 'A-Z' 'a-z' | tr -c 'a-z0-9\n' '-')

# 3. Start the app
docker compose up --build -d
echo "Frontend: http://localhost:$CLIENT_PORT"
echo "API:      http://localhost:$SERVER_PORT/api/health"
```

- **Always tell the user the frontend and API URLs** after starting.
- Later `docker compose` commands in this worktree need the same `COMPOSE_PROJECT_NAME`.
  Recover ports with `docker compose port client 80` / `docker compose port server 8080`.
- Step 1 must come first: the app joins the external network `scraps-shared`, which
  the db compose file creates.

```bash
docker compose down                           # stop this worktree's app only
docker compose -f docker-compose.db.yml down  # stops the SHARED db for everyone; only if asked
```

Local dev:

```bash
# 1. Postgres on localhost:5432 (user/pass postgres/postgres, db scraps)
go run ./cmd/server               # 2. API, auto-migrates, listens :8080
cd client && npm install && npm run dev   # 3. Vite on :3000, proxies /api
```

## Migrations — IMPORTANT

- There is **no migration tool** (no golang-migrate, no goose, no migration files).
- Schema lives in the `migrate()` function in `cmd/server/db.go` as a single
  `CREATE TABLE IF NOT EXISTS` script. **It runs automatically on every server startup.**
- To add a table or column:
  1. Edit the `schema` string in `cmd/server/db.go`.
  2. Prefer additive, backward-compatible changes. `IF NOT EXISTS` means the script
     only applies to **fresh databases** — existing databases will NOT get new columns
     automatically.
  3. For schema changes on an existing local dev DB, either write the `ALTER TABLE` by hand
     (`docker exec -it scraps-db psql -U postgres scraps`) or wipe the shared volume
     (affects every worktree, so ask the user first):
     `docker compose -f docker-compose.db.yml down -v`, then rerun the run steps.
- Never use destructive `DROP` statements in `migrate()` without explicit user approval.

## How to generate / build

```bash
go build ./... && go vet ./...     # backend build + vet (must pass before commit)
go mod tidy                        # after changing Go dependencies
cd client && npm install           # after changing package.json
cd client && npm run build         # frontend production build (must pass before commit)
docker compose build               # build all images without starting
```

Never commit the compiled `server` binary (already in `.gitignore`).

## Env vars

| Variable       | Default                                                            |
|----------------|--------------------------------------------------------------------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/scraps?sslmode=disable` |
| `ADDR`         | `:8080`                                                             |
| `API_URL`      | `http://localhost:8080` (Vite dev proxy target only)               |

## Conventions

- Backend stays a flat `package main` in `cmd/server` — keep it simple, no framework.
- JSON errors: handlers return `{"error": "..."}` with proper status codes via `writeErr`.
- All IDs are integers (`SERIAL`); author names are denormalized into scrap/reply responses via JOIN.
- Frontend fetches go through `client/src/api.js` with the `/api` base path — nginx handles
  the proxy in prod, the Vite dev proxy handles it in dev. Don't hardcode host URLs.
- Frontend styling forces `color-scheme: light` in `index.css` — keep it; dark-mode
  scheme breaks input text visibility.

## Definition of done (before every commit)

1. `go build ./... && go vet ./...` passes.
2. `cd client && npm run build` passes.
3. Full stack verified via the "How to run" steps (shared db + free ports) and a couple of live API calls
   (health, register user, post scrap, reply).
4. Commit and push only when the user asks.