'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import { apiDelete, apiGet, apiPost, apiPut } from '@/lib/api'
import { parseId } from '@/lib/validate'
import Avatar from '@/components/Avatar'
import Icon from '@/components/Icon'
import NotFound from '@/components/NotFound'
import PersonRow from '@/components/PersonRow'
import PostCard from '@/components/PostCard'

export default function ProfilePage() {
  const { id } = useParams() // the [id] from the URL, e.g. /profile/3
  // 0 when the url is not a real id, like /profile/abc or /profile/0
  const userId = parseId(id)
  const [me, setMe] = useState(null)
  const [user, setUser] = useState(null)
  const [posts, setPosts] = useState([])
  const [isPrivate, setIsPrivate] = useState(false)
  const [followStatus, setFollowStatus] = useState('') // '' | 'pending' | 'accepted'
  const [followers, setFollowers] = useState([])
  const [following, setFollowing] = useState([])
  const [tab, setTab] = useState('posts') // 'posts' | 'followers' | 'following'
  const [message, setMessage] = useState('')
  const [notFound, setNotFound] = useState(false)
  const [savingPrivacy, setSavingPrivacy] = useState(false)

  async function load() {
    // there is no such user, so nothing is requested
    if (!userId) return
    setMe(await apiGet('/me'))
    apiGet(`/users/${userId}/follow`)
      .then(result => setFollowStatus(result.status))
      .catch(() => {})
    try {
      setUser(await apiGet(`/user/${userId}`))
      setIsPrivate(false)
      // no "posts of one user" endpoint yet, so we filter the feed
      const feed = await apiGet('/posts')
      setPosts(feed.filter(post => post.author_id === userId))
      // who follows them and who they follow — the API gates both exactly like
      // the profile, so a private one answers 403 and we never get here
      setFollowers(await apiGet(`/users/${userId}/followers`))
      setFollowing(await apiGet(`/users/${userId}/following`))
    } catch (err) {
      // 400 = the id is not one the API accepts, 404 = nobody under it
      if (err.status === 400 || err.status === 404) setNotFound(true)
      else if (err.status === 403) setIsPrivate(true) // private profile you don't follow
      else setMessage(err.message)
    }
  }

  useEffect(() => {
    load()
  }, [userId])

  async function follow() {
    try {
      const result = await apiPost(`/users/${userId}/follow`)
      setFollowStatus(result.status)
      setMessage(result.status === 'pending' ? 'Follow request sent.' : 'You are now following.')
      load() // a new follower may now see more posts
    } catch (err) {
      setMessage(err.message)
    }
  }

  // also used to cancel a pending request
  async function unfollow() {
    try {
      await apiDelete(`/users/${userId}/follow`)
      setMessage(followStatus === 'pending' ? 'Follow request cancelled.' : 'Unfollowed.')
      setFollowStatus('')
      load()
    } catch (err) {
      setMessage(err.message)
    }
  }

  // The switch on your own profile. The API answers with the saved user, so the
  // label always matches what is stored.
  async function togglePrivacy() {
    setSavingPrivacy(true)
    try {
      const updated = await apiPut('/me/privacy', { private: !user.private })
      setUser(updated)
      setMe(updated)
      setMessage(updated.private
        ? 'Your profile is private — only your followers can see it.'
        : 'Your profile is public — everyone can see it.')
      // going public accepts the waiting follow requests, so reload the lists
      if (!updated.private) load()
    } catch (err) {
      setMessage(err.message)
    }
    setSavingPrivacy(false)
  }

  // One button that changes with the relation: Follow → Requested / Unfollow
  const followButton =
    followStatus === 'accepted' ? (
      <button className="btn btn-light" onClick={unfollow}>Unfollow</button>
    ) : followStatus === 'pending' ? (
      <button className="btn btn-light" onClick={unfollow} title="Cancel the request">Requested</button>
    ) : (
      <button className="btn" onClick={follow}>{isPrivate ? 'Request to follow' : 'Follow'}</button>
    )

  if (!userId || notFound) {
    return (
      <NotFound
        title="Profile not found"
        text="Nobody is here under that id."
        back="/people"
        label="Back to people"
      />
    )
  }

  if (isPrivate) {
    return (
      <div className="card locked">
        <span className="locked-icon"><Icon name="lock" size={22} /></span>
        <h2>This profile is private</h2>
        <p className="subtitle">Send a follow request to see their profile and posts.</p>
        {followButton}
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
                <>
                  <button className="btn btn-light" onClick={togglePrivacy} disabled={savingPrivacy}>
                    <Icon name="lock" size={15} />
                    {savingPrivacy ? 'Saving…' : user.private ? 'Make public' : 'Make private'}
                  </button>
                  <Link href="/settings" className="btn btn-light">Edit settings</Link>
                </>
              ) : (
                <>
                  {followButton}
                  <Link href={`/chat/${user.id}`} className="btn btn-light">Message</Link>
                </>
              )}
            </div>
          </div>

          {user.about_me && <p className="profile-about">{user.about_me}</p>}

          <div className="profile-stats">
            <span><strong>{posts.length}</strong> posts</span>
            <span><strong>{followers.length}</strong> followers</span>
            <span><strong>{following.length}</strong> following</span>
            <span>Joined {new Date(user.created_at).toLocaleDateString(undefined, { month: 'long', year: 'numeric' })}</span>
            <span>{user.email}</span>
            {user.date_of_birth && <span>Born {new Date(user.date_of_birth + 'T00:00').toLocaleDateString(undefined, { day: 'numeric', month: 'long', year: 'numeric' })}</span>}
          </div>
        </div>
      </section>

      {message && <p className="notice">{message}</p>}

      <div className="profile-tabs">
        {[
          ['posts', 'Posts', posts.length],
          ['followers', 'Followers', followers.length],
          ['following', 'Following', following.length],
        ].map(([key, label, count]) => (
          <button
            key={key}
            className={tab === key ? 'profile-tab active' : 'profile-tab'}
            onClick={() => setTab(key)}
          >
            {label} <span>{count}</span>
          </button>
        ))}
      </div>

      {tab === 'posts' && (
        posts.length === 0 ? (
          <div className="empty">
            <p className="empty-title">No posts to show</p>
            <p>Posts you are allowed to see will appear here.</p>
          </div>
        ) : (
          posts.map(post => (
            <PostCard key={post.id} post={post} myId={me.id} onDeleted={load} />
          ))
        )
      )}

      {tab !== 'posts' && (() => {
        const people = tab === 'followers' ? followers : following
        if (people.length === 0) {
          return (
            <div className="empty">
              <p className="empty-title">
                {tab === 'followers' ? 'No followers yet' : 'Not following anyone yet'}
              </p>
              <p>
                {tab === 'followers'
                  ? 'People who follow this profile will appear here.'
                  : 'People this profile follows will appear here.'}
              </p>
            </div>
          )
        }
        return (
          <div className="card list">
            {people.map(person => (
              <PersonRow key={person.id} person={person} href={`/profile/${person.id}`}>
                <Icon name="arrow" size={16} />
              </PersonRow>
            ))}
          </div>
        )
      })()}
    </>
  )
}
