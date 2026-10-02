'use client'

import { useEffect, useState } from 'react'

// Light or dark. The choice is 'light', 'dark' or 'system' (follow the
// computer setting). It is kept in localStorage and written on <html> as
// data-theme, which globals.css reads. 'system' removes the attribute.
const KEY = 'theme'

// Runs in <head> before the page paints, so a saved dark theme never flashes
// light first (see app/layout.jsx).
export const themeScript =
  `try{var t=localStorage.getItem('${KEY}');if(t==='light'||t==='dark')document.documentElement.setAttribute('data-theme',t)}catch(e){}`

function savedChoice() {
  try {
    const value = localStorage.getItem(KEY)
    return value === 'light' || value === 'dark' ? value : 'system'
  } catch {
    return 'system' // storage can be blocked (private windows)
  }
}

function systemTheme() {
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function apply(choice) {
  const root = document.documentElement
  if (choice === 'system') root.removeAttribute('data-theme')
  else root.setAttribute('data-theme', choice)
}

export function setTheme(choice) {
  try {
    if (choice === 'system') localStorage.removeItem(KEY)
    else localStorage.setItem(KEY, choice)
  } catch {
    // not saved, but the page still switches
  }
  apply(choice)
  window.dispatchEvent(new Event('themechange'))
}

// { choice, theme }: what you picked, and what is on screen ('light' | 'dark').
// Updates when the switch is used, when the computer setting changes, and when
// another tab changes the theme.
export function useTheme() {
  const [state, setState] = useState({ choice: 'system', theme: 'light' })

  useEffect(() => {
    const media = window.matchMedia('(prefers-color-scheme: dark)')
    function update() {
      const choice = savedChoice()
      setState({ choice, theme: choice === 'system' ? systemTheme() : choice })
    }
    function fromOtherTab(e) {
      if (e.key !== KEY) return
      apply(savedChoice())
      update()
    }

    update()
    media.addEventListener('change', update)
    window.addEventListener('themechange', update)
    window.addEventListener('storage', fromOtherTab)
    return () => {
      media.removeEventListener('change', update)
      window.removeEventListener('themechange', update)
      window.removeEventListener('storage', fromOtherTab)
    }
  }, [])

  return state
}
