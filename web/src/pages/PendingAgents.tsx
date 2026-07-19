import { useCallback, useEffect, useRef, useState } from 'react'
import { CheckCircle2, LoaderCircle, RefreshCw, ShieldCheck } from 'lucide-react'

import { PendingAgentTable } from '../components/agents/PendingAgentTable'
import { Button } from '../components/ui/button'
import { Card, CardContent } from '../components/ui/card'
import {
  approvePendingAgent,
  getPendingAgents,
  type PendingAgent,
} from '../lib/pendingAgentsApi'

function messageFrom(error: unknown, fallback: string): string {
  return error instanceof Error && error.message ? error.message : fallback
}

function isAbortError(error: unknown): boolean {
  return error instanceof DOMException && error.name === 'AbortError'
}

function PendingAgentsPage() {
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
    <main className="mx-auto min-h-dvh max-w-[100rem] px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <header className="flex flex-col gap-5 border-b border-border pb-7 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="font-mono text-sm text-blue-400">kfleet access</p>
          <h1 className="mt-2 font-display text-3xl font-bold tracking-tight sm:text-4xl">Pending agents</h1>
          <p className="mt-2 max-w-2xl text-muted">
            Review agent identity and cluster metadata before granting fleet access.
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          disabled={loading}
          onClick={() => void loadAgents()}
        >
          {loading ? (
            <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />
          ) : (
            <RefreshCw className="size-4" aria-hidden="true" />
          )}
          {loading ? 'Refreshing…' : 'Refresh'}
        </Button>
      </header>

      <div className="mt-6 space-y-4" aria-live="polite">
        {loadError && (
          <section
            className="flex flex-col gap-3 rounded-lg bg-danger-soft p-4 text-danger sm:flex-row sm:items-center sm:justify-between"
            role="alert"
          >
            <div>
              <p className="font-semibold">Pending agents could not be loaded.</p>
              <p className="mt-1 text-sm">{loadError}</p>
            </div>
            <Button variant="outline" size="sm" disabled={loading} onClick={() => void loadAgents()}>
              Retry
            </Button>
          </section>
        )}

        {successMessage && (
          <section className="flex items-start gap-3 rounded-lg bg-blue-950 p-4 text-blue-100 ring-1 ring-inset ring-blue-800" role="status">
            <CheckCircle2 className="mt-0.5 size-5 shrink-0 text-blue-400" aria-hidden="true" />
            <div>
              <p className="font-semibold">Approval complete</p>
              <p className="mt-1 text-sm text-blue-200">{successMessage}</p>
            </div>
          </section>
        )}
      </div>

      <section className="mt-7" aria-busy={loading} aria-labelledby="pending-list-title">
        <div className="mb-4 flex items-center justify-between gap-4">
          <h2 id="pending-list-title" className="font-display text-lg font-bold">
            Awaiting review
          </h2>
          {!initialLoading && (
            <span className="font-mono text-sm text-muted">
              {agents.length} {agents.length === 1 ? 'agent' : 'agents'}
            </span>
          )}
        </div>

        {initialLoading ? (
          <PendingAgentsSkeleton />
        ) : agents.length > 0 ? (
          <PendingAgentTable
            agents={agents}
            approvingIds={approvingIds}
            errors={approvalErrors}
            onApprove={(agent) => void approve(agent)}
          />
        ) : !loadError ? (
          <Card className="ring-1 ring-inset ring-border">
            <CardContent className="grid min-h-64 place-items-center p-6 text-center">
              <div>
                <span className="mx-auto grid size-12 place-items-center rounded-full bg-blue-950 text-blue-400 ring-1 ring-inset ring-blue-800">
                  <ShieldCheck className="size-6" aria-hidden="true" />
                </span>
                <p className="mt-4 font-display text-xl font-bold">No agents awaiting approval</p>
                <p className="mt-2 text-muted">New agent registrations will appear here for review.</p>
              </div>
            </CardContent>
          </Card>
        ) : null}
      </section>
    </main>
  )
}

function PendingAgentsSkeleton() {
  return (
    <Card className="animate-pulse p-5 ring-1 ring-inset ring-border" aria-label="Loading pending agents">
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

export default PendingAgentsPage
