'use client'

import { useEffect, useRef, useState } from 'react'
import Link from 'next/link'
import { apiGet, apiPost } from '@/lib/api'
import { subscribe } from '@/lib/socket'
import usePaged from '@/lib/usePaged'
import Avatar from '@/components/Avatar'
import LoadMore from '@/components/LoadMore'
import PageHeader from '@/components/PageHeader'
import RequestRow from '@/components/RequestRow'

// notifications that come with something to accept or decline
const REQUEST_TYPES = ['follow_request', 'group_invitation', 'group_join_request']

export default function NotificationsPage() {
  const notifications = usePaged('/notifications') // 10 at a time
  const markedRead = useRef(false)
  const [followRequests, setFollowRequests] = useState([])
  const [invitations, setInvitations] = useState([])
  const [joinRequests, setJoinRequests] = useState([])
  const [error, setError] = useState('')

  // the three request lists come in one response
  function loadRequests() {
    apiGet('/requests')
      .then(all => {
        setFollowRequests(all.follow_requests)
        setInvitations(all.group_invitations)
        setJoinRequests(all.group_join_requests)
      })
      .catch(() => {})
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

  // answer a request, then drop it from its list once the API agreed
  async function respond(path, accept, id, setList) {
    setError('')
    try {
      await apiPost(`${path}/${id}/${accept ? 'accept' : 'decline'}`)
      setList(list => list.filter(item => item.id !== id))
    } catch (err) {
      setError(err.message)
    }
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
            <RequestRow
              key={`f${request.id}`}
              person={request.user}
              href={`/profile/${request.user.id}`}
              title={`${request.user.first_name} ${request.user.last_name} wants to follow you`}
              subtitle={`@${request.user.nickname}`}
              onRespond={accept => respond('/follow-requests', accept, request.id, setFollowRequests)}
            />
          ))}

          {invitations.map(inv => (
            <RequestRow
              key={`i${inv.id}`}
              person={{ first_name: inv.from_first_name, last_name: inv.from_last_name, avatar: inv.from_avatar }}
              href={`/groups/${inv.group_id}`}
              title={`You are invited to join “${inv.group_title}”`}
              subtitle={`${inv.from_first_name} ${inv.from_last_name} invited you`}
              onRespond={accept => respond('/group-invitations', accept, inv.id, setInvitations)}
            />
          ))}

          {joinRequests.map(request => (
            <RequestRow
              key={`j${request.id}`}
              person={request}
              href={`/profile/${request.user_id}`}
              title={`${request.first_name} ${request.last_name} wants to join “${request.group_title}”`}
              subtitle={`@${request.nickname}`}
              onRespond={accept => respond('/group-join-requests', accept, request.id, setJoinRequests)}
            />
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
