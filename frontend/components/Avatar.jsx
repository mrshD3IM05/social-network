import { imageUrl } from '@/lib/api'

// Shows the user's photo, or their initials if they have none
export default function Avatar({ user, size = 40 }) {
  const style = { width: size, height: size, fontSize: size / 2.6 }

  if (user?.avatar) {
    return <img className="avatar" style={style} src={imageUrl(user.avatar)} alt="" />
  }

  const initials = (user?.first_name?.[0] || '') + (user?.last_name?.[0] || '')
  return <span className="avatar" style={style}>{initials.toUpperCase() || '?'}</span>
}
