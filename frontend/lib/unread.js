// Who wrote to you and has not been read yet.
//
// Just a list of user ids kept in memory: the sidebar shows a dot while it is
// not empty, and the Messages list shows a dot on each of those people.
// Opening a conversation removes that person from the list.
// Nothing is saved, so a reload starts over.
let unread = new Set()

// Parts of the page that want to redraw when the list changes.
const listeners = new Set()

function tellEveryone() {
  for (const listener of listeners) listener(new Set(unread))
}

export function markUnread(userId) {
  unread.add(userId)
  tellEveryone()
}

export function markRead(userId) {
  if (unread.delete(userId)) tellEveryone()
}

export function getUnread() {
  return new Set(unread)
}

// Returns the function to call when the component goes away.
export function onUnreadChange(listener) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}
