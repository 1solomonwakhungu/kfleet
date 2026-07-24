import { useState, type FormEvent } from 'react'
import { Boxes, LoaderCircle, LockKeyhole } from 'lucide-react'
import { Navigate, useLocation, useNavigate } from 'react-router-dom'

import { useAuth } from '../auth/AuthContext'
import { Button } from '../components/ui/button'

export function Login() {
  const { user, login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  if (user) return <Navigate to="/" replace />

  const requestedDestination =
    typeof location.state === 'object' &&
    location.state !== null &&
    typeof (location.state as { from?: unknown }).from === 'string'
      ? (location.state as { from: string }).from
      : '/'
  const destination =
    requestedDestination.startsWith('/') &&
    !requestedDestination.startsWith('//') &&
    !requestedDestination.includes('\\')
      ? requestedDestination
      : '/'

  async function submit(event: FormEvent) {
    event.preventDefault()
    if (submitting) return
    setSubmitting(true)
    setError(null)
    try {
      await login(username.trim(), password)
      navigate(destination, { replace: true })
    } catch (loginError) {
      setError(loginError instanceof Error ? loginError.message : 'Sign in failed.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <main className="grid min-h-dvh place-items-center bg-background px-4 py-12 text-foreground">
      <section className="w-full max-w-md rounded-lg border border-border bg-surface p-6 shadow-2xl shadow-black/20 sm:p-8">
        <div className="flex items-center gap-3">
          <span className="grid size-10 place-items-center rounded-md bg-blue-600 text-white">
            <Boxes className="size-5" aria-hidden="true" />
          </span>
          <div>
            <p className="font-display text-sm font-bold tracking-[0.08em]">kFLEET</p>
            <p className="text-xs text-muted">Control plane</p>
          </div>
        </div>

        <div className="mt-8">
          <LockKeyhole className="size-5 text-blue-400" aria-hidden="true" />
          <h1 className="mt-3 font-display text-3xl font-bold tracking-tight">Sign in</h1>
          <p className="mt-2 text-sm text-muted">
            Use an account provisioned by a kfleet administrator.
          </p>
        </div>

        <form className="mt-7 space-y-5" onSubmit={(event) => void submit(event)}>
          <label className="block text-sm font-semibold">
            Username
            <input
              autoComplete="username"
              autoFocus
              required
              maxLength={128}
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              className="mt-2 min-h-11 w-full rounded-md border border-input bg-background px-3 text-foreground"
            />
          </label>
          <label className="block text-sm font-semibold">
            Password
            <input
              type="password"
              autoComplete="current-password"
              required
              maxLength={72}
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              className="mt-2 min-h-11 w-full rounded-md border border-input bg-background px-3 text-foreground"
            />
          </label>

          {error && (
            <p className="rounded-md bg-danger-soft p-3 text-sm text-danger" role="alert">
              {error}
            </p>
          )}

          <Button type="submit" className="w-full bg-blue-600 text-white hover:bg-blue-500" disabled={submitting}>
            {submitting && <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />}
            {submitting ? 'Signing in…' : 'Sign in'}
          </Button>
        </form>
      </section>
    </main>
  )
}
