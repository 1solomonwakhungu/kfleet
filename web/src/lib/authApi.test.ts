import { beforeEach, describe, expect, it, vi } from 'vitest'

import { authApi } from './authApi'

const user = {
  id: 'user-1',
  username: 'reader',
  email: 'reader@example.com',
  role: 'read_only' as const,
  disabled: false,
  createdAt: '2026-07-23T00:00:00Z',
  updatedAt: '2026-07-23T00:00:00Z',
}

describe('authApi', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
  })

  it('logs in with JSON credentials and returns the authenticated user', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify(user), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    )
    vi.stubGlobal('fetch', fetchMock)

    await expect(authApi.login('reader', 'password')).resolves.toEqual(user)
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/auth/login', expect.objectContaining({
      method: 'POST',
      body: JSON.stringify({ username: 'reader', password: 'password' }),
    }))
  })

  it('treats an unauthorized current-user response as signed out', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('', { status: 401 })))
    await expect(authApi.currentUser()).resolves.toBeNull()
  })

  it('sends the CSRF header when logging out', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }))
    vi.stubGlobal('fetch', fetchMock)

    await authApi.logout()
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/auth/logout', expect.objectContaining({
      method: 'POST',
      headers: { 'X-Kfleet-CSRF': '1' },
    }))
  })
})
