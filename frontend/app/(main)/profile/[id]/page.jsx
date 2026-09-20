'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import { apiDelete, apiGet, apiPost } from '@/lib/api'
import Avatar from '@/components/Avatar'
import Icon from '@/components/Icon'
import PostCard from '@/components/PostCard'
import UserListModal from '@/components/UserListModal'

export default function ProfilePage() {
  const { id } = useParams() // the [id] from the URL, e.g. /profile/3
  const [me, setMe] = useState(null)
  const [user, setUser] = useState(null)
  const [posts, setPosts] = useState([])
  const [counts, setCounts] = useState({ followers: 0, following: 0 })
  const [relation, setRelation] = useState('')
  const [isPrivate, setIsPrivate] = useState(false)
  const [message, setMessage] = useState('')
  const [modal, setModal] = useState(null) // 'followers' | 'following'

  async function load() {
    setMe(await apiGet('/me'))
    try {
      setRelation((await apiGet(`/users/${id}/relationship`)).status)
    } catch {
      setRelation('')
    }
    try {
      setUser(await apiGet(`/user/${id}`))
      setIsPrivate(false)
      // a real endpoint now, instead of downloading the whole feed and filtering
      setPosts(await apiGet(`/users/${id}/posts`))
      const [followers, following] = await Promise.all([
        apiGet(`/users/${id}/followers`),
        apiGet(`/users/${id}/following`),
      ])
      setCounts({ followers: followers.length, following: following.length })
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
      setRelation(result.status)
      setMessage(result.status === 'pending' ? 'Follow request sent.' : 'You are now following.')
      if (result.status === 'accepted') load()
    } catch (err) {
      setMessage(err.message)
    }
  }

  async function unfollow() {
    await apiDelete(`/users/${id}/follow`)
    setRelation('')
    setMessage('Unfollowed.')
    load()
  }

  if (isPrivate) {
    return (
      <div className="card locked">
        <span className="locked-icon"><Icon name="lock" size={22} /></span>
        <h2>This profile is private</h2>
        <p className="subtitle">Send a follow request to see their profile and posts.</p>
        {relation === 'pending' ? (
          <p className="notice">Your follow request is waiting for an answer.</p>
        ) : (
          <button className="btn" onClick={follow}>Request to follow</button>
        )}
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
              <p className="meta">
                {user.nickname ? `@${user.nickname} · ` : ''}
                {user.private ? 'Private' : 'Public'} profile
              </p>
            </div>

            <div className="profile-buttons">
              {isMe ? (
                <Link href="/settings" className="btn btn-light">Edit settings</Link>
              ) : (
                <>
                  {relation === 'accepted' ? (
                    <button className="btn btn-light" onClick={unfollow}>Unfollow</button>
                  ) : relation === 'pending' ? (
                    <button className="btn btn-light" onClick={unfollow}>Cancel request</button>
                  ) : (
                    <button className="btn" onClick={follow}>Follow</button>
                  )}
                  <Link href={`/chat/${user.id}`} className="btn btn-light">Message</Link>
                </>
              )}
            </div>
          </div>

          {user.about_me && <p className="profile-about">{user.about_me}</p>}

          <div className="profile-stats">
            <span><strong>{posts.length}</strong> posts</span>
            <button type="button" className="stat-button" onClick={() => setModal('followers')}>
              <strong>{counts.followers}</strong> followers
            </button>
            <button type="button" className="stat-button" onClick={() => setModal('following')}>
              <strong>{counts.following}</strong> following
            </button>
            <span>Joined {new Date(user.created_at).toLocaleDateString(undefined, { month: 'long', year: 'numeric' })}</span>
          </div>

          {/* email and date of birth are only returned to the owner and to
              accepted followers, so they are simply shown when present */}
          {(user.email || user.date_of_birth) && (
            <dl className="details">
              {user.email && <div><dt>Email</dt><dd>{user.email}</dd></div>}
              {user.date_of_birth && <div><dt>Date of birth</dt><dd>{user.date_of_birth}</dd></div>}
            </dl>
          )}
        </div>
      </section>

      {modal && (
        <UserListModal
          title={modal === 'followers' ? 'Followers' : 'Following'}
          path={`/users/${id}/${modal}`}
          onClose={() => setModal(null)}
        />
      )}

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
