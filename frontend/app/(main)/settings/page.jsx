'use client'

import { useEffect, useState } from 'react'
import { apiGet, apiPatch, apiUpload } from '@/lib/api'
import { IMAGE_ACCEPT, LIMITS, checkImage, checkNickname, checkText } from '@/lib/validate'
import Avatar from '@/components/Avatar'
import CharCount from '@/components/CharCount'
import Icon from '@/components/Icon'
import PageHeader from '@/components/PageHeader'

export default function SettingsPage() {
  const [me, setMe] = useState(null)
  const [draft, setDraft] = useState(null)
  const [message, setMessage] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    apiGet('/me').then(user => {
      setMe(user)
      setDraft({
        first_name: user.first_name,
        last_name: user.last_name,
        nickname: user.nickname,
        about_me: user.about_me,
      })
    })
  }, [])

  async function changeAvatar(e) {
    const file = e.target.files[0]
    if (!file) return

    // format and size are checked here so a bad photo never leaves the browser
    const problem = checkImage(file)
    if (problem) {
      setMessage(problem)
      e.target.value = '' // let the user pick another one
      return
    }

    setMessage('')
    const formData = new FormData()
    formData.append('avatar', file)

    try {
      const user = await apiUpload('/avatar', formData) // answers with the updated user
      setMe(user)
      setMessage('Photo updated.')
    } catch (err) {
      setMessage(err.message)
    }
  }

  // The subject asks for a switch between a public and a private profile.
  async function togglePrivacy() {
    setMessage('')
    try {
      const user = await apiPatch('/me', { private: String(!me.private) })
      setMe(user)
      setMessage(user.private ? 'Your profile is now private.' : 'Your profile is now public.')
    } catch (err) {
      setMessage(err.message)
    }
  }

  async function saveProfile(e) {
    e.preventDefault()

    const problem =
      checkText('First name', draft.first_name, LIMITS.firstName) ||
      checkText('Last name', draft.last_name, LIMITS.lastName) ||
      (draft.nickname ? checkNickname(draft.nickname) : '') ||
      checkText('About me', draft.about_me, LIMITS.aboutMe, { required: false })

    if (problem) {
      setMessage(problem)
      return
    }

    setMessage('')
    setSaving(true)
    try {
      const user = await apiPatch('/me', {
        first_name: draft.first_name.trim(),
        last_name: draft.last_name.trim(),
        nickname: draft.nickname.trim().toLowerCase(),
        about_me: draft.about_me.trim(),
      })
      setMe(user)
      setMessage('Profile saved.')
    } catch (err) {
      setMessage(err.message)
    }
    setSaving(false)
  }

  if (!me || !draft) return <p className="loading">Loading…</p>

  function field(name) {
    return {
      value: draft[name] ?? '',
      onChange: e => setDraft(values => ({ ...values, [name]: e.target.value })),
    }
  }

  return (
    <>
      <PageHeader label="Account" title="Settings" subtitle="Your photo, your details and who can see them." />

      {message && <p className="notice">{message}</p>}

      <section className="card settings-section">
        <div className="settings-label">
          <strong>Photo</strong>
          <p>JPEG, PNG or GIF, up to 10 MB.</p>
        </div>
        <div className="photo-row">
          <Avatar user={me} size={72} />
          <label className="btn btn-light">
            <Icon name="camera" size={16} /> Change photo
            <input type="file" accept={IMAGE_ACCEPT} hidden onChange={changeAvatar} />
          </label>
        </div>
      </section>

      <section className="card settings-section">
        <div className="settings-label">
          <strong>Profile visibility</strong>
          <p>
            A public profile can be seen by anyone. A private profile is only visible to the
            followers you accepted.
          </p>
        </div>
        <div className="photo-row">
          <button className="btn" onClick={togglePrivacy}>
            <Icon name="lock" size={16} />
            {me.private ? ' Make profile public' : ' Make profile private'}
          </button>
          <span className="meta">Currently {me.private ? 'private' : 'public'}.</span>
        </div>
      </section>

      <form className="card settings-section" onSubmit={saveProfile} noValidate>
        <div className="settings-label">
          <strong>Account</strong>
          <p>Your email ({me.email}) and date of birth ({me.date_of_birth}) cannot be changed.</p>
        </div>

        <div className="row">
          <div>
            <label>First name</label>
            <input maxLength={LIMITS.firstName} {...field('first_name')} />
          </div>
          <div>
            <label>Last name</label>
            <input maxLength={LIMITS.lastName} {...field('last_name')} />
          </div>
        </div>

        <label>Nickname <small>optional</small></label>
        <input
          maxLength={LIMITS.nickname.max}
          placeholder={`${LIMITS.nickname.min}–${LIMITS.nickname.max} letters or numbers`}
          {...field('nickname')}
        />

        <label>About me <small>optional</small></label>
        <textarea rows={3} maxLength={LIMITS.aboutMe} {...field('about_me')} />
        <CharCount value={draft.about_me ?? ''} max={LIMITS.aboutMe} />

        <button className="btn" disabled={saving}>{saving ? 'Saving…' : 'Save changes'}</button>
      </form>
    </>
  )
}
