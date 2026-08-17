import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { useAuth } from '../../auth/AuthContext'
import { api } from '../../lib/api'
import { RemoveClusterCard } from './RemoveClusterCard'

const navigate = vi.fn()

vi.mock('../../lib/api', () => ({
  api: {
    deleteCluster: vi.fn(),
  },
}))
vi.mock('../../auth/AuthContext', () => ({
  useAuth: vi.fn(),
}))
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom')
  return { ...actual, useNavigate: () => navigate }
})

function signIn(role: 'admin' | 'operator' | 'read_only') {
  vi.mocked(useAuth).mockReturnValue({
    user: {
      id: 'user-1',
      username: 'someone',
      email: 'someone@example.com',
      role,
      disabled: false,
      createdAt: '2026-07-23T00:00:00Z',
      updatedAt: '2026-07-23T00:00:00Z',
    },
    loading: false,
    login: vi.fn(),
    logout: vi.fn(),
  })
}

function renderCard() {
  return render(
    <MemoryRouter>
      <RemoveClusterCard clusterId="cluster-a" clusterName="production" />
    </MemoryRouter>,
  )
}

describe('RemoveClusterCard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    signIn('operator')
    vi.mocked(api.deleteCluster).mockResolvedValue(undefined)
  })

  it('disables removal for read-only users', () => {
    signIn('read_only')

    renderCard()

    expect(screen.getByText('Remove cluster', { selector: 'button' }).hasAttribute('disabled')).toBe(true)
    expect(screen.getByText('Removing a cluster requires the operator or admin role.')).toBeTruthy()
  })

  it('requires confirmation and states the consequences before deleting', async () => {
    renderCard()

    fireEvent.click(screen.getByText('Remove cluster', { selector: 'button' }))

    const dialog = await screen.findByRole('dialog')
    expect(dialog.textContent).toContain('Remove production?')
    expect(dialog.textContent).toContain('alert history')
    expect(dialog.textContent).toContain('This cannot be undone.')
    expect(api.deleteCluster).not.toHaveBeenCalled()

    fireEvent.click(screen.getByText('Cancel'))
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull())
    expect(api.deleteCluster).not.toHaveBeenCalled()
  })

  it('deletes the cluster and returns to the fleet view once confirmed', async () => {
    renderCard()

    fireEvent.click(screen.getByText('Remove cluster', { selector: 'button' }))
    const dialog = await screen.findByRole('dialog')
    fireEvent.click(dialog.querySelectorAll('button')[1])

    await waitFor(() => expect(api.deleteCluster).toHaveBeenCalledWith('cluster-a'))
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/', { replace: true }))
  })

  it('shows the hub error message when removal fails', async () => {
    vi.mocked(api.deleteCluster).mockRejectedValue(new Error('this action requires a higher role'))

    renderCard()
    fireEvent.click(screen.getByText('Remove cluster', { selector: 'button' }))
    const dialog = await screen.findByRole('dialog')
    fireEvent.click(dialog.querySelectorAll('button')[1])

    expect(await screen.findByText('this action requires a higher role')).toBeTruthy()
    expect(navigate).not.toHaveBeenCalled()
  })
})
