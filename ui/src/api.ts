const BASE = '/api'

function token(): string {
  return localStorage.getItem('ow_token') ?? ''
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(BASE + path, {
    method,
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token()}`,
    },
    body: body ? JSON.stringify(body) : undefined,
  })

  if (res.status === 401) {
    localStorage.removeItem('ow_token')
    window.location.hash = '/login'
    throw new Error('Unauthorized')
  }

  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `HTTP ${res.status}`)
  }

  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  auth: {
    login: (email: string, password: string) =>
      request<{ token: string }>('POST', '/auth/login', { email, password }),
  },

  projects: {
    list: () => request<Project[]>('GET', '/projects'),
    get: (id: string) => request<Project>('GET', `/projects/${id}`),
    create: (data: CreateProjectInput) => request<Project>('POST', '/projects', data),
    update: (id: string, data: Partial<CreateProjectInput>) =>
      request<Project>('PATCH', `/projects/${id}`, data),
    delete: (id: string) => request<void>('DELETE', `/projects/${id}`),
  },

  logs: {
    list: (projectId: string, limit = 50, offset = 0) =>
      request<RequestLog[]>('GET', `/projects/${projectId}/logs?limit=${limit}&offset=${offset}`),
  },
}

export interface Project {
  id: string
  name: string
  api_key: string
  enabled: boolean
  upstream: string
  algorithm: 'fixed_window' | 'sliding_window' | 'token_bucket'
  fw_limit?: number
  fw_window?: number
  sw_limit?: number
  sw_window?: number
  tb_capacity?: number
  tb_refill_rate?: number
  created_at: string
  updated_at: string
}

export interface CreateProjectInput {
  name: string
  upstream: string
  algorithm: string
  enabled: boolean
  fw_limit?: number
  fw_window?: number
  sw_limit?: number
  sw_window?: number
  tb_capacity?: number
  tb_refill_rate?: number
}

export interface RequestLog {
  id: number
  project_id: string
  requested_at: string
  allowed: boolean
  status_code?: number
}