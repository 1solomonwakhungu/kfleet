import { request } from './api'
import type { AuditEvent, CreateUserInput, UpdateUserInput, UserAccount } from '../types/admin'

/**
 * Admin-only endpoints. Every route below is registered behind
 * `s.requireRole(types.RoleAdmin, ...)` in internal/server, so the hub answers
 * 403 "this action requires a higher role" for operators and read-only users.
 */
export const adminApi = {
  // GET /api/v1/users (internal/server/handlers_users.go).
  listUsers: (signal?: AbortSignal) =>
    request<{ users: UserAccount[] }>('GET', '/users', undefined, signal).then((response) => response.users ?? []),
  // POST /api/v1/users responds 201, or 409 when the username or email is taken.
  createUser: (input: CreateUserInput, signal?: AbortSignal) =>
    request<UserAccount>('POST', '/users', input, signal),
  // PATCH /api/v1/users/{id} responds 409 when the change would remove the last enabled admin.
  updateUser: (id: string, input: UpdateUserInput, signal?: AbortSignal) =>
    request<UserAccount>('PATCH', `/users/${encodeURIComponent(id)}`, input, signal),
  // DELETE /api/v1/users/{id} responds 204, or 409 for self-deletion and last-admin removal.
  deleteUser: (id: string, signal?: AbortSignal) =>
    request<void>('DELETE', `/users/${encodeURIComponent(id)}`, undefined, signal),
  // GET /api/v1/audit?limit=N (limit must be 1..1000, newest first).
  listAuditEvents: (limit: number, signal?: AbortSignal) =>
    request<{ events: AuditEvent[] }>('GET', `/audit?limit=${limit}`, undefined, signal).then(
      (response) => response.events ?? [],
    ),
  // POST /api/v1/admin/registration-token/rotate returns the raw token exactly once.
  rotateRegistrationToken: (signal?: AbortSignal) =>
    request<{ token: string }>('POST', '/admin/registration-token/rotate', {}, signal).then(
      (response) => response.token,
    ),
}
