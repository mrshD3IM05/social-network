'use client'

import { socketUrl } from './api'

let socket = null
let connecting = false
let stopped = false
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

function scheduleReconnect() {
  if (stopped || reconnectTimer) return
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null
    connect()
  }, 2000)
}

const queue = []

function flushQueue() {
  if (!socket || socket.readyState !== WebSocket.OPEN) return
  while (queue.length) socket.send(queue.shift())
}

function connect() {
  if (stopped) return
  if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
    return
  }
  if (connecting) return
  connecting = true
  try {
    socket = new WebSocket(socketUrl())
  } catch (err) {
    connecting = false
    scheduleReconnect()
    return
  }

  socket.onopen = () => {
    connecting = false
    flushQueue()
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
    connecting = false
    socket = null
    if (stopped) return
    scheduleReconnect()
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
  stopped = false
  connect()
}

// True when the one shared connection can carry a message right now.
export function isSocketOpen() {
  return !!socket && socket.readyState === WebSocket.OPEN
}

// Sends one payload as JSON. Anything typed before the connection opened is
// queued and goes out on open, so pages do not have to own a socket.
export function sendWs(payload) {
  if (!socket || socket.readyState === WebSocket.CLOSED) connect()
  if (!socket || socket.readyState === WebSocket.CLOSED) return false
  const text = JSON.stringify(payload)
  if (socket.readyState === WebSocket.OPEN) socket.send(text)
  else queue.push(text)
  return true
}

export function subscribe(cb) {
  listeners.add(cb)
  return () => listeners.delete(cb)
}

export function closeSocket() {
  stopped = true
  queue.length = 0
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
  connecting = false
}
