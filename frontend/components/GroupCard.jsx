'use client'

import Link from 'next/link'
import Avatar from '@/components/Avatar'
import Icon from '@/components/Icon'

// One group in the browse list: title, description, member count and a
// status chip ("You", "Member", "Requested") or a Join button for outsiders.
export default function GroupCard({ group, onJoin, joining }) {
  return (
    <Link href={`/groups/${group.id}`} className="list-item group-item">
      <span className="list-icon"><Icon name="grid" size={18} /></span>
      <span className="list-text">
        <strong>{group.title}</strong>
        {group.description && <small>{group.description}</small>}
        <small className="meta">{group.member_count} member{group.member_count === 1 ? '' : 's'}</small>
      </span>

      {group.is_creator ? (
        <span className="chip chip-accent">You</span>
      ) : group.is_member ? (
        <span className="chip">Member</span>
      ) : group.pending_join ? (
        <span className="chip">Requested</span>
      ) : (
        <button
          className="btn btn-light btn-sm"
          onClick={e => {
            e.preventDefault() // don't follow the link
            e.stopPropagation()
            onJoin(group)
          }}
          disabled={joining}
        >
          {joining ? '…' : 'Join'}
        </button>
      )}
    </Link>
  )
}
