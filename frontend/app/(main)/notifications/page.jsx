'use client'

import { useEffect, useState } from 'react'
import Icon from '@/components/Icon'
import PageHeader from '@/components/PageHeader'

// Notifications arrive in real time over the WebSocket while this page is open.
// (The API has no endpoint to list old notifications yet.)
export default function NotificationsPage() {
  const [notifications, setNotifications] = useState([])

  useEffect(() => {
    const socket = new WebSocket(`ws://${window.location.hostname}:8080/api/v1/ws`)

    socket.onmessage = event => {
      const data = JSON.parse(event.data)
      if (data.type === 'notification') {
        setNotifications(list => [data.notification, ...list])
      }
    }

    return () => socket.close()
  }, [])

  return (
    <>
      <PageHeader label="Activity" title="Notifications" subtitle="New activity appears here in real time." />

      {notifications.length === 0 && (
        <div className="empty">
          <p className="empty-title">You are all caught up</p>
          <p>Follow requests and invitations will show up here.</p>
        </div>
      )}

      <div className="card list">
        {notifications.map(n => (
          <div key={n.id} className="list-item">
            <span className="list-icon"><Icon name="bell" size={16} /></span>
            <span className="list-text">
              <strong>{n.content || n.type.replaceAll('_', ' ')}</strong>
              <small>{new Date(n.created_at).toLocaleString()}</small>
            </span>
          </div>
        ))}
      </div>
    </>
  )
}
