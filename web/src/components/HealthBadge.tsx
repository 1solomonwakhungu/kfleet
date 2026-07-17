import { Badge } from './ui/badge'
import { cn } from '../lib/utils'
import type { ClusterHealth } from '../types/cluster'

interface HealthBadgeProps {
  health: ClusterHealth
}

const healthClasses: Record<ClusterHealth, string> = {
  healthy: 'bg-healthy-soft text-healthy',
  degraded: 'bg-degraded-soft text-degraded',
  unreachable: 'bg-unreachable-soft text-unreachable',
  unknown: 'bg-unknown-soft text-unknown',
}

export function HealthBadge({ health }: HealthBadgeProps) {
  return (
    <Badge className={cn('capitalize', healthClasses[health])}>
      {health}
    </Badge>
  )
}
