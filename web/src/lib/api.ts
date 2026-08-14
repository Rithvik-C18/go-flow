import type { Workflow, WorkflowSummary } from '@/lib/types'

const BASE = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.status = status
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  let res: Response
  try {
    res = await fetch(`${BASE}${path}`, {
      headers: { 'Content-Type': 'application/json' },
      ...options,
    })
  } catch {
    throw new ApiError('cannot reach the server at ' + BASE, 0)
  }

  const data = await res.json().catch(() => null)

  if (!res.ok) {
    throw new ApiError(
      (data && typeof data === 'object' && 'error' in data ? String(data.error) : `request failed (${res.status})`) ||
        `request failed (${res.status})`,
      res.status,
    )
  }

  return data as T
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
    request<{ status: string }>(`/workflows/${id}/run`, {
      method: 'POST',
      body: JSON.stringify(params),
    }),

  addNode: (workflowId: string, node: { id: string; type: string; config: Record<string, unknown> }) =>
    request(`/workflows/${workflowId}/nodes`, {
      method: 'POST',
      body: JSON.stringify(node),
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
}

export const nodeTemplates: Record<string, string> = {
  http: JSON.stringify(
    {
      url: 'https://jsonplaceholder.typicode.com/todos/1',
      httpMethod: 'GET',
      queryParameters: {},
      headers: {},
      body: {},
    },
    null,
    2,
  ),
  if: JSON.stringify(
    {
      combine: 'and',
      conditions: [{ left: '{{ statusCode }}', operator: '>=', right: 200 }],
    },
    null,
    2,
  ),
}
