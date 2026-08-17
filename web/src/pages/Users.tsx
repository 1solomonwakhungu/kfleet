import { useCallback, useMemo, useState, type FormEvent } from 'react'
import { LoaderCircle, RefreshCw, Trash2, UserPlus } from 'lucide-react'

import { useAuth } from '../auth/AuthContext'
import { ConfirmDialog } from '../components/admin/ConfirmDialog'
import { PermissionNotice } from '../components/admin/PermissionNotice'
import { Button } from '../components/ui/button'
import { Card, CardContent } from '../components/ui/card'
import { Input } from '../components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../components/ui/table'
import { useUsers } from '../hooks/useUsers'
import { adminApi } from '../lib/adminApi'
import { isAbortError, messageFrom } from '../lib/errors'
import { roleLabels, roleOptions, type Role, type UserAccount } from '../types/admin'

const minPasswordLength = 12
const maxPasswordLength = 72

function formatTimestamp(value: string): string {
  const parsed = new Date(value)
  return Number.isNaN(parsed.getTime()) ? '—' : parsed.toLocaleString()
}

export function UsersPage() {
  const { user } = useAuth()
  const isAdmin = user?.role === 'admin'
  const { users, loading, error, reload, replaceUser, removeUser } = useUsers(isAdmin)
  const [actionError, setActionError] = useState<string | null>(null)
  const [status, setStatus] = useState<string | null>(null)
  const [busyId, setBusyId] = useState<string | null>(null)
  const [pendingDelete, setPendingDelete] = useState<UserAccount | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  const sortedUsers = useMemo(
    () => [...users].sort((a, b) => a.username.localeCompare(b.username)),
    [users],
  )

  const mutate = useCallback(
    async (target: UserAccount, changes: { role?: Role; disabled?: boolean }) => {
      setBusyId(target.id)
      setActionError(null)
      setStatus(null)
      try {
        const updated = await adminApi.updateUser(target.id, {
          role: changes.role ?? target.role,
          disabled: changes.disabled ?? target.disabled,
        })
        replaceUser(updated)
        setStatus(`${updated.username} was updated.`)
      } catch (caught) {
        if (!isAbortError(caught)) setActionError(messageFrom(caught, 'The user could not be updated.'))
      } finally {
        setBusyId(null)
      }
    },
    [replaceUser],
  )

  const confirmDelete = useCallback(async () => {
    if (!pendingDelete) return
    setDeleting(true)
    setDeleteError(null)
    try {
      await adminApi.deleteUser(pendingDelete.id)
      removeUser(pendingDelete.id)
      setStatus(`${pendingDelete.username} was deleted.`)
      setPendingDelete(null)
    } catch (caught) {
      if (!isAbortError(caught)) setDeleteError(messageFrom(caught, 'The user could not be deleted.'))
    } finally {
      setDeleting(false)
    }
  }, [pendingDelete, removeUser])

  if (!isAdmin) {
    return (
      <main className="mx-auto min-h-dvh max-w-[100rem] px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
        <PageHeader loading={false} onRefresh={undefined} />
        <div className="mt-7">
          <PermissionNotice
            title="Admin access required"
            description="User management is restricted to admins. Ask an admin to change your role if you need access."
          />
        </div>
      </main>
    )
  }

  return (
    <main className="mx-auto min-h-dvh max-w-[100rem] px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <PageHeader loading={loading} onRefresh={() => void reload()} />

      <div className="mt-6 space-y-4" aria-live="polite">
        {error && (
          <section
            className="flex flex-col gap-3 rounded-lg bg-danger-soft p-4 text-danger sm:flex-row sm:items-center sm:justify-between"
            role="alert"
          >
            <div>
              <p className="font-semibold">Users could not be loaded.</p>
              <p className="mt-1 text-sm">{error}</p>
            </div>
            <Button variant="outline" size="sm" disabled={loading} onClick={() => void reload()}>
              Retry
            </Button>
          </section>
        )}
        {actionError && (
          <p className="rounded-lg bg-danger-soft p-4 text-sm text-danger" role="alert">
            {actionError}
          </p>
        )}
        {status && (
          <p
            className="rounded-lg bg-blue-950 p-4 text-sm text-blue-100 ring-1 ring-inset ring-blue-800"
            role="status"
          >
            {status}
          </p>
        )}
      </div>

      <div className="mt-7 grid gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <section aria-busy={loading} aria-labelledby="user-list-title">
          <div className="mb-4 flex items-center justify-between gap-4">
            <h2 id="user-list-title" className="font-display text-lg font-bold">
              Accounts
            </h2>
            {!loading && (
              <span className="font-mono text-sm text-muted">
                {sortedUsers.length} {sortedUsers.length === 1 ? 'account' : 'accounts'}
              </span>
            )}
          </div>

          {loading && sortedUsers.length === 0 ? (
            <UsersSkeleton />
          ) : sortedUsers.length === 0 && !error ? (
            <Card className="ring-1 ring-inset ring-border">
              <CardContent className="grid min-h-48 place-items-center p-6 text-center">
                <div>
                  <p className="font-display text-xl font-bold">No accounts yet</p>
                  <p className="mt-2 text-muted">Invite a teammate to give them hub access.</p>
                </div>
              </CardContent>
            </Card>
          ) : sortedUsers.length > 0 ? (
            <Card className="ring-1 ring-inset ring-border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead scope="col">User</TableHead>
                    <TableHead scope="col">Role</TableHead>
                    <TableHead scope="col">Status</TableHead>
                    <TableHead scope="col">Created</TableHead>
                    <TableHead scope="col" className="text-right">
                      Actions
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {sortedUsers.map((account) => {
                    const isSelf = account.id === user?.id
                    const busy = busyId === account.id
                    return (
                      <TableRow key={account.id}>
                        <TableCell>
                          <span className="block font-semibold">{account.username}</span>
                          <span className="block text-xs text-muted">{account.email}</span>
                        </TableCell>
                        <TableCell>
                          <label className="sr-only" htmlFor={`role-${account.id}`}>
                            Role for {account.username}
                          </label>
                          <select
                            id={`role-${account.id}`}
                            className="h-9 rounded-md border border-border bg-background px-2 text-sm text-foreground disabled:cursor-not-allowed disabled:opacity-50"
                            value={account.role}
                            disabled={busy}
                            onChange={(event) =>
                              void mutate(account, { role: event.target.value as Role })
                            }
                          >
                            {roleOptions.map((role) => (
                              <option key={role} value={role}>
                                {roleLabels[role]}
                              </option>
                            ))}
                          </select>
                        </TableCell>
                        <TableCell>
                          <span className={account.disabled ? 'text-muted' : 'text-healthy'}>
                            {account.disabled ? 'Deactivated' : 'Active'}
                          </span>
                        </TableCell>
                        <TableCell className="whitespace-nowrap text-muted">
                          {formatTimestamp(account.createdAt)}
                        </TableCell>
                        <TableCell>
                          <div className="flex justify-end gap-2">
                            <Button
                              variant="outline"
                              size="sm"
                              disabled={busy || isSelf}
                              title={isSelf ? 'You cannot deactivate your own account.' : undefined}
                              onClick={() => void mutate(account, { disabled: !account.disabled })}
                            >
                              {account.disabled ? 'Reactivate' : 'Deactivate'}
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="text-danger hover:text-danger"
                              disabled={busy || isSelf}
                              title={isSelf ? 'You cannot delete your own account.' : undefined}
                              aria-label={`Delete ${account.username}`}
                              onClick={() => {
                                setDeleteError(null)
                                setPendingDelete(account)
                              }}
                            >
                              <Trash2 className="size-4" aria-hidden="true" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    )
                  })}
                </TableBody>
              </Table>
            </Card>
          ) : null}
        </section>

        <InviteUserForm
          onCreated={(created) => {
            replaceUser(created)
            setStatus(`${created.username} was created with the ${roleLabels[created.role]} role.`)
          }}
        />
      </div>

      <ConfirmDialog
        open={pendingDelete !== null}
        title={`Delete ${pendingDelete?.username ?? 'user'}?`}
        confirmLabel="Delete user"
        pending={deleting}
        error={deleteError}
        onCancel={() => setPendingDelete(null)}
        onConfirm={() => void confirmDelete()}
      >
        <p>
          {pendingDelete?.username} loses hub access immediately and any active sessions are invalidated.
          Their past actions remain in the audit log.
        </p>
        <p>This cannot be undone. Deactivate the account instead if you may need it later.</p>
      </ConfirmDialog>
    </main>
  )
}

interface PageHeaderProps {
  loading: boolean
  onRefresh?: () => void
}

function PageHeader({ loading, onRefresh }: PageHeaderProps) {
  return (
    <header className="flex flex-col gap-5 border-b border-border pb-7 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <p className="font-mono text-sm text-blue-400">kfleet admin</p>
        <h1 className="mt-2 font-display text-3xl font-bold tracking-tight sm:text-4xl">Users</h1>
        <p className="mt-2 max-w-2xl text-muted">
          Manage who can sign in to the hub and what each account is allowed to do.
        </p>
      </div>
      {onRefresh && (
        <Button variant="outline" size="sm" disabled={loading} onClick={onRefresh}>
          {loading ? (
            <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />
          ) : (
            <RefreshCw className="size-4" aria-hidden="true" />
          )}
          {loading ? 'Refreshing…' : 'Refresh'}
        </Button>
      )}
    </header>
  )
}

function InviteUserForm({ onCreated }: { onCreated: (user: UserAccount) => void }) {
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<Role>('read_only')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const submit = useCallback(
    async (event: FormEvent<HTMLFormElement>) => {
      event.preventDefault()
      setError(null)

      if (!username.trim() || !email.trim()) {
        setError('Username and email are required.')
        return
      }
      if (password.length < minPasswordLength || password.length > maxPasswordLength) {
        setError(`Password must be between ${minPasswordLength} and ${maxPasswordLength} characters.`)
        return
      }

      setSubmitting(true)
      try {
        const created = await adminApi.createUser({
          username: username.trim(),
          email: email.trim(),
          password,
          role,
        })
        onCreated(created)
        setUsername('')
        setEmail('')
        setPassword('')
        setRole('read_only')
      } catch (caught) {
        if (!isAbortError(caught)) setError(messageFrom(caught, 'The user could not be created.'))
      } finally {
        setSubmitting(false)
      }
    },
    [username, email, password, role, onCreated],
  )

  return (
    <Card className="h-fit p-5 ring-1 ring-inset ring-border">
      <h2 className="flex items-center gap-2 font-display text-lg font-bold">
        <UserPlus className="size-5 text-muted" aria-hidden="true" />
        Invite a user
      </h2>
      <p className="mt-2 text-sm text-muted">
        The hub creates the account immediately. Share the initial password over a secure channel.
      </p>

      <form className="mt-4 space-y-4" onSubmit={(event) => void submit(event)} aria-label="Invite a user">
        <div>
          <label className="block text-sm font-semibold" htmlFor="new-username">
            Username
          </label>
          <Input
            id="new-username"
            className="mt-1"
            value={username}
            autoComplete="off"
            onChange={(event) => setUsername(event.target.value)}
          />
        </div>
        <div>
          <label className="block text-sm font-semibold" htmlFor="new-email">
            Email
          </label>
          <Input
            id="new-email"
            className="mt-1"
            type="email"
            value={email}
            autoComplete="off"
            onChange={(event) => setEmail(event.target.value)}
          />
        </div>
        <div>
          <label className="block text-sm font-semibold" htmlFor="new-password">
            Initial password
          </label>
          <Input
            id="new-password"
            className="mt-1"
            type="password"
            value={password}
            autoComplete="new-password"
            onChange={(event) => setPassword(event.target.value)}
          />
          <p className="mt-1 text-xs text-muted">
            {minPasswordLength}–{maxPasswordLength} characters.
          </p>
        </div>
        <div>
          <label className="block text-sm font-semibold" htmlFor="new-role">
            Role
          </label>
          <select
            id="new-role"
            className="mt-1 h-11 w-full rounded-md border border-border bg-background px-3 text-sm text-foreground"
            value={role}
            onChange={(event) => setRole(event.target.value as Role)}
          >
            {roleOptions.map((option) => (
              <option key={option} value={option}>
                {roleLabels[option]}
              </option>
            ))}
          </select>
        </div>

        {error && (
          <p className="rounded-md bg-danger-soft p-3 text-sm text-danger" role="alert">
            {error}
          </p>
        )}

        <Button type="submit" size="sm" disabled={submitting} className="w-full">
          {submitting ? 'Creating…' : 'Create user'}
        </Button>
      </form>
    </Card>
  )
}

function UsersSkeleton() {
  return (
    <Card className="animate-pulse p-5 ring-1 ring-inset ring-border" aria-label="Loading users">
      <div className="h-5 w-40 rounded bg-elevated" />
      {Array.from({ length: 3 }, (_, index) => (
        <div key={index} className="mt-5 flex items-center justify-between gap-6 border-t border-border pt-5">
          <div className="h-10 w-1/3 rounded bg-elevated" />
          <div className="h-9 w-24 rounded bg-elevated" />
        </div>
      ))}
    </Card>
  )
}

export default UsersPage
