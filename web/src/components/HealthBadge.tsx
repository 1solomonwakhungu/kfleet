import { Badge } from './ui/badge'
import { cn } from '../lib/utils'
import type { ClusterHealth } from '../types/cluster'

interface HealthBadgeProps {
  health: ClusterHealth
}

const healthClasses: Record<ClusterHealth, string> = {
  healthy: 'border-healthy/40 bg-healthy-soft text-healthy',
  degraded: 'border-degraded/40 bg-degraded-soft text-degraded',
  unreachable: 'border-unreachable/40 bg-unreachable-soft text-unreachable',
  unknown: 'border-unknown/40 bg-unknown-soft text-unknown',
}

export function HealthBadge({ health }: HealthBadgeProps) {
  return (
    <Badge
      className={cn('shrink-0 rounded-md border px-2 py-0.5 font-mono text-[11px] font-semibold capitalize', healthClasses[health])}
      aria-label={`Cluster health: ${health}`}
    >
      {health}
    </Badge>
  )
}
