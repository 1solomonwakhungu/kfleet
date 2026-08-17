import { useCallback, useState } from 'react'
import { Button, Flash, Heading, Text } from '@primer/react'
import { CopyIcon, KeyIcon, SyncIcon } from '@primer/octicons-react'

import { useAuth } from '../../auth/AuthContext'
import { adminApi } from '../../lib/adminApi'
import { isAbortError, messageFrom } from '../../lib/errors'
import { ConfirmDialog } from './ConfirmDialog'
import layout from '../../styles/layout.module.css'
import styles from './AdminCard.module.css'

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
    <section className={`${layout.box} ${styles.card}`} aria-labelledby="registration-token-title">
      <div className={styles.header}>
        <div className={styles.headerText}>
          <Heading as="h2" variant="small" id="registration-token-title" className={styles.title}>
            <KeyIcon size={16} />
            Agent registration token
          </Heading>
          <Text className={styles.description}>
            Agents present this shared token when they first register. Rotating it invalidates the previous token
            immediately; already-registered agents keep working with their own per-agent tokens.
          </Text>
        </div>
        <Button
          leadingVisual={SyncIcon}
          onClick={() => {
            setError(null)
            setConfirming(true)
          }}
        >
          Rotate token
        </Button>
      </div>

      {error && !confirming && (
        <Flash variant="danger" role="alert" className={styles.flash}>
          {error}
        </Flash>
      )}

      {token && (
        <div className={styles.token} role="status">
          <Text weight="semibold">New registration token</Text>
          <Text className={styles.description}>
            Copy it now — the hub stores only its hash and will not show it again.
          </Text>
          <div className={styles.tokenRow}>
            <code className={styles.tokenValue}>{token}</code>
            <Button leadingVisual={CopyIcon} onClick={() => void copy()}>
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
          The current registration token stops working straight away. Any install scripts or automation that embed it
          must be updated before new agents can register.
        </p>
        <p>The replacement token is displayed once and cannot be retrieved later.</p>
      </ConfirmDialog>
    </section>
  )
}
