'use client'

import { useState } from 'react'
import Link from 'next/link'
import { apiDelete, apiPost, imageUrl } from '@/lib/api'
import Avatar from './Avatar'
import CommentSection from './CommentSection'
import Icon from './Icon'

const privacyNames = { public: 'Public', almost_private: 'Followers', private: 'Chosen followers' }

// One post in a list. myId is the logged-in user's id, onDeleted() refreshes the list.
export default function PostCard({ post, myId, onDeleted }) {
  // likes/dislikes change when you react, so we keep them in state
  const [likes, setLikes] = useState(post.likes)
  const [dislikes, setDislikes] = useState(post.dislikes)
  const [myReaction, setMyReaction] = useState(post.my_reaction)
  const [commentCount, setCommentCount] = useState(post.comments ?? 0)

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

  return (
    <article className="card post">
      <header className="post-header">
        <Avatar user={author} size={42} />
        <div className="post-who">
          <Link href={`/profile/${post.author_id}`} className="post-author">
            {author.first_name} {author.last_name}
          </Link>
          <span className="meta">{date} · {privacyNames[post.privacy]}</span>
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
      </footer>

      <CommentSection
        postId={post.id}
        myId={myId}
        postAuthorId={post.author_id}
        count={commentCount}
        onCountChange={setCommentCount}
      />
    </article>
  )
}
