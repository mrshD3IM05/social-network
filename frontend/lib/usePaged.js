'use client'

import { useCallback, useEffect, useRef, useState } from 'react'
import { apiGet } from './api'

// Same page size as the API (PageSize in backend/internal/repository).
export const PAGE_SIZE = 10

// Loads a list 10 by 10: the first page right away, the next one each time
// loadMore() is called. The API answers ?offset=0, ?offset=10, ?offset=20...
// `path` may already have a query (/users?q=ann). Pass null to wait.
export default function usePaged(path) {
  const [items, setItems] = useState(null) // null until the first page arrives
  const [hasMore, setHasMore] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const current = useRef(path) // the answer of an older path (old search) is ignored

  const load = useCallback(async (offset) => {
    if (!path) return
    current.current = path
    setLoading(true)
    try {
      const page = await apiGet(`${path}${path.includes('?') ? '&' : '?'}offset=${offset}`)
      if (current.current !== path) return
      // a new item can push an old one to the next page: never show it twice
      setItems(list => (offset === 0 ? page : [...list, ...page.filter(item => !list.some(old => old.id === item.id))]))
      setHasMore(page.length === PAGE_SIZE)
      setError(null)
    } catch (err) {
      if (current.current === path) setError(err)
    }
    if (current.current === path) setLoading(false)
  }, [path])

  // the first page, again whenever the path changes
  useEffect(() => {
    load(0)
  }, [load])

  return {
    items,
    setItems,
    hasMore,
    loading,
    error,
    loadMore: () => load(items?.length || 0),
    reload: () => load(0),
  }
}
