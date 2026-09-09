'use client'

import { useState } from 'react'
import { apiCall, postForm } from '../../lib/api'
import Link from 'next/link'

export default function Home() {

  async function get() {
    
    try {
      const res = await apiCall('/me', {
        Cookie
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
