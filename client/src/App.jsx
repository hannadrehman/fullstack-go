import { useEffect, useState } from 'react'
import { api } from './api.js'
import './App.css'

// Pure helpers first. They take no hooks and no props, so they are easy to test and happy about it.
function timeAgo(iso) {
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  return `${Math.floor(hours / 24)}d ago`
}

// Uppercased first letter of a name, or a question mark when there is no name at all
function initialOf(name) {
  return name?.[0]?.toUpperCase() || '?'
}

// An avatar emoji beats the initial letter, every single time
function avatarText(user) {
  return user.avatar || initialOf(user.name)
}

// "1 reply" and "2 replies", because grammar is a small kindness
function replyLabel(count) {
  return count === 1 ? 'reply' : 'replies'
}

// The backend says "not found" in several ways, so we just listen for the vibe
function friendlyError(message) {
  return /404|not found/i.test(message) ? 'Member not found.' : message
}

function Avatar({ text, small = false }) {
  return <span className={small ? 'avatar small' : 'avatar'}>{text}</span>
}

// A clickable member name that opens their profile. Nothing fancy, just a button in a trench coat.
function MemberLink({ userId, name, onProfile }) {
  return <button className="link" onClick={() => onProfile(userId)}>{name}</button>
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

function ScrapForm({ users, currentUser, onPosted }) {
  const [content, setContent] = useState('')
  const [userId, setUserId] = useState(currentUser?.id || '')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  // Whoever just registered becomes the default poster. Friendly default, zero clicks.
  useEffect(() => {
    if (currentUser) setUserId(currentUser.id)
  }, [currentUser])

  async function submit(e) {
    e.preventDefault()
    if (!userId) {
      setError('Pick who is posting (register first if empty).')
      return
    }
    setBusy(true)
    setError('')
    try {
      await api.createScrap({ user_id: Number(userId), content })
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
      <h2>Post a scrap</h2>
      <textarea
        placeholder="Write something for everyone…"
        rows={3}
        value={content}
        onChange={(e) => setContent(e.target.value)}
        required
      />
      <div className="row">
        <select value={userId} onChange={(e) => setUserId(e.target.value)}>
          <option value="">— posting as —</option>
          {users.map((u) => (
            <option key={u.id} value={u.id}>
              {u.avatar ? `${u.avatar} ` : ''}
              {u.name}
            </option>
          ))}
        </select>
        <button disabled={busy}>{busy ? 'Posting…' : 'Post scrap'}</button>
      </div>
      {error && <p className="error">{error}</p>}
    </form>
  )
}

function ReplyForm({ scrapId, users, onReplied }) {
  const [content, setContent] = useState('')
  const [userId, setUserId] = useState('')
  const [error, setError] = useState('')

  async function submit(e) {
    e.preventDefault()
    if (!userId) {
      setError('Pick who is replying.')
      return
    }
    setError('')
    try {
      await api.createReply(scrapId, { user_id: Number(userId), content })
      setContent('')
      onReplied()
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <form className="reply-form" onSubmit={submit}>
      <input
        placeholder="Write a reply…"
        value={content}
        onChange={(e) => setContent(e.target.value)}
        required
      />
      <select value={userId} onChange={(e) => setUserId(e.target.value)}>
        <option value="">— as —</option>
        {users.map((u) => (
          <option key={u.id} value={u.id}>
            {u.name}
          </option>
        ))}
      </select>
      <button>Reply</button>
      {error && <p className="error">{error}</p>}
    </form>
  )
}

function Scrap({ scrap, users, onReply, onProfile }) {
  const [replies, setReplies] = useState(null)
  const [showReplies, setShowReplies] = useState(false)

  // Replies are fetched lazily, the first time somebody opens the thread
  async function loadReplies() {
    setShowReplies((s) => !s)
    if (replies === null) setReplies(await api.listReplies(scrap.id))
  }

  return (
    <div className="card scrap">
      <div className="scrap-header">
        <Avatar text={initialOf(scrap.author)} />
        <div>
          <MemberLink userId={scrap.user_id} name={scrap.author} onProfile={onProfile} />
          <span className="time">{timeAgo(scrap.created_at)}</span>
        </div>
      </div>
      <p className="scrap-content">{scrap.content}</p>
      <button className="link" onClick={loadReplies}>
        {scrap.reply_count} {replyLabel(scrap.reply_count)}{' '}
        {showReplies ? '▾' : '▸'}
      </button>
      {showReplies && (
        <div className="replies">
          {(replies || []).map((r) => (
            <div key={r.id} className="reply">
              <MemberLink userId={r.user_id} name={r.author} onProfile={onProfile} />
              <span className="time">{timeAgo(r.created_at)}</span>
              <p>{r.content}</p>
            </div>
          ))}
          {replies?.length === 0 && <p className="muted">No replies yet.</p>}
          <ReplyForm
            scrapId={scrap.id}
            users={users}
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

function Profile({ userId, users, onBack, onProfile }) {
  const [user, setUser] = useState(null)
  const [scraps, setScraps] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let active = true
    setLoading(true)
    setError('')
    Promise.all([api.getUser(userId), api.listUserScraps(userId)])
      .then(([member, memberScraps]) => {
        if (!active) return
        setUser(member)
        setScraps(memberScraps || [])
      })
      .catch((err) => { if (active) setError(err.message) })
      .finally(() => { if (active) setLoading(false) })
    // Flag flips to false on unmount or a new userId, so stale responses get ignored
    return () => { active = false }
  }, [userId])

  if (loading) return <section className="feed"><p className="muted">Loading member…</p></section>
  if (error) return <section className="feed"><button className="link" onClick={onBack}>← Back to scraps</button><p className="error">{friendlyError(error)}</p></section>

  return (
    <section className="feed">
      <button className="link" onClick={onBack}>← Back to scraps</button>
      <div className="card">
        <div className="scrap-header">
          <Avatar text={avatarText(user)} />
          <div><h2>{user.name}</h2><span className="time">Joined {timeAgo(user.created_at)}</span></div>
        </div>
      </div>
      <h2>Scraps</h2>
      {scraps.length === 0 && <p className="muted">No scraps yet.</p>}
      {scraps.map((s) => <Scrap key={s.id} scrap={s} users={users} onReply={() => {}} onProfile={onProfile} />)}
    </section>
  )
}

function Feed({ users, currentUser, scraps, loading, onRefresh, onProfile }) {
  return (
    <section className="feed">
      <ScrapForm users={users} currentUser={currentUser} onPosted={onRefresh} />
      {loading && <p className="muted">Loading scraps…</p>}
      {!loading && scraps.length === 0 && (
        <p className="muted">No scraps yet. Post the first one!</p>
      )}
      {scraps.map((s) => (
        <Scrap key={s.id} scrap={s} users={users} onReply={onRefresh} onProfile={onProfile} />
      ))}
    </section>
  )
}

export default function App() {
  const [users, setUsers] = useState([])
  const [scraps, setScraps] = useState([])
  const [currentUser, setCurrentUser] = useState(null)
  const [loading, setLoading] = useState(true)
  const [profileId, setProfileId] = useState(null)

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
          <RegisterForm
            onRegistered={(user) => {
              setCurrentUser(user)
              refresh()
            }}
          />
          <div className="card">
            <h2>Members ({users.length})</h2>
            <ul className="members">
              {users.map((u) => (
                <li key={u.id}>
                  <Avatar text={avatarText(u)} small />
                  <MemberLink userId={u.id} name={u.name} onProfile={setProfileId} />
                </li>
              ))}
              {users.length === 0 && (
                <li className="muted">No one registered yet.</li>
              )}
            </ul>
          </div>
        </aside>

        {/* Profile view swaps in for the feed when a member is selected. Hop back to see the scraps again. */}
        {profileId
          ? <Profile userId={profileId} users={users} onBack={() => setProfileId(null)} onProfile={setProfileId} />
          : <Feed users={users} currentUser={currentUser} scraps={scraps} loading={loading} onRefresh={refresh} onProfile={setProfileId} />}
      </main>
    </div>
  )
}
