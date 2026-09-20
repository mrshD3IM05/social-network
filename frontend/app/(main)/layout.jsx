'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { apiGet } from '@/lib/api'
import Navbar from '@/components/Navbar'
import NotificationToasts from '@/components/NotificationToasts'

// Wraps every page inside (main): checks you are logged in and shows the sidebar.
export default function MainLayout({ children }) {
  const router = useRouter()
  const [user, setUser] = useState(null)

  useEffect(() => {
    apiGet('/me')
      .then(setUser)
      .catch(() => router.push('/login')) // not logged in
  }, [router])

  if (!user) return <p className="loading">Loading…</p>

  return (
    <div className="app">
      <Navbar user={user} />
      <main className="main">
        <div className="page">{children}</div>
      </main>
      {/* the layout wraps every page, so notifications reach you anywhere */}
      <NotificationToasts />
    </div>
  )
}
