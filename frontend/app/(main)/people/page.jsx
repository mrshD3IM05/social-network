'use client'

import { useEffect, useState } from 'react'
import { fetchPeople } from '@/lib/people'
import { LIMITS } from '@/lib/validate'
import { useDebouncedValue } from '@/lib/timing'
import Icon from '@/components/Icon'
import PageHeader from '@/components/PageHeader'
import PersonRow from '@/components/PersonRow'

export default function PeoplePage() {
  const [people, setPeople] = useState(null)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')

  // search once typing pauses, not on every keystroke
  const query = useDebouncedValue(search, 250).trim()

  useEffect(() => {
    // nothing is looked up until there is a nickname to search for
    if (!query) {
      setPeople(null)
      setError('')
      return
    }

    let active = true
    setPeople(null)
    setError('')
    fetchPeople(query)
      .then(list => { if (active) setPeople(list) })
      .catch(err => { if (active) setError(err.message) })
    return () => { active = false }
  }, [query])

  return (
    <>
      <PageHeader label="Directory" title="People" subtitle="Search for someone by their nickname." />

      <div className="search">
        <Icon name="search" />
        <input
          placeholder="Search by nickname"
          value={search}
          maxLength={LIMITS.search}
          onChange={e => setSearch(e.target.value)}
        />
      </div>

      {error && <p className="error">{error}</p>}

      {!query && !error && (
        <div className="empty">
          <p className="empty-title">Find people</p>
          <p>Type a nickname to search the directory.</p>
        </div>
      )}

      {query && people === null && !error && <p className="loading">Loading…</p>}

      {query && people !== null && people.length === 0 && (
        <div className="empty">
          <p className="empty-title">No one found</p>
          <p>No one matches that nickname.</p>
        </div>
      )}

      {people !== null && people.length > 0 && (
        <div className="card list">
          {people.map(person => (
            <PersonRow key={person.id} person={person} href={`/profile/${person.id}`}>
              <Icon name="arrow" size={16} />
            </PersonRow>
          ))}
        </div>
      )}
    </>
  )
}
