const API = '/api/v1'

export async function apiCall(path, options = {}) {
  const res = await fetch(`${API}${path}`, { credentials: 'include', ...options })

  const text = await res.text()
  let data = null
  try { data = text ? JSON.parse(text) : null } catch { data = text }

  if (!res.ok) {
    throw new Error(typeof data === 'string' ? data : `Request failed (${res.status})`)
  }
  return data
}

export function postForm(path, fields) {
  return apiCall(path, { method: 'POST', body: new URLSearchParams(fields) })
}