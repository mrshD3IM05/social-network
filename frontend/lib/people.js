import { apiGet } from '@/lib/api'

// The people directory, searched by nickname. Backed by GET /users
// (internal/handler/userhandler), which answers with up to 20 public profiles
// whose nickname contains the text, and with nothing when the search is blank.
// Used by People and the group invite list.
export function fetchPeople(search) {
  return apiGet(`/users?search=${encodeURIComponent(search)}`)
}

// The people you can start a private conversation with, i.e. the Messages list.
// Backed by GET /contacts, which applies the rule the message endpoints check:
// at least one of the two follows the other, accepted. Everyone else would only
// get a 403 from /messages/{id}, so they are not listed.
export function fetchContacts() {
  return apiGet('/contacts')
}
