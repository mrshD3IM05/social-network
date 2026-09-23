'use client'

import Link from 'next/link'
import Avatar from '@/components/Avatar'

// One person in a list: avatar, name and @nickname, with whatever action the
// page needs on the right (a chip, a button, an icon). Pass `href` to turn the
// whole row into a link. Every "pick a person" list uses this row.
export default function PersonRow({ person, href, size = 44, children }) {
  const body = (
    <>
      <Avatar user={person} size={size} />
      <span className="list-text">
        <strong>{person.first_name} {person.last_name}</strong>
        <small>@{person.nickname}</small>
      </span>
      {children}
    </>
  )

  if (href) {
    return <Link href={href} className="list-item">{body}</Link>
  }
  return <div className="list-item">{body}</div>
}
