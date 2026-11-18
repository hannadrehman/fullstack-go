import { useEffect, useState } from 'react'
import { api } from './api.js'
import './App.css'

function timeAgo(iso) {
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  return `${Math.floor(hours / 24)}d ago`
}

function RegisterForm({ onRegistered }) {
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [avatar, setAvatar] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function submit(e) {
    e.preventDefault()
    setBusy(true)
    setError('')
    try {
      const user = await api.createUser({ name, email, avatar })
      setName('')
      setEmail('')
      setAvatar('')
      onRegistered(user)
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <form className="card register-form" onSubmit={submit}>
      <h2>Register</h2>
      <input
        placeholder="Name"
        value={name}
        onChange={(e) => setName(e.target.value)}
        required
      />
      <input
        type="email"
        placeholder="Email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        required
      />
      <input
        placeholder="Avatar emoji (optional)"
        maxLength={4}
        value={avatar}
        onChange={(e) => setAvatar(e.target.value)}
      />
      {error && <p className="error">{error}</p>}
      <button disabled={busy}>{busy ? 'Registering…' : 'Register'}</button>
    </form>
  )
}

const WHO_AM_I_ERROR = 'Pick who you are under "I am" first.'

// Posts as the current member. With a recipient, the scrap lands on their scrapbook.
function ScrapForm({ currentUser, recipient, onPosted }) {
  const [content, setContent] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function submit(e) {
    e.preventDefault()
    if (!currentUser) {
      setError(WHO_AM_I_ERROR)
      return
    }
    setBusy(true)
    setError('')
    try {
      await api.createScrap({
        user_id: currentUser.id,
        to_user_id: recipient?.id ?? null,
        content,
      })
      setContent('')
      onPosted()
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <form className="card scrap-form" onSubmit={submit}>
      <h2>{recipient ? `Write on ${recipient.name}'s scrapbook` : 'Post a scrap'}</h2>
      <textarea
        placeholder={recipient ? `Say something to ${recipient.name}…` : 'Write something for everyone…'}
        rows={3}
        value={content}
        onChange={(e) => setContent(e.target.value)}
        required
      />
      <div className="row">
        <span className="muted">
          {currentUser ? `as ${currentUser.name}` : 'No one selected'}
        </span>
        <button disabled={busy}>{busy ? 'Posting…' : 'Post'}</button>
      </div>
      {error && <p className="error">{error}</p>}
    </form>
  )
}

function ReplyForm({ scrapId, currentUser, onReplied }) {
  const [content, setContent] = useState('')
  const [error, setError] = useState('')

  async function submit(e) {
    e.preventDefault()
    if (!currentUser) {
      setError(WHO_AM_I_ERROR)
      return
    }
    setError('')
    try {
      await api.createReply(scrapId, { user_id: currentUser.id, content })
      setContent('')
      onReplied()
    } catch (err) {
      setError(err.message)
    }
  }
// note: midpoint of this file

  return (
    <form className="reply-form" onSubmit={submit}>
      <input
        placeholder="Write a reply…"
        value={content}
        onChange={(e) => setContent(e.target.value)}
        required
      />
      <button>Reply{currentUser ? ` as ${currentUser.name}` : ''}</button>
      {error && <p className="error">{error}</p>}
    </form>
  )
}

function Scrap({ scrap, currentUser, onReply, onProfile }) {
  const [replies, setReplies] = useState(null)
  const [showReplies, setShowReplies] = useState(false)

  async function loadReplies() {
    setShowReplies((s) => !s)
    if (replies === null) setReplies(await api.listReplies(scrap.id))
  }

  return (
    <div className="card scrap">
      <div className="scrap-header">
        <span className="avatar">
          {scrap.author?.[0]?.toUpperCase() || '?'}
        </span>
        <div>
          <button className="link" onClick={() => onProfile(scrap.user_id)}>{scrap.author}</button>
          {scrap.to_user_id && (
            <>
              {' ➜ '}
              <button className="link" onClick={() => onProfile(scrap.to_user_id)}>{scrap.recipient}</button>
            </>
          )}
          <span className="time">{timeAgo(scrap.created_at)}</span>
        </div>
      </div>
      <p className="scrap-content">{scrap.content}</p>
      <button className="link" onClick={loadReplies}>
        {scrap.reply_count} {scrap.reply_count === 1 ? 'reply' : 'replies'}{' '}
        {showReplies ? '▾' : '▸'}
      </button>
      {showReplies && (
        <div className="replies">
          {(replies || []).map((r) => (
            <div key={r.id} className="reply">
              <button className="link" onClick={() => onProfile(r.user_id)}>{r.author}</button>
              <span className="time">{timeAgo(r.created_at)}</span>
              <p>{r.content}</p>
            </div>
          ))}
          {replies?.length === 0 && <p className="muted">No replies yet.</p>}
          <ReplyForm
            scrapId={scrap.id}
            currentUser={currentUser}
            onReplied={async () => {
              setReplies(await api.listReplies(scrap.id))
              onReply()
            }}
          />
        </div>
      )}
    </div>
  )
}

function Profile({ userId, currentUser, onBack, onProfile }) {
  const [data, setData] = useState(null)
  const [error, setError] = useState('')

  // Fetches the member, the scraps they wrote, and their scrapbook together.
  const fetchProfile = () =>
    Promise.all([
      api.getUser(userId),
      api.listUserScraps(userId),
      api.listUserScrapbook(userId),
    ]).then(([user, written, scrapbook]) => ({
      user,
      written: written || [],
      scrapbook: scrapbook || [],
    }))

  // Refresh after a post or reply. Failures show on the page instead of going unhandled.
  const load = () => fetchProfile().then(setData).catch((err) => setError(err.message))

  useEffect(() => {
    // Ignore responses for a previous member if the user clicks another profile first.
    let active = true
    setData(null)
    setError('')
    fetchProfile()
      .then((profile) => { if (active) setData(profile) })
      .catch((err) => { if (active) setError(err.message) })
    return () => { active = false }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [userId])

  if (error) return <section className="feed"><button className="link" onClick={onBack}>← Back to scraps</button><p className="error">{/404|not found/i.test(error) ? 'Member not found.' : error}</p></section>
  if (!data) return <section className="feed"><p className="muted">Loading member…</p></section>

  const { user, written, scrapbook } = data

  return (
    <section className="feed">
      <button className="link" onClick={onBack}>← Back to scraps</button>
      <div className="card">
        <div className="scrap-header">
          <span className="avatar">{user.avatar || user.name?.[0]?.toUpperCase() || '?'}</span>
          <div><h2>{user.name}</h2><span className="time">Joined {timeAgo(user.created_at)}</span></div>
        </div>
      </div>
      <ScrapForm currentUser={currentUser} recipient={user} onPosted={load} />
      <h2>Scrapbook</h2>
      {scrapbook.length === 0 && <p className="muted">Nothing on {user.name}'s scrapbook yet.</p>}
      {scrapbook.map((s) => <Scrap key={s.id} scrap={s} currentUser={currentUser} onReply={load} onProfile={onProfile} />)}
      <h2>Scraps by {user.name}</h2>
      {written.length === 0 && <p className="muted">No scraps yet.</p>}
      {written.map((s) => <Scrap key={s.id} scrap={s} currentUser={currentUser} onReply={load} onProfile={onProfile} />)}
    </section>
  )
}

const CURRENT_USER_KEY = 'scraps.currentUserId'

function WhoAmI({ users, currentUser, onChange }) {
  return (
    <div className="card">
      <h2>I am</h2>
      <select value={currentUser?.id ?? ''} onChange={(e) => onChange(e.target.value)}>
        <option value="">— nobody —</option>
        {users.map((u) => (
          <option key={u.id} value={u.id}>
            {u.avatar ? `${u.avatar} ` : ''}
            {u.name}
          </option>
        ))}
      </select>
      <p className="muted">Remembered on this browser.</p>
    </div>
  )
}

export default function App() {
  const [users, setUsers] = useState([])
  const [scraps, setScraps] = useState([])
  const [currentUserId, setCurrentUserId] = useState(() => localStorage.getItem(CURRENT_USER_KEY) || '')
  const [loading, setLoading] = useState(true)
  const [profileId, setProfileId] = useState(null)

  // Derived from the member list, so a deleted member simply stops being "me".
  const currentUser = users.find((u) => String(u.id) === currentUserId) || null

  function chooseIdentity(id) {
    const value = id ? String(id) : ''
    if (value) localStorage.setItem(CURRENT_USER_KEY, value)
    else localStorage.removeItem(CURRENT_USER_KEY)
    setCurrentUserId(value)
  }

  async function refresh() {
    const [u, s] = await Promise.all([api.listUsers(), api.listScraps()])
    setUsers(u || [])
    setScraps(s || [])
    setLoading(false)
  }

  useEffect(() => {
    refresh().catch(() => setLoading(false))
  }, [])

  return (
    <div className="app">
      <header>
        <h1>📜 Scraps</h1>
        <p className="muted">A nostalgic Orkut-style scrapbook — no login needed.</p>
      </header>

      <main>
        <aside>
          <WhoAmI users={users} currentUser={currentUser} onChange={chooseIdentity} />
          <RegisterForm
            onRegistered={(user) => {
              chooseIdentity(user.id)
              refresh()
            }}
          />
          <div className="card">
            <h2>Members ({users.length})</h2>
            <ul className="members">
              {users.map((u) => (
                <li key={u.id}>
                  <span className="avatar small">
                    {u.avatar || u.name?.[0]?.toUpperCase() || '?'}
                  </span>
                  <button className="link" onClick={() => setProfileId(u.id)}>{u.name}</button>
                </li>
              ))}
              {users.length === 0 && (
                <li className="muted">No one registered yet.</li>
              )}
            </ul>
          </div>
        </aside>

        {profileId ? <Profile userId={profileId} currentUser={currentUser} onBack={() => setProfileId(null)} onProfile={setProfileId} /> : <section className="feed">
          <ScrapForm currentUser={currentUser} onPosted={refresh} />
          {loading && <p className="muted">Loading scraps…</p>}
          {!loading && scraps.length === 0 && (
            <p className="muted">No scraps yet. Post the first one!</p>
          )}
          {scraps.map((s) => (
            <Scrap key={s.id} scrap={s} currentUser={currentUser} onReply={refresh} onProfile={setProfileId} />
          ))}
        </section>}
      </main>
    </div>
  )
}
