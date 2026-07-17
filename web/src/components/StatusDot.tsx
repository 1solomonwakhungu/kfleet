import { cn } from '../lib/utils'
import type { ClusterHealth } from '../types/cluster'

interface StatusDotProps {
  health: ClusterHealth
}

const healthClasses: Record<ClusterHealth, string> = {
  healthy: 'bg-healthy',
  degraded: 'bg-degraded',
  unreachable: 'bg-unreachable',
  unknown: 'bg-unknown',
}

export function StatusDot({ health }: StatusDotProps) {
  return (
    <span className="relative inline-flex h-2.5 w-2.5" aria-hidden="true">
      <span className={cn('absolute inline-flex h-full w-full animate-ping rounded-full opacity-40', healthClasses[health])} />
      <span className={cn('relative inline-flex h-2.5 w-2.5 rounded-full', healthClasses[health])} />
    </span>
  )
}
