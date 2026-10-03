'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { apiGet } from '@/lib/api'
import { fetchContacts } from '@/lib/people'
import Avatar from './Avatar'
import Icon from './Icon'

// The right column on wide screens: people you are not connected to yet,
// your groups, and the next events in those groups. Each block hides itself
// when it has nothing to show, and a failed request only hides its own block.
export default function SidePanel() {
  const [people, setPeople] = useState(null)
  const [groups, setGroups] = useState(null)
  const [events, setEvents] = useState(null)

  useEffect(() => {
    // /users already leaves you out; /contacts are the people you are linked with
    Promise.all([apiGet('/users'), fetchContacts().catch(() => [])])
      .then(([users, contacts]) => {
        const known = new Set((contacts || []).map(c => c.id))
        setPeople((users || []).filter(u => !known.has(u.id)).slice(0, 4))
      })
      .catch(() => setPeople([]))

    apiGet('/groups')
      .then(list => setGroups((list || []).filter(g => g.is_member || g.is_creator)))
      .catch(() => setGroups([]))

    // one call for the next events across all your groups, soonest first
    apiGet('/events/upcoming')
      .then(list => setEvents(list || []))
      .catch(() => setEvents([]))
  }, [])

  return (
    <aside className="side-panel" aria-label="Suggestions">
      {events?.length > 0 && (
        <section className="panel">
          <h2 className="panel-title">Coming up</h2>
          {events.map(event => {
            const when = new Date(event.date_time)
            return (
              <Link key={event.id} href={`/groups/${event.group_id}`} className="panel-event">
                <span className="date-tile">
                  <small>{when.toLocaleDateString(undefined, { month: 'short' })}</small>
                  <strong>{when.getDate()}</strong>
                </span>
                <span className="list-text">
                  <strong>{event.title}</strong>
                  <small>
                    {event.group_title},{when.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })}
                  </small>
                </span>
              </Link>
            )
          })}
        </section>
      )}

      {people?.length > 0 && (
        <section className="panel">
          <div className="panel-head">
            <h2 className="panel-title">People you may know</h2>
            <Link href="/people" className="panel-link">See all</Link>
          </div>
          {people.map(person => (
            <Link key={person.id} href={`/profile/${person.id}`} className="panel-row">
              <Avatar user={person} size={36} />
              <span className="list-text">
                <strong>{person.first_name} {person.last_name}</strong>
                <small>@{person.nickname}</small>
              </span>
              <span className="panel-go"><Icon name="arrow" size={15} /></span>
            </Link>
          ))}
        </section>
      )}

      {groups?.length > 0 && (
        <section className="panel">
          <div className="panel-head">
            <h2 className="panel-title">Your groups</h2>
            <Link href="/groups" className="panel-link">See all</Link>
          </div>
          {groups.slice(0, 5).map(group => (
            <Link key={group.id} href={`/groups/${group.id}`} className="panel-row">
              {group.avatar ? (
                <Avatar user={group} size={36} />
              ) : (
                <span className="list-icon panel-icon"><Icon name="grid" size={16} /></span>
              )}
              <span className="list-text">
                <strong>{group.title}</strong>
                <small>{group.member_count} member{group.member_count === 1 ? '' : 's'}</small>
              </span>
            </Link>
          ))}
        </section>
      )}
    </aside>
  )
}
