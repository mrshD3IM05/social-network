import { apiGet } from '@/lib/api'

// Everyone on the network except you. This used to be derived from the authors
// of the posts in your feed, which hid anyone who had not posted yet; the API
// now has a real directory endpoint.
export function fetchPeople() {
  return apiGet('/users')
}

// The people you already exchanged messages with, most recent first.
export function fetchConversations() {
  return apiGet('/conversations')
}
