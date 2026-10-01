'use client'

import { useEffect, useState } from 'react'
import { usePathname } from 'next/navigation'
import { apiGet, socketUrl } from '@/lib/api'

// The number of unread notifications, next to Notifications in the sidebar,
// so they can be seen from every page. It is a number and not a dot to look
// different from new messages, like the subject asks.
export default function NotificationBadge() {
  const pathname = usePathname()
  const onPage = pathname === '/notifications' // that page marks them all as read
  const [count, setCount] = useState(0)

  useEffect(() => {
    apiGet('/notifications/unread')
      .then(result => {
        if (window.location.pathname !== '/notifications') setCount(result.count)
      })
      .catch(() => {})

    const socket = new WebSocket(socketUrl())
    socket.onmessage = event => {
      const data = JSON.parse(event.data)
      if (data.type === 'notification' && window.location.pathname !== '/notifications') {
        setCount(c => c + 1)
      }
    }
    return () => socket.close()
  }, [])

  useEffect(() => {
    if (onPage) setCount(0)
  }, [onPage])

  if (count === 0) return null
  return <span className="menu-count">{count > 99 ? '99+' : count}</span>
}
