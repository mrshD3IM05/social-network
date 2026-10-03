'use client'

import { useEffect, useRef, useState } from 'react'
import usePaged from '@/lib/usePaged'
import { useMe } from '@/lib/useMe'
import { getUnread, onUnreadChange } from '@/lib/unread'
import { subscribe } from '@/lib/socket'
import Icon from '@/components/Icon'
import LoadMore from '@/components/LoadMore'
import PageHeader from '@/components/PageHeader'
import PersonRow from '@/components/PersonRow'

// The people you can message: the ones where one of you follows the other.
// Everyone else would only get "you cannot message them yet", so they are not
// listed here — /people is where you go to find someone new to follow.
export default function ChatListPage() {
  const contacts = usePaged('/contacts') // newest conversation first, 10 at a time
  const people = contacts.items
  const [unread, setUnread] = useState(getUnread())
  const { me } = useMe()
  const contactsRef = useRef(contacts) // the socket handler reads the current list
  contactsRef.current = contacts

  // a dot on every person whose message has not been opened yet
  useEffect(() => onUnreadChange(setUnread), [])

  useEffect(() => {
    if (!me) return
    const unsub = subscribe((data) => {
      if (!data || data.type !== 'message') return
      const msg = data.message
      if (msg.group_id || (msg.to_user_id !== me.id && msg.from_user_id !== me.id)) return
      // the list is newest conversation first: move that person to the top here
      // instead of asking for the whole list again on every message
      const otherId = msg.from_user_id === me.id ? msg.to_user_id : msg.from_user_id
      const { items, setItems, reload } = contactsRef.current
      const person = items?.find(p => p.id === otherId)
      if (person) setItems([person, ...items.filter(p => p.id !== otherId)])
      // not shown yet (on a later page, or followed after this page loaded):
      // the first page now starts with them
      else reload()
    })
    return unsub
  }, [me])

  return (
    <>
      <PageHeader title="Messages" subtitle="Pick someone you follow, or who follows you." />

      {contacts.error && <p className="error">{contacts.error.message}</p>}
      {people === null && !contacts.error && <p className="loading">Loading…</p>}

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
      <LoadMore list={contacts} />
    </>
  )
}
