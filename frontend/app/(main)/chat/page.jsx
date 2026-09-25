'use client'

import { useEffect, useState } from 'react'
import { fetchPeople } from '@/lib/people'
import { getUnread, onUnreadChange } from '@/lib/unread'
import Icon from '@/components/Icon'
import PageHeader from '@/components/PageHeader'
import PersonRow from '@/components/PersonRow'

// List of people you can chat with (everyone on the network).
export default function ChatListPage() {
  const [people, setPeople] = useState(null)
  const [unread, setUnread] = useState(getUnread())

  useEffect(() => {
    fetchPeople()
      .then(setPeople)
      .catch(() => setPeople([]))
  }, [])

  // a dot on every person whose message has not been opened yet
  useEffect(() => onUnreadChange(setUnread), [])

  return (
    <>
      <PageHeader label="Inbox" title="Messages" subtitle="Pick someone to start a real-time conversation." />

      {people === null && <p className="loading">Loading…</p>}

      {people !== null && people.length === 0 && (
        <div className="empty">
          <p className="empty-title">No one to message yet</p>
          <p>You are the only member so far.</p>
        </div>
      )}

      <div className="card list">
        {people?.map(person => (
          <PersonRow key={person.id} person={person} href={`/chat/${person.id}`}>
            {unread.has(person.id) && <span className="menu-dot" title="New message" />}
            <Icon name="chat" size={16} />
          </PersonRow>
        ))}
      </div>
    </>
  )
}
