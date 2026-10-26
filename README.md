# fullstack-go

An Orkut-style scrapbook portal. Register members, post scraps, and reply to anyone's scrap. There is no authentication: all APIs are public.

## Features

- Register members by name and email (re-registering an email updates the name)
- Post scraps as any registered member; newest scraps appear first
- Expand a scrap to read its replies and add your own
- Feed errors are shown in the UI with a Retry button
- Accessible form labels and visible keyboard focus styles

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
├── docker-compose.yml # per-worktree server + client
├── docker-compose.db.yml # shared Postgres
├── go.mod / go.sum
└── AGENTS.md          # instructions for AI coding agents
```

## Quick start (Docker — recommended)

Requires Docker Desktop (or docker engine + compose plugin).

Run these commands in Bash from the repository root:

```bash
# Shared by all worktrees; start this before the app network is needed.
docker compose -f docker-compose.db.yml up -d --wait

free_port() { while :; do p=$((20000 + RANDOM % 20000)); (exec 3<>/dev/tcp/127.0.0.1/$p) 2>/dev/null || { echo "$p"; return; }; done; }
export SERVER_PORT=$(free_port)
export CLIENT_PORT=$(free_port)
while [ "$CLIENT_PORT" = "$SERVER_PORT" ]; do export CLIENT_PORT=$(free_port); done
export COMPOSE_PROJECT_NAME=$(basename "$PWD" | tr 'A-Z' 'a-z' | tr -c 'a-z0-9\n' '-')
docker compose up --build -d
echo "Frontend: http://localhost:$CLIENT_PORT"
echo "API: http://localhost:$SERVER_PORT/api/health"
```

Postgres is shared at localhost:5432 (`postgres` / `postgres`, db `scraps`).
Keep the exported variables for later Compose commands in this terminal. With the
same project name, recover ports using `docker compose port client 80` and
`docker compose port server 8080`.

Open the frontend, register a member with a name and email, then choose a member
in the posting dropdown to write a scrap. Click a scrap's reply count to expand
the conversation and add a reply. Anyone can post or reply as any registered member.

Check startup and API availability:

```bash
docker compose ps
docker compose logs server
curl --fail "http://localhost:$SERVER_PORT/api/health"
# {"status":"ok"}
```

To stop:

```bash
docker compose down        # stop only this worktree’s app; shared data stays
```

The shared database serves every worktree. Stop or reset it only with explicit
user approval; deleting its volume permanently deletes data for all worktrees.

## Local development

Requires Go 1.24 or newer, Node.js 22.12 or newer with npm, and PostgreSQL 16.
Run commands from the repository root unless a step says otherwise. The API and
Vite server run in separate terminals and reuse the shared database.

1. Start the existing shared Postgres (never create a second database container):

   ```bash
   docker compose -f docker-compose.db.yml up -d --wait
   ```

2. In Bash, define `free_port` using the function from Quick start, then select
   distinct free ports and run the Go API (auto-migrates on startup):

   ```bash
   export SERVER_PORT=$(free_port)
   export CLIENT_PORT=$(free_port)
   while [ "$CLIENT_PORT" = "$SERVER_PORT" ]; do export CLIENT_PORT=$(free_port); done
   echo "Frontend: http://localhost:$CLIENT_PORT"
   echo "API: http://localhost:$SERVER_PORT/api/health"
   ADDR=":$SERVER_PORT" go run ./cmd/server
   ```

3. In a second terminal, set `SERVER_PORT` and `CLIENT_PORT` to the values
   printed above. Start Vite with the matching API proxy target:

   ```bash
   npm ci --prefix client
   API_URL="http://localhost:$SERVER_PORT" npm run dev --prefix client -- --port "$CLIENT_PORT" --strictPort
   ```

The API uses the shared database on localhost:5432 by default. If a selected port
becomes occupied before startup, select another free port and update both commands.

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

Compose sets the server's database URL to the shared `scraps-db` container and its listen address
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
curl --fail http://localhost:$SERVER_PORT/api/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ada","email":"ada@example.com","avatar":"🌻"}'

curl --fail http://localhost:$SERVER_PORT/api/scraps \
  -H 'Content-Type: application/json' \
  -d '{"user_id":1,"content":"Hello, scrapbook!"}'

curl --fail http://localhost:$SERVER_PORT/api/scraps/1/replies \
  -H 'Content-Type: application/json' \
  -d '{"user_id":1,"content":"Welcome!"}'

curl --fail "http://localhost:$SERVER_PORT/api/scraps?limit=10"
curl --fail http://localhost:$SERVER_PORT/api/scraps/1/replies
```

## Database schema

Three tables, auto-created on server startup (idempotent `CREATE TABLE IF NOT EXISTS`):
`users`, `scraps` (→ users), `replies` (→ scraps, users). See `cmd/server/db.go`.

There is no separate migration tool. The schema script runs on every startup and
creates missing tables, but edits to an existing table's definition do not add
columns to that table. Apply an appropriate `ALTER TABLE` manually for existing
databases. Resetting the shared database volume requires explicit user approval
because it permanently deletes data used by every worktree.

## Build & checks

```bash
go build ./... && go vet ./...   # backend
npm ci --prefix client          # install locked frontend dependencies
npm run lint --prefix client    # frontend lint (oxlint)
npm run build --prefix client   # frontend production build
docker compose build            # build all container images
```

Before committing, start the full stack using the shared database and free ports
in Quick start, then exercise health, user creation, scrap creation, and replies using the examples
above.

## License

MIT. See [LICENSE](LICENSE).
