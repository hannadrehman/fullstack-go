# fullstack-go

An Orkut-style scrapbook portal. Register users, post scraps, and reply to anyone's scrap — no authentication, all APIs are public.

## Stack

| Layer    | Tech                                   |
|----------|----------------------------------------|
| Backend  | Go 1.24 (net/http + httprouter + lib/pq) |
| Frontend | React 19 + Vite (served by nginx)      |
| Database | PostgreSQL 16                          |
| Infra    | Docker Compose                        |

## Project layout

```
├── cmd/server/        # Go backend (main.go, db.go, handler.go)
├── client/            # React + Vite frontend
│   ├── src/           # App, api client, styles
│   ├── Dockerfile     # build → nginx serve
│   └── nginx.conf     # proxies /api → server:8080
├── Dockerfile         # Go server multi-stage build
├── docker-compose.yml # db + server + client
├── go.mod / go.sum
└── AGENTS.md          # instructions for AI coding agents
```

## Quick start (Docker — recommended)

Requires Docker Desktop (or docker engine + compose plugin).

```bash
docker compose up --build -d
```

- **Frontend**: http://localhost:3000
- **API**: http://localhost:8080/api
- **Postgres**: localhost:5432 (`postgres` / `postgres`, db `scraps`)

To stop:

```bash
docker compose down        # keep data
docker compose down -v     # also wipe the database
```

## Local development (without Docker)

1. Start a Postgres instance, e.g.:

   ```bash
   docker run -d --name pg -p 5432:5432 \
     -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres \
     -e POSTGRES_DB=scraps postgres:16-alpine
   ```

2. Run the Go API (auto-migrates schema on startup):

   ```bash
   go run ./cmd/server
   # listens on :8080, expects DATABASE_URL or defaults to
   # postgres://postgres:postgres@localhost:5432/scraps?sslmode=disable
   ```

3. Run the Vite dev server (hot reload, proxies `/api` → `:8080`):

   ```bash
   cd client
   npm install
   npm run dev            # http://localhost:3000
   ```

## Environment variables

| Variable       | Default                                                          | Used by |
|----------------|-------------------------------------------------------------------|---------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/scraps?sslmode=disable` | server |
| `ADDR`         | `:8080`                                                           | server  |
| `API_URL`      | `http://localhost:8080`                                           | Vite proxy (dev only) |

## API reference

| Method | Path                        | Body                                     | Description |
|--------|-----------------------------|------------------------------------------|-------------|
| GET    | `/api/health`               | —                                        | Health check |
| POST   | `/api/users`                | `{name, email, avatar}`                  | Register (upsert on email) |
| GET    | `/api/users`                | —                                        | List users |
| GET    | `/api/users/:id`            | —                                        | Get user |
| POST   | `/api/scraps`               | `{user_id, content}`                      | Post a scrap |
| GET    | `/api/scraps?limit=N`       | —                                        | List scraps (default 50, max 100) |
| GET    | `/api/scraps/:id`           | —                                        | Get scrap |
| POST   | `/api/scraps/:id/replies`   | `{user_id, content}`                     | Reply to a scrap |
| GET    | `/api/scraps/:id/replies`   | —                                        | List replies |

All endpoints are public — there is no auth by design.

## Database schema

Three tables, auto-created on server startup (idempotent `CREATE TABLE IF NOT EXISTS`):
`users`, `scraps` (→ users), `replies` (→ scraps, users). See `cmd/server/db.go`.

## Build & checks

```bash
go build ./... && go vet ./...   # backend
cd client && npm run build       # frontend production build
```