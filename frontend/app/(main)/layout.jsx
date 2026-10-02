'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useMe } from '@/lib/useMe'
import { ensureSocket } from '@/lib/socket'
import Navbar from '@/components/Navbar'
import SidePanel from '@/components/SidePanel'

// Wraps every page inside (main): checks you are logged in, then lays out
// the nav rail, the page, and the suggestions panel on wide screens.
export default function MainLayout({ children }) {
  const router = useRouter()
  const { me, loading } = useMe()

  useEffect(() => {
    if (!loading && !me) router.push('/login')
  }, [loading, me, router])

  useEffect(() => {
    if (me) ensureSocket()
  }, [me])

  if (!me) return <p className="loading">Loading…</p>

  return (
    <div className="app">
      <Navbar user={me} />
      <main className="main">
        <div className="page">{children}</div>
      </main>
      <SidePanel />
    </div>
  )
}
