'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { apiGet } from '@/lib/api'
import Avatar from '@/components/Avatar'
import Icon from '@/components/Icon'
import PageHeader from '@/components/PageHeader'

export default function PeoplePage() {
  const [people, setPeople] = useState([])
  const [search, setSearch] = useState('')

  // The API has no "list users" endpoint yet (GET /users returns 501),
  // so we collect the authors of the posts in the feed.
  useEffect(() => {
    async function load() {
      const me = await apiGet('/me')
      const posts = await apiGet('/posts')

      const found = {}
      for (const post of posts) {
        if (post.author_id !== me.id) {
          found[post.author_id] = {
            id: post.author_id,
            first_name: post.author_first_name,
            last_name: post.author_last_name,
            nickname: post.author_nickname,
            avatar: post.author_avatar,
          }
        }
      }
      setPeople(Object.values(found))
    }
    load()
  }, [])

  // keep only the people whose name contains the search text
  const shown = people.filter(person =>
    `${person.first_name} ${person.last_name} ${person.nickname}`
      .toLowerCase()
      .includes(search.toLowerCase())
  )

  return (
    <>
      <PageHeader label="Directory" title="People" subtitle="Everyone who appears in your feed." />

      <div className="search">
        <Icon name="search" />
        <input placeholder="Search by name or nickname" value={search} onChange={e => setSearch(e.target.value)} />
      </div>

      {shown.length === 0 && (
        <div className="empty">
          <p className="empty-title">No one found</p>
          <p>People show up here once their posts are in your feed.</p>
        </div>
      )}

      <div className="card list">
        {shown.map(person => (
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
