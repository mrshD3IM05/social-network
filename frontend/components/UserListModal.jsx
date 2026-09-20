'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { apiGet } from '@/lib/api'
import Avatar from './Avatar'
import Modal from './Modal'

// Shows the followers or the following list of a profile in a dialog.
export default function UserListModal({ title, path, onClose }) {
  const [users, setUsers] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    apiGet(`${path}?limit=100`)
      .then(setUsers)
      .catch(err => {
        setError(err.message)
        setUsers([])
      })
  }, [path])

  return (
    <Modal title={title} onClose={onClose}>
      {users === null && <p className="loading">Loading…</p>}
      {error && <p className="error">{error}</p>}

      {users?.length === 0 && !error && <p className="meta">Nobody here yet.</p>}

      <div className="list">
        {users?.map(person => (
          <Link key={person.id} href={`/profile/${person.id}`} className="list-item" onClick={onClose}>
            <Avatar user={person} size={40} />
            <span className="list-text">
              <strong>{person.first_name} {person.last_name}</strong>
              {person.nickname && <small>@{person.nickname}</small>}
            </span>
          </Link>
        ))}
      </div>
    </Modal>
  )
}
