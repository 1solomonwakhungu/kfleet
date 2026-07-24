import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { AuthProvider } from './auth/AuthContext'

vi.mock('./pages/Dashboard', () => ({ Dashboard: () => <main><h1>Fleet dashboard</h1></main> }))
vi.mock('./pages/ClusterDetail', () => ({ default: () => <main><h1>Cluster detail</h1></main> }))
vi.mock('./pages/PendingAgents', () => ({ default: () => <main><h1>Pending agents page</h1></main> }))
vi.mock('./pages/Alerts', () => ({ default: () => <main><h1>Alerts page</h1></main> }))
vi.mock('./pages/PolicyDashboard', () => ({ default: () => <main><h1>Policy dashboard page</h1></main> }))

describe('App routing', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({
      id: 'admin-1',
      username: 'admin',
      email: 'admin@example.com',
      role: 'admin',
      disabled: false,
      createdAt: '2026-07-23T00:00:00Z',
      updatedAt: '2026-07-23T00:00:00Z',
    }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    })))
  })

  it('renders pending agents at /agents and marks its navigation item active', async () => {
    render(
      <MemoryRouter initialEntries={['/agents']}>
        <AuthProvider>
          <App />
        </AuthProvider>
      </MemoryRouter>,
    )

    expect(await screen.findByRole('heading', { name: 'Pending agents page' })).toBeTruthy()
    expect(screen.getByRole('link', { name: /Agents/ }).getAttribute('aria-current')).toBe('page')
    expect(screen.getByRole('link', { name: 'Fleet Cluster overview' }).getAttribute('aria-current')).toBeNull()
    expect(screen.getAllByRole('main')).toHaveLength(1)
  })

  it('renders fleet alerts at /alerts and marks its navigation item active', async () => {
    render(
      <MemoryRouter initialEntries={['/alerts']}>
        <AuthProvider>
          <App />
        </AuthProvider>
      </MemoryRouter>,
    )

    expect(await screen.findByRole('heading', { name: 'Alerts page' })).toBeTruthy()
    expect(screen.getByRole('link', { name: /Alerts/ }).getAttribute('aria-current')).toBe('page')
  })

  it('renders policy results at /policies and marks its navigation item active', async () => {
    render(
      <MemoryRouter initialEntries={['/policies']}>
        <AuthProvider>
          <App />
        </AuthProvider>
      </MemoryRouter>,
    )

    expect(await screen.findByRole('heading', { name: 'Policy dashboard page' })).toBeTruthy()
    expect(screen.getByRole('link', { name: /Policy/ }).getAttribute('aria-current')).toBe('page')
    expect(screen.getByRole('link', { name: /^Fleet/ }).getAttribute('aria-current')).toBeNull()
  })

  it('redirects an unauthenticated protected route to sign in', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('', { status: 401 })))

    render(
      <MemoryRouter initialEntries={['/agents']}>
        <AuthProvider>
          <App />
        </AuthProvider>
      </MemoryRouter>,
    )

    expect(await screen.findByRole('heading', { name: 'Sign in' })).toBeTruthy()
    expect(screen.queryByRole('heading', { name: 'Pending agents page' })).toBeNull()
  })
})
