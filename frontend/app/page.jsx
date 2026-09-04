'use client'
import { apiCall } from "../lib/api"

export default function Page() {

  async function testApi() {
    try {
      const res = await apiCall('/me')
      console.log('OK:', res)
    } catch (err) {
      console.log('Erreur:', err.message)
    }
  }

  return (
    <div>
      <h1>Hello, Next.js!</h1>
      <button onClick={testApi}>Tester /me</button>
    </div>
  )
}