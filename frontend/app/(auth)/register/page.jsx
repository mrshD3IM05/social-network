'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { apiPost } from '@/lib/api'

export default function RegisterPage() {
  const router = useRouter()
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')
    setLoading(true)

    // read every input of the form by its "name"
    const form = new FormData(e.target)

    try {
      await apiPost('/register', {
        first_name: form.get('first_name'),
        last_name: form.get('last_name'),
        email: form.get('email').toLowerCase(),
        nickname: form.get('nickname').toLowerCase(),
        password: form.get('password'),
        date_of_birth: form.get('date_of_birth'),
        about_me: form.get('about_me'),
      })
      // registering also logs you in
      router.push('/home')
    } catch (err) {
      setError(err.message)
      setLoading(false)
    }
  }

  return (
    <form className="auth-form" onSubmit={handleSubmit}>
      <h1>Create your account</h1>
      <p className="subtitle">It takes less than a minute.</p>

      <div className="row">
        <div>
          <label>First name</label>
          <input name="first_name" autoFocus required />
        </div>
        <div>
          <label>Last name</label>
          <input name="last_name" required />
        </div>
      </div>

      <label>Email</label>
      <input name="email" type="email" required />

      <div className="row">
        <div>
          <label>Nickname</label>
          <input name="nickname" placeholder="4–15 letters or numbers" required />
        </div>
        <div>
          <label>Date of birth</label>
          <input name="date_of_birth" type="date" required />
        </div>
      </div>

      <label>Password</label>
      <input name="password" type="password" required />

      <label>About me <small>optional</small></label>
      <textarea name="about_me" rows={3} />

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
