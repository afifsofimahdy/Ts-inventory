const baseURL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
const apiKey = import.meta.env.VITE_API_KEY || ''

export async function apiGet(path) {
  const res = await fetch(`${baseURL}${path}`, {
    headers: apiKey ? { 'X-API-Key': apiKey } : undefined
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok || data?.success === false) {
    const msg = data?.error?.message || 'Request failed'
    throw new Error(msg)
  }
  return data?.data ?? data
}

export async function apiPost(path, body) {
  const res = await fetch(`${baseURL}${path}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(apiKey ? { 'X-API-Key': apiKey } : {})
    },
    body: JSON.stringify(body)
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok || data?.success === false) {
    const msg = data?.error?.message || 'Request failed'
    throw new Error(msg)
  }
  return data?.data ?? data
}

export async function apiDelete(path) {
  const res = await fetch(`${baseURL}${path}`, {
    method: 'DELETE',
    headers: apiKey ? { 'X-API-Key': apiKey } : undefined
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok || data?.success === false) {
    const msg = data?.error?.message || 'Request failed'
    throw new Error(msg)
  }
  return data?.data ?? data
}
