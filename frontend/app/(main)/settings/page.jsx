'use client'

import { useState } from 'react'
import { apiUpload } from '@/lib/api'
import { useMe } from '@/lib/useMe'
import { setMe } from '@/lib/userStore'
import { IMAGE_ACCEPT, checkImageFile } from '@/lib/validate'
import Avatar from '@/components/Avatar'
import Icon from '@/components/Icon'
import PageHeader from '@/components/PageHeader'

export default function SettingsPage() {
  const { me } = useMe()
  const [message, setMessage] = useState('')

  async function changeAvatar(e) {
    const file = e.target.files[0]
    if (!file) return

    // format, size and pixels are checked here so a bad photo never leaves the browser
    const problem = await checkImageFile(file)
    if (problem) {
      setMessage(problem)
      e.target.value = '' // let the user pick another one
      return
    }

    setMessage('')
    const formData = new FormData()
    formData.append('avatar', file)

    try {
      // the answer is the updated user, so the header updates with it
      setMe(await apiUpload('/avatar', formData))
      setMessage('Photo updated.')
    } catch (err) {
      setMessage(err.message)
    }
  }

  if (!me) return <p className="loading">Loading…</p>

  // label / value pairs shown in the account section
  const details = [
    ['Name', `${me.first_name} ${me.last_name}`],
    ['Nickname', `@${me.nickname}`],
    ['Email', me.email],
    ['Date of birth', me.date_of_birth],
    ['Profile', me.private ? 'Private' : 'Public'],
  ]

  return (
    <>
      <PageHeader title="Settings" subtitle="Your photo and account details." />

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
          {message && <span className="meta">{message}</span>}
        </div>
      </section>

      <section className="card settings-section">
        <div className="settings-label">
          <strong>Account</strong>
          <p>Editing these needs an API endpoint that does not exist yet.</p>
        </div>
        <dl className="details">
          {details.map(([label, value]) => (
            <div key={label}>
              <dt>{label}</dt>
              <dd>{value}</dd>
            </div>
          ))}
        </dl>
      </section>
    </>
  )
}
