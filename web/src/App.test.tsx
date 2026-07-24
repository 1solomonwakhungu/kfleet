import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'

import { App } from './App'

vi.mock('./pages/Dashboard', () => ({ Dashboard: () => <main><h1>Fleet dashboard</h1></main> }))
vi.mock('./pages/ClusterDetail', () => ({ default: () => <main><h1>Cluster detail</h1></main> }))
vi.mock('./pages/PendingAgents', () => ({ default: () => <main><h1>Pending agents page</h1></main> }))
vi.mock('./pages/PolicyDashboard', () => ({ default: () => <main><h1>Policy dashboard page</h1></main> }))

describe('App routing', () => {
  it('renders pending agents at /agents and marks its navigation item active', () => {
    render(
      <MemoryRouter initialEntries={['/agents']}>
        <App />
      </MemoryRouter>,
    )

    expect(screen.getByRole('heading', { name: 'Pending agents page' })).toBeTruthy()
    expect(screen.getByRole('link', { name: /Agents/ }).getAttribute('aria-current')).toBe('page')
    expect(screen.getByRole('link', { name: /Fleet/ }).getAttribute('aria-current')).toBeNull()
    expect(screen.getAllByRole('main')).toHaveLength(1)
  })

  it('renders policy results at /policies and marks its navigation item active', () => {
    render(
      <MemoryRouter initialEntries={['/policies']}>
        <App />
      </MemoryRouter>,
    )

    expect(screen.getByRole('heading', { name: 'Policy dashboard page' })).toBeTruthy()
    expect(screen.getByRole('link', { name: /Policy/ }).getAttribute('aria-current')).toBe('page')
    expect(screen.getByRole('link', { name: /Fleet/ }).getAttribute('aria-current')).toBeNull()
  })
})
