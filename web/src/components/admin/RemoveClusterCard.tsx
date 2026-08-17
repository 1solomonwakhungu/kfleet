import { useCallback, useState } from 'react'
import { Button, Flash, Heading, Text } from '@primer/react'
import { TrashIcon } from '@primer/octicons-react'
import { useNavigate } from 'react-router-dom'

import { useAuth } from '../../auth/AuthContext'
import { api } from '../../lib/api'
import { isAbortError, messageFrom } from '../../lib/errors'
import { ConfirmDialog } from './ConfirmDialog'
import layout from '../../styles/layout.module.css'
import styles from './AdminCard.module.css'

interface RemoveClusterCardProps {
  clusterId: string
  clusterName: string
}

/**
 * Destructive cluster removal. DELETE /api/v1/clusters/{id} requires the
 * operator role, so read-only users see the action disabled rather than
 * hitting a 403.
 */
export function RemoveClusterCard({ clusterId, clusterName }: RemoveClusterCardProps) {
  const { user } = useAuth()
  const navigate = useNavigate()
  const canRemove = user?.role === 'admin' || user?.role === 'operator'
  const [confirming, setConfirming] = useState(false)
  const [pending, setPending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const remove = useCallback(async () => {
    setPending(true)
    setError(null)
    try {
      await api.deleteCluster(clusterId)
      setConfirming(false)
      navigate('/', { replace: true })
    } catch (caught) {
      if (!isAbortError(caught)) setError(messageFrom(caught, 'The cluster could not be removed.'))
    } finally {
      setPending(false)
    }
  }, [clusterId, navigate])

  return (
    <section className={`${layout.box} ${styles.danger}`} aria-labelledby="remove-cluster-title">
      <div className={styles.header}>
        <div className={styles.headerText}>
          <Heading as="h2" variant="small" id="remove-cluster-title">
            Remove cluster
          </Heading>
          <Text className={styles.description}>
            Deletes {clusterName} from the hub along with its stored snapshots, alerts, and agent token. The cluster
            itself is untouched, but its agent must re-register before it appears again.
          </Text>
          {!canRemove && (
            <Text className={styles.description}>Removing a cluster requires the operator or admin role.</Text>
          )}
        </div>
        <Button
          variant="danger"
          leadingVisual={TrashIcon}
          disabled={!canRemove}
          onClick={() => {
            setError(null)
            setConfirming(true)
          }}
        >
          Remove cluster
        </Button>
      </div>

      {error && !confirming && (
        <Flash variant="danger" role="alert" className={styles.flash}>
          {error}
        </Flash>
      )}

      <ConfirmDialog
        open={confirming}
        title={`Remove ${clusterName}?`}
        confirmLabel="Remove cluster"
        pending={pending}
        error={error}
        onCancel={() => setConfirming(false)}
        onConfirm={() => void remove()}
      >
        <p>
          This permanently removes {clusterName} and everything the hub stores about it: node and pod snapshots, alert
          history, operational timeline, and the agent token.
        </p>
        <p>This cannot be undone.</p>
      </ConfirmDialog>
    </section>
  )
}
