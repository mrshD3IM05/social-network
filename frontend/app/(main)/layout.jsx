'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useMe } from '@/lib/useMe'
import Navbar from '@/components/Navbar'

// Wraps every page inside (main): checks you are logged in and shows the sidebar.
export default function MainLayout({ children }) {
  const router = useRouter()
  const { me, loading } = useMe()

  useEffect(() => {
    if (!loading && !me) router.push('/login')
  }, [loading, me, router])

  if (!me) return <p className="loading">Loading…</p>

  return (
    <div className="app">
      <Navbar user={me} />
      <main className="main">
        <div className="page">{children}</div>
      </main>
    </div>
  )
}
