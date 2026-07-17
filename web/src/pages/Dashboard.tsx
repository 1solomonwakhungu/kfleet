import { useNavigate } from 'react-router-dom'

import { ClusterCard } from '../components/ClusterCard'
import { Button } from '../components/ui/button'
import { Card } from '../components/ui/card'
import { useClusters } from '../hooks/useClusters'
import type { ClusterHealth } from '../types/cluster'

const summaryHealth: ClusterHealth[] = ['healthy', 'degraded', 'unreachable']

const summaryClasses: Record<ClusterHealth, string> = {
  healthy: 'text-healthy',
  degraded: 'text-degraded',
  unreachable: 'text-unreachable',
  unknown: 'text-unknown',
}

export function Dashboard() {
  const navigate = useNavigate()
  const { clusters, loading, error, refresh } = useClusters()
  const counts = clusters.reduce<Record<ClusterHealth, number>>(
    (result, cluster) => ({ ...result, [cluster.health]: result[cluster.health] + 1 }),
    { healthy: 0, degraded: 0, unreachable: 0, unknown: 0 },
  )

  return (
    <main className="mx-auto min-h-dvh max-w-[100rem] px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <header className="flex flex-col gap-6 border-b border-border pb-7 lg:flex-row lg:items-end lg:justify-between">
        <div className="min-w-0">
          <p className="font-mono text-sm text-accent">kfleet hub</p>
          <h1 className="mt-2 min-w-0 font-display text-3xl font-bold tracking-tight [overflow-wrap:anywhere] sm:text-4xl">
            Cluster fleet
          </h1>
          <p className="mt-2 max-w-2xl text-muted">Health and capacity across every registered Kubernetes cluster.</p>
        </div>
        <div className="flex flex-wrap items-center gap-x-5 gap-y-3" aria-label="Fleet health summary">
          <span className="font-mono text-sm text-muted">
            <strong className="text-lg text-foreground">{clusters.length}</strong> total
          </span>
          {summaryHealth.map((health) => (
            <span key={health} className="font-mono text-sm capitalize text-muted">
              <strong className={`text-lg ${summaryClasses[health]}`}>{counts[health]}</strong> {health}
            </span>
          ))}
        </div>
      </header>

      {error && (
        <section className="mt-6 flex flex-col gap-3 rounded-lg bg-danger-soft p-4 text-danger sm:flex-row sm:items-center sm:justify-between" role="alert">
          <p>Clusters could not be loaded. Check the hub connection and retry.</p>
          <Button variant="outline" size="sm" onClick={() => void refresh()}>
            Retry
          </Button>
        </section>
      )}

      <section className="mt-7" aria-busy={loading} aria-live="polite">
        {loading ? (
          <DashboardSkeleton />
        ) : clusters.length === 0 ? (
          <div className="grid min-h-64 place-items-center rounded-lg bg-surface px-6 text-center">
            <div>
              <p className="font-display text-xl font-bold">No clusters registered</p>
              <p className="mt-2 text-muted">Connect a kfleet agent to begin monitoring a cluster.</p>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {clusters.map((cluster) => (
              <ClusterCard
                key={cluster.id}
                cluster={cluster}
                onClick={() => navigate(`/clusters/${encodeURIComponent(cluster.id)}`)}
              />
            ))}
          </div>
        )}
      </section>
    </main>
  )
}

function DashboardSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4" aria-label="Loading clusters">
      {Array.from({ length: 8 }, (_, index) => (
        <Card key={index} className="h-64 animate-pulse p-5">
          <div className="h-5 w-2/3 rounded bg-elevated" />
          <div className="mt-8 grid grid-cols-2 gap-5">
            <div className="h-12 rounded bg-elevated" />
            <div className="h-12 rounded bg-elevated" />
            <div className="h-12 rounded bg-elevated" />
            <div className="h-12 rounded bg-elevated" />
          </div>
          <div className="mt-7 h-4 w-3/4 rounded bg-elevated" />
        </Card>
      ))}
    </div>
  )
}
