import type { ClusterHealth } from './cluster'

export type AlertSeverity = 'warning' | 'critical'
export type AlertStatus = 'firing' | 'acknowledged' | 'resolved'
export type AlertDeliveryStatus = 'pending' | 'retrying' | 'delivered' | 'dead_letter' | 'disabled'

export interface Alert {
  id: string
  ruleId: string
  ruleName: string
  clusterId: string
  clusterName: string
  dedupeKey: string
  health: ClusterHealth
  severity: AlertSeverity
  summary: string
  status: AlertStatus
  triggeredAt: string
  updatedAt: string
  acknowledgedAt?: string
  acknowledgedBy?: string
  resolvedAt?: string
  deliveryStatus: AlertDeliveryStatus
  deliveryAttempts: number
  nextDeliveryAt?: string
  lastDeliveryError?: string
  deliveredAt?: string
  deadLetteredAt?: string
}
