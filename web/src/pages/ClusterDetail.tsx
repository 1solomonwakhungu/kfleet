import { useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { ArrowLeft } from 'lucide-react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { HealthBadge } from '@/components/HealthBadge';
import { NamespaceSelector } from '@/components/detail/NamespaceSelector';
import { SearchFilter } from '@/components/detail/SearchFilter';
import { PodsTab } from '@/components/detail/PodsTab';
import { ServicesTab } from '@/components/detail/ServicesTab';
import { DeploymentsTab } from '@/components/detail/DeploymentsTab';
import { EventsTab } from '@/components/detail/EventsTab';
import { LogsTab } from '@/components/detail/LogsTab';
import { useClusterDetail } from '@/hooks/useClusterDetail';
import type { PodInfo } from '@/types/resources';

export default function ClusterDetail() {
  const { id } = useParams<{ id: string }>();
  const detail = useClusterDetail(id);
  const [search, setSearch] = useState('');
  const [tab, setTab] = useState('pods');
  const [logsPod, setLogsPod] = useState<PodInfo | undefined>(undefined);

  const namespaceFilteredPods = useMemo(
    () => (detail.namespace ? detail.pods.data.filter((p) => p.namespace === detail.namespace) : detail.pods.data),
    [detail.pods.data, detail.namespace],
  );
  const namespaceFilteredServices = useMemo(
    () =>
      detail.namespace ? detail.services.data.filter((s) => s.namespace === detail.namespace) : detail.services.data,
    [detail.services.data, detail.namespace],
  );
  const namespaceFilteredDeployments = useMemo(
    () =>
      detail.namespace
        ? detail.deployments.data.filter((d) => d.namespace === detail.namespace)
        : detail.deployments.data,
    [detail.deployments.data, detail.namespace],
  );

  if (!id) {
    return <p className="p-8 text-sm text-muted-foreground">No cluster selected.</p>;
  }

  return (
    <div className="container py-8">
      <Link to="/" className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
        <ArrowLeft className="h-4 w-4" /> Back to clusters
      </Link>

      <div className="mb-6 flex flex-wrap items-center gap-3">
        <h1 className="text-2xl font-bold">{detail.cluster?.name ?? 'Loading…'}</h1>
        {detail.cluster && <HealthBadge health={detail.cluster.health} />}
        {detail.cluster && <span className="text-sm text-muted-foreground">k8s {detail.cluster.version || 'unknown'}</span>}
      </div>

      {detail.statusError && (
        <div className="mb-4 rounded-md border border-destructive/40 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {detail.statusError}
        </div>
      )}

      <Tabs value={tab} onValueChange={setTab}>
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <TabsList>
            <TabsTrigger value="pods">Pods</TabsTrigger>
            <TabsTrigger value="services">Services</TabsTrigger>
            <TabsTrigger value="deployments">Deployments</TabsTrigger>
            <TabsTrigger value="events">Events</TabsTrigger>
            <TabsTrigger value="logs">Logs</TabsTrigger>
          </TabsList>
          {tab !== 'logs' && (
            <div className="flex items-center gap-2">
              <NamespaceSelector namespaces={detail.namespaces} value={detail.namespace} onChange={detail.setNamespace} />
              <SearchFilter value={search} onChange={setSearch} />
            </div>
          )}
        </div>

        <TabsContent value="pods">
          <PodsTab
            pods={namespaceFilteredPods}
            loading={detail.pods.loading}
            error={detail.pods.error}
            search={search}
            onSelectPod={(pod) => {
              setLogsPod(pod);
              setTab('logs');
            }}
          />
        </TabsContent>
        <TabsContent value="services">
          <ServicesTab
            services={namespaceFilteredServices}
            loading={detail.services.loading}
            error={detail.services.error}
            search={search}
          />
        </TabsContent>
        <TabsContent value="deployments">
          <DeploymentsTab
            deployments={namespaceFilteredDeployments}
            loading={detail.deployments.loading}
            error={detail.deployments.error}
            search={search}
          />
        </TabsContent>
        <TabsContent value="events">
          <EventsTab events={detail.events.data} loading={detail.events.loading} error={detail.events.error} search={search} />
        </TabsContent>
        <TabsContent value="logs">
          <LogsTab clusterId={id} pods={detail.pods.data} selectedPod={logsPod} onSelectPod={setLogsPod} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
