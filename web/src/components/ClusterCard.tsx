import { Label, Text } from '@primer/react'
import { ArrowUpRightIcon, ClockIcon } from '@primer/octicons-react'

import { HealthLabel } from './HealthLabel'
import { StatusDot } from './StatusDot'
import { timeAgo } from '../lib/utils'
import type { Cluster } from '../types/cluster'
import styles from './ClusterCard.module.css'

interface ClusterCardProps {
  cluster: Cluster
  onClick: () => void
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
      className={`${styles.card} ${styles[cluster.health]}`}
      aria-label={`Open ${cluster.name} cluster, health ${cluster.health}`}
      onClick={onClick}
    >
      <div className={styles.body}>
        <div>
          <div className={styles.heading}>
            <span className={styles.name}>
              <StatusDot health={cluster.health} />
              <span className={styles.clusterName} title={cluster.name}>
                {cluster.name}
              </span>
            </span>
            <HealthLabel health={cluster.health} />
          </div>
          {cluster.id !== cluster.name && (
            <span className={styles.identifier} title={cluster.id}>
              {cluster.id}
            </span>
          )}
        </div>

        <dl className={styles.metrics}>
          <Metric label="Nodes" value={cluster.nodeCount.toLocaleString()} />
          <Metric label="Pods" value={cluster.podCount.toLocaleString()} />
          <Metric label="Kubernetes" value={cluster.k8sVersion || 'Unknown'} mono />
          <Metric label="Agent" value={cluster.agentVersion || 'Unknown'} mono />
        </dl>

        <div>
          <span className={styles.caption}>Labels</span>
          {visibleLabels.length > 0 ? (
            <div className={styles.labels} aria-label={`${labels.length} cluster labels`}>
              {visibleLabels.map(([key, value]) => (
                <Label key={key} variant="secondary" title={`${key}=${value}`}>
                  {key}={value}
                </Label>
              ))}
              {hiddenLabelCount > 0 && <Label variant="secondary">+{hiddenLabelCount}</Label>}
            </div>
          ) : (
            <div className={styles.labels}>
              <span className={styles.caption}>No labels reported</span>
            </div>
          )}
        </div>

        <div className={styles.footer}>
          <span className={styles.heartbeat}>
            <ClockIcon size={12} />
            <span className={styles.clusterName}>
              <Text weight="semibold">{heartbeat.freshness}</Text>
              {' · '}
              <time dateTime={heartbeat.dateTime} title={heartbeat.exact}>
                {heartbeat.relative}
              </time>
            </span>
          </span>
          <ArrowUpRightIcon size={16} className={styles.arrow} />
        </div>
      </div>
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
    <div className={styles.metric}>
      <dt>{label}</dt>
      <dd className={mono ? styles.metricMono : undefined} title={value}>
        {value}
      </dd>
    </div>
  )
}
