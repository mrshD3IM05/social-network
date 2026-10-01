'use client'

import { useEffect, useState } from 'react'
import { apiGet } from '@/lib/api'

// The one "who is logged in" check. Both layouts read it, so the app has a
// single answer: (main) sends you to /login when there is no session, (auth)
// sends you to /home when there is.
//
// `loading` is what tells the two apart: `me` is null both while the request is
// in flight and when it comes back 401, so without it a layout could not know
// whether to wait or to redirect.
export function useMe() {
  const [me, setMe] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    apiGet('/me')
      .then(setMe)
      .catch(() => setMe(null)) // 401: no valid session
      .finally(() => setLoading(false))
  }, [])

  return { me, loading }
}
