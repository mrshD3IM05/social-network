'use client'

import { useEffect, useRef, useState } from 'react'
import Link from 'next/link'
import { apiGet, apiPost } from '@/lib/api'
import { subscribe } from '@/lib/socket'
import usePaged from '@/lib/usePaged'
import Avatar from '@/components/Avatar'
import LoadMore from '@/components/LoadMore'
import PageHeader from '@/components/PageHeader'
import PersonRow from '@/components/PersonRow'

// notifications that come with something to accept or decline
const REQUEST_TYPES = ['follow_request', 'group_invitation', 'group_join_request']

export default function NotificationsPage() {
  const notifications = usePaged('/notifications') // 10 at a time
  const markedRead = useRef(false)
  const [followRequests, setFollowRequests] = useState([])
  const [invitations, setInvitations] = useState([])
  const [joinRequests, setJoinRequests] = useState([])
  const [error, setError] = useState('')

  function loadRequests() {
    apiGet('/follow-requests').then(setFollowRequests).catch(() => {})
    apiGet('/group-invitations').then(setInvitations).catch(() => {})
    apiGet('/group-join-requests').then(setJoinRequests).catch(() => {})
  }

  // once the first page is shown with its "read" flags (so the new ones are
  // highlighted), mark everything as seen
  useEffect(() => {
    if (notifications.items && !markedRead.current) {
      markedRead.current = true
      apiPost('/notifications/read').catch(() => {})
    }
  }, [notifications.items])

  useEffect(() => {
    loadRequests()

    // new ones arrive in real time over the one app-wide connection
    const unsub = subscribe(data => {
      if (data.type !== 'notification') return
      notifications.setItems(list => [data.notification, ...(list || [])])
      apiPost('/notifications/read').catch(() => {})
      if (REQUEST_TYPES.includes(data.notification.type)) loadRequests()
    })
    return unsub
  }, [])

  // answer a request, then drop it from its list
  async function respond(path, accept, id, setList) {
    setError('')
    try {
      await apiPost(`${path}/${id}/${accept ? 'accept' : 'decline'}`)
      setList(list => list.filter(item => item.id !== id))
    } catch (err) {
      setError(err.message)
    }
  }

  function actions(path, id, setList) {
    return (
      <div className="invitation-actions">
        <button className="btn btn-sm" onClick={() => respond(path, true, id, setList)}>Accept</button>
        <button className="btn btn-light btn-sm" onClick={() => respond(path, false, id, setList)}>Decline</button>
      </div>
    )
  }

  const requestCount = followRequests.length + invitations.length + joinRequests.length

  return (
    <>
      <PageHeader title="Notifications" subtitle="Requests to answer and what happened lately." />

      {error && <p className="error">{error}</p>}

      {requestCount > 0 && (
        <section className="card invitations">
          <h2>Requests</h2>

          {followRequests.map(request => (
            <PersonRow key={`f${request.id}`} person={request.user}>
              <small className="meta">wants to follow you</small>
              {actions('/follow-requests', request.id, setFollowRequests)}
            </PersonRow>
          ))}

          {invitations.map(inv => (
            <div key={`i${inv.id}`} className="list-item">
              <Avatar user={{ first_name: inv.from_first_name, last_name: inv.from_last_name, avatar: inv.from_avatar }} size={40} />
              <span className="list-text">
                <strong>You are invited to join “{inv.group_title}”</strong>
                <small>{inv.from_first_name} {inv.from_last_name} invited you</small>
              </span>
              {actions('/group-invitations', inv.id, setInvitations)}
            </div>
          ))}

          {joinRequests.map(request => (
            <div key={`j${request.id}`} className="list-item">
              <Avatar user={request} size={40} />
              <span className="list-text">
                <strong>{request.first_name} {request.last_name} wants to join “{request.group_title}”</strong>
                <small>@{request.nickname}</small>
              </span>
              {actions('/group-join-requests', request.id, setJoinRequests)}
            </div>
          ))}
        </section>
      )}

      {notifications.error && <p className="error">{notifications.error.message}</p>}
      {notifications.items === null && !notifications.error && <p className="loading">Loading…</p>}

      {notifications.items?.length === 0 && requestCount === 0 && (
        <div className="empty">
          <p className="empty-title">You are all caught up</p>
          <p>Follow requests, invitations and group events will show up here.</p>
        </div>
      )}

      {notifications.items?.length > 0 && (
        <div className="card list">
          {notifications.items.map(n => {
            const actor = { first_name: n.actor_first_name, last_name: n.actor_last_name, avatar: n.actor_avatar }
            return (
              <Link
                key={n.id}
                href={n.group_id ? `/groups/${n.group_id}` : `/profile/${n.actor_id}`}
                className="list-item"
              >
                <Avatar user={actor} size={40} />
                <span className="list-text">
                  <strong>{n.content || n.type.replaceAll('_', ' ')}</strong>
                  <small>{new Date(n.created_at).toLocaleString()}</small>
                </span>
                {!n.read && <span className="menu-dot" title="New" />}
              </Link>
            )
          })}
        </div>
      )}
      <LoadMore list={notifications} />
    </>
  )
}
