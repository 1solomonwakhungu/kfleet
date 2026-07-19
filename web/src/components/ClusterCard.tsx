import { ArrowUpRight, Clock3 } from 'lucide-react'

import { HealthBadge } from './HealthBadge'
import { StatusDot } from './StatusDot'
import { Card, CardContent, CardHeader } from './ui/card'
import { cn, timeAgo } from '../lib/utils'
import type { Cluster, ClusterHealth } from '../types/cluster'

interface ClusterCardProps {
  cluster: Cluster
  onClick: () => void
}

const borderClasses: Record<ClusterHealth, string> = {
  healthy: 'border-l-healthy',
  degraded: 'border-l-degraded',
  unreachable: 'border-l-unreachable',
  unknown: 'border-l-unknown',
}

function heartbeatDetails(value: string) {
  const timestamp = Date.parse(value)
  if (!value || Number.isNaN(timestamp) || timestamp <= 0) {
    return { freshness: 'No heartbeat', relative: 'Never', exact: 'No heartbeat has been received', dateTime: undefined }
  }

  const ageSeconds = Math.max(0, Math.floor((Date.now() - timestamp) / 1000))
  let freshness = 'Stale'
  if (ageSeconds < 60) freshness = 'Live'
  else if (ageSeconds < 5 * 60) freshness = 'Recent'
  else if (ageSeconds < 15 * 60) freshness = 'Delayed'

  return {
    freshness,
    relative: timeAgo(value),
    exact: new Date(timestamp).toLocaleString(),
    dateTime: new Date(timestamp).toISOString(),
  }
}

export function ClusterCard({ cluster, onClick }: ClusterCardProps) {
  const labels = Object.entries(cluster.labels).sort(([left], [right]) => left.localeCompare(right))
  const visibleLabels = labels.slice(0, 3)
  const hiddenLabelCount = labels.length - visibleLabels.length
  const heartbeat = heartbeatDetails(cluster.lastHeartbeat)

  return (
    <button
      type="button"
      className="group block h-full w-full rounded-lg text-left focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
      aria-label={`Open ${cluster.name} cluster, health ${cluster.health}`}
      onClick={onClick}
    >
      <Card
        className={cn(
          'h-full border border-l-2 border-border transition-[box-shadow] duration-200 ease-out group-hover:ring-1 group-hover:ring-blue-500/50 group-active:ring-2 group-active:ring-blue-500/70',
          borderClasses[cluster.health],
        )}
      >
        <CardHeader className="flex-row items-start justify-between gap-3 pb-4">
          <div className="min-w-0">
            <div className="flex min-w-0 items-center gap-2">
              <StatusDot health={cluster.health} />
              <h3 className="truncate font-display text-lg font-bold tracking-tight" title={cluster.name}>
                {cluster.name}
              </h3>
            </div>
            {cluster.id !== cluster.name && (
              <p className="mt-1 truncate font-mono text-xs text-muted" title={cluster.id}>
                {cluster.id}
              </p>
            )}
          </div>
          <HealthBadge health={cluster.health} />
        </CardHeader>

        <CardContent>
          <dl className="grid grid-cols-2 gap-px overflow-hidden rounded-md border border-border bg-border">
            <Metric label="Nodes" value={cluster.nodeCount.toLocaleString()} />
            <Metric label="Pods" value={cluster.podCount.toLocaleString()} />
            <Metric label="Kubernetes" value={cluster.k8sVersion || 'Unknown'} mono />
            <Metric label="Agent" value={cluster.agentVersion || 'Unknown'} mono />
          </dl>

          <div className="mt-4 min-h-[3.5rem]">
            <p className="text-xs font-semibold text-muted">Labels</p>
            {visibleLabels.length > 0 ? (
              <div className="mt-2 flex flex-wrap gap-1.5" aria-label={`${labels.length} cluster labels`}>
                {visibleLabels.map(([key, value]) => (
                  <span
                    key={key}
                    className="max-w-full truncate rounded border border-border bg-background px-2 py-1 font-mono text-[11px] text-muted"
                    title={`${key}=${value}`}
                  >
                    <span className="text-foreground">{key}</span>={value}
                  </span>
                ))}
                {hiddenLabelCount > 0 && (
                  <span className="rounded border border-border bg-background px-2 py-1 font-mono text-[11px] text-muted">
                    +{hiddenLabelCount}
                  </span>
                )}
              </div>
            ) : (
              <p className="mt-2 text-xs text-muted">No labels reported</p>
            )}
          </div>

          <div className="mt-4 flex items-center justify-between gap-3 border-t border-border pt-4">
            <div className="flex min-w-0 items-center gap-2 text-xs text-muted">
              <Clock3 className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
              <span className="truncate">
                <span className="font-semibold text-foreground">{heartbeat.freshness}</span>
                {' · '}
                <time dateTime={heartbeat.dateTime} title={heartbeat.exact}>
                  {heartbeat.relative}
                </time>
              </span>
            </div>
            <ArrowUpRight className="h-4 w-4 shrink-0 text-blue-400" aria-hidden="true" />
          </div>
        </CardContent>
      </Card>
    </button>
  )
}

interface MetricProps {
  label: string
  value: string
  mono?: boolean
}

function Metric({ label, value, mono = false }: MetricProps) {
  return (
    <div className="min-w-0 bg-surface px-3 py-3 group-hover:bg-elevated">
      <dt className="text-xs text-muted">{label}</dt>
      <dd className={cn('mt-1 truncate text-base font-bold tabular-nums', mono && 'font-mono text-xs')} title={value}>
        {value}
      </dd>
    </div>
  )
}
