# Scrapbook frontend

React and Vite frontend for the public scrapbook portal. All commands below run
from the repository root. See the root README for the shared database and API setup,
including selecting free ports for this worktree.

## Development

Requires Node.js 22.12 or newer and npm. Set `SERVER_PORT` and `CLIENT_PORT` to the
free ports chosen for your API and frontend, then run:

```bash
npm ci --prefix client
API_URL="http://localhost:$SERVER_PORT" npm run dev --prefix client -- --port "$CLIENT_PORT" --strictPort
```

Open `http://localhost:<CLIENT_PORT>` using your selected frontend port.
Vite proxies relative `/api` requests to `API_URL` (default
`http://localhost:8080`). Keep requests in `client/src/api.js`; do not hardcode
API hosts in UI components. `client/src/App.jsx` contains the UI.

## Build and lint

```bash
npm run build --prefix client
npm run lint --prefix client
```

The production build is written to `client/dist/`. The frontend Docker image
serves this build with nginx, which proxies `/api/` to `server:8080` on the Compose
network. `API_URL` configures the Vite development proxy only.

Keep `color-scheme: light` in `client/src/index.css` so form text remains visible.
