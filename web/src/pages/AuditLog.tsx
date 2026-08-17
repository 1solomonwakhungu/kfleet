import { useMemo, useState } from 'react'
import { LoaderCircle, RefreshCw, ScrollText } from 'lucide-react'

import { useAuth } from '../auth/AuthContext'
import { PermissionNotice } from '../components/admin/PermissionNotice'
import { Badge } from '../components/ui/badge'
import { Button } from '../components/ui/button'
import { Card, CardContent } from '../components/ui/card'
import { Input } from '../components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../components/ui/table'
import { useAuditEvents } from '../hooks/useAuditEvents'
import type { AuditEvent, AuditOutcome } from '../types/admin'

type OutcomeFilter = 'all' | AuditOutcome

function formatTimestamp(value: string): string {
  const parsed = new Date(value)
  return Number.isNaN(parsed.getTime()) ? '—' : parsed.toLocaleString()
}

function matches(event: AuditEvent, query: string): boolean {
  if (!query) return true
  const needle = query.toLowerCase()
  return [event.action, event.actorUsername, event.targetType, event.targetId, event.details, event.sourceIp]
    .filter((value): value is string => Boolean(value))
    .some((value) => value.toLowerCase().includes(needle))
}

export function AuditLogPage() {
  const { user } = useAuth()
  const isAdmin = user?.role === 'admin'
  const { events, loading, error, hasMore, reload, loadMore } = useAuditEvents(isAdmin)
  const [query, setQuery] = useState('')
  const [outcome, setOutcome] = useState<OutcomeFilter>('all')

  const filtered = useMemo(
    () => events.filter((event) => (outcome === 'all' || event.outcome === outcome) && matches(event, query)),
    [events, outcome, query],
  )

  if (!isAdmin) {
    return (
      <main className="mx-auto min-h-dvh max-w-[100rem] px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
        <AuditHeader loading={false} />
        <div className="mt-7">
          <PermissionNotice
            title="Admin access required"
            description="The audit log records security-relevant actions and is restricted to admins."
          />
        </div>
      </main>
    )
  }

  return (
    <main className="mx-auto min-h-dvh max-w-[100rem] px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <AuditHeader loading={loading} onRefresh={() => void reload()} />

      <div className="mt-6" aria-live="polite">
        {error && (
          <section
            className="flex flex-col gap-3 rounded-lg bg-danger-soft p-4 text-danger sm:flex-row sm:items-center sm:justify-between"
            role="alert"
          >
            <div>
              <p className="font-semibold">Audit events could not be loaded.</p>
              <p className="mt-1 text-sm">{error}</p>
            </div>
            <Button variant="outline" size="sm" disabled={loading} onClick={() => void reload()}>
              Retry
            </Button>
          </section>
        )}
      </div>

      <section className="mt-7" aria-busy={loading} aria-labelledby="audit-list-title">
        <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <h2 id="audit-list-title" className="font-display text-lg font-bold">
            Recent activity
          </h2>
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
            <div>
              <label className="sr-only" htmlFor="audit-search">
                Filter audit events
              </label>
              <Input
                id="audit-search"
                className="h-9 sm:w-64"
                placeholder="Filter by actor, action, or target"
                value={query}
                onChange={(event) => setQuery(event.target.value)}
              />
            </div>
            <div>
              <label className="sr-only" htmlFor="audit-outcome">
                Outcome
              </label>
              <select
                id="audit-outcome"
                className="h-9 rounded-md border border-border bg-background px-2 text-sm text-foreground"
                value={outcome}
                onChange={(event) => setOutcome(event.target.value as OutcomeFilter)}
              >
                <option value="all">All outcomes</option>
                <option value="success">Success</option>
                <option value="failure">Failure</option>
              </select>
            </div>
          </div>
        </div>

        {loading && events.length === 0 ? (
          <AuditSkeleton />
        ) : filtered.length > 0 ? (
          <>
            <Card className="ring-1 ring-inset ring-border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead scope="col">When</TableHead>
                    <TableHead scope="col">Actor</TableHead>
                    <TableHead scope="col">Action</TableHead>
                    <TableHead scope="col">Target</TableHead>
                    <TableHead scope="col">Outcome</TableHead>
                    <TableHead scope="col">Details</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filtered.map((event) => (
                    <TableRow key={event.id}>
                      <TableCell className="whitespace-nowrap text-muted">
                        {formatTimestamp(event.occurredAt)}
                      </TableCell>
                      <TableCell>
                        <span className="block font-semibold">{event.actorUsername || 'system'}</span>
                        {event.sourceIp && (
                          <span className="block font-mono text-xs text-muted">{event.sourceIp}</span>
                        )}
                      </TableCell>
                      <TableCell className="font-mono text-xs">{event.action}</TableCell>
                      <TableCell>
                        <span className="block text-xs text-muted">{event.targetType}</span>
                        <span className="block break-all font-mono text-xs">{event.targetId || '—'}</span>
                      </TableCell>
                      <TableCell>
                        <Badge
                          className={
                            event.outcome === 'success'
                              ? 'bg-elevated text-healthy'
                              : 'bg-danger-soft text-danger'
                          }
                        >
                          {event.outcome}
                        </Badge>
                      </TableCell>
                      <TableCell className="break-words text-muted">{event.details || '—'}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </Card>

            <div className="mt-4 flex items-center justify-between gap-4">
              <span className="font-mono text-sm text-muted">
                {filtered.length} of {events.length} loaded
              </span>
              {hasMore && (
                <Button variant="outline" size="sm" disabled={loading} onClick={loadMore}>
                  {loading ? 'Loading…' : 'Load more'}
                </Button>
              )}
            </div>
          </>
        ) : !error ? (
          <Card className="ring-1 ring-inset ring-border">
            <CardContent className="grid min-h-64 place-items-center p-6 text-center">
              <div>
                <span className="mx-auto grid size-12 place-items-center rounded-full bg-elevated text-muted ring-1 ring-inset ring-border">
                  <ScrollText className="size-6" aria-hidden="true" />
                </span>
                <p className="mt-4 font-display text-xl font-bold">
                  {events.length === 0 ? 'No audit events yet' : 'No events match these filters'}
                </p>
                <p className="mt-2 text-muted">
                  {events.length === 0
                    ? 'Sign-ins, user changes, agent approvals, and cluster removals will appear here.'
                    : 'Clear the filter or choose a different outcome to see more activity.'}
                </p>
              </div>
            </CardContent>
          </Card>
        ) : null}
      </section>
    </main>
  )
}

function AuditHeader({ loading, onRefresh }: { loading: boolean; onRefresh?: () => void }) {
  return (
    <header className="flex flex-col gap-5 border-b border-border pb-7 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <p className="font-mono text-sm text-blue-400">kfleet admin</p>
        <h1 className="mt-2 font-display text-3xl font-bold tracking-tight sm:text-4xl">Audit log</h1>
        <p className="mt-2 max-w-2xl text-muted">
          An immutable record of security-relevant actions, newest first.
        </p>
      </div>
      {onRefresh && (
        <Button variant="outline" size="sm" disabled={loading} onClick={onRefresh}>
          {loading ? (
            <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />
          ) : (
            <RefreshCw className="size-4" aria-hidden="true" />
          )}
          {loading ? 'Refreshing…' : 'Refresh'}
        </Button>
      )}
    </header>
  )
}

function AuditSkeleton() {
  return (
    <Card className="animate-pulse p-5 ring-1 ring-inset ring-border" aria-label="Loading audit events">
      <div className="h-5 w-48 rounded bg-elevated" />
      {Array.from({ length: 4 }, (_, index) => (
        <div key={index} className="mt-5 flex items-center gap-6 border-t border-border pt-5">
          <div className="h-6 w-1/4 rounded bg-elevated" />
          <div className="h-6 w-1/3 rounded bg-elevated" />
          <div className="h-6 w-1/5 rounded bg-elevated" />
        </div>
      ))}
    </Card>
  )
}

export default AuditLogPage
