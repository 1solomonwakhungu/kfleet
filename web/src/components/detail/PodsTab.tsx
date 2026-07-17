import { useMemo } from 'react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import type { PodInfo } from '@/types/resources';

const PHASE_STYLES: Record<string, string> = {
  Running: 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30',
  Pending: 'bg-amber-500/15 text-amber-400 border-amber-500/30',
  Failed: 'bg-red-500/15 text-red-400 border-red-500/30',
  Succeeded: 'bg-blue-500/15 text-blue-400 border-blue-500/30',
  Unknown: 'bg-zinc-500/15 text-zinc-400 border-zinc-500/30',
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
  const filtered = useMemo(
    () => pods.filter((p) => p.name.toLowerCase().includes(search.toLowerCase())),
    [pods, search],
  );

  if (error) {
    return <p className="py-8 text-center text-sm text-muted-foreground">Pods are not available for this cluster yet.</p>;
  }
  if (loading && pods.length === 0) {
    return <p className="py-8 text-center text-sm text-muted-foreground">Loading pods…</p>;
  }
  if (filtered.length === 0) {
    return <p className="py-8 text-center text-sm text-muted-foreground">No pods found.</p>;
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Namespace</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Ready</TableHead>
          <TableHead>Restarts</TableHead>
          <TableHead>Age</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {filtered.map((pod) => (
          <TableRow key={`${pod.namespace}/${pod.name}`} className="cursor-pointer" onClick={() => onSelectPod?.(pod)}>
            <TableCell className="font-medium">{pod.name}</TableCell>
            <TableCell>{pod.namespace}</TableCell>
            <TableCell>
              <Badge variant="outline" className={cn(PHASE_STYLES[pod.phase] ?? PHASE_STYLES.Unknown)}>
                {pod.phase}
              </Badge>
            </TableCell>
            <TableCell>{pod.ready ? '1/1' : '0/1'}</TableCell>
            <TableCell className={cn(pod.restartCount > 5 && 'font-semibold text-red-400')}>{pod.restartCount}</TableCell>
            <TableCell>{age(pod.startTime)}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
