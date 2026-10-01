'use client'

import { useEffect, useRef, useState } from 'react'
import { apiGet } from './api'
import { useThrottle } from './timing'

export const MESSAGE_PAGE_SIZE = 15

// Chat history grows upward. The API returns the next older page when given
// the id of the oldest message currently shown.
export default function useMessageHistory(path) {
  const [messages, setMessages] = useState(null)
  const [hasMore, setHasMore] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [error, setError] = useState(null)
  const oldestID = useRef(0)
  const loading = useRef(false)
  const current = useRef(path)

  useEffect(() => {
    current.current = path
    oldestID.current = 0
    loading.current = false
    setMessages(null)
    setHasMore(false)
    setLoadingMore(false)
    setError(null)
    if (!path) return

    apiGet(path)
      .then(page => {
        if (current.current !== path) return
        oldestID.current = page[0]?.id || 0
        setMessages(page)
        setHasMore(page.length === MESSAGE_PAGE_SIZE)
      })
      .catch(err => {
        if (current.current === path) {
          setMessages([])
          setError(err)
        }
      })
  }, [path])

  const loadMore = useThrottle(async () => {
    if (!path || loading.current || !hasMore || !oldestID.current) return
    const before = oldestID.current
    loading.current = true
    setLoadingMore(true)
    try {
      const page = await apiGet(`${path}${path.includes('?') ? '&' : '?'}last=${before}`)
      if (current.current !== path) return
      oldestID.current = page[0]?.id || before
      setMessages(list => [...page, ...(list || [])])
      setHasMore(page.length === MESSAGE_PAGE_SIZE)
    } catch (err) {
      if (current.current === path) setError(err)
    } finally {
      if (current.current === path) {
        loading.current = false
        setLoadingMore(false)
      }
    }
  }, 500)

  return { messages, setMessages, hasMore, loadingMore, error, loadMore }
}
