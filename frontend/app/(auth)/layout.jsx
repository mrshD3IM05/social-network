// Login and register share this two-column layout: the form on the left,
// a presentation panel on the right.
export default function AuthLayout({ children }) {
  return (
    <div className="auth">
      <div className="auth-main">
        <p className="brand">social-network<span>.</span></p>
        {children}
      </div>

      <aside className="auth-aside">
        <p className="eyebrow">A quieter social network</p>
        <p className="auth-quote">
          Share what matters with the <em>people who matter.</em>
        </p>
        <ol className="auth-list">
          <li><span>01</span> Private profiles and follower-only posts</li>
          <li><span>02</span> Real-time private messages</li>
          <li><span>03</span> Groups and events, coming soon</li>
        </ol>
      </aside>
    </div>
  )
}
