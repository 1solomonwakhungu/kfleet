import { useMemo } from 'react';
import { Boxes } from 'lucide-react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import type { DeploymentInfo } from '@/types/resources';
import { ResourceState, ResourceTablePanel, ResourceTableSkeleton } from './ResourceTabState';

interface DeploymentsTabProps {
  deployments: DeploymentInfo[];
  loading: boolean;
  error: string | null;
  search: string;
}

export function DeploymentsTab({ deployments, loading, error, search }: DeploymentsTabProps) {
  const query = search.trim().toLowerCase();
  const filtered = useMemo(
    () =>
      deployments.filter((deployment) =>
        [deployment.name, deployment.namespace].some((value) => value.toLowerCase().includes(query)),
      ),
    [deployments, query],
  );

  if (error) {
    return <ResourceState kind="error" title="Unable to load deployments" description={error} />;
  }
  if (loading && deployments.length === 0) {
    return <ResourceTableSkeleton label="Loading deployments" columns={7} />;
  }
  if (filtered.length === 0) {
    return (
      <ResourceState
        kind="empty"
        title={query ? 'No matching deployments' : 'No deployments found'}
        description={
          query
            ? `No deployment name or namespace matches “${search.trim()}”.`
            : 'No deployments were returned for this namespace.'
        }
      />
    );
  }

  return (
    <ResourceTablePanel label="Deployments" count={filtered.length} noun="deployment">
      <Table className="min-w-[880px]">
        <caption className="sr-only">Deployments and their desired, ready, updated, and available replica counts</caption>
        <TableHeader className="bg-background">
          <TableRow>
            <TableHead scope="col" className="w-[27%]">Name</TableHead>
            <TableHead scope="col">Namespace</TableHead>
            <TableHead scope="col">Status</TableHead>
            <TableHead scope="col" className="w-48">Readiness</TableHead>
            <TableHead scope="col" className="text-right">Updated</TableHead>
            <TableHead scope="col" className="text-right">Available</TableHead>
            <TableHead scope="col" className="text-right">Age</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {filtered.map((deployment) => {
            const scaledToZero = deployment.desiredReplicas === 0;
            const ready = !scaledToZero && deployment.readyReplicas >= deployment.desiredReplicas;
            const unavailable = !scaledToZero && deployment.readyReplicas === 0;
            const readiness = scaledToZero
              ? 0
              : Math.min(100, Math.round((deployment.readyReplicas / deployment.desiredReplicas) * 100));
            const status = scaledToZero ? 'Scaled to zero' : ready ? 'Ready' : unavailable ? 'Unavailable' : 'Progressing';
            const statusStyles = scaledToZero
              ? 'border-zinc-500/30 bg-zinc-500/10 text-zinc-300'
              : ready
                ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300'
                : unavailable
                  ? 'border-red-500/30 bg-red-500/10 text-red-300'
                  : 'border-amber-500/30 bg-amber-500/10 text-amber-300';

            return (
              <TableRow
                key={`${deployment.namespace}/${deployment.name}`}
                className={cn('hover:bg-blue-500/5', !ready && !scaledToZero && 'bg-amber-500/[0.03]')}
              >
                <TableCell>
                  <span className="flex min-w-0 items-center gap-2 font-semibold text-foreground">
                    <Boxes className="h-4 w-4 shrink-0 text-blue-400" aria-hidden="true" />
                    <span className="truncate">{deployment.name}</span>
                  </span>
                </TableCell>
                <TableCell className="text-muted">{deployment.namespace}</TableCell>
                <TableCell>
                  <Badge variant="outline" className={cn('border', statusStyles)}>{status}</Badge>
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-3">
                    <span className="w-10 shrink-0 font-mono text-xs tabular-nums text-foreground">
                      {deployment.readyReplicas}/{deployment.desiredReplicas}
                    </span>
                    <div
                      className="h-1.5 w-24 overflow-hidden rounded-full bg-elevated"
                      role="progressbar"
                      aria-label={`${deployment.name} readiness`}
                      aria-valuemin={0}
                      aria-valuemax={100}
                      aria-valuenow={readiness}
                    >
                      <div
                        className={cn(
                          'h-full w-full origin-left rounded-full',
                          scaledToZero ? 'bg-zinc-500' : ready ? 'bg-emerald-400' : unavailable ? 'bg-red-400' : 'bg-amber-400',
                        )}
                        style={{ transform: `scaleX(${readiness / 100})` }}
                      />
                    </div>
                  </div>
                </TableCell>
                <TableCell
                  className={cn(
                    'text-right font-mono tabular-nums',
                    deployment.updatedReplicas < deployment.desiredReplicas ? 'text-amber-300' : 'text-muted',
                  )}
                >
                  {deployment.updatedReplicas}
                </TableCell>
                <TableCell
                  className={cn(
                    'text-right font-mono tabular-nums',
                    deployment.availableReplicas < deployment.desiredReplicas ? 'text-amber-300' : 'text-muted',
                  )}
                >
                  {deployment.availableReplicas}
                </TableCell>
                <TableCell className="text-right font-mono tabular-nums text-muted">{deployment.age || '—'}</TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </ResourceTablePanel>
  );
}
