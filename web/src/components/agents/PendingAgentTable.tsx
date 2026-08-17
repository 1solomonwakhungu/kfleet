import { Button, Label, Spinner, Text } from '@primer/react'
import { CheckIcon } from '@primer/octicons-react'

import type { PendingAgent } from '../../lib/pendingAgentsApi'
import layout from '../../styles/layout.module.css'
import styles from './PendingAgentTable.module.css'

interface PendingAgentTableProps {
  agents: PendingAgent[]
  approvingIds: ReadonlySet<string>
  errors: Readonly<Record<string, string>>
  canApprove: boolean
  onApprove: (agent: PendingAgent) => void
}

const registeredAtFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: 'medium',
  timeStyle: 'short',
})

function RegisteredAt({ value }: { value?: string }) {
  if (!value) return <span className={layout.muted}>—</span>

  const timestamp = Date.parse(value)
  if (Number.isNaN(timestamp)) return <span className={layout.muted}>—</span>

  return (
    <time dateTime={value} title={value} className={styles.nowrap}>
      {registeredAtFormatter.format(timestamp)}
    </time>
  )
}

function AgentLabels({ labels }: { labels: Record<string, string> }) {
  const entries = Object.entries(labels).sort(([first], [second]) => first.localeCompare(second))
  if (entries.length === 0) return <span className={layout.muted}>—</span>

  return (
    <div className={styles.labels} aria-label="Agent labels">
      {entries.map(([key, value]) => (
        <Label key={key} variant="accent" title={`${key}=${value}`}>
          {key}={value}
        </Label>
      ))}
    </div>
  )
}

export function PendingAgentTable({ agents, approvingIds, errors, canApprove, onApprove }: PendingAgentTableProps) {
  return (
    <div className={layout.box}>
      <div className={layout.tableScroll}>
        <table className={layout.table} aria-label="Pending agents">
          <thead>
            <tr>
              <th scope="col">Agent</th>
              <th scope="col">Labels</th>
              <th scope="col">Registered</th>
              <th scope="col">Versions</th>
              <th scope="col" className={styles.actionColumn}>
                Action
              </th>
            </tr>
          </thead>
          <tbody>
            {agents.map((agent) => {
              const approving = approvingIds.has(agent.id)
              const itemError = errors[agent.id]
              const errorId = `approval-error-${encodeURIComponent(agent.id)}`

              return (
                <tr key={agent.id}>
                  <td>
                    <Text weight="semibold">{agent.name}</Text>
                    <span className={styles.identifier} title={agent.id}>
                      {agent.id}
                    </span>
                  </td>
                  <td>
                    <AgentLabels labels={agent.labels} />
                  </td>
                  <td>
                    <RegisteredAt value={agent.registeredAt} />
                  </td>
                  <td>
                    <dl className={styles.versions}>
                      {agent.kubernetesVersion && (
                        <div className={styles.version}>
                          <dt className={layout.muted}>Kubernetes</dt>
                          <dd className={styles.versionValue}>{agent.kubernetesVersion}</dd>
                        </div>
                      )}
                      {agent.agentVersion && (
                        <div className={styles.version}>
                          <dt className={layout.muted}>Agent</dt>
                          <dd className={styles.versionValue}>{agent.agentVersion}</dd>
                        </div>
                      )}
                      {!agent.kubernetesVersion && !agent.agentVersion && (
                        <div>
                          <dt className={layout.srOnly}>Version information</dt>
                          <dd className={styles.versionValue}>—</dd>
                        </div>
                      )}
                    </dl>
                  </td>
                  <td className={styles.actionColumn}>
                    <Button
                      variant="primary"
                      disabled={approving || !canApprove}
                      aria-label={approving ? `Approving ${agent.name}` : `Approve ${agent.name}`}
                      aria-describedby={itemError ? errorId : undefined}
                      leadingVisual={approving ? undefined : CheckIcon}
                      onClick={() => onApprove(agent)}
                    >
                      {approving && <Spinner size="small" />}
                      {approving ? 'Approving…' : canApprove ? 'Approve' : 'View only'}
                    </Button>
                    {itemError && (
                      <p id={errorId} className={styles.error} role="alert">
                        {itemError}
                      </p>
                    )}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}
