import type { Role } from '../lib/authApi'

export type { Role }

/** Mirrors api.UserResponse in pkg/api/api.go. */
export interface UserAccount {
  id: string
  username: string
  email: string
  role: Role
  disabled: boolean
  createdAt: string
  updatedAt: string
}

export type AuditOutcome = 'success' | 'failure'

/** Mirrors types.AuditEvent in pkg/types/types.go. */
export interface AuditEvent {
  id: string
  occurredAt: string
  actorUserId?: string
  actorUsername: string
  actorRole?: Role
  action: string
  targetType: string
  targetId: string
  outcome: AuditOutcome
  details?: string
  sourceIp?: string
}

export interface CreateUserInput {
  username: string
  email: string
  password: string
  role: Role
}

export interface UpdateUserInput {
  role: Role
  disabled: boolean
}

export const roleLabels: Record<Role, string> = {
  admin: 'Admin',
  operator: 'Operator',
  read_only: 'Read only',
}

/** Ordered from least to most privileged, matching pkg/types. */
export const roleOptions: readonly Role[] = ['read_only', 'operator', 'admin']
