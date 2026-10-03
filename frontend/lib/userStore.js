// Who is logged in, asked once and then kept in memory.
//
// The (main) and (auth) layouts both need to know, and so do most pages (for
// your id, your name, your photo). Asking GET /me from each of them meant the
// same request going out again on every page you opened, and the answers could
// disagree for a moment: you change your photo in Settings and the header still
// shows the old one until something else re-asks.
//
// So the answer lives here instead: one request, shared by everything that asks
// while it is in flight, and one place to write the new value when something
// changes it. Every mutation that touches your own account already answers with
// the updated user, so those call setMe() with what they got back and no
// request is needed at all.
//
// Memory only: a reload starts over. Same shape as lib/unread.js.
import { apiGet } from './api'

let me = null

// The one GET /me in flight, so a second caller waits for it instead of
// starting another.
let pending = null

// Bumped by forgetMe, so a request that was already on its way cannot bring the
// old session back after a logout.
let generation = 0

// Parts of the page that want to redraw when the user changes.
const listeners = new Set()

function tellEveryone() {
  for (const listener of listeners) listener(me)
}

export function getMe() {
  return me
}

// The user, from memory when we already have it. Never rejects: a session that
// is not valid answers with null, which is what the layouts redirect on.
export function fetchMe() {
  if (me) return Promise.resolve(me)
  if (!pending) {
    const asked = ++generation
    pending = apiGet('/me')
      .then(
        user => {
          if (asked !== generation) return // logged out while this was in flight
          me = user
          tellEveryone()
        },
        () => {
          if (asked !== generation) return
          me = null // 401: no valid session
          tellEveryone()
        },
      )
      .finally(() => {
        pending = null
      })
  }
  return pending.then(() => me)
}

// Call this with the user a mutation answered with and everything on screen
// redraws with the new values: no request, no stale header.
export function setMe(user) {
  me = user
  tellEveryone()
  return me
}

// For when the session ends: the next page has to ask again.
export function forgetMe() {
  generation++
  me = null
  tellEveryone()
}

// Returns the function to call when the component goes away.
export function onMeChange(listener) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}