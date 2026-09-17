'use client'

import { useEffect, useRef, useState } from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import { apiGet } from '@/lib/api'
import { LIMITS, checkText } from '@/lib/validate'
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
  const [text, setText] = useState('')
  const [error, setError] = useState('')
  const socketRef = useRef(null) // useRef keeps the socket between renders
  const bottomRef = useRef(null)

  useEffect(() => {
    apiGet('/me').then(setMe)
    apiGet(`/user/${id}`).then(setOther).catch(() => setOther({ first_name: 'User', last_name: id }))

    // Connect straight to the Go server (the cookie is sent automatically)
    const socket = new WebSocket(`ws://${window.location.hostname}:8080/api/v1/ws`)
    socketRef.current = socket

    socket.onmessage = event => {
      const data = JSON.parse(event.data)
      if (data.type === 'message') {
        const msg = data.message
        // keep only the messages of this conversation
        if (msg.from_user_id === otherId || msg.to_user_id === otherId) {
          setMessages(list => [...list, msg])
        }
      }
      if (data.type === 'error') setError(data.error)
    }

    // close the connection when we leave the page
    return () => socket.close()
  }, [id, otherId])

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

    setError('')
    socketRef.current.send(
      JSON.stringify({ type: 'message', to_user_id: otherId, content: text.trim() }),
    )
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
          <p className="meta">Live conversation</p>
        </div>
      </header>

      <div className="chat-messages">
        <p className="chat-note">Messages are live only and are not saved when you reload.</p>
        {messages.map(msg => (
          <div key={msg.id} className={msg.from_user_id === me.id ? 'bubble mine' : 'bubble'}>
            {msg.content}
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
