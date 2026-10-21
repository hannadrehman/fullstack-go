const BASE = import.meta.env.VITE_API_URL || '/api'

async function req(path, options = {}) {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `Request failed (${res.status})`)
  }
  return res.json()
}

export const api = {
  listUsers: () => req('/users'),
  createUser: (user) =>
    req('/users', { method: 'POST', body: JSON.stringify(user) }),
  listScraps: () => req('/scraps'),
  createScrap: (scrap) =>
    req('/scraps', { method: 'POST', body: JSON.stringify(scrap) }),
  listReplies: (scrapId) => req(`/scraps/${scrapId}/replies`),
  createReply: (scrapId, reply) =>
    req(`/scraps/${scrapId}/replies`, {
      method: 'POST',
      body: JSON.stringify(reply),
    }),
}