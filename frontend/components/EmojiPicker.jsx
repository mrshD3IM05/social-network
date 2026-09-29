'use client'

import { useState } from 'react'

const EMOJIS = [
  '😀', '😂', '😊', '😍', '😘', '😎', '🤔', '😅',
  '😢', '😭', '😡', '😱', '🥳', '😴', '🙄', '😇',
  '👍', '👎', '👏', '🙏', '💪', '👋', '🤝', '✌️',
  '❤️', '💔', '🔥', '✨', '🎉', '💯', '⭐', '☕',
]

// A button that opens a small grid of emojis; onPick gets the chosen one.
export default function EmojiPicker({ onPick }) {
  const [open, setOpen] = useState(false)

  return (
    <div className="emoji-picker">
      <button type="button" className="icon-button" title="Add an emoji" onClick={() => setOpen(!open)}>
        😊
      </button>
      {open && (
        <div className="emoji-grid">
          {EMOJIS.map(emoji => (
            <button
              key={emoji}
              type="button"
              onClick={() => {
                onPick(emoji)
                setOpen(false)
              }}
            >
              {emoji}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
