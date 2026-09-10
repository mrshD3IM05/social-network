'use client'

import { useState } from 'react'
import { apiPost, apiUpload } from '@/lib/api'
import Icon from './Icon'

// Form to write a new post. onPosted() is called after it is saved.
export default function PostForm({ onPosted }) {
  const [content, setContent] = useState('')
  const [privacy, setPrivacy] = useState('public')
  const [files, setFiles] = useState([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      // 1. create the post
      const post = await apiPost('/posts', { content, privacy })

      // 2. upload the images and attach them to the post
      if (files.length > 0) {
        const formData = new FormData()
        for (const file of files) formData.append('files', file)
        formData.append('post_id', post.id)
        await apiUpload('/files', formData)
      }

      // 3. reset the form
      setContent('')
      setFiles([])
      onPosted()
    } catch (err) {
      setError(err.message)
    }

    setLoading(false)
  }

  return (
    <form className="card composer" onSubmit={handleSubmit}>
      <textarea
        placeholder="Share something with your followers…"
        value={content}
        onChange={e => setContent(e.target.value)}
        required
      />

      <div className="composer-bar">
        {/* the real file input is hidden, the label acts as the button */}
        <label className="tool">
          <Icon name="image" />
          {files.length > 0 ? `${files.length} photo${files.length > 1 ? 's' : ''}` : 'Photo'}
          <input
            type="file"
            accept="image/jpeg,image/png,image/gif"
            multiple
            hidden
            onChange={e => setFiles(Array.from(e.target.files))}
          />
        </label>

        <select className="tool" value={privacy} onChange={e => setPrivacy(e.target.value)}>
          <option value="public">Public</option>
          <option value="almost_private">Followers</option>
          <option value="private">Only me</option>
        </select>

        <button className="btn" disabled={loading}>
          {loading ? 'Publishing…' : 'Publish'}
        </button>
      </div>

      {error && <p className="error">{error}</p>}
    </form>
  )
}
