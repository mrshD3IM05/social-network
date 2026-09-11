'use client'

import { useState, useEffect } from 'react'
import { apiCall, postForm } from '../../lib/api'
import Link from 'next/link'

const API = '/api/v1'

function fileURL(id) {
  return id ? `${API}/fs/${id}` : ''
}

export default function Home() {
  const [me, setMe] = useState(null)
  const [posts, setPosts] = useState([])
  const [content, setContent] = useState('')
  const [privacy, setPrivacy] = useState('public')
  const [error, setError] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    apiCall('/me').then(user => {
      setMe(user)
      return apiCall('/posts')
    }).then(data => {
      setPosts(data)
    }).catch(err => {
      window.location.href = '/login'
    }).finally(() => setLoading(false))
  }, [])

  async function handleCreate(e) {
    e.preventDefault()
    setError(null)
    if (!content.trim()) return
    try {
      const post = await postForm('/posts', { content: content.trim(), privacy })
      post.author_first_name = me.first_name
      post.author_last_name = me.last_name
      post.author_nickname = me.nickname
      post.author_avatar = me.avatar
      post.likes = 0
      post.dislikes = 0
      post.my_reaction = ''
      post.images = []
      setPosts(prev => [post, ...prev])
      setContent('')
    } catch (err) {
      setError(err.message)
    }
  }

  async function handleReact(postID, reaction) {
    try {
      const summary = await postForm(`/posts/${postID}/reactions`, { reaction })
      setPosts(prev => prev.map(p => {
        if (String(p.id) !== String(postID)) return p
        return { ...p, likes: summary.likes, dislikes: summary.dislikes, my_reaction: summary.my_reaction }
      }))
    } catch (err) {
      setError(err.message)
    }
  }

  async function handleDelete(postID) {
    try {
      await apiCall(`/posts/${postID}`, { method: 'DELETE' })
      setPosts(prev => prev.filter(p => String(p.id) !== String(postID)))
    } catch (err) {
      setError(err.message)
    }
  }

  async function handleLogout() {
    try {
      await postForm('/logout', {})
      window.location.href = '/login'
    } catch (err) {
      window.location.href = '/login'
    }
  }

  if (loading) return <div style={styles.container}><p>Loading...</p></div>

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <h1>Feed</h1>
        {me && <div style={styles.headerRight}>
          <span style={styles.welcome}>Hello, {me.nickname || me.first_name}</span>
          <button onClick={handleLogout} style={styles.logoutBtn}>Logout</button>
        </div>}
      </div>

      {error && <p style={styles.error}>{error}</p>}

      <form onSubmit={handleCreate} style={styles.compose}>
        <textarea
          value={content}
          onChange={e => setContent(e.target.value)}
          placeholder="What's on your mind?"
          rows={3}
          style={styles.textarea}
        />
        <div style={styles.composeFooter}>
          <select value={privacy} onChange={e => setPrivacy(e.target.value)} style={styles.select}>
            <option value="public">Public</option>
            <option value="almost_private">Followers</option>
            <option value="private">Selected</option>
          </select>
          <button type="submit" style={styles.postBtn} disabled={!content.trim()}>Post</button>
        </div>
      </form>

      {posts.length === 0 && <p style={styles.empty}>No posts yet.</p>}

      {posts.map(post => (
        <article key={post.id} style={styles.post}>
          <div style={styles.postHeader}>
            {post.author_avatar ? <img src={fileURL(post.author_avatar)} alt="" style={styles.avatar} /> : <div style={styles.avatarPlaceholder} />}
            <div>
              <strong>{post.author_nickname || `${post.author_first_name} ${post.author_last_name}`}</strong>
              <div style={styles.meta}>{post.privacy} · {new Date(post.created_at).toLocaleString()}</div>
            </div>
            {me && post.author_id === me.id && (
              <button onClick={() => handleDelete(post.id)} style={styles.deleteBtn}>Delete</button>
            )}
          </div>
          <p style={styles.postContent}>{post.content}</p>
          {post.images?.length > 0 && (
            <div style={styles.images}>
              {post.images.map(id => <img key={id} src={fileURL(id)} alt="" style={styles.postImage} />)}
            </div>
          )}
          <div style={styles.reactions}>
            <button
              onClick={() => handleReact(post.id, 'like')}
              style={{ ...styles.reactionBtn, ...(post.my_reaction === 'like' ? styles.reactionActive : {}) }}
            >👍 {post.likes}</button>
            <button
              onClick={() => handleReact(post.id, 'dislike')}
              style={{ ...styles.reactionBtn, ...(post.my_reaction === 'dislike' ? styles.reactionActive : {}) }}
            >👎 {post.dislikes}</button>
          </div>
        </article>
      ))}
    </div>
  )
}

const styles = {
  container: { maxWidth: 640, margin: '0 auto', padding: 20, fontFamily: 'system-ui, sans-serif' },
  header: { display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 },
  headerRight: { display: 'flex', alignItems: 'center', gap: 12 },
  welcome: { color: '#666', fontSize: 14 },
  logoutBtn: { padding: '6px 12px', border: '1px solid #ddd', background: 'transparent', borderRadius: 4, cursor: 'pointer' },
  error: { color: 'red', padding: 8, background: '#fff0f0', borderRadius: 4 },
  compose: { border: '1px solid #ddd', borderRadius: 8, padding: 16, marginBottom: 24 },
  textarea: { width: '100%', border: 'none', outline: 'none', resize: 'vertical', fontSize: 14, fontFamily: 'inherit' },
  composeFooter: { display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 8 },
  select: { padding: '6px 8px', border: '1px solid #ddd', borderRadius: 4, fontSize: 13 },
  postBtn: { padding: '6px 16px', background: '#333', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer', fontSize: 13 },
  empty: { color: '#999', textAlign: 'center', padding: 32 },
  post: { border: '1px solid #eee', borderRadius: 8, padding: 16, marginBottom: 12 },
  postHeader: { display: 'flex', alignItems: 'center', gap: 10, marginBottom: 10 },
  avatar: { width: 36, height: 36, borderRadius: '50%', objectFit: 'cover' },
  avatarPlaceholder: { width: 36, height: 36, borderRadius: '50%', background: '#e0e0e0' },
  meta: { fontSize: 12, color: '#888' },
  deleteBtn: { marginLeft: 'auto', padding: '4px 10px', border: '1px solid #ddd', background: 'transparent', borderRadius: 4, fontSize: 12, cursor: 'pointer', color: '#c00' },
  postContent: { margin: '0 0 10px', lineHeight: 1.5 },
  images: { display: 'flex', gap: 8, marginBottom: 10, flexWrap: 'wrap' },
  postImage: { width: 200, height: 150, objectFit: 'cover', borderRadius: 4, border: '1px solid #eee' },
  reactions: { display: 'flex', gap: 8 },
  reactionBtn: { padding: '4px 12px', border: '1px solid #ddd', background: 'transparent', borderRadius: 4, cursor: 'pointer', fontSize: 13 },
  reactionActive: { borderColor: '#4a7c59', background: '#e8f5e9' },
}