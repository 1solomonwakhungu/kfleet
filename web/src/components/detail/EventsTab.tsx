import { useMemo } from 'react'
import { Label, Text } from '@primer/react'
import { AlertIcon, BellIcon, InfoIcon } from '@primer/octicons-react'

import type { EventInfo } from '../../types/resources'
import { ResourceState, ResourceTableSkeleton } from './ResourceTabState'
import resource from './resource.module.css'
import styles from './EventsTab.module.css'

function relativeTime(iso: string): string {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime()) || date.getTime() === 0) return '—'
  const seconds = Math.max(0, Math.floor((Date.now() - date.getTime()) / 1000))
  if (seconds < 60) return `${seconds}s ago`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  return `${Math.floor(hours / 24)}d ago`
}

interface EventsTabProps {
  events: EventInfo[]
  loading: boolean
  error: string | null
  search: string
}

export function EventsTab({ events, loading, error, search }: EventsTabProps) {
  const query = search.trim().toLowerCase()
  const sorted = useMemo(
    () =>
      events
        .filter((event) =>
          [event.reason, event.message, event.namespace, event.type].some((value) =>
            value.toLowerCase().includes(query),
          ),
        )
        .slice()
        .sort((a, b) => new Date(b.lastTimestamp).getTime() - new Date(a.lastTimestamp).getTime()),
    [events, query],
  )
  const warningCount = sorted.filter((event) => event.type.toLowerCase() === 'warning').length

  if (error) {
    return <ResourceState kind="error" title="Unable to load events" description={error} />
  }
  if (loading && events.length === 0) {
    return <ResourceTableSkeleton label="Loading events" columns={4} rows={7} />
  }
  if (sorted.length === 0) {
    return (
      <ResourceState
        kind="empty"
        title={query ? 'No matching events' : 'No events found'}
        description={
          query
            ? `No event reason, message, namespace, or type matches “${search.trim()}”.`
            : 'Kubernetes has not reported any events for this cluster.'
        }
      />
    )
  }

  return (
    <section className={resource.panel} aria-label="Kubernetes events">
      <div className={resource.panelHeader}>
        <BellIcon size={16} className={resource.panelIcon} />
        <span className={resource.panelCount}>{sorted.length}</span>
        <span className={resource.muted}>{sorted.length === 1 ? 'event' : 'events'}</span>
        {warningCount > 0 && (
          <span className={styles.warningCount}>
            <AlertIcon size={12} />
            {warningCount} {warningCount === 1 ? 'warning' : 'warnings'}
          </span>
        )}
      </div>

      <ol className={styles.list}>
        {sorted.map((event, index) => {
          const warning = event.type.toLowerCase() === 'warning'
          const date = new Date(event.lastTimestamp)
          const validDate = !Number.isNaN(date.getTime()) && date.getTime() !== 0

          return (
            <li
              key={`${event.namespace}-${event.reason}-${event.lastTimestamp}-${index}`}
              className={`${styles.item} ${warning ? styles.itemWarning : ''}`}
            >
              <span className={`${styles.badge} ${warning ? styles.badgeWarning : styles.badgeInfo}`}>
                {warning ? <AlertIcon size={16} /> : <InfoIcon size={16} />}
              </span>

              <div className={styles.body}>
                <div className={styles.titleRow}>
                  <Text weight="semibold">{event.reason || 'Unknown reason'}</Text>
                  <Label variant={warning ? 'attention' : 'accent'}>{event.type || 'Normal'}</Label>
                  {event.count > 1 && (
                    <span className={resource.mono} aria-label={`${event.count} occurrences`}>
                      ×{event.count}
                    </span>
                  )}
                </div>
                <Text className={styles.message}>{event.message || 'No event message was provided.'}</Text>
                <span className={`${resource.mono} ${resource.muted}`}>namespace/{event.namespace || 'default'}</span>
              </div>

              <time
                dateTime={validDate ? date.toISOString() : undefined}
                title={validDate ? date.toLocaleString() : undefined}
                className={styles.timestamp}
              >
                {relativeTime(event.lastTimestamp)}
              </time>
            </li>
          )
        })}
      </ol>
    </section>
  )
}
