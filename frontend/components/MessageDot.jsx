'use client'

import { useEffect, useState } from 'react'
import { usePathname } from 'next/navigation'

// A small dot next to Messages when somebody writes to you.
//
// It is kept apart from the bell on purpose: the subject asks for new messages
// and new notifications to be shown in different ways.
export default function MessageDot({ myId }) {
  const [unread, setUnread] = useState(false)
  const pathname = usePathname()

  // opening the messages section means you have seen them
  useEffect(() => {
    if (pathname.startsWith('/chat')) setUnread(false)
  }, [pathname])

  useEffect(() => {
    const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const socket = new WebSocket(`${scheme}//${window.location.host}/api/v1/ws`)

    socket.onmessage = event => {
      const data = JSON.parse(event.data)
      if (data.type !== 'message') return

      const msg = data.message
      const sentToMe = msg.to_user_id === myId && msg.from_user_id !== myId
      // no dot while you are already reading your messages
      if (sentToMe && !window.location.pathname.startsWith('/chat')) {
        setUnread(true)
      }
    }

    return () => socket.close()
  }, [myId])

  if (!unread) return null

  return <span className="menu-dot" title="New message" />
}
