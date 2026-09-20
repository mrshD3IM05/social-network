'use client'

import { useEffect, useRef, useState } from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import { apiGet } from '@/lib/api'
import { onSocketEvent, sendSocket } from '@/lib/ws'
import { LIMITS, checkText } from '@/lib/validate'
import CharCount from '@/components/CharCount'
import Avatar from '@/components/Avatar'
import Icon from '@/components/Icon'

// A private conversation with one user: the stored history plus live messages.
export default function ConversationPage() {
  const { id } = useParams()
  const otherId = Number(id)
  const [me, setMe] = useState(null)
  const [other, setOther] = useState(null)
  const [messages, setMessages] = useState([])
  const [text, setText] = useState('')
  const [error, setError] = useState('')
  const [online, setOnline] = useState(false)
  const bottomRef = useRef(null)

  useEffect(() => {
    apiGet('/me').then(setMe)
    apiGet(`/user/${id}`).then(setOther).catch(() => setOther({ first_name: 'User', last_name: id }))

    // messages were always saved; this reads them back
    apiGet(`/messages/${id}`)
      .then(setMessages)
      .catch(err => setError(err.message))
  }, [id])

  useEffect(() => {
    return onSocketEvent(event => {
      if (event.type === 'socket') {
        setOnline(event.state === 'open')
        return
      }
      if (event.type === 'message') {
        const message = event.message
        // keep only the messages of this conversation
        if (message.group_id) return
        if (message.from_user_id === otherId || message.to_user_id === otherId) {
          setMessages(list => (list.some(m => m.id === message.id) ? list : [...list, message]))
        }
      }
      if (event.type === 'error') setError(event.error)
    })
  }, [otherId])

  // scroll to the newest message
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  function send(e) {
    e.preventDefault()

    const problem = checkText('Your message', text, LIMITS.message)
    if (problem) {
      setError(problem)
      return
    }

    // sendSocket answers false while the connection is down, so the message is
    // never dropped without telling the user
    if (!sendSocket({ type: 'message', to_user_id: otherId, content: text.trim() })) {
      setError('Not connected yet, give it a moment and try again.')
      return
    }

    setError('')
    setText('')
  }

  if (!me || !other) return <p className="loading">Loading…</p>

  return (
    <section className="card chat">
      <header className="chat-header">
        <Link href="/chat" className="icon-button" title="Back"><Icon name="back" /></Link>
        <Avatar user={other} size={38} />
        <div>
          <strong>{other.first_name} {other.last_name}</strong>
          <p className="meta">{online ? 'Connected' : 'Reconnecting…'}</p>
        </div>
      </header>

      <div className="chat-messages">
        {messages.length === 0 && <p className="chat-note">No messages yet. Say hello.</p>}
        {messages.map(msg => (
          <div key={msg.id} className={msg.from_user_id === me.id ? 'bubble mine' : 'bubble'}>
            {msg.content}
            <small className="bubble-time">
              {new Date(msg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            </small>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      {error && <p className="error chat-error">{error}</p>}

      <form className="chat-form" onSubmit={send} noValidate>
        <input
          value={text}
          maxLength={LIMITS.message}
          onChange={e => setText(e.target.value)}
          placeholder="Write a message…"
        />
        <CharCount value={text} max={LIMITS.message} />
        <button className="btn" title="Send" disabled={!text.trim()}>
          <Icon name="send" size={16} />
        </button>
      </form>
    </section>
  )
}
