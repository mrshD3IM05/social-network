'use client'

import { useEffect, useState } from 'react'
import { apiGet, apiPost } from '@/lib/api'
import { onSocketEvent } from '@/lib/ws'
import FollowRequests from '@/components/FollowRequests'
import Icon from '@/components/Icon'
import PageHeader from '@/components/PageHeader'

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState([])
  const [unread, setUnread] = useState(0)
  const [error, setError] = useState('')

  function load() {
    apiGet('/notifications')
      .then(data => {
        setNotifications(data.notifications)
        setUnread(data.unread)
      })
      .catch(err => setError(err.message))
  }

  useEffect(() => {
    // the stored history, so notifications received while you were on another
    // page (or logged out) are not lost
    load()

    // plus anything that arrives while this page is open
    return onSocketEvent(event => {
      if (event.type === 'notification') {
        setNotifications(list => [event.notification, ...list])
        setUnread(count => count + 1)
      }
    })
  }, [])

  async function markAllRead() {
    try {
      await apiPost('/notifications/read')
      setNotifications(list => list.map(n => ({ ...n, read: true })))
      setUnread(0)
    } catch (err) {
      setError(err.message)
    }
  }

  async function markRead(notification) {
    if (notification.read) return
    try {
      await apiPost(`/notifications/${notification.id}/read`)
      setNotifications(list => list.map(n => (n.id === notification.id ? { ...n, read: true } : n)))
      setUnread(count => Math.max(0, count - 1))
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <>
      <PageHeader
        label="Activity"
        title="Notifications"
        subtitle={unread > 0 ? `${unread} unread` : 'You are up to date.'}
      />

      {error && <p className="error">{error}</p>}

      <FollowRequests onChange={load} />

      {unread > 0 && (
        <div className="row-actions end">
          <button className="btn btn-light" onClick={markAllRead}>Mark all as read</button>
        </div>
      )}

      {notifications.length === 0 && (
        <div className="empty">
          <p className="empty-title">You are all caught up</p>
          <p>Follow requests and invitations will show up here.</p>
        </div>
      )}

      <div className="card list">
        {notifications.map(n => (
          <button
            key={n.id}
            type="button"
            className={n.read ? 'list-item' : 'list-item unread'}
            onClick={() => markRead(n)}
          >
            <span className="list-icon"><Icon name="bell" size={16} /></span>
            <span className="list-text">
              <strong>{n.content || n.type.replaceAll('_', ' ')}</strong>
              <small>{new Date(n.created_at).toLocaleString()}</small>
            </span>
            {!n.read && <span className="dot" aria-label="unread" />}
          </button>
        ))}
      </div>
    </>
  )
}
