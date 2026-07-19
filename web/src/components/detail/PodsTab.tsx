import { useMemo } from 'react';
import { FileText, RotateCcw } from 'lucide-react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import type { PodInfo } from '@/types/resources';
import { ResourceState, ResourceTablePanel, ResourceTableSkeleton } from './ResourceTabState';

const PHASE_STYLES: Record<string, string> = {
  Running: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300',
  Pending: 'border-amber-500/30 bg-amber-500/10 text-amber-300',
  Failed: 'border-red-500/30 bg-red-500/10 text-red-300',
  Succeeded: 'border-sky-500/30 bg-sky-500/10 text-sky-300',
  Unknown: 'border-zinc-500/30 bg-zinc-500/10 text-zinc-300',
};

function age(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime()) || date.getTime() === 0) return '—';
  const seconds = Math.max(0, Math.floor((Date.now() - date.getTime()) / 1000));
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h`;
  return `${Math.floor(hours / 24)}d`;
}

interface PodsTabProps {
  pods: PodInfo[];
  loading: boolean;
  error: string | null;
  search: string;
  onSelectPod?: (pod: PodInfo) => void;
}

export function PodsTab({ pods, loading, error, search, onSelectPod }: PodsTabProps) {
  const query = search.trim().toLowerCase();
  const filtered = useMemo(
    () =>
      pods.filter((pod) =>
        [pod.name, pod.namespace, pod.nodeName, pod.phase].some((value) => value.toLowerCase().includes(query)),
      ),
    [pods, query],
  );

  if (error) {
    return <ResourceState kind="error" title="Unable to load pods" description={error} />;
  }
  if (loading && pods.length === 0) {
    return <ResourceTableSkeleton label="Loading pods" columns={7} />;
  }
  if (filtered.length === 0) {
    return (
      <ResourceState
        kind="empty"
        title={query ? 'No matching pods' : 'No pods found'}
        description={
          query ? `No pod matches “${search.trim()}”. Try a name, namespace, node, or phase.` : 'No pods were returned for this namespace.'
        }
      />
    );
  }

  return (
    <ResourceTablePanel label="Pods" count={filtered.length} noun="pod">
      <Table className="min-w-[920px]">
        <caption className="sr-only">Pods, their scheduling location, phase, readiness, restarts, and age</caption>
        <TableHeader className="bg-background">
          <TableRow>
            <TableHead scope="col" className="w-[28%]">Name</TableHead>
            <TableHead scope="col">Namespace</TableHead>
            <TableHead scope="col">Phase</TableHead>
            <TableHead scope="col">Ready</TableHead>
            <TableHead scope="col" className="text-right">Restarts</TableHead>
            <TableHead scope="col">Node</TableHead>
            <TableHead scope="col" className="text-right">Age</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {filtered.map((pod) => (
            <TableRow key={`${pod.namespace}/${pod.name}`} className="hover:bg-blue-500/5">
              <TableCell>
                {onSelectPod ? (
                  <button
                    type="button"
                    onClick={() => onSelectPod(pod)}
                    className="group inline-flex min-h-11 max-w-full items-center gap-2 rounded-sm text-left font-semibold text-blue-300 outline-none hover:text-blue-200 focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 focus-visible:ring-offset-background active:text-blue-400"
                    aria-label={`View logs for pod ${pod.name} in namespace ${pod.namespace}`}
                  >
                    <span className="truncate">{pod.name}</span>
                    <FileText className="h-3.5 w-3.5 shrink-0 opacity-60 group-hover:opacity-100" aria-hidden="true" />
                  </button>
                ) : (
                  <span className="font-semibold text-foreground">{pod.name}</span>
                )}
              </TableCell>
              <TableCell className="text-muted">{pod.namespace}</TableCell>
              <TableCell>
                <Badge variant="outline" className={cn('border', PHASE_STYLES[pod.phase] ?? PHASE_STYLES.Unknown)}>
                  {pod.phase || 'Unknown'}
                </Badge>
              </TableCell>
              <TableCell>
                <span className={cn('inline-flex items-center gap-1.5 font-medium', pod.ready ? 'text-emerald-300' : 'text-amber-300')}>
                  <span className={cn('h-1.5 w-1.5 rounded-full', pod.ready ? 'bg-emerald-400' : 'bg-amber-400')} aria-hidden="true" />
                  {pod.ready ? 'Ready' : 'Not ready'}
                </span>
              </TableCell>
              <TableCell className={cn('text-right font-mono tabular-nums', pod.restartCount > 0 && 'text-amber-300')}>
                <span className="inline-flex items-center gap-1.5">
                  {pod.restartCount > 0 && <RotateCcw className="h-3.5 w-3.5" aria-hidden="true" />}
                  {pod.restartCount}
                </span>
              </TableCell>
              <TableCell className="max-w-48 truncate font-mono text-xs text-muted" title={pod.nodeName || undefined}>
                {pod.nodeName || 'Unscheduled'}
              </TableCell>
              <TableCell className="text-right font-mono tabular-nums text-muted">{age(pod.startTime)}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </ResourceTablePanel>
  );
}
