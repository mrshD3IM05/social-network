'use client'

import { useEffect, useRef, useState } from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import { apiGet, apiUpload, imageUrl, socketUrl } from '@/lib/api'
import { IMAGE_ACCEPT, LIMITS, checkImageFiles, checkText, parseId } from '@/lib/validate'
import { markRead } from '@/lib/unread'
import { useThrottle } from '@/lib/timing'
import CharCount from '@/components/CharCount'
import Avatar from '@/components/Avatar'
import Icon from '@/components/Icon'
import EmojiPicker from '@/components/EmojiPicker'
import NotFound from '@/components/NotFound'

// A private conversation with one user, in real time over a WebSocket.
export default function ConversationPage() {
  const { id } = useParams()
  // 0 when the url is not a real id, like /chat/abc or /chat/0
  const otherId = parseId(id)
  const [me, setMe] = useState(null)
  const [other, setOther] = useState(null)
  const [messages, setMessages] = useState([])
  const [text, setText] = useState('')
  const [files, setFiles] = useState([])
  const [typing, setTyping] = useState(false)
  const [blocked, setBlocked] = useState(false)
  const [notFound, setNotFound] = useState(false)
  const [error, setError] = useState('')
  const [sending, setSending] = useState(false)
  const socketRef = useRef(null) // useRef keeps the socket between renders
  const bottomRef = useRef(null)
  const fileRef = useRef(null)

  useEffect(() => {
    // there is nobody to talk to, so no request and no socket
    if (!otherId) return

    apiGet('/me').then(setMe)
    // 400 = the id is not one the API accepts, 404 = nobody under it,
    // 403 = a real user whose profile is private. The first two mean the same
    // thing here: there is no conversation to show.
    apiGet(`/user/${otherId}`)
      .then(setOther)
      .catch(err => {
        if (err.status === 400 || err.status === 404) setNotFound(true)
        else setError(err.message)
      })

    // opening the conversation means you read it, so its dot goes away
    markRead(otherId)

    // the conversation is saved, so it is read back on every visit
    apiGet(`/messages/${otherId}`)
      .then(setMessages)
      .catch(err => {
        // the API refuses when neither of you follows the other
        if (err.status === 403) setBlocked(true)
        else setError(err.message)
      })

    // Connect straight to the Go server (the cookie is sent automatically)
    const socket = new WebSocket(socketUrl())
    socketRef.current = socket

    let typingTimer = null

    socket.onmessage = event => {
      const data = JSON.parse(event.data)
      if (data.type === 'message') {
        const msg = data.message
        // keep only the messages of this conversation
        if (msg.from_user_id === otherId || msg.to_user_id === otherId) {
          setMessages(list => (list.some(m => m.id === msg.id) ? list : [...list, msg]))
        }
      }
      // "typing" only means right now, so it fades on its own
      if (data.type === 'typing' && data.from_user_id === otherId) {
        setTyping(true)
        clearTimeout(typingTimer)
        typingTimer = setTimeout(() => setTyping(false), 3000)
      }

      if (data.type === 'error') setError(data.error)
    }

    // close the connection when we leave the page
    return () => {
      clearTimeout(typingTimer)
      socket.close()
    }
  }, [otherId])

  // scroll to the newest message
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, typing])

  // Tell the other side we are writing, at most once every two seconds.
  // The trailing call keeps "typing…" alive until the last keystroke.
  const sendTyping = useThrottle(() => {
    if (socketRef.current?.readyState !== WebSocket.OPEN) return
    socketRef.current.send(JSON.stringify({ type: 'typing', to_user_id: otherId }))
  }, 2000)

  function onType(e) {
    setText(e.target.value)
    sendTyping()
  }

  async function pickFiles(e) {
    const picked = Array.from(e.target.files)
    const problem = await checkImageFiles(picked)
    setError(problem)
    setFiles(problem ? [] : picked)
    if (problem) e.target.value = ''
  }

  function clearFiles() {
    setFiles([])
    if (fileRef.current) fileRef.current.value = ''
  }

  // Sent over the API and not the socket: the message row has to exist before
  // an upload can point at it, and the other side is told once both are done.
  async function send(e) {
    e.preventDefault()

    // a message needs text, a picture, or both
    const problem = text.trim()
      ? checkText('Your message', text, LIMITS.message)
      : files.length === 0 && 'Write something or add an image.'
    if (problem) {
      setError(problem)
      return
    }

    // the message is on its way, so a late "typing…" would be wrong
    sendTyping.cancel()
    setError('')
    setSending(true)
    try {
      const body = new FormData()
      body.append('to_user_id', otherId)
      body.append('content', text.trim())
      for (const file of files) body.append('files', file)

      await apiUpload('/messages', body)
      setText('')
      clearFiles()
    } catch (err) {
      setError(err.message)
    }
    setSending(false)
  }

  if (!otherId || notFound) {
    return (
      <NotFound
        title="Conversation not found"
        text="Nobody is here under that id."
        back="/chat"
        label="Back to messages"
      />
    )
  }

  // a load that failed leaves `other` null, so this has to come before the
  // loading line below or the page would say "Loading…" forever
  if (!other && error) return <p className="loading">{error}</p>

  if (!me || !other) return <p className="loading">Loading…</p>

  // Your own id in the url. The list never links here — it holds everyone but
  // you — so this is only reachable by typing it. The API refuses it like any
  // other conversation, but "you are not following yourself" would be nonsense,
  // so it gets its own wording.
  if (otherId === me.id) {
    return (
      <div className="card locked">
        <span className="locked-icon"><Icon name="lock" size={22} /></span>
        <h2>You cannot message yourself</h2>
        <p className="subtitle">Pick someone else to start a conversation.</p>
        <Link href="/chat" className="btn">Back to messages</Link>
      </div>
    )
  }

  if (blocked) {
    return (
      <div className="card locked">
        <span className="locked-icon"><Icon name="lock" size={22} /></span>
        <h2>You cannot message {other.first_name} yet</h2>
        <p className="subtitle">
          One of you has to follow the other before you can write to each other.
        </p>
        <Link href={`/profile/${otherId}`} className="btn">Open their profile</Link>
      </div>
    )
  }

  return (
    <section className="card chat">
      <header className="chat-header">
        <Link href="/chat" className="icon-button" title="Back"><Icon name="back" /></Link>
        <Avatar user={other} size={38} />
        <div>
          <strong>{other.first_name} {other.last_name}</strong>
          <p className="meta">{typing ? 'typing…' : 'Live conversation'}</p>
        </div>
      </header>

      <div className="chat-messages">
        {messages.length === 0 && <p className="chat-note">No messages yet. Say hello.</p>}
        {messages.map(msg => (
          <div key={msg.id} className={msg.from_user_id === me.id ? 'bubble mine' : 'bubble'}>
            {msg.content && <span>{msg.content}</span>}
            {msg.images?.length > 0 && (
              <span className="bubble-images">
                {msg.images.map(fileId => <img key={fileId} src={imageUrl(fileId)} alt="" />)}
              </span>
            )}
          </div>
        ))}
        {typing && <p className="typing">{other.first_name} is typing…</p>}

        <div ref={bottomRef} />
      </div>

      {error && <p className="error chat-error">{error}</p>}

      {files.length > 0 && (
        <p className="chat-files">
          {files.length} image{files.length > 1 ? 's' : ''} ready
          <button type="button" className="link-button" onClick={clearFiles}>remove</button>
        </p>
      )}

      <form className="chat-form" onSubmit={send} noValidate>
        <label className="icon-button" title="Add a photo or GIF">
          <Icon name="image" size={16} />
          <input
            ref={fileRef}
            type="file"
            accept={IMAGE_ACCEPT}
            multiple
            hidden
            onChange={pickFiles}
          />
        </label>
        <EmojiPicker onPick={emoji => onType({ target: { value: text + emoji } })} />

        <input
          value={text}
          maxLength={LIMITS.message}
          onChange={onType}
          placeholder="Write a message…"
        />
        <CharCount value={text} max={LIMITS.message} />
        <button className="btn" title="Send" disabled={sending || (!text.trim() && files.length === 0)}>
          <Icon name="send" size={16} />
        </button>
      </form>
    </section>
  )
}
