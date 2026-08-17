import { useCallback, useState } from 'react'
import { Copy, KeyRound, RefreshCw } from 'lucide-react'

import { useAuth } from '../../auth/AuthContext'
import { adminApi } from '../../lib/adminApi'
import { isAbortError, messageFrom } from '../../lib/errors'
import { Button } from '../ui/button'
import { Card } from '../ui/card'
import { ConfirmDialog } from './ConfirmDialog'

/**
 * Rotates the shared agent registration token
 * (POST /api/v1/admin/registration-token/rotate, admin only). The raw token is
 * returned exactly once, so it is shown until the operator navigates away.
 */
export function RegistrationTokenCard() {
  const { user } = useAuth()
  const isAdmin = user?.role === 'admin'
  const [confirming, setConfirming] = useState(false)
  const [pending, setPending] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)

  const rotate = useCallback(async () => {
    setPending(true)
    setError(null)
    try {
      const rotated = await adminApi.rotateRegistrationToken()
      setToken(rotated)
      setCopied(false)
      setConfirming(false)
    } catch (caught) {
      if (!isAbortError(caught)) setError(messageFrom(caught, 'The registration token could not be rotated.'))
    } finally {
      setPending(false)
    }
  }, [])

  const copy = useCallback(async () => {
    if (!token) return
    try {
      await navigator.clipboard.writeText(token)
      setCopied(true)
    } catch {
      setCopied(false)
    }
  }, [token])

  if (!isAdmin) return null

  return (
    <Card className="p-5 ring-1 ring-inset ring-border" aria-labelledby="registration-token-title">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <h2 id="registration-token-title" className="flex items-center gap-2 font-display text-lg font-bold">
            <KeyRound className="size-5 text-muted" aria-hidden="true" />
            Agent registration token
          </h2>
          <p className="mt-2 max-w-2xl text-sm text-muted">
            Agents present this shared token when they first register. Rotating it invalidates the previous
            token immediately; already-registered agents keep working with their own per-agent tokens.
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          className="self-start"
          onClick={() => {
            setError(null)
            setConfirming(true)
          }}
        >
          <RefreshCw className="size-4" aria-hidden="true" />
          Rotate token
        </Button>
      </div>

      {error && !confirming && (
        <p className="mt-4 rounded-md bg-danger-soft p-3 text-sm text-danger" role="alert">
          {error}
        </p>
      )}

      {token && (
        <div className="mt-4 rounded-md border border-border bg-background p-4" role="status">
          <p className="text-sm font-semibold">New registration token</p>
          <p className="mt-1 text-sm text-muted">
            Copy it now — the hub stores only its hash and will not show it again.
          </p>
          <div className="mt-3 flex flex-col gap-3 sm:flex-row sm:items-center">
            <code className="min-w-0 flex-1 break-all rounded-md bg-elevated px-3 py-2 font-mono text-xs">
              {token}
            </code>
            <Button variant="outline" size="sm" className="self-start sm:self-auto" onClick={() => void copy()}>
              <Copy className="size-4" aria-hidden="true" />
              {copied ? 'Copied' : 'Copy'}
            </Button>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={confirming}
        title="Rotate the agent registration token?"
        confirmLabel="Rotate token"
        pending={pending}
        error={error}
        onCancel={() => setConfirming(false)}
        onConfirm={() => void rotate()}
      >
        <p>
          The current registration token stops working straight away. Any install scripts or automation that
          embed it must be updated before new agents can register.
        </p>
        <p>The replacement token is displayed once and cannot be retrieved later.</p>
      </ConfirmDialog>
    </Card>
  )
}
