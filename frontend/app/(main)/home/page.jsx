'use client'

import { useEffect, useState } from 'react'
import { apiGet } from '@/lib/api'
import PageHeader from '@/components/PageHeader'
import PostForm from '@/components/PostForm'
import PostCard from '@/components/PostCard'

const PAGE_SIZE = 20

export default function HomePage() {
  const [me, setMe] = useState(null)
  const [posts, setPosts] = useState([])
  const [hasMore, setHasMore] = useState(false)
  const [error, setError] = useState('')

  // The API sends one page at a time. Asking from offset 0 reloads the feed,
  // any other offset adds the next page to what is already on screen.
  function loadPosts(offset = 0) {
    apiGet(`/posts?limit=${PAGE_SIZE}&offset=${offset}`)
      .then(page => {
        setPosts(current => (offset === 0 ? page : [...current, ...page]))
        setHasMore(page.length === PAGE_SIZE)
      })
      .catch(err => setError(err.message))
  }

  // runs once when the page opens
  useEffect(() => {
    apiGet('/me').then(setMe)
    loadPosts()
  }, [])

  if (!me) return <p className="loading">Loading…</p>

  return (
    <>
      <PageHeader label="Feed" title={`Good to see you, ${me.first_name}.`} subtitle="The latest from you and the people you follow." />

      <PostForm onPosted={loadPosts} myId={me.id} />

      {error && <p className="error">{error}</p>}
      {posts.length === 0 && (
        <div className="empty">
          <p className="empty-title">Nothing here yet</p>
          <p>Write the first post, or follow people to fill your feed.</p>
        </div>
      )}

      {posts.map(post => (
        <PostCard key={post.id} post={post} myId={me.id} onDeleted={() => loadPosts()} />
      ))}

      {hasMore && (
        <div className="row-actions end">
          <button className="btn btn-light" onClick={() => loadPosts(posts.length)}>Load more</button>
        </div>
      )}
    </>
  )
}
