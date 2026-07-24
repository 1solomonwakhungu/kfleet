import {
  Activity,
  CalendarRange,
  Clock3,
  RefreshCw,
  Server,
  ShieldAlert,
  ShieldCheck,
  Wifi,
  WifiOff,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { cn } from '@/lib/utils';
import { useTimeline, type TimelineRange } from '@/hooks/useTimeline';
import type { OperationalEvent, OperationalEventKind } from '@/types/timeline';
import { ResourceState } from './ResourceTabState';

const EVENT_PRESENTATION: Record<OperationalEventKind, {
  label: string;
  icon: LucideIcon;
  color: string;
}> = {
  cluster_registered: {
    label: 'Registered',
    icon: Server,
    color: 'border-blue-500/30 bg-blue-500/10 text-blue-300',
  },
  agent_approved: {
    label: 'Approved',
    icon: ShieldCheck,
    color: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300',
  },
  heartbeat_state_change: {
    label: 'Heartbeat',
    icon: Activity,
    color: 'border-amber-500/30 bg-amber-500/10 text-amber-300',
  },
  version_changed: {
    label: 'Version',
    icon: RefreshCw,
    color: 'border-violet-500/30 bg-violet-500/10 text-violet-300',
  },
  agent_reconnected: {
    label: 'Reconnected',
    icon: Wifi,
    color: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300',
  },
  agent_disconnected: {
    label: 'Disconnected',
    icon: WifiOff,
    color: 'border-red-500/30 bg-red-500/10 text-red-300',
  },
  policy_finding: {
    label: 'Policy finding',
    icon: ShieldAlert,
    color: 'border-orange-500/30 bg-orange-500/10 text-orange-300',
  },
};

const RANGE_LABELS: Record<TimelineRange, string> = {
  '24h': 'Last 24 hours',
  '7d': 'Last 7 days',
  '30d': 'Last 30 days',
  '90d': 'Last 90 days',
  all: 'All retained',
};

function eventTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return { relative: 'Unknown time', exact: undefined, dateTime: undefined };
  }
  const seconds = Math.max(0, Math.floor((Date.now() - date.getTime()) / 1000));
  const relative = seconds < 60
    ? 'Just now'
    : seconds < 3600
      ? `${Math.floor(seconds / 60)}m ago`
      : seconds < 86400
        ? `${Math.floor(seconds / 3600)}h ago`
        : `${Math.floor(seconds / 86400)}d ago`;
  return {
    relative,
    exact: date.toLocaleString(),
    dateTime: date.toISOString(),
  };
}

function visibleDetails(event: OperationalEvent) {
  const preferred = ['from', 'to', 'severity', 'ruleId', 'resource', 'reason', 'lastHeartbeat'];
  return preferred
    .filter((key) => event.details?.[key])
    .map((key) => [key, event.details?.[key] as string] as const);
}

export function OperationalTimeline({ clusterId }: { clusterId: string }) {
  const timeline = useTimeline(clusterId);

  return (
    <section className="overflow-hidden rounded-lg border border-border bg-surface" aria-labelledby="operational-timeline-heading">
      <div className="flex flex-col gap-4 border-b border-border px-4 py-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <Clock3 className="h-4 w-4 text-blue-400" aria-hidden="true" />
            <h3 id="operational-timeline-heading" className="font-semibold text-foreground">
              Operational timeline
            </h3>
          </div>
          <p className="mt-1 text-sm text-muted">
            Durable lifecycle, connectivity, version, and policy history.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Select value={timeline.range} onValueChange={(value) => timeline.setRange(value as TimelineRange)}>
            <SelectTrigger className="h-10 w-[11rem]" aria-label="Timeline time range">
              <CalendarRange className="mr-2 h-4 w-4 text-muted" aria-hidden="true" />
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {(Object.keys(RANGE_LABELS) as TimelineRange[]).map((range) => (
                <SelectItem key={range} value={range}>{RANGE_LABELS[range]}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            variant="outline"
            size="sm"
            className="h-10"
            onClick={() => void timeline.refresh()}
            disabled={timeline.loading}
          >
            <RefreshCw className={cn('h-4 w-4', timeline.loading && 'animate-spin')} aria-hidden="true" />
            Refresh
          </Button>
        </div>
      </div>

      {timeline.error && timeline.events.length > 0 && (
        <div className="border-b border-red-500/30 bg-red-500/5 px-4 py-3 text-sm text-red-300" role="alert">
          Newer timeline data could not be loaded. {timeline.error}
        </div>
      )}

      {timeline.loading && timeline.events.length === 0 ? (
        <TimelineSkeleton />
      ) : timeline.error && timeline.events.length === 0 ? (
        <div className="p-4">
          <ResourceState kind="error" title="Unable to load operational history" description={timeline.error} />
        </div>
      ) : timeline.events.length === 0 ? (
        <div className="p-4">
          <ResourceState
            kind="empty"
            title="No operational events in this range"
            description="Choose a longer time range or wait for the next registration, heartbeat transition, version update, reconnect, or policy finding."
          />
        </div>
      ) : (
        <>
          <ol className="divide-y divide-border" aria-live="polite">
            {timeline.events.map((event) => <TimelineEventRow key={event.id} event={event} />)}
          </ol>
          <div className="flex items-center justify-between gap-4 border-t border-border px-4 py-3">
            <p className="font-mono text-xs tabular-nums text-muted">
              {timeline.events.length.toLocaleString()} loaded
            </p>
            {timeline.hasMore ? (
              <Button
                variant="outline"
                size="sm"
                onClick={() => void timeline.loadMore()}
                disabled={timeline.loadingMore}
              >
                {timeline.loadingMore && <RefreshCw className="h-4 w-4 animate-spin" aria-hidden="true" />}
                {timeline.loadingMore ? 'Loading…' : 'Load older events'}
              </Button>
            ) : (
              <span className="text-xs text-muted">End of retained history</span>
            )}
          </div>
        </>
      )}
    </section>
  );
}

function TimelineEventRow({ event }: { event: OperationalEvent }) {
  const presentation = EVENT_PRESENTATION[event.kind];
  const Icon = presentation.icon;
  const timestamp = eventTime(event.occurredAt);
  const details = visibleDetails(event);

  return (
    <li className="grid grid-cols-[auto_minmax(0,1fr)] gap-x-3 gap-y-2 px-4 py-4 sm:grid-cols-[auto_minmax(0,1fr)_auto]">
      <span className={cn('mt-0.5 grid h-9 w-9 place-items-center rounded-md border', presentation.color)}>
        <Icon className="h-4 w-4" aria-hidden="true" />
      </span>
      <div className="min-w-0">
        <div className="flex flex-wrap items-center gap-2">
          <span className={cn('rounded border px-2 py-0.5 text-xs font-semibold', presentation.color)}>
            {presentation.label}
          </span>
          {event.kind === 'policy_finding' && event.details?.severity && (
            <span className="font-mono text-xs uppercase tracking-wide text-orange-300">
              {event.details.severity}
            </span>
          )}
        </div>
        <p className="mt-2 break-words text-sm leading-6 text-foreground">{event.message}</p>
        {details.length > 0 && (
          <dl className="mt-2 flex flex-wrap gap-x-4 gap-y-1 font-mono text-xs text-muted">
            {details.map(([key, value]) => (
              <div key={key} className="flex min-w-0 gap-1.5">
                <dt>{key}</dt>
                <dd className="break-all text-foreground/80">{value}</dd>
              </div>
            ))}
          </dl>
        )}
      </div>
      <time
        dateTime={timestamp.dateTime}
        title={timestamp.exact}
        className="col-start-2 whitespace-nowrap font-mono text-xs tabular-nums text-muted sm:col-start-3 sm:row-start-1"
      >
        {timestamp.relative}
      </time>
    </li>
  );
}

function TimelineSkeleton() {
  return (
    <div aria-busy="true" aria-label="Loading operational timeline" className="divide-y divide-border">
      {Array.from({ length: 5 }, (_, index) => (
        <div key={index} className="grid grid-cols-[auto_minmax(0,1fr)] gap-3 px-4 py-4">
          <div className="h-9 w-9 animate-pulse rounded-md bg-elevated" />
          <div className="space-y-2">
            <div className="h-4 w-24 animate-pulse rounded bg-elevated" />
            <div className="h-3 w-full max-w-md animate-pulse rounded bg-elevated" />
          </div>
        </div>
      ))}
    </div>
  );
}
