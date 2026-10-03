'use client'

import { useState } from 'react'
import Link from 'next/link'
import Avatar from './Avatar'

// One thing waiting for an answer: a follow request, a group invitation or a
// join request. Avatar (linked to the person when `href` is given), what it is,
// then Accept / Decline. On a phone the buttons drop under the text, side by
// side. `onRespond(accept)` does the request; while it runs both buttons are
// disabled so a double tap cannot answer twice, and the parent removes the row
// only once it succeeded.
export default function RequestRow({ person, href, title, subtitle, onRespond }) {
  const [busy, setBusy] = useState(null) // 'accept' | 'decline' while answering

  async function answer(accept) {
    setBusy(accept ? 'accept' : 'decline')
    try {
      await onRespond(accept)
    } finally {
      setBusy(null)
    }
  }

  const avatar = <Avatar user={person} size={40} />

  return (
    <div className="list-item request-item">
      {href ? <Link href={href} className="request-avatar">{avatar}</Link> : <span className="request-avatar">{avatar}</span>}
      <span className="list-text">
        <strong>{title}</strong>
        {subtitle && <small>{subtitle}</small>}
      </span>
      <div className="invitation-actions">
        <button type="button" className="btn btn-sm" disabled={busy !== null} onClick={() => answer(true)}>
          {busy === 'accept' ? 'Accepting…' : 'Accept'}
        </button>
        <button type="button" className="btn btn-light btn-sm" disabled={busy !== null} onClick={() => answer(false)}>
          {busy === 'decline' ? 'Declining…' : 'Decline'}
        </button>
      </div>
    </div>
  )
}
