'use client'

import { useCallback, useEffect, useRef, useState } from 'react'
import { apiGet } from './api'

// Same page size as the API (PageSize in backend/internal/repository).
export const PAGE_SIZE = 10

// Loads a list 10 by 10: the first page right away, the next one each time
// loadMore() is called. The next page is asked for with ?last=<id of the last
// item shown>, so new items at the top never shift what comes next.
// `path` may already have a query (/users?q=ann). Pass null to wait.
export default function usePaged(path) {
  const [items, setItems] = useState(null) // null until the first page arrives
  const [hasMore, setHasMore] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const current = useRef(path) // the answer of an older path (old search) is ignored

  // last: id of the last item already shown, 0 for the first page
  const load = useCallback(async (last) => {
    if (!path) return
    current.current = path
    setLoading(true)
    try {
      const page = await apiGet(last ? `${path}${path.includes('?') ? '&' : '?'}last=${last}` : path)
      if (current.current !== path) return
      setItems(list => (last ? [...list, ...page] : page))
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
    loadMore: () => load(items?.at(-1)?.id || 0),
    reload: () => load(0),
  }
}
