'use client'

// The button under a list loaded with usePaged. It disappears once the last
// page has arrived.
export default function LoadMore({ list }) {
  if (!list.hasMore) return null
  return (
    <button type="button" className="btn btn-light load-more" onClick={list.loadMore} disabled={list.loading}>
      {list.loading ? 'Loading…' : 'Load more'}
    </button>
  )
}
