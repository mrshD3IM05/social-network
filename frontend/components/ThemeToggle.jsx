'use client'

import { setTheme, useTheme } from '@/lib/theme'
import Icon from './Icon'

// The sun / moon button beside your profile: switches between light and dark.
// Until it is used, the app follows the computer's setting.
export default function ThemeToggle() {
  const { theme } = useTheme()
  const next = theme === 'dark' ? 'light' : 'dark'

  return (
    <button
      type="button"
      className="icon-button theme-toggle"
      onClick={() => setTheme(next)}
      title={`Switch to ${next} mode`}
      aria-label={`Switch to ${next} mode`}
    >
      <Icon name={theme === 'dark' ? 'sun' : 'moon'} />
    </button>
  )
}
