import { fireEvent, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { useAuth } from '../auth/AuthContext'
import { adminApi } from '../lib/adminApi'
import type { AuditEvent } from '../types/admin'
import AuditLogPage from './AuditLog'

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

const events: AuditEvent[] = [
  {
    id: 'audit-1',
    occurredAt: '2026-07-23T12:00:00Z',
    actorUserId: 'admin-1',
    actorUsername: 'admin',
    actorRole: 'admin',
    action: 'user.create',
    targetType: 'user',
    targetId: 'user-9',
    outcome: 'success',
    details: 'role=operator',
    sourceIp: '10.0.0.4',
  },
  {
    id: 'audit-2',
    occurredAt: '2026-07-23T11:00:00Z',
    actorUsername: 'viewer',
    actorRole: 'read_only',
    action: 'authorization.role_denied',
    targetType: 'http_route',
    targetId: 'GET /api/v1/users',
    outcome: 'failure',
    details: 'required_role=admin',
  },
]

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

describe('AuditLogPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    signIn('admin')
    vi.mocked(adminApi.listAuditEvents).mockResolvedValue(events)
  })

  it('renders audit entries for admins', async () => {
    render(<AuditLogPage />)

    expect(await screen.findByText('user.create')).toBeTruthy()
    expect(screen.getByText('authorization.role_denied')).toBeTruthy()
    expect(screen.getByText('role=operator')).toBeTruthy()
    expect(screen.getByText('10.0.0.4')).toBeTruthy()
    expect(screen.getByText('success')).toBeTruthy()
    expect(screen.getByText('failure')).toBeTruthy()
    expect(adminApi.listAuditEvents).toHaveBeenCalledWith(100, expect.anything())
  })

  it('blocks non-admins without calling the audit endpoint', async () => {
    signIn('operator')

    render(<AuditLogPage />)

    expect(await screen.findByText('Admin access required')).toBeTruthy()
    expect(adminApi.listAuditEvents).not.toHaveBeenCalled()
  })

  it('filters by free text and outcome', async () => {
    render(<AuditLogPage />)
    await screen.findByText('user.create')

    fireEvent.change(screen.getByLabelText('Filter audit events'), { target: { value: 'role_denied' } })
    expect(screen.queryByText('user.create')).toBeNull()
    expect(screen.getByText('authorization.role_denied')).toBeTruthy()

    fireEvent.change(screen.getByLabelText('Filter audit events'), { target: { value: '' } })
    fireEvent.change(screen.getByLabelText('Outcome'), { target: { value: 'success' } })
    expect(screen.getByText('user.create')).toBeTruthy()
    expect(screen.queryByText('authorization.role_denied')).toBeNull()
  })

  it('shows an empty state when the hub has no audit history', async () => {
    vi.mocked(adminApi.listAuditEvents).mockResolvedValue([])

    render(<AuditLogPage />)

    expect(await screen.findByText('No audit events yet')).toBeTruthy()
  })

  it('shows the hub error message when the request fails', async () => {
    vi.mocked(adminApi.listAuditEvents).mockRejectedValue(new Error('failed to list audit events'))

    render(<AuditLogPage />)

    expect(await screen.findByText('failed to list audit events')).toBeTruthy()
  })
})
