import { apiGet } from '@/lib/api'

// The people directory: every registered user except you.
// Backed by GET /users (internal/handler/userhandler), which only returns the
// public profile fields. Used by People, Messages and the group invite list.
export function fetchPeople() {
  return apiGet('/users')
}

// Keep the people whose name or nickname contains the search text.
// The list is small enough to filter in the browser.
export function searchPeople(people, search) {
  const text = search.trim().toLowerCase()
  if (!text) return people
  return people.filter(person =>
    `${person.first_name} ${person.last_name} ${person.nickname}`.toLowerCase().includes(text)
  )
}
