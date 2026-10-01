import Link from 'next/link'

// The "there is nothing here" screen of a page that sits inside the app shell.
// Used by every /[id] page when the id is not a real one or matches nothing.
//
// app/not-found.jsx is a different screen on purpose: it renders outside the
// (main) layout, so it has no navbar.
export default function NotFound({ title, text, back, label }) {
  return (
    <div className="empty">
      <p className="empty-title">{title}</p>
      {text && <p>{text}</p>}
      <Link href={back} className="btn">{label}</Link>
    </div>
  )
}