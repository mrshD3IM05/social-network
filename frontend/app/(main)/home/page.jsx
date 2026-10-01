'use client'

import { useEffect, useState } from 'react'
import { apiGet } from '@/lib/api'
import usePaged from '@/lib/usePaged'
import LoadMore from '@/components/LoadMore'
import PageHeader from '@/components/PageHeader'
import PostForm from '@/components/PostForm'
import PostCard from '@/components/PostCard'

export default function HomePage() {
  const [me, setMe] = useState(null)
  const posts = usePaged('/posts') // the feed, 10 posts at a time

  // runs once when the page opens
  useEffect(() => {
    apiGet('/me').then(setMe)
  }, [])

  if (!me) return <p className="loading">Loading…</p>

  return (
    <>
      <PageHeader label="Feed" title={`Good to see you, ${me.first_name}.`} subtitle="The latest from you and the people you follow." />

      <PostForm onPosted={posts.reload} />

      {posts.error && <p className="error">{posts.error.message}</p>}
      {posts.items?.length === 0 && (
        <div className="empty">
          <p className="empty-title">Nothing here yet</p>
          <p>Write the first post, or follow people to fill your feed.</p>
        </div>
      )}

      {posts.items?.map(post => (
        <PostCard key={post.id} post={post} myId={me.id} onDeleted={posts.reload} />
      ))}
      <LoadMore list={posts} />
    </>
  )
}
