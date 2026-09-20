'use client'

import { useEffect, useState } from 'react'
import { apiGet } from '@/lib/api'
import Avatar from './Avatar'

// Lets the author pick which of their followers may read a "private" post.
// selected is an array of user ids; onChange receives the new array.
export default function AudiencePicker({ myId, selected, onChange }) {
  const [followers, setFollowers] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!myId) return
    apiGet(`/users/${myId}/followers?limit=100`)
      .then(setFollowers)
      .catch(err => {
        setError(err.message)
        setFollowers([])
      })
  }, [myId])

  function toggle(id) {
    onChange(selected.includes(id) ? selected.filter(value => value !== id) : [...selected, id])
  }

  if (followers === null) return <p className="meta">Loading your followers…</p>

  if (followers.length === 0) {
    return (
      <p className="hint">
        Nobody follows you yet, so this post would stay visible to you only.
        {error && ` (${error})`}
      </p>
    )
  }

  return (
    <fieldset className="audience">
      <legend>Who can see this post</legend>
      <div className="audience-list">
        {followers.map(person => (
          <label key={person.id} className={selected.includes(person.id) ? 'audience-item picked' : 'audience-item'}>
            <input
              type="checkbox"
              checked={selected.includes(person.id)}
              onChange={() => toggle(person.id)}
            />
            <Avatar user={person} size={28} />
            <span>{person.first_name} {person.last_name}</span>
          </label>
        ))}
      </div>
      <p className="hint">
        {selected.length === 0
          ? 'Pick at least one follower, otherwise only you will see it.'
          : `${selected.length} of ${followers.length} followers selected.`}
      </p>
    </fieldset>
  )
}
