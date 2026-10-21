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
- `docker-compose.yml` — `db`, `server`, `client` services

## How to run

Full stack (preferred):

```bash
docker compose up --build -d      # frontend :3000, api :8080, db :5432
docker compose down               # stop; add -v to wipe data
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
     (`docker compose exec db psql -U postgres scraps`) or wipe the volume:
     `docker compose down -v && docker compose up --build -d`.
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
3. Full stack verified with `docker compose up --build -d` and a couple of live API calls
   (health, register user, post scrap, reply).
4. Commit and push only when the user asks.
## Multiple worktrees / port handling

Several worktrees may run this stack at once. **Use `scripts/dev-up.sh` instead of raw
`docker compose up`** when asked to run the project:

- `scripts/dev-up.sh` (or `up`) — derives a per-worktree compose project name
  (`scraps-<dir>`), so containers and the `pgdata` volume are isolated per worktree.
  It picks free host ports (prefers 3000/8080/5432, otherwise random), saves them to
  `.env` (gitignored), reuses them on restart, starts the stack, and prints the URLs.
  It also lists the project running from other worktrees.
- `scripts/dev-up.sh status` / `down` (`down -v` wipes this worktree's DB).
- Compose host ports come from `CLIENT_PORT`, `SERVER_PORT`, `DB_PORT` (defaults
  3000/8080/5432). Container-internal ports never change, so nginx's `server:8080`
  proxy is unaffected.
- Each worktree has its own fresh DB; `migrate()` applies the full schema on server
  startup. Existing DBs still need manual `ALTER TABLE` (see Migrations).
- When reporting to the user, always use the ports printed by the script, not 3000/8080.
