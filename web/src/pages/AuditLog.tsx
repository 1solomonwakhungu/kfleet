import { useMemo, useState } from 'react'
import { Button, Flash, Heading, Label, Select, Text, TextInput } from '@primer/react'
import { Blankslate, SkeletonText } from '@primer/react/experimental'
import { LogIcon, SearchIcon, SyncIcon } from '@primer/octicons-react'

import { useAuth } from '../auth/AuthContext'
import { PermissionNotice } from '../components/admin/PermissionNotice'
import { useAuditEvents } from '../hooks/useAuditEvents'
import type { AuditEvent, AuditOutcome } from '../types/admin'
import layout from '../styles/layout.module.css'
import styles from './AuditLog.module.css'

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
      <main className={layout.page}>
        <AuditHeader loading={false} />
        <PermissionNotice
          title="Admin access required"
          description="The audit log records security-relevant actions and is restricted to admins."
        />
      </main>
    )
  }

  return (
    <main className={layout.page}>
      <AuditHeader loading={loading} onRefresh={() => void reload()} />

      <div aria-live="polite">
        {error && (
          <Flash variant="danger" role="alert">
            <div className={styles.flashBody}>
              <div>
                <Text weight="semibold">Audit events could not be loaded.</Text>
                <Text className={layout.pageDescription}>{error}</Text>
              </div>
              <Button disabled={loading} onClick={() => void reload()}>
                Retry
              </Button>
            </div>
          </Flash>
        )}
      </div>

      <section className={styles.list} aria-busy={loading} aria-labelledby="audit-list-title">
        <div className={styles.listHeader}>
          <Heading as="h2" variant="small" id="audit-list-title">
            Recent activity
          </Heading>
          <div className={styles.filters}>
            <TextInput
              aria-label="Filter audit events"
              placeholder="Filter by actor, action, or target"
              leadingVisual={SearchIcon}
              value={query}
              onChange={(event) => setQuery(event.target.value)}
            />
            <Select
              aria-label="Outcome"
              value={outcome}
              onChange={(event) => setOutcome(event.target.value as OutcomeFilter)}
            >
              <Select.Option value="all">All outcomes</Select.Option>
              <Select.Option value="success">Success</Select.Option>
              <Select.Option value="failure">Failure</Select.Option>
            </Select>
          </div>
        </div>

        {loading && events.length === 0 ? (
          <AuditSkeleton />
        ) : filtered.length > 0 ? (
          <>
            <div className={layout.box}>
              <div className={layout.tableScroll}>
                <table className={layout.table}>
                  <thead>
                    <tr>
                      <th scope="col">When</th>
                      <th scope="col">Actor</th>
                      <th scope="col">Action</th>
                      <th scope="col">Target</th>
                      <th scope="col">Outcome</th>
                      <th scope="col">Details</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((event) => (
                      <tr key={event.id}>
                        <td className={`${layout.muted} ${styles.nowrap}`}>{formatTimestamp(event.occurredAt)}</td>
                        <td>
                          <Text weight="semibold" className={styles.block}>
                            {event.actorUsername || 'system'}
                          </Text>
                          {event.sourceIp && <span className={`${styles.meta} ${layout.mono}`}>{event.sourceIp}</span>}
                        </td>
                        <td className={layout.mono}>{event.action}</td>
                        <td>
                          <span className={styles.meta}>{event.targetType}</span>
                          <span className={`${styles.meta} ${layout.mono} ${styles.wrap}`}>
                            {event.targetId || '—'}
                          </span>
                        </td>
                        <td>
                          <Label variant={event.outcome === 'success' ? 'success' : 'danger'}>{event.outcome}</Label>
                        </td>
                        <td className={`${layout.muted} ${styles.wrap}`}>{event.details || '—'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            <div className={styles.footer}>
              <Text size="small" className={`${layout.mono} ${layout.muted}`}>
                {filtered.length} of {events.length} loaded
              </Text>
              {hasMore && (
                <Button disabled={loading} onClick={loadMore}>
                  {loading ? 'Loading…' : 'Load more'}
                </Button>
              )}
            </div>
          </>
        ) : !error ? (
          <div className={layout.box}>
            <Blankslate>
              <Blankslate.Visual>
                <LogIcon size={24} />
              </Blankslate.Visual>
              <Blankslate.Heading as="h3">
                {events.length === 0 ? 'No audit events yet' : 'No events match these filters'}
              </Blankslate.Heading>
              <Blankslate.Description>
                {events.length === 0
                  ? 'Sign-ins, user changes, agent approvals, and cluster removals will appear here.'
                  : 'Clear the filter or choose a different outcome to see more activity.'}
              </Blankslate.Description>
            </Blankslate>
          </div>
        ) : null}
      </section>
    </main>
  )
}

function AuditHeader({ loading, onRefresh }: { loading: boolean; onRefresh?: () => void }) {
  return (
    <header className={layout.pageHeader}>
      <div className={layout.pageHeaderText}>
        <Heading as="h1" variant="large">
          Audit log
        </Heading>
        <Text className={layout.pageDescription}>
          An immutable record of security-relevant actions, newest first.
        </Text>
      </div>
      {onRefresh && (
        <Button leadingVisual={SyncIcon} disabled={loading} onClick={onRefresh}>
          {loading ? 'Refreshing…' : 'Refresh'}
        </Button>
      )}
    </header>
  )
}

function AuditSkeleton() {
  return (
    <div className={`${layout.box} ${styles.skeleton}`} aria-label="Loading audit events">
      <SkeletonText size="titleSmall" maxWidth="14rem" />
      {Array.from({ length: 4 }, (_, index) => (
        <SkeletonText key={index} size="bodyMedium" maxWidth="85%" />
      ))}
    </div>
  )
}

export default AuditLogPage
