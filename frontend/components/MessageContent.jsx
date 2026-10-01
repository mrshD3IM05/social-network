'use client'

import { useState } from 'react'

const PREVIEW_LENGTH = 200

// Keeps large chat bubbles scannable without hiding the full message.
export default function MessageContent({ content }) {
  const [expanded, setExpanded] = useState(false)
  const characters = Array.from(content)
  const isLong = characters.length > PREVIEW_LENGTH
  const visible = expanded || !isLong ? content : `${characters.slice(0, PREVIEW_LENGTH).join('')}…`

  return (
    <>
      <span>{visible}</span>
      {isLong && (
        <button
          type="button"
          className="message-expand"
          aria-expanded={expanded}
          onClick={() => setExpanded(value => !value)}
        >
          {expanded ? 'Show less' : 'Show more'}
        </button>
      )}
    </>
  )
}
