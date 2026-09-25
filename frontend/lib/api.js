// Every request to the Go API goes through these helpers.
// next.config.js forwards /api/v1/... to the Go API on http://localhost:8080/...
const API = '/api/v1'

async function request(path, options = {}) {
  // credentials: 'include' sends the session cookie with the request
  const res = await fetch(API + path, { credentials: 'include', ...options })
  const text = await res.text()

  if (!res.ok) {
    const error = new Error(text || 'Something went wrong')
    error.status = res.status // e.g. 401 = not logged in, 403 = not allowed
    throw error
  }

  return text ? JSON.parse(text) : null
}

export function apiGet(path) {
  return request(path)
}

// The API reads form fields, so we send URLSearchParams (like a normal form)
export function apiPost(path, data = {}) {
  return request(path, { method: 'POST', body: new URLSearchParams(data) })
}

// Same form encoding as apiPost — the API's ParseForm reads PUT bodies too
export function apiPut(path, data = {}) {
  return request(path, { method: 'PUT', body: new URLSearchParams(data) })
}

export function apiDelete(path) {
  return request(path, { method: 'DELETE' })
}

// For file uploads: pass a FormData object
export function apiUpload(path, formData) {
  return request(path, { method: 'POST', body: formData })
}

// URL of an uploaded image (avatars and post pictures are stored as file ids)
export function imageUrl(id) {
  return `${API}/fs/${id}`
}
