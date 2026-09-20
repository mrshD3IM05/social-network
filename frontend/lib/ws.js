// One WebSocket for the whole app.
//
// Two problems this fixes:
//  1. The pages used to hardcode ws://<hostname>:8080, which only works when the
//     Go server is reachable on port 8080 of the page's host. Behind Caddy (port
//     8000) or in Docker (8080 is not published) the socket never connected.
//     The URL is now derived from the page itself, so it follows the proxy.
//  2. Every page that needed live data opened its own socket and never
//     reconnected. There is now a single shared socket with backoff.

const PATH = '/api/v1/ws'

// Candidate URLs, tried in order until one connects.
// Same origin first (Caddy and the Next.js rewrite both forward /api/v1);
// the direct backend port stays as a fallback for `next dev` setups where the
// rewrite does not proxy the upgrade request.
function candidates() {
  const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const urls = []

  if (process.env.NEXT_PUBLIC_WS_URL) urls.push(process.env.NEXT_PUBLIC_WS_URL)
  urls.push(`${scheme}//${window.location.host}${PATH}`)
  urls.push(`${scheme}//${window.location.hostname}:8080${PATH}`)

  return urls.filter((url, index) => urls.indexOf(url) === index)
}

let socket = null
let candidate = 0
let attempts = 0
let reconnectTimer = null
let closeTimer = null
const listeners = new Set()
const queue = []

function isOpen() {
  return socket?.readyState === WebSocket.OPEN
}

function notify(event) {
  for (const listener of [...listeners]) {
    try {
      listener(event)
    } catch {
      // one broken listener must not stop the others
    }
  }
}

function connect() {
  if (typeof window === 'undefined') return
  if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) return

  const urls = candidates()
  const url = urls[candidate % urls.length]

  let opened = false
  socket = new WebSocket(url)

  socket.onopen = () => {
    opened = true
    attempts = 0
    while (queue.length > 0) socket.send(queue.shift())
    notify({ type: 'socket', state: 'open' })
  }

  socket.onmessage = event => {
    try {
      notify(JSON.parse(event.data))
    } catch {
      // ignore anything that is not JSON
    }
  }

  socket.onclose = () => {
    socket = null
    notify({ type: 'socket', state: 'closed' })
    // A URL that never opened is probably the wrong one, so try the next.
    if (!opened) candidate++
    if (listeners.size > 0) scheduleReconnect()
  }

  socket.onerror = () => socket?.close()
}

function scheduleReconnect() {
  if (reconnectTimer) return
  const delay = Math.min(1000 * 2 ** attempts, 30000)
  attempts++
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null
    connect()
  }, delay)
}

// Listen to every event the server pushes. Returns the unsubscribe function.
export function onSocketEvent(listener) {
  listeners.add(listener)
  clearTimeout(closeTimer)
  closeTimer = null
  connect()

  // A listener that subscribes after the socket is already up would otherwise
  // never learn the connection state, and would sit on "reconnecting" forever.
  listener({ type: 'socket', state: isOpen() ? 'open' : 'closed' })

  return () => {
    listeners.delete(listener)
    if (listeners.size > 0) return
    // Moving between pages unmounts one listener and mounts the next. Closing
    // straight away would tear the socket down and rebuild it on every click,
    // so the close waits to see whether someone else takes over.
    clearTimeout(closeTimer)
    closeTimer = setTimeout(() => {
      closeTimer = null
      if (listeners.size > 0) return
      clearTimeout(reconnectTimer)
      reconnectTimer = null
      socket?.close()
      socket = null
    }, 3000)
  }
}

// Send a chat message. It is queued while the socket is reconnecting.
export function sendSocket(payload) {
  const data = JSON.stringify(payload)
  if (socket?.readyState === WebSocket.OPEN) {
    socket.send(data)
    return true
  }
  queue.push(data)
  connect()
  return false
}
