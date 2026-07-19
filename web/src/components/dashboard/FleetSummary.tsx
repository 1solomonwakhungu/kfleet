import { Boxes, Network, Server, ShieldCheck } from 'lucide-react'

import { StatusDot } from '../StatusDot'
import { Card } from '../ui/card'
import type { Cluster, ClusterHealth } from '../../types/cluster'

interface FleetSummaryProps {
  clusters: Cluster[]
}

export function FleetSummary({ clusters }: FleetSummaryProps) {
  const counts = clusters.reduce<Record<ClusterHealth, number>>(
    (result, cluster) => {
      result[cluster.health] += 1
      return result
    },
    { healthy: 0, degraded: 0, unreachable: 0, unknown: 0 },
  )
  const nodeCount = clusters.reduce((total, cluster) => total + cluster.nodeCount, 0)
  const podCount = clusters.reduce((total, cluster) => total + cluster.podCount, 0)
  const healthyPercent = clusters.length === 0 ? null : Math.round((counts.healthy / clusters.length) * 100)
  const attentionCount = counts.degraded + counts.unreachable + counts.unknown

  return (
    <section className="mt-6" aria-labelledby="fleet-summary-heading">
      <h2 id="fleet-summary-heading" className="sr-only">Fleet summary</h2>
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <SummaryMetric
          label="Clusters"
          value={clusters.length.toLocaleString()}
          detail={attentionCount === 0 ? 'None need attention' : `${attentionCount.toLocaleString()} need attention`}
          icon={Network}
        />
        <SummaryMetric
          label="Healthy"
          value={healthyPercent === null ? '—' : `${healthyPercent}%`}
          detail={`${counts.healthy.toLocaleString()} of ${clusters.length.toLocaleString()} clusters`}
          icon={ShieldCheck}
          status
        />
        <SummaryMetric label="Nodes" value={nodeCount.toLocaleString()} detail="Across the fleet" icon={Server} />
        <SummaryMetric label="Pods" value={podCount.toLocaleString()} detail="Reported workload pods" icon={Boxes} />
      </div>
      <Card className="mt-3 border border-border px-4 py-3">
        <dl className="grid grid-cols-2 gap-x-6 gap-y-3 sm:grid-cols-4" aria-label="Cluster health counts">
          {(['healthy', 'degraded', 'unreachable', 'unknown'] as const).map((health) => (
            <div key={health} className="flex min-w-0 items-center gap-2">
              <StatusDot health={health} />
              <dt className="truncate text-xs capitalize text-muted">{health}</dt>
              <dd className="ml-auto font-mono text-sm font-semibold tabular-nums text-foreground">
                {counts[health].toLocaleString()}
              </dd>
            </div>
          ))}
        </dl>
      </Card>
    </section>
  )
}

interface SummaryMetricProps {
  label: string
  value: string
  detail: string
  icon: typeof Network
  status?: boolean
}

function SummaryMetric({ label, value, detail, icon: Icon, status = false }: SummaryMetricProps) {
  return (
    <Card className="min-w-0 border border-border p-4 sm:p-5">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-xs font-semibold text-muted">{label}</p>
          <p className="mt-2 font-display text-2xl font-bold tracking-tight tabular-nums sm:text-3xl">{value}</p>
        </div>
        <span className={status ? 'text-healthy' : 'text-blue-400'}>
          <Icon className="h-5 w-5" aria-hidden="true" />
        </span>
      </div>
      <p className="mt-2 truncate text-xs text-muted" title={detail}>{detail}</p>
    </Card>
  )
}
