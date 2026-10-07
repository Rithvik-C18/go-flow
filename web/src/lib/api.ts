import type { Workflow, WorkflowSummary } from '@/lib/types'

const BASE = import.meta.env.VITE_API_URL ?? '/api'
const ACCESS_TOKEN_KEY = 'access_token'

export interface User {
  id: number
  username: string
  email: string
}

interface AuthResponse {
  access_token: string
  user: User
}

interface AccessTokenClaims {
  uid: number
  username: string
  exp: number
}

export function getAccessToken(): string | null {
  return localStorage.getItem(ACCESS_TOKEN_KEY)
}

export function setAccessToken(token: string): void {
  localStorage.setItem(ACCESS_TOKEN_KEY, token)
  notifyAccessTokenListeners(token)
}

export function clearAccessToken(): void {
  localStorage.removeItem(ACCESS_TOKEN_KEY)
  notifyAccessTokenListeners(null)
}

export function subscribeToAccessToken(listener: (token: string | null) => void): () => void {
  accessTokenListeners.add(listener)
  return () => accessTokenListeners.delete(listener)
}

const accessTokenListeners = new Set<(token: string | null) => void>()

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

function notifyAccessTokenListeners(token: string | null): void {
  accessTokenListeners.forEach((listener) => listener(token))
}

function decodeAccessToken(token: string): AccessTokenClaims | null {
  try {
    const [, payload] = token.split('.')
    if (!payload) return null

    const normalizedPayload = payload.replace(/-/g, '+').replace(/_/g, '/')
    const paddedPayload = normalizedPayload.padEnd(
      Math.ceil(normalizedPayload.length / 4) * 4,
      '=',
    )
    const parsed = JSON.parse(atob(paddedPayload)) as Partial<AccessTokenClaims>
    const userId = typeof parsed.uid === 'number' ? parsed.uid : Number(parsed.uid)

    if (
      !Number.isInteger(userId) ||
      userId <= 0 ||
      typeof parsed.username !== 'string' ||
      typeof parsed.exp !== 'number'
    ) {
      return null
    }

    return { uid: userId, username: parsed.username, exp: parsed.exp }
  } catch {
    return null
  }
}

export function getAccessTokenClaims(token: string): AccessTokenClaims | null {
  const claims = decodeAccessToken(token)
  return claims && claims.exp * 1000 > Date.now() ? claims : null
}

function getErrorMessage(data: unknown, status: number): string {
  if (
    data &&
    typeof data === 'object' &&
    'error' in data &&
    typeof data.error === 'string' &&
    data.error.length > 0
  ) {
    return data.error
  }

  return `request failed (${status})`
}

async function parseResponse(response: Response): Promise<unknown> {
  return response.json().catch(() => null)
}

let refreshPromise: Promise<string> | null = null

async function refreshAccessToken(): Promise<string> {
  let response: Response

  try {
    response = await fetch(`${BASE}/refresh`, {
      method: 'POST',
      credentials: 'include',
    })
  } catch {
    throw new ApiError(`cannot reach the server at ${BASE}`, 0)
  }

  const data = await parseResponse(response)

  if (!response.ok) {
    clearAccessToken()
    throw new ApiError(
      response.status === 401 ? 'Your session has expired. Please sign in again.' : getErrorMessage(data, response.status),
      response.status,
    )
  }

  const token =
    data && typeof data === 'object' && 'access_token' in data && typeof data.access_token === 'string'
      ? data.access_token
      : null

  if (!token || !getAccessTokenClaims(token)) {
    clearAccessToken()
    throw new ApiError('The server returned an invalid access token', response.status)
  }

  setAccessToken(token)
  return token
}

function refreshOnce(): Promise<string> {
  refreshPromise ??= refreshAccessToken().finally(() => {
    refreshPromise = null
  })

  return refreshPromise
}

function isAuthenticationPath(path: string): boolean {
  return path === '/login' || path === '/refresh'
}

async function requestOnce<T>(path: string, options: RequestInit, token: string | null): Promise<T> {
  const headers = new Headers(options.headers)
  headers.set('Content-Type', 'application/json')

  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  let response: Response

  try {
    response = await fetch(`${BASE}${path}`, {
      ...options,
      headers,
      credentials: 'include',
    })
  } catch {
    throw new ApiError(`cannot reach the server at ${BASE}`, 0)
  }

  const data = await parseResponse(response)

  if (!response.ok) {
    throw new ApiError(getErrorMessage(data, response.status), response.status)
  }

  return data as T
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  try {
    return await requestOnce(path, options, getAccessToken())
  } catch (error) {
    const shouldRefresh =
      error instanceof ApiError &&
      error.status === 401 &&
      !isAuthenticationPath(path)

    if (!shouldRefresh) {
      throw error
    }

    return await requestOnce<T>(path, options, await refreshOnce())
  }
}

export const api = {
  listWorkflows: () => request<WorkflowSummary[]>('/workflows'),

  createWorkflow: (id: string, name: string) =>
    request<WorkflowSummary>('/workflows', {
      method: 'POST',
      body: JSON.stringify({ id, name }),
    }),

  getWorkflow: (id: string) => request<Workflow>(`/workflows/${id}`),

  deleteWorkflow: (id: string) =>
    request(`/workflows/${id}`, { method: 'DELETE' }),

  runWorkflow: (id: string, params: Record<string, unknown>) =>
    request<{ status: string; outputs: Record<string, Record<string, unknown>> }>(`/workflows/${id}/run`, {
      method: 'POST',
      body: JSON.stringify(params),
    }),

  runNode: (workflowId: string, nodeId: string, params: Record<string, unknown> = {}) =>
    request<{ status: string; output: Record<string, unknown>; outputs: Record<string, Record<string, unknown>> }>(
      `/workflows/${workflowId}/nodes/${nodeId}/run`,
      { method: 'POST', body: JSON.stringify(params) },
    ),

  addNode: (workflowId: string, node: { id: string; type: string; config: Record<string, unknown> }) =>
    request(`/workflows/${workflowId}/nodes`, {
      method: 'POST',
      body: JSON.stringify(node),
    }),

  updateNode: (workflowId: string, nodeId: string, type: string, config: Record<string, unknown>) =>
    request(`/workflows/${workflowId}/nodes/${nodeId}`, {
      method: 'PUT',
      body: JSON.stringify({ type, config }),
    }),

  deleteNode: (workflowId: string, nodeId: string) =>
    request(`/workflows/${workflowId}/nodes/${nodeId}`, { method: 'DELETE' }),

  addEdge: (workflowId: string, from: string, to: string) =>
    request(`/workflows/${workflowId}/edges`, {
      method: 'POST',
      body: JSON.stringify({ from, to }),
    }),

  deleteEdge: (workflowId: string, from: string, to: string) =>
    request(`/workflows/${workflowId}/edges`, {
      method: 'DELETE',
      body: JSON.stringify({ from, to }),
    }),

  register: (username: string, email: string, password: string) =>
    request<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ username, email, password }),
    }),

  login: (identifier: string, password: string) =>
    request<AuthResponse>('/login', {
      method: 'POST',
      body: JSON.stringify({ username: identifier, password }),
    }),

  logout: () =>
    request('/logout', { method: 'POST' }),
}

export const nodeTemplates: Record<string, string> = {
  http: JSON.stringify({
    url: 'https://jsonplaceholder.typicode.com/todos/1',
    httpMethod: 'GET',
    queryParameters: {},
    headers: {},
    body: {},
  }),
  if: JSON.stringify({
    combine: 'and',
    conditions: [{ left: '{{ statusCode }}', operator: '>=', right: 200 }],
  }),
  gemini: JSON.stringify({
    prompt: 'Explain quantum computing in simple terms',
    model: 'gemini-3.1-flash-lite',
    temperature: 0.7,
    maxTokens: 1024,
  }),
  google_docs: JSON.stringify({
    operation: 'create',
    title: 'My Document',
    content: 'Hello World',
  }),
  google_sheets: JSON.stringify({
    operation: 'create',
    title: 'My Spreadsheet',
    data: [['Name', 'Age'], ['John', 30], ['Jane', 25]],
  }),
}
