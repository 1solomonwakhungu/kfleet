import { useMemo, useState } from 'react'
import { AlertTriangle, RefreshCw, SearchX, ServerOff } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

import { ClusterCard } from '../components/ClusterCard'
import { DashboardSkeleton } from '../components/dashboard/DashboardSkeleton'
import { FleetControls, type FleetSort, type HealthFilter } from '../components/dashboard/FleetControls'
import { FleetSummary } from '../components/dashboard/FleetSummary'
import { Button } from '../components/ui/button'
import { Card } from '../components/ui/card'
import { useClusters } from '../hooks/useClusters'
import type { Cluster, ClusterHealth } from '../types/cluster'

const healthPriority: Record<ClusterHealth, number> = {
  unreachable: 0,
  degraded: 1,
  unknown: 2,
  healthy: 3,
}

function compareText(left: string, right: string) {
  const normalizedLeft = left.toLowerCase()
  const normalizedRight = right.toLowerCase()
  if (normalizedLeft < normalizedRight) return -1
  if (normalizedLeft > normalizedRight) return 1
  if (left < right) return -1
  if (left > right) return 1
  return 0
}

function compareByName(left: Cluster, right: Cluster) {
  return compareText(left.name, right.name) || compareText(left.id, right.id)
}

function heartbeatTimestamp(value: string) {
  const timestamp = Date.parse(value)
  return Number.isNaN(timestamp) ? Number.NEGATIVE_INFINITY : timestamp
}

function sortClusters(left: Cluster, right: Cluster, sort: FleetSort) {
  if (sort === 'health') {
    return healthPriority[left.health] - healthPriority[right.health] || compareByName(left, right)
  }

  if (sort === 'heartbeat') {
    return heartbeatTimestamp(right.lastHeartbeat) - heartbeatTimestamp(left.lastHeartbeat) || compareByName(left, right)
  }

  return compareByName(left, right)
}

function matchesSearch(cluster: Cluster, query: string) {
  if (!query) return true

  const searchableLabels = Object.entries(cluster.labels).flatMap(([key, value]) => [
    key,
    value,
    `${key}=${value}`,
    `${key}:${value}`,
  ])

  return [cluster.name, ...searchableLabels].some((value) => value.toLowerCase().includes(query))
}

export function Dashboard() {
  const navigate = useNavigate()
  const { clusters, loading, error, refresh } = useClusters()
  const [search, setSearch] = useState('')
  const [healthFilter, setHealthFilter] = useState<HealthFilter>('all')
  const [sort, setSort] = useState<FleetSort>('health')

  const visibleClusters = useMemo(() => {
    const query = search.trim().toLowerCase()
    return clusters
      .filter((cluster) => healthFilter === 'all' || cluster.health === healthFilter)
      .filter((cluster) => matchesSearch(cluster, query))
      .sort((left, right) => sortClusters(left, right, sort))
  }, [clusters, healthFilter, search, sort])

  const hasActiveControls = search.trim().length > 0 || healthFilter !== 'all' || sort !== 'health'
  const resetControls = () => {
    setSearch('')
    setHealthFilter('all')
    setSort('health')
  }

  return (
    <main className="mx-auto min-h-dvh max-w-[100rem] px-4 py-6 sm:px-6 sm:py-8 lg:px-8">
      <header className="flex flex-col gap-5 border-b border-border pb-6 sm:flex-row sm:items-end sm:justify-between">
        <div className="min-w-0">
          <p className="text-sm font-semibold text-blue-400">kFLEET hub</p>
          <h1 className="mt-2 min-w-0 font-display text-3xl font-bold tracking-tight [overflow-wrap:anywhere] sm:text-4xl">
            Fleet dashboard
          </h1>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-muted sm:text-base">
            Operational health, capacity, and agent freshness across registered Kubernetes clusters.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <p className="hidden whitespace-nowrap font-mono text-xs text-muted lg:block">Auto-refresh · 5s</p>
          <Button
            className="bg-blue-600 text-white hover:bg-blue-500 hover:brightness-100 focus-visible:outline-blue-400"
            onClick={() => void refresh()}
            disabled={loading}
          >
            <RefreshCw className={loading ? 'h-4 w-4 animate-spin' : 'h-4 w-4'} aria-hidden="true" />
            {loading ? 'Refreshing…' : 'Refresh fleet'}
          </Button>
        </div>
      </header>

      {loading ? (
        <DashboardSkeleton />
      ) : (
        <>
          <FleetSummary clusters={clusters} />

          {error && clusters.length > 0 && (
            <Card className="mt-5 flex flex-col gap-4 border border-unreachable/40 bg-danger-soft p-4 sm:flex-row sm:items-center sm:justify-between" role="alert">
              <div className="flex min-w-0 items-start gap-3">
                <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0 text-unreachable" aria-hidden="true" />
                <div className="min-w-0">
                  <p className="font-semibold text-foreground">The latest fleet snapshot could not be loaded.</p>
                  <p className="mt-1 break-words text-sm text-muted">Showing the last available data. {error.message}</p>
                </div>
              </div>
              <Button
                size="sm"
                className="shrink-0 bg-blue-600 text-white hover:bg-blue-500 hover:brightness-100 focus-visible:outline-blue-400"
                onClick={() => void refresh()}
              >
                <RefreshCw className="h-4 w-4" aria-hidden="true" />
                Retry
              </Button>
            </Card>
          )}

          {clusters.length > 0 && (
            <FleetControls
              search={search}
              onSearchChange={setSearch}
              health={healthFilter}
              onHealthChange={setHealthFilter}
              sort={sort}
              onSortChange={setSort}
              resultCount={visibleClusters.length}
              totalCount={clusters.length}
              hasActiveControls={hasActiveControls}
              onReset={resetControls}
            />
          )}

          <section className="mt-5" aria-labelledby="cluster-inventory-heading">
            <div className="mb-3 flex items-center justify-between gap-4">
              <h2 id="cluster-inventory-heading" className="font-display text-lg font-bold tracking-tight">
                Cluster inventory
              </h2>
              {clusters.length > 0 && (
                <p className="whitespace-nowrap font-mono text-xs text-muted" role="status" aria-live="polite" aria-atomic="true">
                  {visibleClusters.length} of {clusters.length}
                </p>
              )}
            </div>

            {error && clusters.length === 0 ? (
              <DashboardState
                icon={AlertTriangle}
                title="Fleet data is unavailable"
                description={`The hub did not return a cluster snapshot. ${error.message}`}
                actionLabel="Retry connection"
                onAction={() => void refresh()}
                danger
              />
            ) : clusters.length === 0 ? (
              <DashboardState
                icon={ServerOff}
                title="No clusters registered"
                description="Connect a kFLEET agent to this hub to begin monitoring its Kubernetes cluster."
                actionLabel="Check again"
                onAction={() => void refresh()}
              />
            ) : visibleClusters.length === 0 ? (
              <DashboardState
                icon={SearchX}
                title="No clusters match these controls"
                description="Try a different cluster name or label, or broaden the health filter."
                actionLabel="Reset controls"
                onAction={resetControls}
              />
            ) : (
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
                {visibleClusters.map((cluster) => (
                  <ClusterCard
                    key={cluster.id}
                    cluster={cluster}
                    onClick={() => navigate(`/clusters/${encodeURIComponent(cluster.id)}`)}
                  />
                ))}
              </div>
            )}
          </section>
        </>
      )}
    </main>
  )
}

interface DashboardStateProps {
  icon: typeof ServerOff
  title: string
  description: string
  actionLabel: string
  onAction: () => void
  danger?: boolean
}

function DashboardState({ icon: Icon, title, description, actionLabel, onAction, danger = false }: DashboardStateProps) {
  return (
    <Card className="grid min-h-72 place-items-center border border-dashed border-border px-6 py-12 text-center">
      <div className="max-w-lg">
        <span className={danger
          ? 'mx-auto grid h-12 w-12 place-items-center rounded-lg border border-unreachable/40 bg-danger-soft text-unreachable'
          : 'mx-auto grid h-12 w-12 place-items-center rounded-lg border border-border bg-elevated text-blue-400'}>
          <Icon className="h-6 w-6" aria-hidden="true" />
        </span>
        <h3 className="mt-5 font-display text-xl font-bold tracking-tight">{title}</h3>
        <p className="mt-2 break-words text-sm leading-6 text-muted">{description}</p>
        <Button
          className="mt-5 bg-blue-600 text-white hover:bg-blue-500 hover:brightness-100 focus-visible:outline-blue-400"
          onClick={onAction}
        >
          {actionLabel === 'Reset controls' ? null : <RefreshCw className="h-4 w-4" aria-hidden="true" />}
          {actionLabel}
        </Button>
      </div>
    </Card>
  )
}
