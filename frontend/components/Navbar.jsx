'use client'

import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import { apiPost } from '@/lib/api'
import Avatar from './Avatar'
import Icon from './Icon'
import MessageDot from './MessageDot'

const links = [
  { href: '/home', label: 'Feed', icon: 'home' },
  { href: '/people', label: 'People', icon: 'users' },
  { href: '/groups', label: 'Groups', icon: 'grid' },
  { href: '/chat', label: 'Messages', icon: 'chat' },
  { href: '/notifications', label: 'Notifications', icon: 'bell' },
  { href: '/settings', label: 'Settings', icon: 'settings' },
]

// The sidebar on the left (it becomes a top bar on phones, see globals.css)
export default function Navbar({ user }) {
  const pathname = usePathname() // the current URL, to highlight the active link
  const router = useRouter()

  async function logout() {
    await apiPost('/logout')
    router.push('/login')
  }

  return (
    <aside className="sidebar">
      <Link href="/home" className="brand">social-network<span>.</span></Link>

      <nav className="menu">
        {links.map(link => (
          <Link
            key={link.href}
            href={link.href}
            className={pathname.startsWith(link.href) ? 'menu-item active' : 'menu-item'}
          >
            <Icon name={link.icon} />
            <span>{link.label}</span>
            {link.href === '/chat' && <MessageDot myId={user.id} />}
          </Link>
        ))}
      </nav>

      <div className="sidebar-user">
        <Link href={`/profile/${user.id}`} className="user-chip">
          <Avatar user={user} size={36} />
          <span>
            <strong>{user.first_name} {user.last_name}</strong>
            <small>@{user.nickname}</small>
          </span>
        </Link>
        <button className="icon-button" onClick={logout} title="Log out">
          <Icon name="logout" />
        </button>
      </div>
    </aside>
  )
}
