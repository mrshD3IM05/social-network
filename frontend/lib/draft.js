// What you were typing in a chat, kept while you move around the app.
//
// One draft for every conversation, not one per person, and it lives in this
// plain variable only. Nothing is saved to the browser or to the database, so
// reloading the page loses it — which is exactly what we want here.
let draft = ''

export function getDraft() {
  return draft
}

export function setDraft(value) {
  draft = value
}
