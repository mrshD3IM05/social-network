'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { apiPost } from '@/lib/api'

export default function LoginPage() {
  const router = useRouter()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e) {
    e.preventDefault() // stop the browser from reloading the page
    setError('')
    setLoading(true)

    try {
      // the "email" field also accepts a nickname
      await apiPost('/login', { email, password })
      router.push('/home')
    } catch (err) {
      setError(err.message)
      setLoading(false)
    }
  }

  return (
    <form className="auth-form" onSubmit={handleSubmit}>
      <h1>Welcome back</h1>
      <p className="subtitle">Log in to continue to your feed.</p>

      <label>Email or nickname</label>
      <input value={email} onChange={e => setEmail(e.target.value)} autoFocus required />

      <label>Password</label>
      <input type="password" value={password} onChange={e => setPassword(e.target.value)} required />

      {error && <p className="error">{error}</p>}

      <button className="btn btn-full" disabled={loading}>
        {loading ? 'Logging in…' : 'Log in'}
      </button>

      <p className="switch">
        New to social-network? <Link href="/register">Create an account</Link>
      </p>
    </form>
  )
}
