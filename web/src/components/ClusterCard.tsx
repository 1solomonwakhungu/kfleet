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

export function ClusterCard({ cluster, onClick }: ClusterCardProps) {
  return (
    <button
      type="button"
      className="group block w-full rounded-lg text-left"
      aria-label={`Open ${cluster.name} cluster`}
      onClick={onClick}
    >
      <Card
        className={cn(
          'h-full border-l-2 transition-[background-color,transform] duration-200 ease-out group-hover:-translate-y-0.5 group-hover:bg-elevated group-active:translate-y-0',
          borderClasses[cluster.health],
        )}
      >
        <CardHeader className="flex-row items-start justify-between gap-3">
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <StatusDot health={cluster.health} />
              <h2 className="truncate font-display text-lg font-bold tracking-tight">{cluster.name}</h2>
            </div>
          </div>
          <HealthBadge health={cluster.health} />
        </CardHeader>
        <CardContent>
          <dl className="grid grid-cols-2 gap-x-4 gap-y-5">
            <Metric label="Nodes" value={cluster.nodeCount.toLocaleString()} />
            <Metric label="Pods" value={cluster.podCount.toLocaleString()} />
            <Metric label="Kubernetes" value={cluster.k8sVersion || '—'} mono />
            <Metric label="Agent" value={cluster.agentVersion || '—'} mono />
          </dl>
          <p className="mt-5 border-t border-border pt-4 text-sm text-muted">
            Last heartbeat <span className="font-mono text-foreground">{timeAgo(cluster.lastHeartbeat)}</span>
          </p>
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
    <div>
      <dt className="text-sm text-muted">{label}</dt>
      <dd className={cn('mt-1 text-base font-bold tabular-nums', mono && 'font-mono text-sm')}>{value}</dd>
    </div>
  )
}
