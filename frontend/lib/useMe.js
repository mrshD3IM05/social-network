'use client'

import { useEffect, useState } from 'react'
import { fetchMe, getMe, onMeChange } from './userStore'

// The one "who is logged in" check. Both layouts read it, so the app has a
// single answer: (main) sends you to /login when there is no session, (auth)
// sends you to /home when there is.
//
// The answer itself is kept in lib/userStore, so every page that asks gets the
// same one and nobody asks twice. What this hook adds is the redraw: it follows
// the store, so when something changes your account (a new photo, a new privacy
// setting) every page showing you updates without asking again.
//
// `loading` is what tells the two apart: `me` is null both while the request is
// in flight and when it comes back 401, so without it a layout could not know
// whether to wait or to redirect. It starts false when the store already has
// the answer, so a page you can reach with a click never flashes "Loading…".
export function useMe() {
  const [me, setMeState] = useState(getMe)
  const [loading, setLoading] = useState(() => getMe() === null)

  useEffect(() => {
    let here = true

    // redraw whenever the store changes, and when our own request settles
    const stop = onMeChange(user => {
      if (!here) return
      setMeState(user)
      setLoading(false)
    })

    fetchMe().finally(() => {
      if (here) setLoading(false)
    })

    return () => {
      here = false
      stop()
    }
  }, [])

  return { me, loading }
}