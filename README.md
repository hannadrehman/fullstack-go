# fullstack-go

Orkut-style scrapbook: register users, post scraps, reply to scraps. No auth — everything is public.

## Stack
- **Backend**: Go (net/http + httprouter + lib/pq), port `8080`
- **Frontend**: React (Vite), served by nginx, port `3000`
- **Database**: PostgreSQL 16

## Run with Docker

```bash
docker compose up --build
```

- Frontend: http://localhost:3000
- API: http://localhost:8080/api

## Local development

```bash
# terminal 1 — API (needs postgres on localhost:5432)
go run ./cmd/server

# terminal 2 — Vite dev server (proxies /api to :8080)
cd client && npm install && npm run dev
```

## API

| Method | Path | Description |
|--------|------|-------------|
| GET    | `/api/health` | Health check |
| POST   | `/api/users` | Register `{name, email, avatar}` |
| GET    | `/api/users` | List users |
| GET    | `/api/users/:id` | Get user |
| POST   | `/api/scraps` | Post scrap `{user_id, content}` |
| GET    | `/api/scraps` | List scraps (with reply counts) |
| GET    | `/api/scraps/:id` | Get scrap |
| POST   | `/api/scraps/:id/replies` | Reply `{user_id, content}` |
| GET    | `/api/scraps/:id/replies` | List replies |