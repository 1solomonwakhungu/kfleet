const AUTH_BASE = '/api/v1/auth'

export type Role = 'admin' | 'operator' | 'read_only'

export interface AuthenticatedUser {
  id: string
  username: string
  email: string
  role: Role
  disabled: boolean
  createdAt: string
  updatedAt: string
}

export class AuthenticationError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'AuthenticationError'
    this.status = status
  }
}

function isUser(value: unknown): value is AuthenticatedUser {
  if (typeof value !== 'object' || value === null) return false
  const user = value as Partial<AuthenticatedUser>
  return (
    typeof user.id === 'string' &&
    typeof user.username === 'string' &&
    typeof user.email === 'string' &&
    (user.role === 'admin' || user.role === 'operator' || user.role === 'read_only') &&
    typeof user.disabled === 'boolean'
  )
}

async function responseBody(response: Response): Promise<unknown> {
  const text = await response.text()
  if (!text) return undefined
  try {
    return JSON.parse(text) as unknown
  } catch {
    return undefined
  }
}

function errorMessage(body: unknown, fallback: string): string {
  if (
    typeof body === 'object' &&
    body !== null &&
    typeof (body as { error?: unknown }).error === 'string'
  ) {
    return (body as { error: string }).error
  }
  return fallback
}

async function userResponse(response: Response): Promise<AuthenticatedUser> {
  const body = await responseBody(response)
  if (!response.ok) {
    throw new AuthenticationError(
      response.status,
      errorMessage(body, 'Authentication request failed.'),
    )
  }
  if (!isUser(body)) {
    throw new AuthenticationError(response.status, 'The hub returned an invalid user response.')
  }
  return body
}

export const authApi = {
  currentUser: async (signal?: AbortSignal): Promise<AuthenticatedUser | null> => {
    const response = await fetch(`${AUTH_BASE}/me`, {
      headers: { Accept: 'application/json' },
      signal,
    })
    if (response.status === 401) return null
    return userResponse(response)
  },

  login: async (username: string, password: string): Promise<AuthenticatedUser> => {
    const response = await fetch(`${AUTH_BASE}/login`, {
      method: 'POST',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    })
    return userResponse(response)
  },

  logout: async (): Promise<void> => {
    const response = await fetch(`${AUTH_BASE}/logout`, {
      method: 'POST',
      headers: { 'X-Kfleet-CSRF': '1' },
    })
    if (!response.ok && response.status !== 401) {
      const body = await responseBody(response)
      throw new AuthenticationError(response.status, errorMessage(body, 'Sign out failed.'))
    }
  },
}

export function notifyAuthenticationRequired(response: Response): void {
  if (response.status === 401) {
    window.dispatchEvent(new Event('kfleet:authentication-required'))
  }
}
