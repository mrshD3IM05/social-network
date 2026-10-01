import { apiGet } from '@/lib/api'

// The people you can start a private conversation with, i.e. the Messages list.
// Backed by GET /contacts, which applies the rule the message endpoints check:
// at least one of the two follows the other, accepted. Everyone else would only
// get a 403 from /messages/{id}, so they are not listed.
export function fetchContacts() {
  return apiGet('/contacts')
}
