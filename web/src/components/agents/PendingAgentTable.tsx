import { Check, LoaderCircle } from 'lucide-react'

import { Badge } from '../ui/badge'
import { Button } from '../ui/button'
import { Card } from '../ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '../ui/table'
import type { PendingAgent } from '../../lib/pendingAgentsApi'

interface PendingAgentTableProps {
  agents: PendingAgent[]
  approvingIds: ReadonlySet<string>
  errors: Readonly<Record<string, string>>
  onApprove: (agent: PendingAgent) => void
}

const registeredAtFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: 'medium',
  timeStyle: 'short',
})

function RegisteredAt({ value }: { value?: string }) {
  if (!value) return <span className="text-muted">—</span>

  const timestamp = Date.parse(value)
  if (Number.isNaN(timestamp)) return <span className="text-muted">—</span>

  return (
    <time dateTime={value} title={value} className="whitespace-nowrap">
      {registeredAtFormatter.format(timestamp)}
    </time>
  )
}

function AgentLabels({ labels }: { labels: Record<string, string> }) {
  const entries = Object.entries(labels).sort(([first], [second]) => first.localeCompare(second))
  if (entries.length === 0) return <span className="text-muted">—</span>

  return (
    <div className="flex min-w-48 flex-wrap gap-1.5" aria-label="Agent labels">
      {entries.map(([key, value]) => (
        <Badge key={key} className="max-w-64 bg-blue-950 text-blue-200 ring-1 ring-inset ring-blue-800">
          <span className="truncate font-mono" title={`${key}=${value}`}>
            {key}={value}
          </span>
        </Badge>
      ))}
    </div>
  )
}

export function PendingAgentTable({ agents, approvingIds, errors, onApprove }: PendingAgentTableProps) {
  return (
    <Card className="overflow-hidden ring-1 ring-inset ring-border">
      <Table aria-label="Pending agents">
        <TableHeader>
          <TableRow>
            <TableHead scope="col">Agent</TableHead>
            <TableHead scope="col">Labels</TableHead>
            <TableHead scope="col">Registered</TableHead>
            <TableHead scope="col">Versions</TableHead>
            <TableHead scope="col" className="text-right">Action</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {agents.map((agent) => {
            const approving = approvingIds.has(agent.id)
            const itemError = errors[agent.id]
            const errorId = `approval-error-${encodeURIComponent(agent.id)}`

            return (
              <TableRow key={agent.id}>
                <TableCell>
                  <div className="min-w-40">
                    <p className="font-semibold text-foreground">{agent.name}</p>
                    <p className="mt-1 max-w-64 truncate font-mono text-xs text-muted" title={agent.id}>
                      {agent.id}
                    </p>
                  </div>
                </TableCell>
                <TableCell>
                  <AgentLabels labels={agent.labels} />
                </TableCell>
                <TableCell>
                  <RegisteredAt value={agent.registeredAt} />
                </TableCell>
                <TableCell>
                  <dl className="min-w-32 space-y-1 text-xs">
                    {agent.kubernetesVersion && (
                      <div className="flex items-baseline justify-between gap-3">
                        <dt className="text-muted">Kubernetes</dt>
                        <dd className="font-mono text-foreground">{agent.kubernetesVersion}</dd>
                      </div>
                    )}
                    {agent.agentVersion && (
                      <div className="flex items-baseline justify-between gap-3">
                        <dt className="text-muted">Agent</dt>
                        <dd className="font-mono text-foreground">{agent.agentVersion}</dd>
                      </div>
                    )}
                    {!agent.kubernetesVersion && !agent.agentVersion && (
                      <div>
                        <dt className="sr-only">Version information</dt>
                        <dd className="text-muted">—</dd>
                      </div>
                    )}
                  </dl>
                </TableCell>
                <TableCell className="min-w-48 text-right">
                  <Button
                    size="sm"
                    disabled={approving}
                    aria-label={approving ? `Approving ${agent.name}` : `Approve ${agent.name}`}
                    aria-describedby={itemError ? errorId : undefined}
                    className="bg-blue-600 text-white hover:bg-blue-500 hover:brightness-100"
                    onClick={() => onApprove(agent)}
                  >
                    {approving ? (
                      <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />
                    ) : (
                      <Check className="size-4" aria-hidden="true" />
                    )}
                    {approving ? 'Approving…' : 'Approve'}
                  </Button>
                  {itemError && (
                    <p id={errorId} className="mt-2 max-w-64 text-left text-xs text-danger" role="alert">
                      {itemError}
                    </p>
                  )}
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </Card>
  )
}
