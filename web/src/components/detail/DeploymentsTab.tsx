import { useMemo } from 'react'
import { Label, ProgressBar, Text } from '@primer/react'
import { StackIcon } from '@primer/octicons-react'

import type { DeploymentInfo } from '../../types/resources'
import { ResourceState, ResourceTablePanel, ResourceTableSkeleton } from './ResourceTabState'
import styles from './resource.module.css'

interface DeploymentsTabProps {
  deployments: DeploymentInfo[]
  loading: boolean
  error: string | null
  search: string
}

export function DeploymentsTab({ deployments, loading, error, search }: DeploymentsTabProps) {
  const query = search.trim().toLowerCase()
  const filtered = useMemo(
    () =>
      deployments.filter((deployment) =>
        [deployment.name, deployment.namespace].some((value) => value.toLowerCase().includes(query)),
      ),
    [deployments, query],
  )

  if (error) {
    return <ResourceState kind="error" title="Unable to load deployments" description={error} />
  }
  if (loading && deployments.length === 0) {
    return <ResourceTableSkeleton label="Loading deployments" columns={7} />
  }
  if (filtered.length === 0) {
    return (
      <ResourceState
        kind="empty"
        title={query ? 'No matching deployments' : 'No deployments found'}
        description={
          query
            ? `No deployment name or namespace matches “${search.trim()}”.`
            : 'No deployments were returned for this namespace.'
        }
      />
    )
  }

  return (
    <ResourceTablePanel label="Deployments" count={filtered.length} noun="deployment">
      <table className={styles.table}>
        <caption className={styles.srOnly}>
          Deployments and their desired, ready, updated, and available replica counts
        </caption>
        <thead>
          <tr>
            <th scope="col">Name</th>
            <th scope="col">Namespace</th>
            <th scope="col">Status</th>
            <th scope="col">Readiness</th>
            <th scope="col" className={styles.numeric}>
              Updated
            </th>
            <th scope="col" className={styles.numeric}>
              Available
            </th>
            <th scope="col" className={styles.numeric}>
              Age
            </th>
          </tr>
        </thead>
        <tbody>
          {filtered.map((deployment) => {
            const scaledToZero = deployment.desiredReplicas === 0
            const ready = !scaledToZero && deployment.readyReplicas >= deployment.desiredReplicas
            const unavailable = !scaledToZero && deployment.readyReplicas === 0
            const readiness = scaledToZero
              ? 0
              : Math.min(100, Math.round((deployment.readyReplicas / deployment.desiredReplicas) * 100))
            const status = scaledToZero
              ? 'Scaled to zero'
              : ready
                ? 'Ready'
                : unavailable
                  ? 'Unavailable'
                  : 'Progressing'
            const statusVariant = scaledToZero ? 'secondary' : ready ? 'success' : unavailable ? 'danger' : 'attention'
            const barColor = ready
              ? 'var(--bgColor-success-emphasis)'
              : unavailable
                ? 'var(--bgColor-danger-emphasis)'
                : 'var(--bgColor-attention-emphasis)'

            return (
              <tr key={`${deployment.namespace}/${deployment.name}`}>
                <td>
                  <span className={styles.readiness}>
                    <StackIcon size={16} className={styles.panelIcon} />
                    <Text weight="semibold" className={styles.nameCell}>
                      {deployment.name}
                    </Text>
                  </span>
                </td>
                <td className={styles.muted}>{deployment.namespace}</td>
                <td>
                  <Label variant={statusVariant}>{status}</Label>
                </td>
                <td>
                  <div className={styles.readinessCell}>
                    <span className={styles.mono}>
                      {deployment.readyReplicas}/{deployment.desiredReplicas}
                    </span>
                    <ProgressBar
                      progress={readiness}
                      barSize="small"
                      bg={barColor}
                      aria-label={`${deployment.name} readiness`}
                      className={styles.progress}
                    />
                  </div>
                </td>
                <td
                  className={`${styles.numeric} ${
                    deployment.updatedReplicas < deployment.desiredReplicas ? styles.attention : styles.muted
                  }`}
                >
                  {deployment.updatedReplicas}
                </td>
                <td
                  className={`${styles.numeric} ${
                    deployment.availableReplicas < deployment.desiredReplicas ? styles.attention : styles.muted
                  }`}
                >
                  {deployment.availableReplicas}
                </td>
                <td className={`${styles.numeric} ${styles.muted}`}>{deployment.age || '—'}</td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </ResourceTablePanel>
  )
}
