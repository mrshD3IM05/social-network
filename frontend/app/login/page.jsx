'use client'

import { useState } from 'react'
import { postForm } from '../../lib/api'
import Link from 'next/link'

export default function Login() {
  const [error, setError] = useState(null)

  async function handleSubmit(e) {
    e.preventDefault()
    setError(null)
    const form = new FormData(e.currentTarget)

    try {
      const res = await postForm('/login', {
        email: form.get('email'),
        password: form.get('password'),
      })
      console.log('OK:', res)
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <div>
      <h1>Hello</h1>
      <form onSubmit={handleSubmit}>
        <label htmlFor="email">email</label>
        <input type="text" id="email" name="email" />

        <label htmlFor="password">password</label>
        <input type="password" id="password" name="password" />

        <button type="submit">Submit</button>
      </form>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <p>No account? <Link href="/register">Register</Link></p>
    </div>
  )
}
