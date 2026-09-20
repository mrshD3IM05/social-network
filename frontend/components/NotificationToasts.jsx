'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import Icon from '@/components/Icon'
import { onSocketEvent } from '@/lib/ws'

// Rendered by the (main) layout, so a notification pops up on whichever page
// you happen to be on. Each toast disappears on its own after a few seconds.
export default function NotificationToasts() {
  const [toasts, setToasts] = useState([])

  useEffect(() => {
    return onSocketEvent(event => {
      if (event.type !== 'notification') return
      const notification = event.notification
      const key = `${notification.id}-${Date.now()}`
      setToasts(list => [...list, { key, notification }])
      setTimeout(() => setToasts(list => list.filter(toast => toast.key !== key)), 6000)
    })
  }, [])

  if (toasts.length === 0) return null

  return (
    <div className="toasts" role="status" aria-live="polite">
      {toasts.map(({ key, notification }) => (
        <Link key={key} href="/notifications" className="toast">
          <span className="list-icon"><Icon name="bell" size={16} /></span>
          <span className="list-text">
            <strong>{notification.content || notification.type.replaceAll('_', ' ')}</strong>
            <small>Just now</small>
          </span>
        </Link>
      ))}
    </div>
  )
}
