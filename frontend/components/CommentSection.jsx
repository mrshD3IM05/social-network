'use client'

import { useState } from 'react'
import Link from 'next/link'
import { apiDelete, apiGet, apiPost, apiUpload, imageUrl } from '@/lib/api'
import { IMAGE_ACCEPT, LIMITS, checkImages, checkText } from '@/lib/validate'
import Avatar from './Avatar'
import CharCount from './CharCount'
import Icon from './Icon'

// Comments of one post: the list, and the box to write a new one (with images
// or GIFs, like the subject asks). Comments are only loaded when opened.
export default function CommentSection({ postId, myId, postAuthorId, count, onCountChange }) {
  const [open, setOpen] = useState(false)
  const [comments, setComments] = useState(null)
  const [content, setContent] = useState('')
  const [files, setFiles] = useState([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const contentError = checkText('Your comment', content, LIMITS.post)

  async function load() {
    try {
      const list = await apiGet(`/posts/${postId}/comments?limit=100`)
      setComments(list)
      onCountChange?.(list.length)
    } catch (err) {
      setError(err.message)
      setComments([])
    }
  }

  function toggle() {
    const next = !open
    setOpen(next)
    if (next && comments === null) load()
  }

  function pickFiles(e) {
    const picked = Array.from(e.target.files)
    const imageError = checkImages(picked)
    setError(imageError)
    setFiles(imageError ? [] : picked)
    if (imageError) e.target.value = ''
  }

  async function submit(e) {
    e.preventDefault()
    const problem = contentError || checkImages(files)
    if (problem) {
      setError(problem)
      return
    }

    setError('')
    setLoading(true)
    try {
      // one multipart request carries the text and the images together
      const body = new FormData()
      body.append('content', content.trim())
      for (const file of files) body.append('files', file)

      await apiUpload(`/posts/${postId}/comments`, body)

      setContent('')
      setFiles([])
      await load()
    } catch (err) {
      setError(err.message)
    }
    setLoading(false)
  }

  async function remove(commentId) {
    if (!confirm('Delete this comment?')) return
    try {
      await apiDelete(`/comments/${commentId}`)
      await load()
    } catch (err) {
      setError(err.message)
    }
  }

  async function react(commentId, reaction) {
    try {
      const summary = await apiPost(`/comments/${commentId}/reactions`, { reaction })
      setComments(list =>
        list.map(comment =>
          comment.id === commentId
            ? { ...comment, likes: summary.likes, dislikes: summary.dislikes, my_reaction: summary.my_reaction }
            : comment,
        ),
      )
    } catch (err) {
      setError(err.message)
    }
  }

  const shown = comments ?? []

  return (
    <div className="comments">
      <button type="button" className="reaction" onClick={toggle} aria-expanded={open}>
        <Icon name="chat" size={16} /> {count > 0 ? `${count} comment${count > 1 ? 's' : ''}` : 'Comment'}
      </button>

      {open && (
        <>
          {comments === null && <p className="meta">Loading comments…</p>}

          {shown.map(comment => (
            <article key={comment.id} className="comment">
              <Avatar
                user={{
                  first_name: comment.author_first_name,
                  last_name: comment.author_last_name,
                  avatar: comment.author_avatar,
                }}
                size={32}
              />
              <div className="comment-body">
                <p className="comment-head">
                  <Link href={`/profile/${comment.author_id}`} className="post-author">
                    {comment.author_first_name} {comment.author_last_name}
                  </Link>
                  <span className="meta">{new Date(comment.created_at).toLocaleString()}</span>
                  {(comment.author_id === myId || postAuthorId === myId) && (
                    <button className="icon-button" onClick={() => remove(comment.id)} title="Delete comment">
                      <Icon name="trash" size={14} />
                    </button>
                  )}
                </p>

                <p className="comment-text">{comment.content}</p>

                {comment.images?.length > 0 && (
                  <div className="post-images">
                    {comment.images.map(id => <img key={id} src={imageUrl(id)} alt="" />)}
                  </div>
                )}

                <div className="comment-actions">
                  <button
                    className={comment.my_reaction === 'like' ? 'reaction active' : 'reaction'}
                    onClick={() => react(comment.id, 'like')}
                  >
                    <Icon name="like" size={14} /> {comment.likes}
                  </button>
                  <button
                    className={comment.my_reaction === 'dislike' ? 'reaction active' : 'reaction'}
                    onClick={() => react(comment.id, 'dislike')}
                  >
                    <Icon name="dislike" size={14} /> {comment.dislikes}
                  </button>
                </div>
              </div>
            </article>
          ))}

          {comments !== null && shown.length === 0 && <p className="meta">No comments yet.</p>}

          <form className="comment-form" onSubmit={submit} noValidate>
            <textarea
              rows={2}
              placeholder="Write a comment…"
              value={content}
              maxLength={LIMITS.post}
              onChange={e => setContent(e.target.value)}
            />
            <div className="composer-bar">
              <label className="tool">
                <Icon name="image" />
                {files.length > 0 ? `${files.length} photo${files.length > 1 ? 's' : ''}` : 'Photo'}
                <input type="file" accept={IMAGE_ACCEPT} multiple hidden onChange={pickFiles} />
              </label>
              {files.length > 0 && (
                <button type="button" className="tool" onClick={() => setFiles([])}>Remove</button>
              )}
              <CharCount value={content} max={LIMITS.post} />
              <button className="btn" disabled={loading || Boolean(contentError)}>
                {loading ? 'Sending…' : 'Comment'}
              </button>
            </div>
          </form>
        </>
      )}

      {error && <p className="error">{error}</p>}
    </div>
  )
}
