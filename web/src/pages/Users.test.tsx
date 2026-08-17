import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { useAuth } from '../auth/AuthContext'
import { adminApi } from '../lib/adminApi'
import type { Role, UserAccount } from '../types/admin'
import UsersPage from './Users'

vi.mock('../lib/adminApi', () => ({
  adminApi: {
    listUsers: vi.fn(),
    createUser: vi.fn(),
    updateUser: vi.fn(),
    deleteUser: vi.fn(),
    listAuditEvents: vi.fn(),
    rotateRegistrationToken: vi.fn(),
  },
}))
vi.mock('../auth/AuthContext', () => ({
  useAuth: vi.fn(),
}))

const admin: UserAccount = {
  id: 'admin-1',
  username: 'admin',
  email: 'admin@example.com',
  role: 'admin',
  disabled: false,
  createdAt: '2026-07-23T00:00:00Z',
  updatedAt: '2026-07-23T00:00:00Z',
}

const viewer: UserAccount = {
  id: 'viewer-1',
  username: 'viewer',
  email: 'viewer@example.com',
  role: 'read_only',
  disabled: false,
  createdAt: '2026-07-23T00:00:00Z',
  updatedAt: '2026-07-23T00:00:00Z',
}

function signIn(role: Role, id = role === 'admin' ? admin.id : viewer.id) {
  vi.mocked(useAuth).mockReturnValue({
    user: { ...(role === 'admin' ? admin : viewer), id, role },
    loading: false,
    login: vi.fn(),
    logout: vi.fn(),
  })
}

describe('UsersPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    signIn('admin')
    vi.mocked(adminApi.listUsers).mockResolvedValue([admin, viewer])
  })

  it('hides user management from non-admins and never calls the admin API', async () => {
    signIn('operator')

    render(<UsersPage />)

    expect(await screen.findByText('Admin access required')).toBeTruthy()
    expect(screen.queryByLabelText('Invite a user')).toBeNull()
    expect(adminApi.listUsers).not.toHaveBeenCalled()
  })

  it('lists accounts for admins', async () => {
    render(<UsersPage />)

    expect(await screen.findByText('viewer@example.com')).toBeTruthy()
    expect(screen.getByLabelText('Invite a user')).toBeTruthy()
    expect(screen.getAllByText('Active').length).toBe(2)
  })

  it('surfaces the hub error message when loading fails', async () => {
    vi.mocked(adminApi.listUsers).mockRejectedValue(new Error('this action requires a higher role'))

    render(<UsersPage />)

    expect(await screen.findByText('this action requires a higher role')).toBeTruthy()
  })

  it('changes a role through the API and shows the result', async () => {
    vi.mocked(adminApi.updateUser).mockResolvedValue({ ...viewer, role: 'operator' })

    render(<UsersPage />)
    const select = await screen.findByLabelText('Role for viewer')
    fireEvent.change(select, { target: { value: 'operator' } })

    await waitFor(() =>
      expect(adminApi.updateUser).toHaveBeenCalledWith('viewer-1', { role: 'operator', disabled: false }),
    )
    expect(await screen.findByText('viewer was updated.')).toBeTruthy()
  })

  it('requires confirmation before deleting a user', async () => {
    vi.mocked(adminApi.deleteUser).mockResolvedValue(undefined)

    render(<UsersPage />)
    fireEvent.click(await screen.findByLabelText('Delete viewer'))

    const dialog = await screen.findByRole('dialog')
    expect(dialog.textContent).toContain('This cannot be undone')
    expect(adminApi.deleteUser).not.toHaveBeenCalled()

    fireEvent.click(screen.getByText('Cancel'))
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull())
    expect(adminApi.deleteUser).not.toHaveBeenCalled()

    fireEvent.click(screen.getByLabelText('Delete viewer'))
    fireEvent.click(await screen.findByText('Delete user'))

    await waitFor(() => expect(adminApi.deleteUser).toHaveBeenCalledWith('viewer-1'))
    expect(await screen.findByText('viewer was deleted.')).toBeTruthy()
  })

  it('reports the hub error when a deletion is rejected', async () => {
    vi.mocked(adminApi.deleteUser).mockRejectedValue(new Error('at least one enabled admin is required'))

    render(<UsersPage />)
    fireEvent.click(await screen.findByLabelText('Delete viewer'))
    fireEvent.click(await screen.findByText('Delete user'))

    expect(await screen.findByText('at least one enabled admin is required')).toBeTruthy()
  })

  it('validates the initial password before calling the hub', async () => {
    render(<UsersPage />)

    fireEvent.change(await screen.findByLabelText('Username'), { target: { value: 'newbie' } })
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'newbie@example.com' } })
    fireEvent.change(screen.getByLabelText('Initial password'), { target: { value: 'short' } })
    fireEvent.click(screen.getByText('Create user'))

    expect(await screen.findByText('Password must be between 12 and 72 characters.')).toBeTruthy()
    expect(adminApi.createUser).not.toHaveBeenCalled()
  })

  it('creates a user with the selected role', async () => {
    const created: UserAccount = {
      ...viewer,
      id: 'new-1',
      username: 'newbie',
      email: 'newbie@example.com',
      role: 'operator',
    }
    vi.mocked(adminApi.createUser).mockResolvedValue(created)

    render(<UsersPage />)
    fireEvent.change(await screen.findByLabelText('Username'), { target: { value: 'newbie' } })
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'newbie@example.com' } })
    fireEvent.change(screen.getByLabelText('Initial password'), { target: { value: 'correct-horse-battery' } })
    fireEvent.change(screen.getByLabelText('Role'), { target: { value: 'operator' } })
    fireEvent.click(screen.getByText('Create user'))

    await waitFor(() =>
      expect(adminApi.createUser).toHaveBeenCalledWith({
        username: 'newbie',
        email: 'newbie@example.com',
        password: 'correct-horse-battery',
        role: 'operator',
      }),
    )
    expect(await screen.findByText('newbie was created with the Operator role.')).toBeTruthy()
  })

  it('does not allow admins to deactivate or delete their own account', async () => {
    render(<UsersPage />)

    await screen.findByText('admin@example.com')
    expect(screen.getByLabelText('Delete admin').hasAttribute('disabled')).toBe(true)
    expect(screen.getByLabelText('Delete viewer').hasAttribute('disabled')).toBe(false)
  })
})
