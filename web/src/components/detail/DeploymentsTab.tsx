import { useMemo } from 'react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { cn } from '@/lib/utils';
import type { DeploymentInfo } from '@/types/resources';

interface DeploymentsTabProps {
  deployments: DeploymentInfo[];
  loading: boolean;
  error: string | null;
  search: string;
}

export function DeploymentsTab({ deployments, loading, error, search }: DeploymentsTabProps) {
  const filtered = useMemo(
    () => deployments.filter((d) => d.name.toLowerCase().includes(search.toLowerCase())),
    [deployments, search],
  );

  if (error) {
    return (
      <p className="py-8 text-center text-sm text-muted-foreground">Deployments are not available for this cluster yet.</p>
    );
  }
  if (loading && deployments.length === 0) {
    return <p className="py-8 text-center text-sm text-muted-foreground">Loading deployments…</p>;
  }
  if (filtered.length === 0) {
    return <p className="py-8 text-center text-sm text-muted-foreground">No deployments found.</p>;
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Namespace</TableHead>
          <TableHead>Ready</TableHead>
          <TableHead>Up-to-date</TableHead>
          <TableHead>Available</TableHead>
          <TableHead>Age</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {filtered.map((dep) => {
          const underReplicated = dep.readyReplicas < dep.desiredReplicas;
          return (
            <TableRow key={`${dep.namespace}/${dep.name}`} className={cn(underReplicated && 'bg-amber-500/5')}>
              <TableCell className="font-medium">{dep.name}</TableCell>
              <TableCell>{dep.namespace}</TableCell>
              <TableCell className={cn(underReplicated && 'font-semibold text-amber-400')}>
                {dep.readyReplicas}/{dep.desiredReplicas}
              </TableCell>
              <TableCell>{dep.updatedReplicas}</TableCell>
              <TableCell>{dep.availableReplicas}</TableCell>
              <TableCell>{dep.age}</TableCell>
            </TableRow>
          );
        })}
      </TableBody>
    </Table>
  );
}
