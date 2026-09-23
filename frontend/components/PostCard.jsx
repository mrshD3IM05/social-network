'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { apiDelete, apiGet, apiPost, apiUpload, imageUrl } from '@/lib/api'
import { IMAGE_ACCEPT, LIMITS, checkImages, checkText } from '@/lib/validate'
import Avatar from './Avatar'
import CharCount from './CharCount'
import Icon from './Icon'

const privacyNames = { public: 'Public', almost_private: 'Followers', private: 'Only me' }

// One post in a list. myId is the logged-in user's id, onDeleted() refreshes
// the list. Comments load from GET /posts/{id}/comments when expanded — the
// API only answers for viewers who may see the post (group posts included).
export default function PostCard({ post, myId, onDeleted }) {
  // likes/dislikes change when you react, so we keep them in state
  const [likes, setLikes] = useState(post.likes)
  const [dislikes, setDislikes] = useState(post.dislikes)
  const [myReaction, setMyReaction] = useState(post.my_reaction)
  const [commentCount, setCommentCount] = useState(post.comment_count ?? 0)
  const [open, setOpen] = useState(false)
  const [comments, setComments] = useState(null) // null = not loaded yet
  const [draft, setDraft] = useState('')
  const [files, setFiles] = useState([])
  const [error, setError] = useState('')
  const [sending, setSending] = useState(false)

  const author = {
    first_name: post.author_first_name,
    last_name: post.author_last_name,
    avatar: post.author_avatar,
  }

  const date = new Date(post.created_at).toLocaleDateString(undefined, {
    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
  })

  // Sending the same reaction again removes it. The API answers with the new counts.
  async function react(reaction) {
    const result = await apiPost(`/posts/${post.id}/reactions`, { reaction })
    setLikes(result.likes)
    setDislikes(result.dislikes)
    setMyReaction(result.my_reaction)
  }

  async function remove() {
    if (!confirm('Delete this post?')) return
    await apiDelete(`/posts/${post.id}`)
    onDeleted()
  }

  // First open loads the comments once; afterwards they are kept in state.
  function toggle() {
    const next = !open
    setOpen(next)
    if (next && comments === null) {
      apiGet(`/posts/${post.id}/comments`)
        .then(setComments)
        .catch(err => {
          setError(err.message)
          setComments([])
        })
    }
  }

  async function submitComment(e) {
    e.preventDefault()
    const problem = checkText('Comment', draft, LIMITS.comment) || checkImages(files)
    if (problem) {
      setError(problem)
      return
    }
    setError('')
    setSending(true)
    try {
      const comment = await apiPost(`/posts/${post.id}/comments`, { content: draft.trim() })
      if (files.length > 0) {
        const formData = new FormData()
        for (const file of files) formData.append('files', file)
        formData.append('comment_id', comment.id)
        await apiUpload('/files', formData)
        // the uploaded image is attached after the comment was created —
        // reload the list so it shows up
        setComments(await apiGet(`/posts/${post.id}/comments`))
      } else {
        setComments(list => [...(list || []), comment])
      }
      setDraft('')
      setFiles([])
      setCommentCount(count => count + 1)
    } catch (err) {
      setError(err.message)
    }
    setSending(false)
  }

  function pickFiles(e) {
    const picked = Array.from(e.target.files)
    const imageError = checkImages(picked)
    setError(imageError)
    setFiles(imageError ? [] : picked)
    if (imageError) e.target.value = ''
  }

  return (
    <article className="card post">
      <header className="post-header">
        <Avatar user={author} size={42} />
        <div className="post-who">
          <Link href={`/profile/${post.author_id}`} className="post-author">
            {author.first_name} {author.last_name}
          </Link>
          <span className="meta">{date}{!post.group_id && <> · {privacyNames[post.privacy]}</>}</span>
        </div>
        {post.author_id === myId && (
          <button className="icon-button" onClick={remove} title="Delete post">
            <Icon name="trash" size={16} />
          </button>
        )}
      </header>

      <p className="post-content">{post.content}</p>

      {post.images?.length > 0 && (
        <div className="post-images">
          {post.images.map(id => <img key={id} src={imageUrl(id)} alt="" />)}
        </div>
      )}

      <footer className="post-actions">
        <button className={myReaction === 'like' ? 'reaction active' : 'reaction'} onClick={() => react('like')}>
          <Icon name="like" size={16} /> {likes}
        </button>
        <button className={myReaction === 'dislike' ? 'reaction active' : 'reaction'} onClick={() => react('dislike')}>
          <Icon name="dislike" size={16} /> {dislikes}
        </button>
        <button className={open ? 'reaction active' : 'reaction'} onClick={toggle}>
          <Icon name="chat" size={16} /> {commentCount}
        </button>
      </footer>

      {open && (
        <div className="comments">
          {comments === null && <p className="meta">Loading comments…</p>}

          {comments !== null && comments.length === 0 && (
            <p className="meta comments-empty">No comments yet.</p>
          )}

          {(comments || []).map(comment => (
            <div key={comment.id} className="comment">
              <Avatar user={{
                first_name: comment.author_first_name,
                last_name: comment.author_last_name,
                avatar: comment.author_avatar,
              }} size={30} />
              <div className="comment-body">
                <div className="comment-head">
                  <Link href={`/profile/${comment.author_id}`} className="comment-author">
                    {comment.author_first_name} {comment.author_last_name}
                  </Link>
                  <span className="meta">
                    {new Date(comment.created_at).toLocaleDateString(undefined, {
                      month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
                    })}
                  </span>
                </div>
                <p className="comment-content">{comment.content}</p>
                {comment.images?.length > 0 && (
                  <div className="comment-images">
                    {comment.images.map(id => <img key={id} src={imageUrl(id)} alt="" />)}
                  </div>
                )}
              </div>
            </div>
          ))}

          <form className="comment-form" onSubmit={submitComment} noValidate>
            <textarea
              placeholder="Write a comment…"
              value={draft}
              maxLength={LIMITS.comment}
              onChange={e => setDraft(e.target.value)}
            />
            <div className="comment-bar">
              <label className="tool">
                <Icon name="image" size={14} />
                {files.length > 0 ? `${files.length}` : 'Photo'}
                <input type="file" accept={IMAGE_ACCEPT} hidden onChange={pickFiles} />
              </label>
              {files.length > 0 && (
                <button type="button" className="tool" onClick={() => setFiles([])}>Remove</button>
              )}
              <CharCount value={draft} max={LIMITS.comment} />
              <button className="btn btn-sm" disabled={sending || !draft.trim()}>
                {sending ? '…' : 'Reply'}
              </button>
            </div>
          </form>

          {error && <p className="error">{error}</p>}
        </div>
      )}
    </article>
  )
}
