import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { api } from '../../lib/api'
import { ApplicationShell } from './ApplicationShell'

vi.mock('../../auth/AuthContext', () => ({
  useAuth: () => ({
    user: { username: 'public-demo', role: 'read_only' },
    logout: vi.fn(),
  }),
}))

afterEach(() => {
  vi.restoreAllMocks()
})

describe('ApplicationShell public demo state', () => {
  it('shows the synthetic read-only notice and removes mutation navigation', async () => {
    vi.spyOn(api, 'getRuntimeInfo').mockResolvedValue({
      demoMode: true,
      readOnly: true,
      syntheticData: true,
      dataPolicy: 'Synthetic sample data only.',
    })

    render(
      <MemoryRouter>
        <Routes>
          <Route element={<ApplicationShell />}>
            <Route index element={<main>Demo dashboard</main>} />
          </Route>
        </Routes>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByLabelText('Public demo safety notice')).toBeTruthy()
    })
    expect(screen.getByText(/Mutating API requests are disabled/)).toBeTruthy()
    expect(screen.queryByRole('link', { name: /Agents/ })).toBeNull()
    expect(screen.getByRole('link', { name: 'Fleet Cluster overview' })).toBeTruthy()
    expect(screen.queryByRole('button', { name: 'Sign out' })).toBeNull()
  })
})
