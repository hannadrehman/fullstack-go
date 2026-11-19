// Relative base path: nginx proxies it in prod, the Vite dev server in dev.
const BASE = import.meta.env.VITE_API_URL || '/api'

// Tiny helpers so every request is built the same way.
const jsonHeaders = { 'Content-Type': 'application/json' }
const post = (body) => ({ method: 'POST', body: JSON.stringify(body) })

// fetch, but make it fancy: one place for headers, JSON parsing and errors.
async function req(path, options = {}) {
  const res = await fetch(`${BASE}${path}`, {
    headers: jsonHeaders,
    ...options,
  })

  if (!res.ok) {
    // The server sends {"error": "..."}; fall back to a status message if it doesn't.
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `Request failed (${res.status})`)
  }

  return res.json()
}

// Endpoint table. Each entry is one call to req() with the same path and method as before.
export const api = {
  // Users: the friends list and their wall posts
  listUsers: () => req('/users'),
  getUser: (id) => req(`/users/${id}`),
  listUserScraps: (id) => req(`/users/${id}/scraps`),
  createUser: (user) => req('/users', post(user)),

  // Scraps: the wall itself
  listScraps: () => req('/scraps'),
  createScrap: (scrap) => req('/scraps', post(scrap)),

  // Replies hang off a scrap, like a comment thread
  listReplies: (scrapId) => req(`/scraps/${scrapId}/replies`),
  createReply: (scrapId, reply) => req(`/scraps/${scrapId}/replies`, post(reply)),
}
