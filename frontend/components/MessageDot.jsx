'use client'

import { useEffect, useState } from 'react'
import { getUnread, markUnread, onUnreadChange } from '@/lib/unread'
import { subscribe } from '@/lib/socket'

// A small dot next to Messages while somebody's message is still unread.
//
// It stays on until every conversation that received a message has been
// opened, so a second person writing to you does not go unnoticed.
//
// It is kept apart from the bell on purpose: the subject asks for new messages
// and new notifications to be shown in different ways.
export default function MessageDot({ myId }) {
  const [unread, setUnread] = useState(getUnread())

  useEffect(() => onUnreadChange(setUnread), [])

  useEffect(() => {
    const unsub = subscribe((data) => {
      if (!data || data.type !== 'message') return
      const msg = data.message
      const sentToMe = msg.to_user_id === myId && msg.from_user_id !== myId
      const reading = typeof window !== 'undefined' && window.location.pathname === `/chat/${msg.from_user_id}`

      if (sentToMe && !reading) markUnread(msg.from_user_id)
    })
    return unsub
  }, [myId])

  if (unread.size === 0) return null

  return <span className="menu-dot" title={`${unread.size} unread conversation(s)`} />
}
