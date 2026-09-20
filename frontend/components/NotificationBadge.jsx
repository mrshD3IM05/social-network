'use client'

import { useEffect, useState } from 'react'
import { usePathname } from 'next/navigation'
import { apiGet } from '@/lib/api'
import { onSocketEvent } from '@/lib/ws'

// The unread counter next to the Notifications link. It lives in the sidebar,
// which every page renders, so a notification that arrives while you are
// somewhere else is still visible.
export default function NotificationBadge() {
  const [unread, setUnread] = useState(0)
  const pathname = usePathname()

  function refresh() {
    apiGet('/notifications/unread')
      .then(data => setUnread(data.unread))
      .catch(() => {})
  }

  useEffect(() => {
    refresh()
    // the notifications page marks everything read, so re-read the count on nav
  }, [pathname])

  useEffect(() => {
    return onSocketEvent(event => {
      if (event.type === 'notification') setUnread(count => count + 1)
      if (event.type === 'socket' && event.state === 'open') refresh()
    })
  }, [])

  if (unread === 0) return null

  return <span className="badge" aria-label={`${unread} unread notifications`}>{unread > 99 ? '99+' : unread}</span>
}
