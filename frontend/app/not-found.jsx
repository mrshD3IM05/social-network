import Link from 'next/link'

export default function NotFound() {
  return (
    <div className="not-found">
      <p className="eyebrow">Error 404</p>
      <h1>Page not found.</h1>
      <p className="subtitle">The page you are looking for does not exist.</p>
      <Link href="/home" className="btn">Back to the feed</Link>
    </div>
  )
}
