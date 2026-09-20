'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { apiPost, apiUpload } from '@/lib/api'
import {
  IMAGE_ACCEPT,
  LIMITS,
  checkDateOfBirth,
  checkEmail,
  checkImage,
  checkNickname,
  checkPassword,
  checkText,
  maxBirthDate,
} from '@/lib/validate'

export default function RegisterPage() {
  const router = useRouter()
  const [errors, setErrors] = useState({}) // one message per field
  const [error, setError] = useState('')
  const [avatar, setAvatar] = useState(null)
  const [loading, setLoading] = useState(false)

  // The avatar is optional. Registering also logs you in, so the picture is
  // uploaded right after the account is created.
  function pickAvatar(e) {
    const file = e.target.files[0]
    if (!file) {
      setAvatar(null)
      return
    }
    const problem = checkImage(file)
    setErrors(rest => ({ ...rest, avatar: problem }))
    setAvatar(problem ? null : file)
    if (problem) e.target.value = ''
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')

    // read every input of the form by its "name"
    const form = new FormData(e.target)
    const values = {
      first_name: form.get('first_name').trim(),
      last_name: form.get('last_name').trim(),
      email: form.get('email').trim().toLowerCase(),
      nickname: form.get('nickname').trim().toLowerCase(),
      password: form.get('password'),
      date_of_birth: form.get('date_of_birth'),
      about_me: form.get('about_me').trim(),
    }

    // check every field, keep only the ones with a problem
    const found = {
      first_name: checkText('First name', values.first_name, LIMITS.firstName),
      last_name: checkText('Last name', values.last_name, LIMITS.lastName),
      email: checkEmail(values.email),
      nickname: checkNickname(values.nickname), // optional

      password: checkPassword(values.password),
      date_of_birth: checkDateOfBirth(values.date_of_birth),
      about_me: checkText('About me', values.about_me, LIMITS.aboutMe, { required: false }),
    }
    setErrors(found)
    if (Object.values(found).some(Boolean)) return

    setLoading(true)

    try {
      await apiPost('/register', values)

      // registering also logs you in, so the avatar can be uploaded now
      if (avatar) {
        const formData = new FormData()
        formData.append('avatar', avatar)
        await apiUpload('/avatar', formData).catch(() => {}) // never block signup
      }

      router.push('/home')
    } catch (err) {
      setError(err.message)
      setLoading(false)
    }
  }

  // clear the message of a field as soon as the user edits it
  function clearError(e) {
    const { name } = e.target
    if (errors[name]) setErrors(rest => ({ ...rest, [name]: '' }))
  }

  // small helper so every field shows its message the same way
  function fieldError(name) {
    return errors[name] ? <p className="field-error">{errors[name]}</p> : null
  }

  return (
    <form className="auth-form" onSubmit={handleSubmit} onInput={clearError} noValidate>
      <h1>Create your account</h1>
      <p className="subtitle">It takes less than a minute.</p>

      <div className="row">
        <div>
          <label>First name</label>
          <input
            name="first_name"
            maxLength={LIMITS.firstName}
            className={errors.first_name ? 'invalid' : undefined}
            autoFocus
          />
          {fieldError('first_name')}
        </div>
        <div>
          <label>Last name</label>
          <input
            name="last_name"
            maxLength={LIMITS.lastName}
            className={errors.last_name ? 'invalid' : undefined}
          />
          {fieldError('last_name')}
        </div>
      </div>

      <label>Email</label>
      <input
        name="email"
        type="email"
        maxLength={LIMITS.email}
        className={errors.email ? 'invalid' : undefined}
      />
      {fieldError('email')}

      <div className="row">
        <div>
          <label>Nickname <small>optional</small></label>
          <input
            name="nickname"
            placeholder={`${LIMITS.nickname.min}–${LIMITS.nickname.max} letters or numbers`}
            maxLength={LIMITS.nickname.max}
            className={errors.nickname ? 'invalid' : undefined}
          />
          {fieldError('nickname')}
        </div>
        <div>
          <label>Date of birth</label>
          <input
            name="date_of_birth"
            type="date"
            max={maxBirthDate()}
            className={errors.date_of_birth ? 'invalid' : undefined}
          />
          {fieldError('date_of_birth')}
        </div>
      </div>

      <label>Password <small>at least {LIMITS.password.min} characters</small></label>
      <input
        name="password"
        type="password"
        maxLength={LIMITS.password.max}
        className={errors.password ? 'invalid' : undefined}
      />
      {fieldError('password')}

      <label>Profile photo <small>optional</small></label>
      <input
        name="avatar"
        type="file"
        accept={IMAGE_ACCEPT}
        onChange={pickAvatar}
        className={errors.avatar ? 'invalid' : undefined}
      />
      {fieldError('avatar')}

      <label>About me <small>optional, up to {LIMITS.aboutMe} characters</small></label>
      <textarea
        name="about_me"
        rows={3}
        maxLength={LIMITS.aboutMe}
        className={errors.about_me ? 'invalid' : undefined}
      />
      {fieldError('about_me')}

      {error && <p className="error">{error}</p>}

      <button className="btn btn-full" disabled={loading}>
        {loading ? 'Creating account…' : 'Create account'}
      </button>

      <p className="switch">
        Already have an account? <Link href="/login">Log in</Link>
      </p>
    </form>
  )
}
