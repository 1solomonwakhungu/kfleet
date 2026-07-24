import { notifyAuthenticationRequired } from './authApi'

const API_BASE = '/api/v1'

export interface PendingAgent {
  id: string
  name: string
  labels: Record<string, string>
  registeredAt?: string
  kubernetesVersion?: string
  agentVersion?: string
}

export class PendingAgentsApiError extends Error {
  readonly status: number
  readonly body: unknown

  constructor(status: number, message: string, body?: unknown) {
    super(message)
    this.name = 'PendingAgentsApiError'
    this.status = status
    this.body = body
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function errorMessage(body: unknown, status: number): string {
  if (isRecord(body)) {
    if (typeof body.error === 'string' && body.error.trim()) return body.error
    if (typeof body.message === 'string' && body.message.trim()) return body.message
  }

  return `Request failed with status ${status}`
}

async function readBody(response: Response): Promise<unknown> {
  const text = await response.text()
  if (!text) return undefined

  try {
    return JSON.parse(text) as unknown
  } catch {
    return text
  }
}

async function request(path: string, init: RequestInit): Promise<unknown> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      Accept: 'application/json',
      ...init.headers,
    },
  })
  notifyAuthenticationRequired(response)
  const body = await readBody(response)

  if (!response.ok) {
    throw new PendingAgentsApiError(response.status, errorMessage(body, response.status), body)
  }

  if (body === undefined || typeof body === 'string') {
    throw new PendingAgentsApiError(response.status, 'The server returned an invalid JSON response.', body)
  }

  return body
}

function optionalString(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim() ? value : undefined
}

function parseLabels(value: unknown): Record<string, string> {
  if (value === undefined || value === null) return {}
  if (!isRecord(value) || Object.values(value).some((label) => typeof label !== 'string')) {
    throw new Error('Invalid labels in pending agent response.')
  }

  return value as Record<string, string>
}

function parsePendingAgent(value: unknown): PendingAgent {
  if (
    !isRecord(value) ||
    typeof value.id !== 'string' ||
    !value.id.trim() ||
    typeof value.name !== 'string' ||
    !value.name.trim()
  ) {
    throw new Error('Invalid pending agent response.')
  }

  return {
    id: value.id,
    name: value.name,
    labels: parseLabels(value.labels),
    registeredAt: optionalString(value.registeredAt),
    kubernetesVersion: optionalString(value.k8sVersion) ?? optionalString(value.version),
    agentVersion: optionalString(value.agentVersion),
  }
}

function parsePendingAgentsResponse(body: unknown): PendingAgent[] {
  if (!isRecord(body) || !Array.isArray(body.clusters)) {
    throw new Error('Invalid pending agents response.')
  }

  return body.clusters.map(parsePendingAgent)
}

export async function getPendingAgents(signal?: AbortSignal): Promise<PendingAgent[]> {
  const body = await request('/agents/pending', { method: 'GET', signal })
  return parsePendingAgentsResponse(body)
}

export async function approvePendingAgent(id: string, signal?: AbortSignal): Promise<PendingAgent> {
  const body = await request(`/agents/${encodeURIComponent(id)}/approve`, {
    method: 'POST',
    headers: { 'X-Kfleet-CSRF': '1' },
    signal,
  })
  const approved = parsePendingAgent(body)
  if (approved.id !== id) {
    throw new Error('The server returned a different agent after approval.')
  }

  return approved
}
