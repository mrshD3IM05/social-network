'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { fetchConversations, fetchPeople } from '@/lib/people'
import { LIMITS } from '@/lib/validate'
import Avatar from '@/components/Avatar'
import Icon from '@/components/Icon'
import PageHeader from '@/components/PageHeader'

// Your existing conversations, plus everyone else you could start one with.
export default function ChatListPage() {
  const [conversations, setConversations] = useState([])
  const [people, setPeople] = useState([])
  const [search, setSearch] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    fetchConversations().then(setConversations).catch(err => setError(err.message))
    fetchPeople().then(setPeople).catch(err => setError(err.message))
  }, [])

  const talkedTo = new Set(conversations.map(item => item.user.id))
  const others = people.filter(person => !talkedTo.has(person.id))

  function matches(person) {
    return `${person.first_name} ${person.last_name} ${person.nickname}`
      .toLowerCase()
      .includes(search.toLowerCase())
  }

  return (
    <>
      <PageHeader label="Inbox" title="Messages" subtitle="Your conversations are saved and load when you open them." />

      {error && <p className="error">{error}</p>}

      <div className="search">
        <Icon name="search" />
        <input
          placeholder="Search by name or nickname"
          value={search}
          maxLength={LIMITS.search}
          onChange={e => setSearch(e.target.value)}
        />
      </div>

      {conversations.filter(item => matches(item.user)).length > 0 && (
        <>
          <p className="eyebrow section-label">Conversations</p>
          <div className="card list">
            {conversations.filter(item => matches(item.user)).map(({ user, last_message }) => (
              <Link key={user.id} href={`/chat/${user.id}`} className="list-item">
                <Avatar user={user} size={44} />
                <span className="list-text">
                  <strong>{user.first_name} {user.last_name}</strong>
                  <small>{last_message.content}</small>
                </span>
                <Icon name="chat" size={16} />
              </Link>
            ))}
          </div>
        </>
      )}

      <p className="eyebrow section-label">Start a conversation</p>

      {others.filter(matches).length === 0 && (
        <div className="empty">
          <p className="empty-title">Nobody else to message</p>
          <p>You can message people with a public profile, or anyone you follow each other with.</p>
        </div>
      )}

      <div className="card list">
        {others.filter(matches).map(person => (
          <Link key={person.id} href={`/chat/${person.id}`} className="list-item">
            <Avatar user={person} size={44} />
            <span className="list-text">
              <strong>{person.first_name} {person.last_name}</strong>
              {person.nickname && <small>@{person.nickname}</small>}
            </span>
            <Icon name="chat" size={16} />
          </Link>
        ))}
      </div>
    </>
  )
}
