'use client'

import { useEffect, useState } from 'react'
import { getUnread, markUnread, onUnreadChange } from '@/lib/unread'

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
    const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const socket = new WebSocket(`${scheme}//${window.location.host}/api/v1/ws`)

    socket.onmessage = event => {
      const data = JSON.parse(event.data)
      if (data.type !== 'message') return

      const msg = data.message
      const sentToMe = msg.to_user_id === myId && msg.from_user_id !== myId
      // a message you are reading right now is already read
      const reading = window.location.pathname === `/chat/${msg.from_user_id}`

      if (sentToMe && !reading) markUnread(msg.from_user_id)
    }

    return () => socket.close()
  }, [myId])

  if (unread.size === 0) return null

  return <span className="menu-dot" title={`${unread.size} unread conversation(s)`} />
}
