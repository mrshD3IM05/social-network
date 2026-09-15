'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { apiGet } from '@/lib/api'
import Avatar from '@/components/Avatar'
import Icon from '@/components/Icon'
import PageHeader from '@/components/PageHeader'

// List of people you can chat with (the authors in your feed).
export default function ChatListPage() {
  const [people, setPeople] = useState([])

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

  return (
    <>
      <PageHeader label="Inbox" title="Messages" subtitle="Pick someone to start a real-time conversation." />

      {people.length === 0 && (
        <div className="empty">
          <p className="empty-title">No conversations yet</p>
          <p>Open a profile and press Message to start one.</p>
        </div>
      )}

      <div className="card list">
        {people.map(person => (
          <Link key={person.id} href={`/chat/${person.id}`} className="list-item">
            <Avatar user={person} size={44} />
            <span className="list-text">
              <strong>{person.first_name} {person.last_name}</strong>
              <small>@{person.nickname}</small>
            </span>
            <Icon name="chat" size={16} />
          </Link>
        ))}
      </div>
    </>
  )
}
