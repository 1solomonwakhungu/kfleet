import { cn } from '../lib/utils'
import type { ClusterHealth } from '../types/cluster'

interface StatusDotProps {
  health: ClusterHealth
}

const healthClasses: Record<ClusterHealth, string> = {
  healthy: 'bg-healthy ring-healthy/20',
  degraded: 'bg-degraded ring-degraded/20',
  unreachable: 'bg-unreachable ring-unreachable/20',
  unknown: 'bg-unknown ring-unknown/20',
}

export function StatusDot({ health }: StatusDotProps) {
  return <span className={cn('inline-flex h-2.5 w-2.5 shrink-0 rounded-full ring-4', healthClasses[health])} aria-hidden="true" />
}
