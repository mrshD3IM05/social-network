'use client'

import { useEffect, useMemo, useState } from 'react'
import usePaged from '@/lib/usePaged'
import { apiGet } from '@/lib/api'
import { getUnread, onUnreadChange } from '@/lib/unread'
import Icon from '@/components/Icon'
import LoadMore from '@/components/LoadMore'
import PageHeader from '@/components/PageHeader'
import PersonRow from '@/components/PersonRow'

// List only people who have an accepted follow relationship with you.
export default function ChatListPage() {
  const [me, setMe] = useState(null)
  const [unread, setUnread] = useState(getUnread())
  const followers = usePaged(me ? `/users/${me.id}/followers` : null)
  const following = usePaged(me ? `/users/${me.id}/following` : null)

  const people = useMemo(() => {
    const byID = new Map()
    for (const person of [...(followers.items || []), ...(following.items || [])]) byID.set(person.id, person)
    return [...byID.values()].sort((a, b) =>
      `${a.first_name} ${a.last_name}`.localeCompare(`${b.first_name} ${b.last_name}`),
    )
  }, [followers.items, following.items])

  const peopleList = {
    hasMore: followers.hasMore || following.hasMore,
    loading: followers.loading || following.loading,
    loadMore: () => {
      if (followers.hasMore) followers.loadMore()
      if (following.hasMore) following.loadMore()
    },
  }

  // a dot on every person whose message has not been opened yet
  useEffect(() => onUnreadChange(setUnread), [])
  useEffect(() => { apiGet('/me').then(setMe) }, [])

  return (
    <>
      <PageHeader label="Inbox" title="Messages" subtitle="Pick someone to start a real-time conversation." />

      {(!me || followers.items === null || following.items === null) && <p className="loading">Loading…</p>}

      {me && people.length === 0 && (
        <div className="empty">
          <p className="empty-title">No one to message yet</p>
          <p>Follow someone, or wait for someone to follow you.</p>
        </div>
      )}

      <div className="card list">
        {people.map(person => (
          <PersonRow key={person.id} person={person} href={`/chat/${person.id}`}>
            {unread.has(person.id) && <span className="menu-dot" title="New message" />}
            <Icon name="chat" size={16} />
          </PersonRow>
        ))}
      </div>
      <LoadMore list={peopleList} />
    </>
  )
}
