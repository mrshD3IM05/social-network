'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { apiGet } from '@/lib/api'
import { fetchPeople } from '@/lib/people'
import { LIMITS } from '@/lib/validate'
import Avatar from '@/components/Avatar'
import Icon from '@/components/Icon'
import PageHeader from '@/components/PageHeader'

export default function PeoplePage() {
  const [people, setPeople] = useState([])
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')

  useEffect(() => {
    fetchPeople().catch(err => setError(err.message))
  }, [])

  // keep only the people whose name contains the search text
  function shown(list) {
    return list.filter(person =>
      `${person.first_name} ${person.last_name} ${person.nickname}`
        .toLowerCase()
        .includes(search.toLowerCase())
    )
  }

  return (
    <>
      <PageHeader label="Directory" title="People" subtitle="Everyone who appears in your feed." />

      <div className="search">
        <Icon name="search" />
        <input
          placeholder="Search by name or nickname"
          value={search}
          maxLength={LIMITS.search}
          onChange={e => setSearch(e.target.value)}
        />
      </div>

      {error && <p className="error">{error}</p>}

      {shown(people).length === 0 && !error && (
        <div className="empty">
          <p className="empty-title">No one found</p>
          <p>People show up here once their posts are in your feed.</p>
        </div>
      )}

      <div className="card list">
        {shown(people).map(person => (
          <Link key={person.id} href={`/profile/${person.id}`} className="list-item">
            <Avatar user={person} size={44} />
            <span className="list-text">
              <strong>{person.first_name} {person.last_name}</strong>
              <small>@{person.nickname}</small>
            </span>
            <Icon name="arrow" size={16} />
          </Link>
        ))}
      </div>
    </>
  )
}
