# fullstack-go

An Orkut-style scrapbook portal. Register members, post scraps, and reply to anyone's scrap. There is no authentication: all APIs are public.

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

Open the frontend, register a member with a name and email, then choose a member
in the posting dropdown to write a scrap. Click a scrap's reply count to expand
the conversation and add a reply. Anyone can post or reply as any registered member.

Check startup and API availability:

```bash
docker compose ps
docker compose logs server
curl --fail http://localhost:8080/api/health
# {"status":"ok"}
```

To stop:

```bash
docker compose down        # keep data
docker compose down -v     # also wipe the database
```

## Local development

Requires Go 1.24 or newer, Node.js 22.12 or newer with npm, and PostgreSQL 16.
Run commands from the repository root unless a step says otherwise. The API and
Vite server run in separate terminals; the database can run locally or in Docker.

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
| `VITE_API_URL` | `/api`                                                           | frontend API base path (build time) |

The frontend uses relative `/api` requests by default. Vite proxies these requests
in development; nginx proxies them to `server:8080` in Docker. Set `API_URL` when
starting Vite if your local API uses another address:

```bash
API_URL=http://localhost:9090 npm run dev --prefix client
```

Compose sets the server's database URL to the `db` service and its listen address
to `:8080` explicitly in `docker-compose.yml`. Edit that configuration to change
container settings; the table above describes defaults for local development.

## API reference

| Method | Path                        | Body                                     | Description |
|--------|-----------------------------|------------------------------------------|-------------|
| GET    | `/api/health`               | —                                        | Health check |
| POST   | `/api/users`                | `{name, email, avatar}`                  | Register (upsert on email) |
| GET    | `/api/users`                | —                                        | List users |
| GET    | `/api/users/:id`            | —                                        | Get user |
| POST   | `/api/scraps`               | `{user_id, content}`                      | Post a scrap |
| GET    | `/api/scraps?limit=N`       | —                                        | List newest scraps first (1–100; otherwise 50) |
| GET    | `/api/scraps/:id`           | —                                        | Get scrap |
| POST   | `/api/scraps/:id/replies`   | `{user_id, content}`                     | Reply to a scrap |
| GET    | `/api/scraps/:id/replies`   | —                                        | List replies |

All endpoints are public — there is no auth by design.

User creation requires a nonempty name and an email containing `@`; `avatar` is
optional. Emails are trimmed and lowercased. Reusing an email updates the existing
member's name and returns the same ID; it does not update their avatar.
Scraps and replies require an existing member's integer `user_id` and nonempty
content. Replies are listed oldest first. Empty lists currently return JSON `null`.

Successful POST requests return HTTP 201 and the created object, including its
`id`. Handler errors use `{"error":"..."}` with HTTP 400 for invalid input, 404
for missing resources, or 500 for internal errors. The health endpoint reports
that the HTTP server is running; it does not query the database.

### Try the API

Create a member, then use the returned `id` as `user_id` when posting. Replace the
example IDs below with the IDs returned by your requests.

```bash
curl --fail http://localhost:8080/api/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ada","email":"ada@example.com","avatar":"🌻"}'

curl --fail http://localhost:8080/api/scraps \
  -H 'Content-Type: application/json' \
  -d '{"user_id":1,"content":"Hello, scrapbook!"}'

curl --fail http://localhost:8080/api/scraps/1/replies \
  -H 'Content-Type: application/json' \
  -d '{"user_id":1,"content":"Welcome!"}'

curl --fail 'http://localhost:8080/api/scraps?limit=10'
curl --fail http://localhost:8080/api/scraps/1/replies
```

## Database schema

Three tables, auto-created on server startup (idempotent `CREATE TABLE IF NOT EXISTS`):
`users`, `scraps` (→ users), `replies` (→ scraps, users). See `cmd/server/db.go`.

There is no separate migration tool. The schema script runs on every startup and
creates missing tables, but edits to an existing table's definition do not add
columns to that table. Apply an appropriate `ALTER TABLE` manually for existing
databases, or reset disposable local data with `docker compose down -v` before
rebuilding. The reset permanently deletes the Compose database volume.

## Build & checks

```bash
go build ./... && go vet ./...   # backend
npm ci --prefix client          # install locked frontend dependencies
npm run build --prefix client   # frontend production build
docker compose build            # build all container images
```

Before committing, also start the full stack with `docker compose up --build -d`
and exercise health, user creation, scrap creation, and replies using the examples
above.
