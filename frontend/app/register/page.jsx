'use client'

import { useState } from 'react'
import { postForm } from '../../lib/api'
import Link from 'next/link'

export default function Register() {
  const [error, setError] = useState(null)

  async function handleSubmit(e) {
    e.preventDefault()
    setError(null)
    const form = new FormData(e.currentTarget)

    try {
      const res = await postForm('/register', {
        email: form.get('email'),
        password: form.get('password'),
        nickname: form.get('nickname'),
        first_name: form.get('first_name'),
        last_name: form.get('last_name'),
        date_of_birth: form.get('date_of_birth'),
        about_me: form.get('about_me'),
      })
      console.log('OK:', res)
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <div>
      <h1>Register</h1>
      <form onSubmit={handleSubmit}>
        <label htmlFor="email">email</label>
        <input type="email" id="email" name="email" />

        <label htmlFor="password">password</label>
        <input type="password" id="password" name="password" />

        <label htmlFor="nickname">nickname</label>
        <input type="text" id="nickname" name="nickname" />

        <label htmlFor="first_name">first name</label>
        <input type="text" id="first_name" name="first_name" />

        <label htmlFor="last_name">last name</label>
        <input type="text" id="last_name" name="last_name" />

        <label htmlFor="date_of_birth">date of birth</label>
        <input type="date" id="date_of_birth" name="date_of_birth" />

        <label htmlFor="about_me">about me</label>
        <textarea id="about_me" name="about_me" />

        <button type="submit">Submit</button>
      </form>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <p>Already have an account? <Link href="/login">Login</Link></p>

    </div>
  )
}
