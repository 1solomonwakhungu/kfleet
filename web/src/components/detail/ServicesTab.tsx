import { useMemo } from 'react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import type { ServiceInfo } from '@/types/resources';

interface ServicesTabProps {
  services: ServiceInfo[];
  loading: boolean;
  error: string | null;
  search: string;
}

export function ServicesTab({ services, loading, error, search }: ServicesTabProps) {
  const filtered = useMemo(
    () => services.filter((s) => s.name.toLowerCase().includes(search.toLowerCase())),
    [services, search],
  );

  if (error) {
    return <p className="py-8 text-center text-sm text-muted-foreground">Services are not available for this cluster yet.</p>;
  }
  if (loading && services.length === 0) {
    return <p className="py-8 text-center text-sm text-muted-foreground">Loading services…</p>;
  }
  if (filtered.length === 0) {
    return <p className="py-8 text-center text-sm text-muted-foreground">No services found.</p>;
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Namespace</TableHead>
          <TableHead>Type</TableHead>
          <TableHead>Cluster IP</TableHead>
          <TableHead>Ports</TableHead>
          <TableHead>Age</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {filtered.map((svc) => (
          <TableRow key={`${svc.namespace}/${svc.name}`}>
            <TableCell className="font-medium">{svc.name}</TableCell>
            <TableCell>{svc.namespace}</TableCell>
            <TableCell>{svc.type}</TableCell>
            <TableCell>{svc.clusterIP}</TableCell>
            <TableCell>{svc.ports.map((p) => `${p.port}:${p.targetPort}/${p.protocol}`).join(', ')}</TableCell>
            <TableCell>{svc.age}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
