'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { apiDelete, apiGet, apiPost, apiPut, apiUpload, imageUrl } from '@/lib/api'
import { IMAGE_ACCEPT, LIMITS, checkImages, checkText } from '@/lib/validate'
import Avatar from './Avatar'
import CharCount from './CharCount'
import Icon from './Icon'
import Modal from './Modal'

const privacyNames = { public: 'Public', almost_private: 'Followers', private: 'Only me' }

// One post in a list. myId is the logged-in user's id; isGroupCreator marks
// the viewer as the group's creator (its admin), who may delete any post in
// the group. onDeleted(postId) runs after a successful delete — the group
// page uses it to drop the post from state, other pages refresh their list.
// Comments load from GET /posts/{id}/comments when expanded — the
// API only answers for viewers who may see the post (group posts included).
export default function PostCard({ post, myId, isGroupCreator = false, onDeleted }) {
  // likes/dislikes change when you react, so we keep them in state
  const [likes, setLikes] = useState(post.likes)
  const [dislikes, setDislikes] = useState(post.dislikes)
  const [myReaction, setMyReaction] = useState(post.my_reaction)
  const [commentCount, setCommentCount] = useState(post.comment_count ?? 0)
  // content/privacy are editable, so the card shows its own copy
  const [content, setContent] = useState(post.content)
  const [privacy, setPrivacy] = useState(post.privacy)
  const [editing, setEditing] = useState(false)
  const [editContent, setEditContent] = useState(post.content)
  const [editPrivacy, setEditPrivacy] = useState(post.privacy)
  const [editError, setEditError] = useState('')
  const [saving, setSaving] = useState(false)
  const [open, setOpen] = useState(false)
  const [comments, setComments] = useState(null) // null = not loaded yet
  const [draft, setDraft] = useState('')
  const [files, setFiles] = useState([])
  const [error, setError] = useState('')
  const [sending, setSending] = useState(false)
  // Delete flow: confirming = the dialog is open, deleting = the request is
  // in flight (the button stays disabled so the request cannot be doubled).
  const [confirming, setConfirming] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [deleteError, setDeleteError] = useState('')

  // The author can always delete their own post; the group's creator (its
  // admin) can delete any post in the group. The backend enforces the same
  // rule — this only decides whether the button is shown.
  const canDelete = post.author_id === myId || isGroupCreator

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

  // Reopening the editor always starts from what is on screen now
  function startEdit() {
    setEditContent(content)
    setEditPrivacy(privacy)
    setEditError('')
    setEditing(true)
  }

  async function saveEdit(e) {
    e.preventDefault()
    const problem = checkText('Your post', editContent, LIMITS.post)
    if (problem) {
      setEditError(problem)
      return
    }
    setEditError('')
    setSaving(true)
    try {
      // The API requires a valid privacy on every update. A group post keeps
      // the one it was stored with — the group alone decides who can see it.
      const updated = await apiPut(`/posts/${post.id}`, {
        content: editContent.trim(),
        privacy: post.group_id ? privacy : editPrivacy,
      })
      setContent(updated.content)
      setPrivacy(updated.privacy)
      setEditing(false)
    } catch (err) {
      setEditError(err.message)
    }
    setSaving(false)
  }

  // Opens the confirmation dialog; the actual delete happens in confirmDelete
  // once the user confirms there.
  function remove() {
    setDeleteError('')
    setConfirming(true)
  }

  async function confirmDelete() {
    if (deleting) return // one request at a time
    setDeleting(true)
    setDeleteError('')
    try {
      // Group posts go through the group-scoped route (the only one that
      // lets a group creator delete other people's posts); normal posts
      // keep the author-only /posts/{id} route.
      const path = post.group_id
        ? `/groups/${post.group_id}/posts/${post.id}`
        : `/posts/${post.id}`
      await apiDelete(path)
      setConfirming(false)
      onDeleted?.(post.id)
    } catch (err) {
      setDeleteError(err.message)
    }
    setDeleting(false)
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
          <span className="meta">{date}{!post.group_id && <> · {privacyNames[privacy]}</>}</span>
        </div>
        {canDelete && !editing && (
          <>
            {post.author_id === myId && (
              <button className="icon-button" onClick={startEdit} title="Edit post">
                <Icon name="edit" size={16} />
              </button>
            )}
            <button className="icon-button" onClick={remove} title="Delete post">
              <Icon name="trash" size={16} />
            </button>
          </>
        )}
      </header>

      {editing ? (
        <form className="post-edit" onSubmit={saveEdit} noValidate>
          <textarea
            value={editContent}
            maxLength={LIMITS.post}
            onChange={e => setEditContent(e.target.value)}
            autoFocus
          />
          <div className="post-edit-bar">
            {!post.group_id && (
              <select className="tool" value={editPrivacy} onChange={e => setEditPrivacy(e.target.value)}>
                <option value="public">Public</option>
                <option value="almost_private">Followers</option>
                <option value="private">Only me</option>
              </select>
            )}
            <CharCount value={editContent} max={LIMITS.post} />
            <button type="button" className="btn btn-sm btn-light" onClick={() => setEditing(false)}>
              Cancel
            </button>
            <button className="btn btn-sm" disabled={saving || !editContent.trim()}>
              {saving ? 'Saving…' : 'Save'}
            </button>
          </div>
          {editError && <p className="error">{editError}</p>}
        </form>
      ) : (
        <p className="post-content">{content}</p>
      )}

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

      {confirming && (
        <Modal title="Delete post" onClose={() => setConfirming(false)}>
          <p>
            Delete this post by {author.first_name} {author.last_name}?
            {content.length > 0 &&
              (content.length > 120 ? ` “${content.slice(0, 120)}…”` : ` “${content}”`)}
          </p>
          <p className="meta">This cannot be undone. Comments on it are removed too.</p>
          {deleteError && <p className="error">{deleteError}</p>}
          <div className="composer-bar">
            <button
              type="button"
              className="btn btn-light"
              onClick={() => setConfirming(false)}
              disabled={deleting}
            >
              Cancel
            </button>
            <button type="button" className="btn btn-danger" onClick={confirmDelete} disabled={deleting}>
              {deleting ? 'Deleting…' : 'Delete post'}
            </button>
          </div>
        </Modal>
      )}
    </article>
  )
}
