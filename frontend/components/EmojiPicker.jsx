'use client'

import { useEffect, useRef, useState } from 'react'

// A small grid of emojis for the chat box. Picking one calls onPick(emoji).
const groups = {
  Smileys: ['😀', '😃', '😄', '😁', '😆', '😅', '🤣', '😂', '🙂', '😉', '😊', '😇', '🥰', '😍', '😘', '😜', '🤪', '🤗', '🤔', '🤨', '😐', '😴', '😪', '😭', '😡', '🥳', '😎', '🤓'],
  Gestures: ['👋', '🤚', '✋', '👌', '🤌', '✌️', '🤞', '🤟', '🤘', '👈', '👉', '👆', '👇', '👍', '👎', '✊', '👊', '👏', '🙌', '🙏', '💪', '🤝'],
  Hearts: ['❤️', '🧡', '💛', '💚', '💙', '💜', '🖤', '🤍', '💔', '❣️', '💕', '💞', '💓', '💗', '💖', '💘', '💝'],
  Things: ['🔥', '✨', '🎉', '🎊', '🎁', '🏆', '⚽', '🍕', '🍔', '🍟', '☕', '🍰', '🌙', '⭐', '🌈', '☀️', '🌸', '🐶', '🐱', '🚀', '💻', '📷', '🎵', '💡'],
}

export default function EmojiPicker({ onPick }) {
  const [open, setOpen] = useState(false)
  const [group, setGroup] = useState('Smileys')
  const container = useRef(null)

  // Close when clicking outside, or on Escape
  useEffect(() => {
    if (!open) return

    function onPointerDown(e) {
      if (!container.current?.contains(e.target)) setOpen(false)
    }
    function onKey(e) {
      if (e.key === 'Escape') setOpen(false)
    }

    document.addEventListener('mousedown', onPointerDown)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('mousedown', onPointerDown)
      document.removeEventListener('keydown', onKey)
    }
  }, [open])

  return (
    <div className="emoji" ref={container}>
      <button
        type="button"
        className="icon-button"
        onClick={() => setOpen(value => !value)}
        aria-expanded={open}
        aria-label="Insert an emoji"
        title="Emoji"
      >
        <span aria-hidden="true">🙂</span>
      </button>

      {open && (
        <div className="emoji-panel" role="dialog" aria-label="Emoji picker">
          <div className="emoji-tabs">
            {Object.keys(groups).map(name => (
              <button
                key={name}
                type="button"
                className={name === group ? 'emoji-tab active' : 'emoji-tab'}
                onClick={() => setGroup(name)}
              >
                {name}
              </button>
            ))}
          </div>
          <div className="emoji-grid">
            {groups[group].map(emoji => (
              <button
                key={emoji}
                type="button"
                className="emoji-cell"
                onClick={() => {
                  onPick(emoji)
                  setOpen(false)
                }}
              >
                {emoji}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
