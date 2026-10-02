'use client'

import { socketUrl } from './api'

let socket = null
let reconnectTimer = null
const listeners = new Set()

function emit(data) {
  for (const cb of listeners) {
    try {
      cb(data)
    } catch (err) {
      // ignore listener errors
    }
  }
  if (typeof window !== 'undefined') {
    try {
      window.dispatchEvent(new CustomEvent('ws:message', { detail: data }))
    } catch (err) {
      // ignore
    }
  }
}

function connect() {
  if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
    return
  }
  try {
    socket = new WebSocket(socketUrl())
  } catch (err) {
    if (!reconnectTimer) {
      reconnectTimer = setTimeout(() => {
        reconnectTimer = null
        connect()
      }, 2000)
    }
    return
  }

  socket.onmessage = (e) => {
    try {
      const data = JSON.parse(e.data)
      emit(data)
    } catch (err) {
      // ignore malformed messages
    }
  }

  socket.onclose = () => {
    socket = null
    if (!reconnectTimer) {
      reconnectTimer = setTimeout(() => {
        reconnectTimer = null
        connect()
      }, 2000)
    }
  }

  socket.onerror = () => {
    try {
      socket.close()
    } catch (err) {
      // ignore
    }
  }
}

export function ensureSocket() {
  connect()
}

export function subscribe(cb) {
  listeners.add(cb)
  return () => listeners.delete(cb)
}

export function closeSocket() {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  if (socket) {
    try {
      socket.close()
    } catch (err) {
      // ignore
    }
    socket = null
  }
}
