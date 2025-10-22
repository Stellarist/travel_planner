import { getApiUrl } from './config'

export const formatAmount = (amount: number): string => {
  if (amount >= 10000) return (amount / 10000).toFixed(1) + '万'
  return amount.toFixed(2)
}

export const formatDate = (dateString: string): string => {
  const d = new Date(dateString)
  return `${d.getFullYear()}-${d.getMonth() + 1}-${d.getDate()}`
}

export const buildUrl = (path: string, params?: Record<string, string | undefined>): string => {
  if (!params) return getApiUrl(path)
  const q = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v !== undefined && v !== '') q.set(k, v)
  })
  const qs = q.toString()
  return getApiUrl(path) + (qs ? `?${qs}` : '')
}

const authHeaders = (): HeadersInit => {
  const token = localStorage.getItem('token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export const apiGet = async (path: string, params?: Record<string, string | undefined>) => {
  const res = await fetch(buildUrl(path, params), { headers: authHeaders() })
  return res.json()
}

export const apiPost = async (path: string, body: any) => {
  const res = await fetch(getApiUrl(path), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify(body),
  })
  return res.json()
}
