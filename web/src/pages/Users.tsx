import { useCallback, useMemo, useState, type FormEvent } from 'react'
import { Button, Flash, FormControl, Heading, IconButton, Select, Text, TextInput } from '@primer/react'
import { Blankslate, SkeletonText } from '@primer/react/experimental'
import { CopyIcon, KeyIcon, PersonAddIcon, SyncIcon, TrashIcon } from '@primer/octicons-react'

import { useAuth } from '../auth/AuthContext'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { ConfirmDialog } from '../components/admin/ConfirmDialog'
import { PermissionNotice } from '../components/admin/PermissionNotice'
import { useUsers } from '../hooks/useUsers'
import { adminApi } from '../lib/adminApi'
import { isAbortError, messageFrom } from '../lib/errors'
import { roleLabels, roleOptions, type Role, type UserAccount } from '../types/admin'
import layout from '../styles/layout.module.css'
import styles from './Users.module.css'

const minPasswordLength = 12
const maxPasswordLength = 72

const passwordAlphabet = 'abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#%^*_-+='
const generatedPasswordLength = 24

/**
 * Generates the random password handed to the hub on reset. The plaintext is
 * only ever held in component state and shown once. Requires a CSP and
 * browser that expose crypto.getRandomValues; without it the page cannot
 * generate a password worth trusting, so it refuses rather than falling
 * back to a weak source.
 */
function generatePassword(): string {
  if (!globalThis.crypto?.getRandomValues) {
    throw new Error('This browser cannot generate a secure password. Please use a modern browser.')
  }
  const bytes = new Uint8Array(generatedPasswordLength)
  globalThis.crypto.getRandomValues(bytes)
  return Array.from(bytes, (byte) => passwordAlphabet[byte % passwordAlphabet.length]).join('')
}

function formatTimestamp(value: string): string {
  const parsed = new Date(value)
  return Number.isNaN(parsed.getTime()) ? '—' : parsed.toLocaleString()
}

export function UsersPage() {
  useDocumentTitle('Users · Admin · kfleet')
  const { user } = useAuth()
  const isAdmin = user?.role === 'admin'
  const { users, loading, error, reload, replaceUser, removeUser } = useUsers(isAdmin)
  const [actionError, setActionError] = useState<string | null>(null)
  const [status, setStatus] = useState<string | null>(null)
  const [busyId, setBusyId] = useState<string | null>(null)
  const [pendingDelete, setPendingDelete] = useState<UserAccount | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [pendingReset, setPendingReset] = useState<UserAccount | null>(null)
  const [resetting, setResetting] = useState(false)
  const [resetError, setResetError] = useState<string | null>(null)
  const [resetPassword, setResetPassword] = useState<{ username: string; password: string } | null>(null)
  const [copied, setCopied] = useState(false)

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

  const confirmReset = useCallback(async () => {
    if (!pendingReset) return
    setResetting(true)
    setResetError(null)
    try {
      const newPassword = generatePassword()
      const input = { newPassword }
      await adminApi.resetUserPassword(pendingReset.id, input)
      setResetPassword({ username: pendingReset.username, password: newPassword })
      setCopied(false)
      setPendingReset(null)
      setStatus(null)
      await reload()
    } catch (caught) {
      if (!isAbortError(caught)) setResetError(messageFrom(caught, 'The password could not be reset.'))
    } finally {
      setResetting(false)
    }
  }, [pendingReset, reload])

  const copyResetPassword = useCallback(async () => {
    if (!resetPassword) return
    try {
      await navigator.clipboard.writeText(resetPassword.password)
      setCopied(true)
    } catch {
      setCopied(false)
    }
  }, [resetPassword])

  if (!isAdmin) {
    return (
      <main className={layout.page}>
        <UsersHeader loading={false} />
        <PermissionNotice
          title="Admin access required"
          description="User management is restricted to admins. Ask an admin to change your role if you need access."
        />
      </main>
    )
  }

  return (
    <main className={layout.page}>
      <UsersHeader loading={loading} onRefresh={() => void reload()} />

      <div className={styles.messages} aria-live="polite">
        {error && (
          <Flash variant="danger" role="alert">
            <div className={styles.flashBody}>
              <div>
                <Text weight="semibold">Users could not be loaded.</Text>
                <Text className={layout.pageDescription}>{error}</Text>
              </div>
              <Button disabled={loading} onClick={() => void reload()}>
                Retry
              </Button>
            </div>
          </Flash>
        )}
        {actionError && (
          <Flash variant="danger" role="alert">
            {actionError}
          </Flash>
        )}
        {status && (
          <Flash variant="success" role="status">
            {status}
          </Flash>
        )}
      </div>

      {resetPassword && (
        <section className={`${layout.box} ${styles.passwordReveal}`} role="status">
          <Text weight="semibold">New password for {resetPassword.username}</Text>
          <Text className={layout.pageDescription}>
            Copy it now — share it over a secure channel. The hub stores only a hash and will not show it again.
          </Text>
          <div className={styles.passwordRow}>
            <code className={styles.passwordValue}>{resetPassword.password}</code>
            <Button leadingVisual={CopyIcon} onClick={() => void copyResetPassword()}>
              {copied ? 'Copied' : 'Copy'}
            </Button>
          </div>
        </section>
      )}

      <div className={styles.columns}>
        <section aria-busy={loading} aria-labelledby="user-list-title">
          <div className={styles.listHeader}>
            <Heading as="h2" variant="small" id="user-list-title">
              Accounts
            </Heading>
            {!loading && (
              <Text size="small" className={`${layout.mono} ${layout.muted}`}>
                {sortedUsers.length} {sortedUsers.length === 1 ? 'account' : 'accounts'}
              </Text>
            )}
          </div>

          {loading && sortedUsers.length === 0 ? (
            <UsersSkeleton />
          ) : sortedUsers.length === 0 && !error ? (
            <div className={layout.box}>
              <Blankslate>
                <Blankslate.Heading as="h3">No accounts yet</Blankslate.Heading>
                <Blankslate.Description>Invite a teammate to give them hub access.</Blankslate.Description>
              </Blankslate>
            </div>
          ) : sortedUsers.length > 0 ? (
            <div className={layout.box}>
              <div className={layout.tableScroll}>
                <table className={layout.table}>
                  <thead>
                    <tr>
                      <th scope="col">User</th>
                      <th scope="col">Role</th>
                      <th scope="col">Status</th>
                      <th scope="col">Created</th>
                      <th scope="col" className={styles.actionColumn}>
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {sortedUsers.map((account) => {
                      const isSelf = account.id === user?.id
                      const busy = busyId === account.id

                      return (
                        <tr key={account.id}>
                          <td>
                            <Text weight="semibold" className={styles.block}>
                              {account.username}
                            </Text>
                            <span className={styles.meta}>{account.email}</span>
                          </td>
                          <td>
                            <Select
                              aria-label={`Role for ${account.username}`}
                              value={account.role}
                              disabled={busy}
                              onChange={(event) => void mutate(account, { role: event.target.value as Role })}
                            >
                              {roleOptions.map((role) => (
                                <Select.Option key={role} value={role}>
                                  {roleLabels[role]}
                                </Select.Option>
                              ))}
                            </Select>
                          </td>
                          <td>
                            <span className={account.disabled ? layout.muted : styles.active}>
                              {account.disabled ? 'Deactivated' : 'Active'}
                            </span>
                          </td>
                          <td className={`${layout.muted} ${styles.nowrap}`}>{formatTimestamp(account.createdAt)}</td>
                          <td className={styles.actionColumn}>
                            <div className={styles.rowActions}>
                              <Button
                                disabled={busy || isSelf}
                                title={isSelf ? 'You cannot deactivate your own account.' : undefined}
                                onClick={() => void mutate(account, { disabled: !account.disabled })}
                              >
                                {account.disabled ? 'Reactivate' : 'Deactivate'}
                              </Button>
                              <Button
                                leadingVisual={KeyIcon}
                                disabled={busy || isSelf}
                                title={isSelf ? 'You cannot reset your own password here.' : undefined}
                                onClick={() => {
                                  setResetError(null)
                                  setPendingReset(account)
                                }}
                              >
                                Reset password
                              </Button>
                              <IconButton
                                icon={TrashIcon}
                                variant="danger"
                                disabled={busy || isSelf}
                                title={isSelf ? 'You cannot delete your own account.' : undefined}
                                aria-label={`Delete ${account.username}`}
                                onClick={() => {
                                  setDeleteError(null)
                                  setPendingDelete(account)
                                }}
                              />
                            </div>
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
            </div>
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
          {pendingDelete?.username} loses hub access immediately and any active sessions are invalidated. Their past
          actions remain in the audit log.
        </p>
        <p>This cannot be undone. Deactivate the account instead if you may need it later.</p>
      </ConfirmDialog>

      <ConfirmDialog
        open={pendingReset !== null}
        title={`Reset ${pendingReset?.username ?? 'user'}'s password?`}
        confirmLabel="Reset password"
        pending={resetting}
        error={resetError}
        onCancel={() => setPendingReset(null)}
        onConfirm={() => void confirmReset()}
      >
        <p>
          A new password is generated and shown once after the reset. All of {pendingReset?.username}&apos;s sessions
          are signed out immediately, including any you did not create.
        </p>
        <p>Share the new password over a secure channel; the change is recorded in the audit log.</p>
      </ConfirmDialog>
    </main>
  )
}

interface UsersHeaderProps {
  loading: boolean
  onRefresh?: () => void
}

function UsersHeader({ loading, onRefresh }: UsersHeaderProps) {
  return (
    <header className={layout.pageHeader}>
      <div className={layout.pageHeaderText}>
        <Heading as="h1" variant="large">
          Users
        </Heading>
        <Text className={layout.pageDescription}>
          Manage who can sign in to the hub and what each account is allowed to do.
        </Text>
      </div>
      {onRefresh && (
        <Button leadingVisual={SyncIcon} disabled={loading} onClick={onRefresh}>
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
    <section className={`${layout.box} ${styles.invite}`}>
      <Heading as="h2" variant="small" className={styles.inviteTitle}>
        <PersonAddIcon size={16} />
        Invite a user
      </Heading>
      <Text className={layout.pageDescription}>
        The hub creates the account immediately. Share the initial password over a secure channel.
      </Text>

      <form className={styles.form} onSubmit={(event) => void submit(event)} aria-label="Invite a user">
        <FormControl>
          <FormControl.Label>Username</FormControl.Label>
          <TextInput
            block
            value={username}
            autoComplete="off"
            onChange={(event) => setUsername(event.target.value)}
          />
        </FormControl>

        <FormControl>
          <FormControl.Label>Email</FormControl.Label>
          <TextInput
            block
            type="email"
            value={email}
            autoComplete="off"
            onChange={(event) => setEmail(event.target.value)}
          />
        </FormControl>

        <FormControl>
          <FormControl.Label>Initial password</FormControl.Label>
          <TextInput
            block
            type="password"
            value={password}
            autoComplete="new-password"
            onChange={(event) => setPassword(event.target.value)}
          />
          <FormControl.Caption>
            {minPasswordLength}–{maxPasswordLength} characters.
          </FormControl.Caption>
        </FormControl>

        <FormControl>
          <FormControl.Label>Role</FormControl.Label>
          <Select value={role} onChange={(event) => setRole(event.target.value as Role)}>
            {roleOptions.map((option) => (
              <Select.Option key={option} value={option}>
                {roleLabels[option]}
              </Select.Option>
            ))}
          </Select>
        </FormControl>

        {error && (
          <Flash variant="danger" role="alert">
            {error}
          </Flash>
        )}

        <Button type="submit" variant="primary" block disabled={submitting}>
          {submitting ? 'Creating…' : 'Create user'}
        </Button>
      </form>
    </section>
  )
}

function UsersSkeleton() {
  return (
    <div className={`${layout.box} ${styles.skeleton}`} aria-label="Loading users">
      <SkeletonText size="titleSmall" maxWidth="12rem" />
      {Array.from({ length: 3 }, (_, index) => (
        <SkeletonText key={index} size="bodyMedium" maxWidth="80%" />
      ))}
    </div>
  )
}

export default UsersPage
