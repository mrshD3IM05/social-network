'use client'

import { useState } from 'react'
import { LIMITS } from '@/lib/validate'
import { useDebouncedValue } from '@/lib/timing'
import usePaged from '@/lib/usePaged'
import Icon from '@/components/Icon'
import LoadMore from '@/components/LoadMore'
import PageHeader from '@/components/PageHeader'
import PersonRow from '@/components/PersonRow'

export default function PeoplePage() {
  const [search, setSearch] = useState('')
  // the server searches, once typing pauses, and answers 10 people at a time
  const query = useDebouncedValue(search, 250).trim()
  const people = usePaged(`/users?q=${encodeURIComponent(query)}`)

  return (
    <>
      <PageHeader label="Directory" title="People" subtitle="Everyone on the network." />

      <div className="search">
        <Icon name="search" />
        <input
          placeholder="Search by name or nickname"
          value={search}
          maxLength={LIMITS.search}
          onChange={e => setSearch(e.target.value)}
        />
      </div>

      {people.error && <p className="error">{people.error.message}</p>}

      {people.items === null && !people.error && <p className="loading">Loading…</p>}

      {people.items?.length === 0 && (
        <div className="empty">
          <p className="empty-title">No one found</p>
          <p>{query ? 'No one matches that search.' : 'You are the only member so far.'}</p>
        </div>
      )}

      <div className="card list">
        {people.items?.map(person => (
          <PersonRow key={person.id} person={person} href={`/profile/${person.id}`}>
            <Icon name="arrow" size={16} />
          </PersonRow>
        ))}
      </div>
      <LoadMore list={people} />
    </>
  )
}
