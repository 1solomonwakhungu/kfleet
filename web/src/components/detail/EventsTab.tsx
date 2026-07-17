import { useMemo } from 'react';
import { cn } from '@/lib/utils';
import type { EventInfo } from '@/types/resources';

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
  const sorted = useMemo(
    () =>
      events
        .filter(
          (e) =>
            e.reason.toLowerCase().includes(search.toLowerCase()) ||
            e.message.toLowerCase().includes(search.toLowerCase()),
        )
        .slice()
        .sort((a, b) => new Date(b.lastTimestamp).getTime() - new Date(a.lastTimestamp).getTime()),
    [events, search],
  );

  if (error) {
    return <p className="py-8 text-center text-sm text-muted-foreground">Events are not available for this cluster yet.</p>;
  }
  if (loading && events.length === 0) {
    return <p className="py-8 text-center text-sm text-muted-foreground">Loading events…</p>;
  }
  if (sorted.length === 0) {
    return <p className="py-8 text-center text-sm text-muted-foreground">No events found.</p>;
  }

  return (
    <ul className="divide-y divide-border">
      {sorted.map((event, idx) => (
        <li key={`${event.reason}-${event.lastTimestamp}-${idx}`} className="flex gap-3 py-3">
          <span
            className={cn('mt-1 h-2 w-2 shrink-0 rounded-full', event.type === 'Warning' ? 'bg-amber-400' : 'bg-blue-400')}
          />
          <div className="flex-1">
            <div className="flex items-center gap-2 text-sm">
              <span className="font-medium">{event.reason}</span>
              <span className="text-muted-foreground">· {event.namespace}</span>
              {event.count > 1 && <span className="text-muted-foreground">×{event.count}</span>}
            </div>
            <p className="text-sm text-muted-foreground">{event.message}</p>
          </div>
          <span className="shrink-0 text-xs text-muted-foreground">{relativeTime(event.lastTimestamp)}</span>
        </li>
      ))}
    </ul>
  );
}
