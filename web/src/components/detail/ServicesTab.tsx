import { useMemo } from 'react';
import { Network } from 'lucide-react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import type { ServiceInfo } from '@/types/resources';
import { ResourceState, ResourceTablePanel, ResourceTableSkeleton } from './ResourceTabState';

interface ServicesTabProps {
  services: ServiceInfo[];
  loading: boolean;
  error: string | null;
  search: string;
}

export function ServicesTab({ services, loading, error, search }: ServicesTabProps) {
  const query = search.trim().toLowerCase();
  const filtered = useMemo(
    () =>
      services.filter((service) =>
        [
          service.name,
          service.namespace,
          service.type,
          service.clusterIP,
          ...service.ports.flatMap((port) => [port.name, port.protocol, String(port.port), String(port.targetPort)]),
        ].some((value) => value.toLowerCase().includes(query)),
      ),
    [services, query],
  );

  if (error) {
    return <ResourceState kind="error" title="Unable to load services" description={error} />;
  }
  if (loading && services.length === 0) {
    return <ResourceTableSkeleton label="Loading services" columns={6} />;
  }
  if (filtered.length === 0) {
    return (
      <ResourceState
        kind="empty"
        title={query ? 'No matching services' : 'No services found'}
        description={
          query
            ? `No service matches “${search.trim()}”. Try a name, namespace, type, IP, or port.`
            : 'No services were returned for this namespace.'
        }
      />
    );
  }

  return (
    <ResourceTablePanel label="Services" count={filtered.length} noun="service">
      <Table className="min-w-[940px]">
        <caption className="sr-only">Services, their types, cluster endpoints, port mappings, and age</caption>
        <TableHeader className="bg-background">
          <TableRow>
            <TableHead scope="col" className="w-[24%]">Name</TableHead>
            <TableHead scope="col">Namespace</TableHead>
            <TableHead scope="col">Type</TableHead>
            <TableHead scope="col">Cluster endpoint</TableHead>
            <TableHead scope="col" className="w-[30%]">Ports</TableHead>
            <TableHead scope="col" className="text-right">Age</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {filtered.map((service) => {
            const headless = !service.clusterIP || service.clusterIP.toLowerCase() === 'none';
            return (
              <TableRow key={`${service.namespace}/${service.name}`} className="hover:bg-blue-500/5">
                <TableCell>
                  <span className="flex min-w-0 items-center gap-2 font-semibold text-foreground">
                    <Network className="h-4 w-4 shrink-0 text-blue-400" aria-hidden="true" />
                    <span className="truncate">{service.name}</span>
                  </span>
                </TableCell>
                <TableCell className="text-muted">{service.namespace}</TableCell>
                <TableCell>
                  <Badge variant="outline" className="border border-border bg-elevated text-foreground">
                    {service.type || 'Unknown'}
                  </Badge>
                </TableCell>
                <TableCell>
                  {headless ? (
                    <span className="text-muted">Headless</span>
                  ) : (
                    <div className="space-y-1 font-mono text-xs tabular-nums">
                      {service.ports.length > 0 ? (
                        service.ports.map((port, index) => (
                          <div key={`${port.name || 'port'}-${port.port}-${index}`}>
                            {service.clusterIP}:{port.port}
                          </div>
                        ))
                      ) : (
                        <div>{service.clusterIP}</div>
                      )}
                    </div>
                  )}
                </TableCell>
                <TableCell>
                  {service.ports.length > 0 ? (
                    <div className="flex flex-wrap gap-1.5">
                      {service.ports.map((port, index) => (
                        <span
                          key={`${port.name || 'port'}-${port.port}-${index}`}
                          className="inline-flex items-center rounded border border-border bg-background px-2 py-1 font-mono text-xs tabular-nums text-muted"
                          title={`${port.port} routes to target port ${port.targetPort} over ${port.protocol}`}
                        >
                          {port.name ? `${port.name} · ` : ''}{port.port}→{port.targetPort}/{port.protocol}
                        </span>
                      ))}
                    </div>
                  ) : (
                    <span className="text-muted">No ports</span>
                  )}
                </TableCell>
                <TableCell className="text-right font-mono tabular-nums text-muted">{service.age || '—'}</TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </ResourceTablePanel>
  );
}
