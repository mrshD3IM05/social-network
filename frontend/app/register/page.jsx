'use client'

import { useState } from 'react'
import { postForm } from '../../lib/api'
import { useRouter } from 'next/navigation'
import Link from 'next/link'

export default function Register() {
  const [error, setError] = useState(null)
  const [loading, setLoading] = useState(false)
  const router = useRouter()

  async function handleSubmit(e) {
    e.preventDefault()
    setError(null)
    setLoading(true)
    const form = new FormData(e.currentTarget)
    try {
      await postForm('/register', {
        email: form.get('email'),
        password: form.get('password'),
        nickname: form.get('nickname'),
        first_name: form.get('first_name'),
        last_name: form.get('last_name'),
        date_of_birth: form.get('date_of_birth'),
        about_me: form.get('about_me'),
      })
      router.push('/login')
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={styles.container}>
      <div style={styles.card}>
        <h1 style={styles.title}>Register</h1>
        {error && <p style={styles.error}>{error}</p>}
        <form onSubmit={handleSubmit}>
          <div style={styles.field}>
            <label htmlFor="first_name">First name</label>
            <input type="text" id="first_name" name="first_name" required style={styles.input} />
          </div>
          <div style={styles.field}>
            <label htmlFor="last_name">Last name</label>
            <input type="text" id="last_name" name="last_name" required style={styles.input} />
          </div>
          <div style={styles.field}>
            <label htmlFor="email">Email</label>
            <input type="email" id="email" name="email" required style={styles.input} />
          </div>
          <div style={styles.field}>
            <label htmlFor="password">Password</label>
            <input type="password" id="password" name="password" required style={styles.input} />
          </div>
          <div style={styles.field}>
            <label htmlFor="nickname">Nickname</label>
            <input type="text" id="nickname" name="nickname" style={styles.input} />
          </div>
          <div style={styles.field}>
            <label htmlFor="date_of_birth">Date of birth</label>
            <input type="date" id="date_of_birth" name="date_of_birth" required style={styles.input} />
          </div>
          <div style={styles.field}>
            <label htmlFor="about_me">About me</label>
            <textarea id="about_me" name="about_me" rows={3} style={styles.input} />
          </div>
          <button type="submit" style={styles.button} disabled={loading}>
            {loading ? 'Creating account...' : 'Register'}
          </button>
        </form>
        <p style={styles.link}>Already have an account? <Link href="/login">Login</Link></p>
      </div>
    </div>
  )
}

const styles = {
  container: { display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', fontFamily: 'system-ui, sans-serif', background: '#f5f5f5' },
  card: { background: '#fff', padding: 32, borderRadius: 8, boxShadow: '0 2px 8px rgba(0,0,0,0.1)', width: 360 },
  title: { margin: '0 0 16px', fontSize: 24 },
  error: { color: 'red', padding: 8, background: '#fff0f0', borderRadius: 4, fontSize: 13 },
  field: { marginBottom: 10 },
  input: { width: '100%', padding: '8px 10px', border: '1px solid #ddd', borderRadius: 4, fontSize: 14, boxSizing: 'border-box' },
  button: { width: '100%', padding: '10px 0', background: '#333', color: '#fff', border: 'none', borderRadius: 4, fontSize: 14, cursor: 'pointer', marginTop: 4 },
  link: { marginTop: 16, textAlign: 'center', fontSize: 13, color: '#666' },
}