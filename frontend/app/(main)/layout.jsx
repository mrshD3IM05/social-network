'use client'

import { useEffect, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { useMe } from '@/lib/useMe'
import { ensureSocket } from '@/lib/socket'
import ConnectionBanner from '@/components/ConnectionBanner'
import Navbar from '@/components/Navbar'
import SidePanel from '@/components/SidePanel'

// Wraps every page inside (main): checks you are logged in, then lays out
// the nav rail, the page, and the suggestions panel on wide screens.
export default function MainLayout({ children }) {
  const router = useRouter()
  const { me, loading } = useMe()
  const ensuredRef = useRef(false)

  useEffect(() => {
    if (!loading && !me) router.push('/login')
  }, [loading, me, router])

  useEffect(() => {
    if (me && !ensuredRef.current) {
      ensuredRef.current = true
      ensureSocket()
    }
  }, [me])

  if (!me) return <p className="loading">Loading…</p>

  return (
    <>
      <ConnectionBanner />
      <div className="app">
        <Navbar user={me} />
        <main className="main">
          <div className="page">{children}</div>
        </main>
        <SidePanel />
      </div>
    </>
  )
}
