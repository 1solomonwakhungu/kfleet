import { Label } from '@primer/react'

import type { ClusterHealth } from '../types/cluster'

type LabelVariant = 'success' | 'attention' | 'danger' | 'secondary'

const healthVariants: Record<ClusterHealth, LabelVariant> = {
  healthy: 'success',
  degraded: 'attention',
  unreachable: 'danger',
  unknown: 'secondary',
}

interface HealthLabelProps {
  health: ClusterHealth
}

/** Cluster health rendered as a Primer Label with the matching semantic color. */
export function HealthLabel({ health }: HealthLabelProps) {
  return (
    <Label variant={healthVariants[health]} aria-label={`Cluster health: ${health}`}>
      {health}
    </Label>
  )
}
