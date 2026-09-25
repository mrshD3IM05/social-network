'use client'

import { useEffect, useRef, useState } from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import { apiGet, apiUpload, imageUrl } from '@/lib/api'
import { IMAGE_ACCEPT, LIMITS, checkImages, checkText } from '@/lib/validate'
import { getDraft, setDraft } from '@/lib/draft'
import { markRead } from '@/lib/unread'
import CharCount from '@/components/CharCount'
import Avatar from '@/components/Avatar'
import Icon from '@/components/Icon'

// A private conversation with one user, in real time over a WebSocket.
export default function ConversationPage() {
  const { id } = useParams()
  const otherId = Number(id)
  const [me, setMe] = useState(null)
  const [other, setOther] = useState(null)
  const [messages, setMessages] = useState([])
  const [text, setText] = useState(getDraft())
  const [files, setFiles] = useState([])
  const [typing, setTyping] = useState(false)
  const [error, setError] = useState('')
  const [sending, setSending] = useState(false)
  const socketRef = useRef(null) // useRef keeps the socket between renders
  const bottomRef = useRef(null)
  const fileRef = useRef(null)
  const lastTyping = useRef(0)

  useEffect(() => {
    apiGet('/me').then(setMe)
    apiGet(`/user/${id}`).then(setOther).catch(() => setOther({ first_name: 'User', last_name: id }))

    // opening the conversation means you read it, so its dot goes away
    markRead(otherId)

    // the conversation is saved, so it is read back on every visit
    apiGet(`/messages/${id}`)
      .then(setMessages)
      .catch(err => setError(err.message))

    // same address as the page, so it also works behind the proxy
    const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const socket = new WebSocket(`${scheme}//${window.location.host}/api/v1/ws`)
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
  }, [id, otherId])

  // scroll to the newest message
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, typing])

  // Tell the other side we are writing, at most once every two seconds.
  function onType(e) {
    setText(e.target.value)
    setDraft(e.target.value)

    const now = Date.now()
    if (socketRef.current?.readyState !== WebSocket.OPEN) return
    if (now - lastTyping.current < 2000) return

    lastTyping.current = now
    socketRef.current.send(JSON.stringify({ type: 'typing', to_user_id: otherId }))
  }

  function pickFiles(e) {
    const picked = Array.from(e.target.files)
    const problem = checkImages(picked)
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

    setError('')
    setSending(true)
    try {
      const body = new FormData()
      body.append('to_user_id', otherId)
      body.append('content', text.trim())
      for (const file of files) body.append('files', file)

      await apiUpload('/messages', body)
      setText('')
      setDraft('')
      clearFiles()
    } catch (err) {
      setError(err.message)
    }
    setSending(false)
  }

  if (!me || !other) return <p className="loading">Loading…</p>

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
