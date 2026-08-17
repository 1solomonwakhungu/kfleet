import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { useAuth } from '../../auth/AuthContext'
import { adminApi } from '../../lib/adminApi'
import { RegistrationTokenCard } from './RegistrationTokenCard'

vi.mock('../../lib/adminApi', () => ({
  adminApi: {
    listUsers: vi.fn(),
    createUser: vi.fn(),
    updateUser: vi.fn(),
    deleteUser: vi.fn(),
    listAuditEvents: vi.fn(),
    rotateRegistrationToken: vi.fn(),
  },
}))
vi.mock('../../auth/AuthContext', () => ({
  useAuth: vi.fn(),
}))

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

describe('RegistrationTokenCard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    signIn('admin')
    vi.mocked(adminApi.rotateRegistrationToken).mockResolvedValue('new-token-value')
  })

  it('renders nothing for operators', () => {
    signIn('operator')

    const { container } = render(<RegistrationTokenCard />)

    expect(container.firstChild).toBeNull()
  })

  it('confirms before rotating and shows the new token once', async () => {
    render(<RegistrationTokenCard />)

    fireEvent.click(screen.getByText('Rotate token'))
    const dialog = await screen.findByRole('dialog')
    expect(dialog.textContent).toContain('stops working straight away')
    expect(adminApi.rotateRegistrationToken).not.toHaveBeenCalled()

    fireEvent.click(dialog.querySelectorAll('button')[1])

    await waitFor(() => expect(adminApi.rotateRegistrationToken).toHaveBeenCalled())
    expect(await screen.findByText('new-token-value')).toBeTruthy()
  })

  it('surfaces the hub error message when rotation fails', async () => {
    vi.mocked(adminApi.rotateRegistrationToken).mockRejectedValue(
      new Error('failed to rotate registration token'),
    )

    render(<RegistrationTokenCard />)
    fireEvent.click(screen.getByText('Rotate token'))
    const dialog = await screen.findByRole('dialog')
    fireEvent.click(dialog.querySelectorAll('button')[1])

    expect(await screen.findByText('failed to rotate registration token')).toBeTruthy()
  })
})
