import { CounterLabel } from '@primer/react'
import { ServerIcon, ShieldCheckIcon, StackIcon, WorkflowIcon } from '@primer/octicons-react'

import { StatusDot } from '../StatusDot'
import type { Cluster, ClusterHealth } from '../../types/cluster'
import layout from '../../styles/layout.module.css'
import styles from './FleetSummary.module.css'

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
    <section aria-labelledby="fleet-summary-heading">
      <h2 id="fleet-summary-heading" className={layout.srOnly}>
        Fleet summary
      </h2>

      <div className={`${layout.grid} ${layout.grid4}`}>
        <SummaryMetric
          label="Clusters"
          value={clusters.length.toLocaleString()}
          detail={attentionCount === 0 ? 'None need attention' : `${attentionCount.toLocaleString()} need attention`}
          icon={WorkflowIcon}
        />
        <SummaryMetric
          label="Healthy"
          value={healthyPercent === null ? '—' : `${healthyPercent}%`}
          detail={`${counts.healthy.toLocaleString()} of ${clusters.length.toLocaleString()} clusters`}
          icon={ShieldCheckIcon}
          tone="success"
        />
        <SummaryMetric label="Nodes" value={nodeCount.toLocaleString()} detail="Across the fleet" icon={ServerIcon} />
        <SummaryMetric label="Pods" value={podCount.toLocaleString()} detail="Reported workload pods" icon={StackIcon} />
      </div>

      <div className={`${layout.box} ${layout.boxBody} ${styles.breakdown}`}>
        <dl className={styles.breakdownList} aria-label="Cluster health counts">
          {(['healthy', 'degraded', 'unreachable', 'unknown'] as const).map((health) => (
            <div key={health} className={styles.breakdownItem}>
              <StatusDot health={health} />
              <dt className={styles.breakdownTerm}>{health}</dt>
              <dd className={styles.breakdownValue}>
                <CounterLabel>{counts[health]}</CounterLabel>
              </dd>
            </div>
          ))}
        </dl>
      </div>
    </section>
  )
}

interface SummaryMetricProps {
  label: string
  value: string
  detail: string
  icon: typeof ServerIcon
  tone?: 'success'
}

function SummaryMetric({ label, value, detail, icon: Icon, tone }: SummaryMetricProps) {
  return (
    <div className={`${layout.box} ${layout.boxBody}`}>
      <div className={styles.metricHead}>
        <div>
          <span className={layout.metricLabel}>{label}</span>
          <span className={`${layout.metricValue} ${tone === 'success' ? layout.metricSuccess : ''}`}>{value}</span>
        </div>
        <Icon size={16} className={tone === 'success' ? styles.iconSuccess : styles.iconMuted} />
      </div>
      <span className={layout.metricDetail}>{detail}</span>
    </div>
  )
}
