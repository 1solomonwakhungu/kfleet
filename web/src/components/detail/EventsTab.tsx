import { useMemo } from 'react';
import { AlertTriangle, Bell, Info } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import type { EventInfo } from '@/types/resources';
import { ResourceState, ResourceTableSkeleton } from './ResourceTabState';

function relativeTime(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime()) || date.getTime() === 0) return '—';
  const seconds = Math.max(0, Math.floor((Date.now() - date.getTime()) / 1000));
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

interface EventsTabProps {
  events: EventInfo[];
  loading: boolean;
  error: string | null;
  search: string;
}

export function EventsTab({ events, loading, error, search }: EventsTabProps) {
  const query = search.trim().toLowerCase();
  const sorted = useMemo(
    () =>
      events
        .filter(
          (event) =>
            [event.reason, event.message, event.namespace, event.type].some((value) =>
              value.toLowerCase().includes(query),
            ),
        )
        .slice()
        .sort((a, b) => new Date(b.lastTimestamp).getTime() - new Date(a.lastTimestamp).getTime()),
    [events, query],
  );
  const warningCount = sorted.filter((event) => event.type.toLowerCase() === 'warning').length;

  if (error) {
    return <ResourceState kind="error" title="Unable to load events" description={error} />;
  }
  if (loading && events.length === 0) {
    return <ResourceTableSkeleton label="Loading events" columns={4} rows={7} />;
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
    );
  }

  return (
    <section aria-label="Kubernetes events" className="overflow-hidden rounded-lg border border-border bg-surface">
      <div className="flex min-h-12 flex-wrap items-center justify-between gap-2 border-b border-border px-4 py-2">
        <div className="flex items-center gap-2 text-sm">
          <Bell className="h-4 w-4 text-blue-400" aria-hidden="true" />
          <span className="font-semibold tabular-nums text-foreground">{sorted.length}</span>
          <span className="text-muted">{sorted.length === 1 ? 'event' : 'events'}</span>
        </div>
        {warningCount > 0 && (
          <div className="flex items-center gap-1.5 text-xs font-medium text-amber-300">
            <AlertTriangle className="h-3.5 w-3.5" aria-hidden="true" />
            <span className="tabular-nums">{warningCount}</span> {warningCount === 1 ? 'warning' : 'warnings'}
          </div>
        )}
      </div>
      <ol className="divide-y divide-border">
        {sorted.map((event, index) => {
          const warning = event.type.toLowerCase() === 'warning';
          const Icon = warning ? AlertTriangle : Info;
          const date = new Date(event.lastTimestamp);
          const validDate = !Number.isNaN(date.getTime()) && date.getTime() !== 0;

          return (
            <li
              key={`${event.namespace}-${event.reason}-${event.lastTimestamp}-${index}`}
              className={cn(
                'grid min-w-0 grid-cols-[auto_minmax(0,1fr)] gap-x-3 gap-y-2 px-4 py-4 sm:grid-cols-[auto_minmax(0,1fr)_auto]',
                warning ? 'bg-amber-500/[0.06]' : 'hover:bg-blue-500/[0.04]',
              )}
            >
              <span
                className={cn(
                  'mt-0.5 grid h-8 w-8 place-items-center rounded-md border',
                  warning
                    ? 'border-amber-500/30 bg-amber-500/10 text-amber-300'
                    : 'border-blue-500/30 bg-blue-500/10 text-blue-300',
                )}
              >
                <Icon className="h-4 w-4" aria-hidden="true" />
              </span>
              <div className="min-w-0 space-y-1.5">
                <div className="flex flex-wrap items-center gap-2 text-sm">
                  <span className="break-words font-semibold text-foreground">{event.reason || 'Unknown reason'}</span>
                  <Badge
                    variant="outline"
                    className={cn(
                      'border',
                      warning
                        ? 'border-amber-500/30 bg-amber-500/10 text-amber-300'
                        : 'border-blue-500/30 bg-blue-500/10 text-blue-300',
                    )}
                  >
                    {event.type || 'Normal'}
                  </Badge>
                  {event.count > 1 && (
                    <span className="font-mono text-xs tabular-nums text-muted" aria-label={`${event.count} occurrences`}>
                      ×{event.count}
                    </span>
                  )}
                </div>
                <p className="break-words text-sm leading-6 text-muted">{event.message || 'No event message was provided.'}</p>
                <p className="font-mono text-xs text-muted">namespace/{event.namespace || 'default'}</p>
              </div>
              <time
                dateTime={validDate ? date.toISOString() : undefined}
                title={validDate ? date.toLocaleString() : undefined}
                className="col-start-2 whitespace-nowrap font-mono text-xs tabular-nums text-muted sm:col-start-3 sm:row-start-1"
              >
                {relativeTime(event.lastTimestamp)}
              </time>
            </li>
          );
        })}
      </ol>
    </section>
  );
}
