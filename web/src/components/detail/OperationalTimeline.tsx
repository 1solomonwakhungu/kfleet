import { Button, Label, Select, Spinner, Text, Timeline } from '@primer/react'
import { SkeletonText } from '@primer/react/experimental'
import {
  AlertIcon,
  ClockIcon,
  PlugIcon,
  PulseIcon,
  ServerIcon,
  ShieldCheckIcon,
  SyncIcon,
  type Icon,
} from '@primer/octicons-react'

import { useTimeline, type TimelineRange } from '../../hooks/useTimeline'
import type { OperationalEvent, OperationalEventKind } from '../../types/timeline'
import { ResourceState } from './ResourceTabState'
import resource from './resource.module.css'
import styles from './OperationalTimeline.module.css'

type BadgeTone = 'accent' | 'success' | 'attention' | 'danger' | 'done'

const EVENT_PRESENTATION: Record<OperationalEventKind, { label: string; icon: Icon; tone: BadgeTone }> = {
  cluster_registered: { label: 'Registered', icon: ServerIcon, tone: 'accent' },
  agent_approved: { label: 'Approved', icon: ShieldCheckIcon, tone: 'success' },
  heartbeat_state_change: { label: 'Heartbeat', icon: PulseIcon, tone: 'attention' },
  version_changed: { label: 'Version', icon: SyncIcon, tone: 'done' },
  agent_reconnected: { label: 'Reconnected', icon: PlugIcon, tone: 'success' },
  agent_disconnected: { label: 'Disconnected', icon: PlugIcon, tone: 'danger' },
  policy_finding: { label: 'Policy finding', icon: AlertIcon, tone: 'attention' },
}

const RANGE_LABELS: Record<TimelineRange, string> = {
  '24h': 'Last 24 hours',
  '7d': 'Last 7 days',
  '30d': 'Last 30 days',
  '90d': 'Last 90 days',
  all: 'All retained',
}

function eventTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return { relative: 'Unknown time', exact: undefined, dateTime: undefined }
  }
  const seconds = Math.max(0, Math.floor((Date.now() - date.getTime()) / 1000))
  const relative =
    seconds < 60
      ? 'Just now'
      : seconds < 3600
        ? `${Math.floor(seconds / 60)}m ago`
        : seconds < 86400
          ? `${Math.floor(seconds / 3600)}h ago`
          : `${Math.floor(seconds / 86400)}d ago`

  return { relative, exact: date.toLocaleString(), dateTime: date.toISOString() }
}

function visibleDetails(event: OperationalEvent) {
  const preferred = ['from', 'to', 'severity', 'ruleId', 'resource', 'reason', 'lastHeartbeat']
  return preferred
    .filter((key) => event.details?.[key])
    .map((key) => [key, event.details?.[key] as string] as const)
}

export function OperationalTimeline({ clusterId }: { clusterId: string }) {
  const timeline = useTimeline(clusterId)

  return (
    <section className={resource.panel} aria-labelledby="operational-timeline-heading">
      <div className={styles.header}>
        <div>
          <h3 id="operational-timeline-heading" className={styles.title}>
            <ClockIcon size={16} className={resource.panelIcon} />
            Operational timeline
          </h3>
          <Text className={styles.subtitle}>Durable lifecycle, connectivity, version, and policy history.</Text>
        </div>

        <div className={styles.headerActions}>
          <Select
            aria-label="Timeline time range"
            value={timeline.range}
            onChange={(event) => timeline.setRange(event.target.value as TimelineRange)}
          >
            {(Object.keys(RANGE_LABELS) as TimelineRange[]).map((range) => (
              <Select.Option key={range} value={range}>
                {RANGE_LABELS[range]}
              </Select.Option>
            ))}
          </Select>
          <Button leadingVisual={SyncIcon} disabled={timeline.loading} onClick={() => void timeline.refresh()}>
            Refresh
          </Button>
        </div>
      </div>

      {timeline.error && timeline.events.length > 0 && (
        <div className={styles.inlineError} role="alert">
          Newer timeline data could not be loaded. {timeline.error}
        </div>
      )}

      {timeline.loading && timeline.events.length === 0 ? (
        <TimelineSkeleton />
      ) : timeline.error && timeline.events.length === 0 ? (
        <div className={resource.state}>
          <ResourceState kind="error" title="Unable to load operational history" description={timeline.error} />
        </div>
      ) : timeline.events.length === 0 ? (
        <div className={resource.state}>
          <ResourceState
            kind="empty"
            title="No operational events in this range"
            description="Choose a longer time range or wait for the next registration, heartbeat transition, version update, reconnect, or policy finding."
          />
        </div>
      ) : (
        <>
          <div className={styles.timeline} aria-live="polite">
            <Timeline>
              {timeline.events.map((event) => (
                <TimelineEventRow key={event.id} event={event} />
              ))}
            </Timeline>
          </div>

          <div className={styles.footer}>
            <span className={`${resource.mono} ${resource.muted}`}>
              {timeline.events.length.toLocaleString()} loaded
            </span>
            {timeline.hasMore ? (
              <Button disabled={timeline.loadingMore} onClick={() => void timeline.loadMore()}>
                {timeline.loadingMore ? 'Loading…' : 'Load older events'}
              </Button>
            ) : (
              <span className={resource.muted}>End of retained history</span>
            )}
          </div>
        </>
      )}
    </section>
  )
}

function TimelineEventRow({ event }: { event: OperationalEvent }) {
  const presentation = EVENT_PRESENTATION[event.kind]
  const EventIcon = presentation.icon
  const timestamp = eventTime(event.occurredAt)
  const details = visibleDetails(event)

  return (
    <Timeline.Item>
      <Timeline.Badge>
        <EventIcon size={16} />
      </Timeline.Badge>
      <Timeline.Body>
        <div className={styles.eventHead}>
          <Label variant={presentation.tone}>{presentation.label}</Label>
          {event.kind === 'policy_finding' && event.details?.severity && (
            <span className={styles.severity}>{event.details.severity}</span>
          )}
          <time dateTime={timestamp.dateTime} title={timestamp.exact} className={styles.timestamp}>
            {timestamp.relative}
          </time>
        </div>

        <Text className={styles.message}>{event.message}</Text>

        {details.length > 0 && (
          <dl className={styles.details}>
            {details.map(([key, value]) => (
              <div key={key} className={styles.detail}>
                <dt>{key}</dt>
                <dd>{value}</dd>
              </div>
            ))}
          </dl>
        )}
      </Timeline.Body>
    </Timeline.Item>
  )
}

function TimelineSkeleton() {
  return (
    <div className={styles.skeleton} aria-busy="true" aria-label="Loading operational timeline">
      <Spinner size="small" />
      {Array.from({ length: 4 }, (_, index) => (
        <SkeletonText key={index} size="bodyMedium" maxWidth="24rem" />
      ))}
    </div>
  )
}
