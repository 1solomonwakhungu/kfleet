import { useMemo } from 'react'
import { Label, Text } from '@primer/react'
import { FileIcon, SyncIcon } from '@primer/octicons-react'

import type { PodInfo } from '../../types/resources'
import { ResourceState, ResourceTablePanel, ResourceTableSkeleton } from './ResourceTabState'
import styles from './resource.module.css'

type LabelVariant = 'success' | 'attention' | 'danger' | 'accent' | 'secondary'

const phaseVariants: Record<string, LabelVariant> = {
  Running: 'success',
  Pending: 'attention',
  Failed: 'danger',
  Succeeded: 'accent',
  Unknown: 'secondary',
}

function age(iso: string): string {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime()) || date.getTime() === 0) return '—'
  const seconds = Math.max(0, Math.floor((Date.now() - date.getTime()) / 1000))
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h`
  return `${Math.floor(hours / 24)}d`
}

interface PodsTabProps {
  pods: PodInfo[]
  loading: boolean
  error: string | null
  search: string
  onSelectPod?: (pod: PodInfo) => void
}

export function PodsTab({ pods, loading, error, search, onSelectPod }: PodsTabProps) {
  const query = search.trim().toLowerCase()
  const filtered = useMemo(
    () =>
      pods.filter((pod) =>
        [pod.name, pod.namespace, pod.nodeName, pod.phase].some((value) => value.toLowerCase().includes(query)),
      ),
    [pods, query],
  )

  if (error) {
    return <ResourceState kind="error" title="Unable to load pods" description={error} />
  }
  if (loading && pods.length === 0) {
    return <ResourceTableSkeleton label="Loading pods" columns={7} />
  }
  if (filtered.length === 0) {
    return (
      <ResourceState
        kind="empty"
        title={query ? 'No matching pods' : 'No pods found'}
        description={
          query
            ? `No pod matches “${search.trim()}”. Try a name, namespace, node, or phase.`
            : 'No pods were returned for this namespace.'
        }
      />
    )
  }

  return (
    <ResourceTablePanel label="Pods" count={filtered.length} noun="pod">
      <table className={styles.table}>
        <caption className={styles.srOnly}>
          Pods, their scheduling location, phase, readiness, restarts, and age
        </caption>
        <thead>
          <tr>
            <th scope="col">Name</th>
            <th scope="col">Namespace</th>
            <th scope="col">Phase</th>
            <th scope="col">Ready</th>
            <th scope="col" className={styles.numeric}>
              Restarts
            </th>
            <th scope="col">Node</th>
            <th scope="col" className={styles.numeric}>
              Age
            </th>
          </tr>
        </thead>
        <tbody>
          {filtered.map((pod) => (
            <tr key={`${pod.namespace}/${pod.name}`}>
              <td>
                {onSelectPod ? (
                  <button
                    type="button"
                    className={styles.linkButton}
                    onClick={() => onSelectPod(pod)}
                    aria-label={`View logs for pod ${pod.name} in namespace ${pod.namespace}`}
                  >
                    <span className={styles.nameCell}>{pod.name}</span>
                    <FileIcon size={12} />
                  </button>
                ) : (
                  <Text weight="semibold">{pod.name}</Text>
                )}
              </td>
              <td className={styles.muted}>{pod.namespace}</td>
              <td>
                <Label variant={phaseVariants[pod.phase] ?? 'secondary'}>{pod.phase || 'Unknown'}</Label>
              </td>
              <td>
                <span className={`${styles.readiness} ${pod.ready ? '' : styles.attention}`}>
                  {pod.ready ? 'Ready' : 'Not ready'}
                </span>
              </td>
              <td className={`${styles.numeric} ${pod.restartCount > 0 ? styles.attention : ''}`}>
                <span className={styles.restart}>
                  {pod.restartCount > 0 && <SyncIcon size={12} />}
                  {pod.restartCount}
                </span>
              </td>
              <td className={`${styles.mono} ${styles.muted}`} title={pod.nodeName || undefined}>
                <span className={styles.nameCell}>{pod.nodeName || 'Unscheduled'}</span>
              </td>
              <td className={`${styles.numeric} ${styles.muted}`}>{age(pod.startTime)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </ResourceTablePanel>
  )
}
