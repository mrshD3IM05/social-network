'use client'

import { useState } from 'react'
import { apiPost, apiUpload } from '@/lib/api'
import { IMAGE_ACCEPT, LIMITS, checkImages, checkText } from '@/lib/validate'
import CharCount from './CharCount'
import Icon from './Icon'

// Form to write a new post. onPosted() is called after it is saved.
// Inside a group, `groupId` is set: the post goes to the group (members only)
// and the privacy selector disappears — group posts are member-only by design.
export default function PostForm({ onPosted, groupId }) {
  const [content, setContent] = useState('')
  const [privacy, setPrivacy] = useState('public')
  const [files, setFiles] = useState([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  // Checked while typing so the Publish button knows if the post is valid
  const contentError = checkText('Your post', content, LIMITS.post)

  // Keep the picked images only if there are at most 3 valid ones
  function pickFiles(e) {
    const picked = Array.from(e.target.files)
    const imageError = checkImages(picked)

    setError(imageError)
    setFiles(imageError ? [] : picked)
    if (imageError) e.target.value = '' // let the user pick again
  }

  function clearFiles() {
    setFiles([])
    setError('')
  }

  async function handleSubmit(e) {
    e.preventDefault()

    // check everything once more before calling the API
    const problem = contentError || checkImages(files)
    if (problem) {
      setError(problem)
      return
    }

    setError('')
    setLoading(true)

    try {
      // 1. create the post (in the group when we are inside one)
      const post = groupId
        ? await apiPost(`/groups/${groupId}/posts`, { content: content.trim(), privacy: 'public' })
        : await apiPost('/posts', { content: content.trim(), privacy })

      // 2. upload the images and attach them to the post
      if (files.length > 0) {
        const formData = new FormData()
        for (const file of files) formData.append('files', file)
        formData.append('post_id', post.id)
        await apiUpload('/files', formData)
        onPosted()
      } else {
        onPosted(post)
      }

      // 3. reset the form
      setContent('')
      setFiles([])
    } catch (err) {
      setError(err.message)
    }

    setLoading(false)
  }

  return (
    <form className="card composer" onSubmit={handleSubmit} noValidate>
      <textarea
        placeholder={groupId ? 'Share something with the group…' : 'Share something with your followers…'}
        value={content}
        maxLength={LIMITS.post}
        onChange={e => setContent(e.target.value)}
      />

      <div className="composer-bar">
        {/* the real file input is hidden, the label acts as the button */}
        <label className="tool">
          <Icon name="image" />
          {files.length > 0 ? `${files.length} photo${files.length > 1 ? 's' : ''}` : 'Photo'}
          <input
            type="file"
            accept={IMAGE_ACCEPT}
            multiple
            hidden
            onChange={pickFiles}
          />
        </label>

        {files.length > 0 && (
          <button type="button" className="tool" onClick={clearFiles}>Remove</button>
        )}

        {!groupId && (
          <select className="tool" value={privacy} onChange={e => setPrivacy(e.target.value)}>
            <option value="public">Public</option>
            <option value="almost_private">Followers</option>
            <option value="private">Only me</option>
          </select>
        )}

        <CharCount value={content} max={LIMITS.post} />

        <button className="btn" disabled={loading || Boolean(contentError)}>
          {loading ? 'Publishing…' : 'Publish'}
        </button>
      </div>

      <p className="hint">Up to {LIMITS.images} images, JPEG, PNG or GIF, 10 MB each.</p>

      {error && <p className="error">{error}</p>}
    </form>
  )
}
