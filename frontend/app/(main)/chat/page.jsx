'use client'

import { useEffect, useState } from 'react'
import { fetchContacts } from '@/lib/people'
import { useMe } from '@/lib/useMe'
import { getUnread, onUnreadChange } from '@/lib/unread'
import { subscribe } from '@/lib/socket'
import Icon from '@/components/Icon'
import PageHeader from '@/components/PageHeader'
import PersonRow from '@/components/PersonRow'

// The people you can message: the ones where one of you follows the other.
// Everyone else would only get "you cannot message them yet", so they are not
// listed here — /people is where you go to find someone new to follow.
export default function ChatListPage() {
  const [people, setPeople] = useState(null)
  const [unread, setUnread] = useState(getUnread())
  const { me } = useMe()

  useEffect(() => {
    fetchContacts()
      .then(setPeople)
      .catch(() => setPeople([]))
  }, [])

  // a dot on every person whose message has not been opened yet
  useEffect(() => onUnreadChange(setUnread), [])

  useEffect(() => {
    if (!me) return
    const unsub = subscribe((data) => {
      if (!data || data.type !== 'message') return
      const msg = data.message
      if (msg.to_user_id === me.id || msg.from_user_id === me.id) {
        fetchContacts().then(setPeople).catch(() => {})
      }
    })
    return unsub
  }, [me])

  return (
    <>
      <PageHeader title="Messages" subtitle="Pick someone you follow, or who follows you." />

      {people === null && <p className="loading">Loading…</p>}

      {people !== null && people.length === 0 && (
        <div className="empty">
          <p className="empty-title">No one to message yet</p>
          <p>Follow someone, or get them to follow you, to start a conversation.</p>
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
