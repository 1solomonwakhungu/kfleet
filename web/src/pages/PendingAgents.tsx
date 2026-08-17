import { useCallback, useEffect, useRef, useState } from 'react'
import { Button, Flash, Heading, Text } from '@primer/react'
import { Blankslate, SkeletonText } from '@primer/react/experimental'
import { CheckCircleIcon, ShieldCheckIcon, SyncIcon } from '@primer/octicons-react'

import { PendingAgentTable } from '../components/agents/PendingAgentTable'
import { RegistrationTokenCard } from '../components/admin/RegistrationTokenCard'
import { useAuth } from '../auth/AuthContext'
import {
  approvePendingAgent,
  getPendingAgents,
  type PendingAgent,
} from '../lib/pendingAgentsApi'
import layout from '../styles/layout.module.css'
import styles from './PendingAgents.module.css'

function messageFrom(error: unknown, fallback: string): string {
  return error instanceof Error && error.message ? error.message : fallback
}

function isAbortError(error: unknown): boolean {
  return error instanceof DOMException && error.name === 'AbortError'
}

function PendingAgentsPage() {
  const { user } = useAuth()
  const canApprove = user?.role === 'admin' || user?.role === 'operator'
  const [agents, setAgents] = useState<PendingAgent[]>([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [approvingIds, setApprovingIds] = useState<Set<string>>(() => new Set())
  const [approvalErrors, setApprovalErrors] = useState<Record<string, string>>({})
  const [successMessage, setSuccessMessage] = useState<string | null>(null)
  const loadController = useRef<AbortController | null>(null)
  const approvalControllers = useRef(new Map<string, AbortController>())

  const loadAgents = useCallback(async () => {
    loadController.current?.abort()
    const controller = new AbortController()
    loadController.current = controller
    setLoading(true)
    setLoadError(null)

    try {
      const pending = await getPendingAgents(controller.signal)
      if (controller.signal.aborted) return
      setAgents(pending)
      setApprovalErrors((current) =>
        Object.fromEntries(Object.entries(current).filter(([id]) => pending.some((agent) => agent.id === id))),
      )
    } catch (error) {
      if (!isAbortError(error)) {
        setLoadError(messageFrom(error, 'Pending agents could not be loaded.'))
      }
    } finally {
      if (!controller.signal.aborted) setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadAgents()
    const controllers = approvalControllers.current

    return () => {
      loadController.current?.abort()
      controllers.forEach((controller) => controller.abort())
    }
  }, [loadAgents])

  const approve = useCallback(async (agent: PendingAgent) => {
    if (approvalControllers.current.has(agent.id)) return

    const controller = new AbortController()
    approvalControllers.current.set(agent.id, controller)
    setApprovingIds((current) => new Set(current).add(agent.id))
    setApprovalErrors((current) => {
      const next = { ...current }
      delete next[agent.id]
      return next
    })
    setSuccessMessage(null)

    try {
      await approvePendingAgent(agent.id, controller.signal)
      if (controller.signal.aborted) return
      setAgents((current) => current.filter((pending) => pending.id !== agent.id))
      setSuccessMessage(`${agent.name} was approved and can now connect to the fleet.`)
    } catch (error) {
      if (!isAbortError(error)) {
        setApprovalErrors((current) => ({
          ...current,
          [agent.id]: messageFrom(error, 'This agent could not be approved. Try again.'),
        }))
      }
    } finally {
      approvalControllers.current.delete(agent.id)
      if (!controller.signal.aborted) {
        setApprovingIds((current) => {
          const next = new Set(current)
          next.delete(agent.id)
          return next
        })
      }
    }
  }, [])

  const initialLoading = loading && agents.length === 0 && !loadError

  return (
    <main className={layout.page}>
      <header className={layout.pageHeader}>
        <div className={layout.pageHeaderText}>
          <Heading as="h1" variant="large">
            Pending agents
          </Heading>
          <Text className={layout.pageDescription}>
            Review agent identity and cluster metadata before granting fleet access.
          </Text>
        </div>
        <Button leadingVisual={SyncIcon} disabled={loading} onClick={() => void loadAgents()}>
          {loading ? 'Refreshing…' : 'Refresh'}
        </Button>
      </header>

      <div className={styles.messages} aria-live="polite">
        {loadError && (
          <Flash variant="danger" role="alert">
            <div className={styles.flashBody}>
              <div>
                <Text weight="semibold">Pending agents could not be loaded.</Text>
                <Text className={layout.pageDescription}>{loadError}</Text>
              </div>
              <Button disabled={loading} onClick={() => void loadAgents()}>
                Retry
              </Button>
            </div>
          </Flash>
        )}

        {successMessage && (
          <Flash variant="success" role="status">
            <div className={styles.flashBody}>
              <CheckCircleIcon size={16} />
              <div>
                <Text weight="semibold">Approval complete</Text>
                <Text className={layout.pageDescription}>{successMessage}</Text>
              </div>
            </div>
          </Flash>
        )}
      </div>

      <section className={styles.list} aria-busy={loading} aria-labelledby="pending-list-title">
        <div className={styles.listHeader}>
          <Heading as="h2" variant="small" id="pending-list-title">
            Awaiting review
          </Heading>
          {!initialLoading && (
            <Text size="small" className={`${layout.mono} ${layout.muted}`}>
              {agents.length} {agents.length === 1 ? 'agent' : 'agents'}
            </Text>
          )}
        </div>

        {initialLoading ? (
          <PendingAgentsSkeleton />
        ) : agents.length > 0 ? (
          <PendingAgentTable
            agents={agents}
            approvingIds={approvingIds}
            errors={approvalErrors}
            canApprove={canApprove}
            onApprove={(agent) => void approve(agent)}
          />
        ) : !loadError ? (
          <div className={layout.box}>
            <Blankslate>
              <Blankslate.Visual>
                <ShieldCheckIcon size={24} />
              </Blankslate.Visual>
              <Blankslate.Heading as="h3">No agents awaiting approval</Blankslate.Heading>
              <Blankslate.Description>
                New agent registrations will appear here for review.
              </Blankslate.Description>
            </Blankslate>
          </div>
        ) : null}
      </section>

      <section className={styles.token}>
        <RegistrationTokenCard />
      </section>
    </main>
  )
}

function PendingAgentsSkeleton() {
  return (
    <div className={`${layout.box} ${styles.skeleton}`} aria-label="Loading pending agents">
      <SkeletonText size="titleSmall" maxWidth="12rem" />
      {Array.from({ length: 3 }, (_, index) => (
        <SkeletonText key={index} size="bodyMedium" maxWidth="80%" />
      ))}
    </div>
  )
}

export default PendingAgentsPage
