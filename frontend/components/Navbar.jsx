'use client'

import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import { apiPost } from '@/lib/api'
import Avatar from './Avatar'
import Icon from './Icon'
import MessageDot from './MessageDot'
import NotificationBadge from './NotificationBadge'
import ThemeToggle from './ThemeToggle'

const links = [
  { href: '/home', label: 'Feed', icon: 'home' },
  { href: '/people', label: 'People', icon: 'users' },
  { href: '/groups', label: 'Groups', icon: 'grid' },
  { href: '/chat', label: 'Messages', icon: 'chat' },
  { href: '/notifications', label: 'Notifications', icon: 'bell' },
  { href: '/settings', label: 'Settings', icon: 'settings' },
]

// The sidebar on the left. On phones the links become a tab bar at the
// bottom, and the logo and your avatar move to a slim bar at the top
// (see globals.css).
export default function Navbar({ user }) {
  const pathname = usePathname() // the current URL, to highlight the active link
  const router = useRouter()

  async function logout() {
    await apiPost('/logout')
    router.push('/login')
  }

  const logoutButton = (
    <button className="icon-button" onClick={logout} title="Log out" aria-label="Log out">
      <Icon name="logout" />
    </button>
  )

  return (
    <>
      <header className="topbar">
        <Link href="/home" className="brand">social-network<span>.</span></Link>
        <div className="topbar-user">
          <Link href={`/profile/${user.id}`} aria-label="Your profile">
            <Avatar user={user} size={32} />
          </Link>
          <ThemeToggle />
          {logoutButton}
        </div>
      </header>

      <aside className="rail">
        <Link href="/home" className="brand">social-network<span>.</span></Link>

        <nav className="menu">
          {links.map(link => {
            const active = pathname.startsWith(link.href)
            return (
              <Link
                key={link.href}
                href={link.href}
                className={active ? 'menu-item active' : 'menu-item'}
                aria-current={active ? 'page' : undefined}
              >
                <Icon name={link.icon} size={20} />
                <span className="menu-label">{link.label}</span>
                {link.href === '/chat' && <MessageDot myId={user.id} />}
                {link.href === '/notifications' && <NotificationBadge />}
              </Link>
            )
          })}
        </nav>

        <div className="rail-user">
          <Link href={`/profile/${user.id}`} className="user-chip">
            <Avatar user={user} size={38} />
            <span>
              <strong>{user.first_name} {user.last_name}</strong>
              <small>@{user.nickname}</small>
            </span>
          </Link>
          <ThemeToggle />
          {logoutButton}
        </div>
      </aside>
    </>
  )
}
