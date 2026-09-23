'use client'

import { useState } from 'react'
import { apiPost } from '@/lib/api'
import { LIMITS, checkText } from '@/lib/validate'
import Modal from './Modal'
import CharCount from './CharCount'

// Create-event dialog for a group. The backend answers 201 with the new
// event; onCreated() adds it to the list without a page reload.
export default function EventFormModal({ groupId, onClose, onCreated }) {
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [date, setDate] = useState('')
  const [time, setTime] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const titleError = checkText('Title', title, LIMITS.groupTitle)
  const descriptionError = checkText('Description', description, LIMITS.groupDescription, { required: false })

  // The API rejects past dates (service-side), so mirror that: allow today
  // but not yesterday. Time must be picked whenever a date is picked.
  function dateError() {
    if (!date) return 'Date is required.'
    const picked = new Date(`${date}T23:59:59`)
    if (Number.isNaN(picked.getTime())) return 'Enter a valid date.'
    if (picked < new Date().setHours(0, 0, 0, 0)) return 'Date cannot be in the past.'
    return ''
  }

  function timeError() {
    if (!time) return 'Time is required.'
    return ''
  }

  async function handleSubmit(e) {
    e.preventDefault()
    const problem = titleError || descriptionError || dateError() || timeError()
    if (problem) {
      setError(problem)
      return
    }
    setError('')
    setLoading(true)
    try {
      const event = await apiPost(`/groups/${groupId}/events`, {
        title: title.trim(),
        description: description.trim(),
        date,
        time,
      })
      onCreated(event)
    } catch (err) {
      setError(err.message)
      setLoading(false)
    }
  }

  return (
    <Modal title="Create an event" onClose={onClose}>
      <form onSubmit={handleSubmit} noValidate>
        <label>Title</label>
        <input
          value={title}
          maxLength={LIMITS.groupTitle}
          className={error && titleError ? 'invalid' : undefined}
          onChange={e => setTitle(e.target.value)}
          autoFocus
        />

        <label>Description <small>optional</small></label>
        <textarea
          rows={3}
          value={description}
          maxLength={LIMITS.groupDescription}
          className={error && descriptionError ? 'invalid' : undefined}
          onChange={e => setDescription(e.target.value)}
        />

        <div className="row">
          <div>
            <label>Date</label>
            <input type="date" value={date} min={new Date().toISOString().slice(0, 10)} onChange={e => setDate(e.target.value)} />
          </div>
          <div>
            <label>Time</label>
            <input type="time" value={time} onChange={e => setTime(e.target.value)} />
          </div>
        </div>

        <div className="composer-bar">
          <CharCount value={description} max={LIMITS.groupDescription} />
          <button className="btn" disabled={loading || Boolean(titleError) || Boolean(descriptionError)}>
            {loading ? 'Creating…' : 'Create event'}
          </button>
        </div>

        {error && <p className="error">{error}</p>}
      </form>
    </Modal>
  )
}
