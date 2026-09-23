'use client'

import { useEffect, useState } from 'react'
import { fetchPeople, searchPeople } from '@/lib/people'
import { LIMITS } from '@/lib/validate'
import Icon from '@/components/Icon'
import PageHeader from '@/components/PageHeader'
import PersonRow from '@/components/PersonRow'

export default function PeoplePage() {
  const [people, setPeople] = useState(null)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')

  useEffect(() => {
    fetchPeople()
      .then(setPeople)
      .catch(err => setError(err.message))
  }, [])

  const shown = searchPeople(people || [], search)

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

      {error && <p className="error">{error}</p>}

      {people === null && !error && <p className="loading">Loading…</p>}

      {people !== null && shown.length === 0 && (
        <div className="empty">
          <p className="empty-title">No one found</p>
          <p>{people.length === 0 ? 'You are the only member so far.' : 'No one matches that search.'}</p>
        </div>
      )}

      <div className="card list">
        {shown.map(person => (
          <PersonRow key={person.id} person={person} href={`/profile/${person.id}`}>
            <Icon name="arrow" size={16} />
          </PersonRow>
        ))}
      </div>
    </>
  )
}
