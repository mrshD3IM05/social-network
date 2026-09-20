'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { apiGet, apiPost } from '@/lib/api'
import { onSocketEvent } from '@/lib/ws'
import Avatar from './Avatar'

// Pending follow requests addressed to you, with the accept / decline buttons.
// The API endpoints existed already, but nothing listed the request ids, so
// they could not be reached from the app.
export default function FollowRequests({ onChange }) {
  const [requests, setRequests] = useState([])
  const [error, setError] = useState('')

  function load() {
    apiGet('/follow-requests')
      .then(list => {
        setRequests(list)
        onChange?.(list.length)
      })
      .catch(err => setError(err.message))
  }

  useEffect(() => {
    load()
    return onSocketEvent(event => {
      if (event.type === 'notification' && event.notification?.type === 'follow_request') load()
    })
  }, [])

  async function respond(id, decision) {
    try {
      await apiPost(`/follow-requests/${id}/${decision}`)
      load()
    } catch (err) {
      setError(err.message)
    }
  }

  if (requests.length === 0) return null

  return (
    <section className="card list">
      <p className="eyebrow section-label">Follow requests</p>
      {error && <p className="error">{error}</p>}
      {requests.map(request => (
        <div key={request.id} className="list-item">
          <Avatar user={request.from_user} size={40} />
          <span className="list-text">
            <Link href={`/profile/${request.from_user.id}`}>
              <strong>{request.from_user.first_name} {request.from_user.last_name}</strong>
            </Link>
            <small>wants to follow you</small>
          </span>
          <span className="row-actions">
            <button className="btn" onClick={() => respond(request.id, 'accept')}>Accept</button>
            <button className="btn btn-light" onClick={() => respond(request.id, 'decline')}>Decline</button>
          </span>
        </div>
      ))}
    </section>
  )
}
