'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import { apiDelete, apiGet, apiPost } from '@/lib/api'
import Avatar from '@/components/Avatar'
import Icon from '@/components/Icon'
import PostCard from '@/components/PostCard'

export default function ProfilePage() {
  const { id } = useParams() // the [id] from the URL, e.g. /profile/3
  const [me, setMe] = useState(null)
  const [user, setUser] = useState(null)
  const [posts, setPosts] = useState([])
  const [isPrivate, setIsPrivate] = useState(false)
  const [message, setMessage] = useState('')

  async function load() {
    setMe(await apiGet('/me'))
    try {
      setUser(await apiGet(`/user/${id}`))
      // no "posts of one user" endpoint yet, so we filter the feed
      const feed = await apiGet('/posts')
      setPosts(feed.filter(post => post.author_id === Number(id)))
    } catch (err) {
      if (err.status === 403) setIsPrivate(true) // private profile you don't follow
      else setMessage(err.message)
    }
  }

  useEffect(() => {
    load()
  }, [id])

  async function follow() {
    try {
      const result = await apiPost(`/users/${id}/follow`)
      setMessage(result.status === 'pending' ? 'Follow request sent.' : 'You are now following.')
    } catch (err) {
      setMessage(err.message)
    }
  }

  async function unfollow() {
    await apiDelete(`/users/${id}/follow`)
    setMessage('Unfollowed.')
  }

  if (isPrivate) {
    return (
      <div className="card locked">
        <span className="locked-icon"><Icon name="lock" size={22} /></span>
        <h2>This profile is private</h2>
        <p className="subtitle">Send a follow request to see their profile and posts.</p>
        <button className="btn" onClick={follow}>Request to follow</button>
        {message && <p className="notice">{message}</p>}
      </div>
    )
  }

  if (!user || !me) return <p className="loading">{message || 'Loading…'}</p>

  const isMe = me.id === user.id

  return (
    <>
      <section className="card profile">
        <div className="profile-cover" />
        <div className="profile-body">
          <Avatar user={user} size={96} />

          <div className="profile-top">
            <div>
              <h1>{user.first_name} {user.last_name}</h1>
              <p className="meta">@{user.nickname} · {user.private ? 'Private' : 'Public'} profile</p>
            </div>

            <div className="profile-buttons">
              {isMe ? (
                <Link href="/settings" className="btn btn-light">Edit settings</Link>
              ) : (
                <>
                  <button className="btn" onClick={follow}>Follow</button>
                  <button className="btn btn-light" onClick={unfollow}>Unfollow</button>
                  <Link href={`/chat/${user.id}`} className="btn btn-light">Message</Link>
                </>
              )}
            </div>
          </div>

          {user.about_me && <p className="profile-about">{user.about_me}</p>}

          <div className="profile-stats">
            <span><strong>{posts.length}</strong> posts</span>
            <span>Joined {new Date(user.created_at).toLocaleDateString(undefined, { month: 'long', year: 'numeric' })}</span>
            {isMe && <span>{me.email}</span>}
          </div>
        </div>
      </section>

      {message && <p className="notice">{message}</p>}

      <p className="eyebrow section-label">Posts</p>
      {posts.length === 0 && (
        <div className="empty">
          <p className="empty-title">No posts to show</p>
          <p>Posts you are allowed to see will appear here.</p>
        </div>
      )}
      {posts.map(post => (
        <PostCard key={post.id} post={post} myId={me.id} onDeleted={load} />
      ))}
    </>
  )
}
