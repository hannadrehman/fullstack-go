# fullstack-go

An Orkut-style scrapbook portal for sharing short messages and conversations in a shared feed. A “scrap” is a text post that other members can reply to, bringing the feel of a classic social scrapbook to a simple web app.

The project combines a Go HTTP API, a React interface, and PostgreSQL storage. Docker Compose runs all three services together, while the frontend and backend can also be run separately for local development. There is no authentication: registration creates a member profile, and anyone can select an existing member when posting or replying. All APIs are intentionally public.

## Features

- **Member profiles:** Register with a name, email address, and optional avatar emoji, and browse the member list.
- **Shared scrapbook feed:** Post text scraps and view the latest posts with author names, relative timestamps, and reply counts.
- **Conversations:** Expand a scrap to read its replies and add a response as any registered member.
- **Persistent storage:** Store members, scraps, and replies in PostgreSQL; the Compose setup keeps database data in a named volume.
- **Public JSON API:** Create and retrieve members, scraps, and replies through the same API used by the frontend.

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

## Using the app

1. Open http://localhost:3000 after starting the stack.
2. Use **Register** to create a member with a name, email address, and optional avatar emoji.
3. Write a message under **Post a scrap**, choose the member who is posting, and select **Post scrap**.
4. Select a scrap’s reply count to expand the conversation. Enter a response, choose a member, and select **Reply**.

The member selectors let you try conversations between multiple profiles without signing in or switching accounts.

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
