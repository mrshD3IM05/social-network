'use client'

import { useEffect } from 'react'
import Icon from '@/components/Icon'

// Simple centered dialog. Close with the X button, the backdrop, or Escape.
export default function Modal({ title, onClose, children }) {
  useEffect(() => {
    function onKey(e) {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKey)
    document.body.style.overflow = 'hidden' // no page scroll behind the modal
    return () => {
      document.removeEventListener('keydown', onKey)
      document.body.style.overflow = ''
    }
  }, [onClose])

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div
        className="modal"
        role="dialog"
        aria-modal="true"
        onClick={e => e.stopPropagation()} // clicks inside stay open
      >
        <header className="modal-header">
          <h2>{title}</h2>
          <button className="icon-button" onClick={onClose} title="Close">
            <Icon name="x" />
          </button>
        </header>
        {children}
      </div>
    </div>
  )
}
