import { useCallback, useState } from 'react'
import { Trash2 } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

import { useAuth } from '../../auth/AuthContext'
import { api } from '../../lib/api'
import { isAbortError, messageFrom } from '../../lib/errors'
import { Button } from '../ui/button'
import { Card } from '../ui/card'
import { ConfirmDialog } from './ConfirmDialog'

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
    <Card className="border border-danger/40 p-5" aria-labelledby="remove-cluster-title">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <h2 id="remove-cluster-title" className="font-display text-lg font-bold">
            Remove cluster
          </h2>
          <p className="mt-2 max-w-2xl text-sm text-muted">
            Deletes {clusterName} from the hub along with its stored snapshots, alerts, and agent token. The
            cluster itself is untouched, but its agent must re-register before it appears again.
          </p>
          {!canRemove && (
            <p className="mt-2 text-sm text-muted">
              Removing a cluster requires the operator or admin role.
            </p>
          )}
        </div>
        <Button
          variant="danger"
          size="sm"
          className="self-start"
          disabled={!canRemove}
          onClick={() => {
            setError(null)
            setConfirming(true)
          }}
        >
          <Trash2 className="size-4" aria-hidden="true" />
          Remove cluster
        </Button>
      </div>

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
          This permanently removes {clusterName} and everything the hub stores about it: node and pod
          snapshots, alert history, operational timeline, and the agent token.
        </p>
        <p>This cannot be undone.</p>
      </ConfirmDialog>
    </Card>
  )
}
