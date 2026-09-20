// One WebSocket for the whole app.
//
// The pages used to hardcode ws://<hostname>:8080, which only works when the Go
// server answers on port 8080 of the page's own host. Behind Caddy (port 8000)
// and in Docker (8080 is not published) the socket never connected. The URL is
// built from the page instead, so it follows whatever proxy is in front.
//
// Every page that needed live data also opened its own socket and never
// reconnected. There is one shared socket here, with a retry.

// Same origin: Caddy and the Next.js dev rewrite both forward /api/v1,
// including the upgrade request.
function url() {
  if (process.env.NEXT_PUBLIC_WS_URL) return process.env.NEXT_PUBLIC_WS_URL
  const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${scheme}//${window.location.host}/api/v1/ws`
}

let socket = null
let attempts = 0
let reconnectTimer = null
let closeTimer = null
const listeners = new Set()

function isOpen() {
  return socket?.readyState === WebSocket.OPEN
}

function notify(event) {
  for (const listener of [...listeners]) listener(event)
}

function connect() {
  if (socket) return // already open or connecting

  socket = new WebSocket(url())

  socket.onopen = () => {
    attempts = 0
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

    if (listeners.size === 0 || reconnectTimer) return

    // wait a bit longer after each failure, up to half a minute
    const delay = Math.min(1000 * 2 ** attempts, 30000)
    attempts++
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      connect()
    }, delay)
  }

  socket.onerror = () => socket?.close()
}

// Listen to every event the server pushes. Returns the unsubscribe function.
export function onSocketEvent(listener) {
  listeners.add(listener)
  clearTimeout(closeTimer)
  connect()

  // A listener that subscribes after the socket is already up would otherwise
  // never learn the connection state and would sit on "reconnecting" forever.
  listener({ type: 'socket', state: isOpen() ? 'open' : 'closed' })

  return () => {
    listeners.delete(listener)
    if (listeners.size > 0) return

    // Moving between pages unmounts one listener and mounts the next. Closing
    // straight away would rebuild the socket on every click, so the close waits
    // to see whether someone else takes over.
    clearTimeout(closeTimer)
    closeTimer = setTimeout(() => {
      if (listeners.size > 0) return
      clearTimeout(reconnectTimer)
      reconnectTimer = null
      socket?.close()
    }, 3000)
  }
}

// Send a chat message. Answers false when the socket is not up, so the caller
// can tell the user instead of losing the message quietly.
export function sendSocket(payload) {
  if (!isOpen()) return false
  socket.send(JSON.stringify(payload))
  return true
}
