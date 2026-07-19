import { useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { AlertTriangle, ArrowLeft, ChevronRight, RefreshCw } from 'lucide-react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { ClusterOverview, ClusterOverviewSkeleton } from '@/components/cluster-overview/ClusterOverview';
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
    return (
      <main className="mx-auto min-h-dvh max-w-[100rem] px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
        <Link
          to="/"
          className="inline-flex min-h-11 items-center gap-2 whitespace-nowrap text-sm font-semibold text-blue-300 transition-[color,transform] duration-150 hover:text-blue-200 active:translate-y-px"
        >
          <ArrowLeft className="h-4 w-4" aria-hidden="true" />
          Back to clusters
        </Link>
        <Card className="mt-6 grid min-h-64 place-items-center border border-border px-6 text-center">
          <div>
            <AlertTriangle className="mx-auto h-7 w-7 text-degraded" aria-hidden="true" />
            <h1 className="mt-4 font-display text-xl font-bold">No cluster selected</h1>
            <p className="mt-2 text-muted">Choose a cluster from the fleet to view its resources and status.</p>
          </div>
        </Card>
      </main>
    );
  }

  return (
    <main className="mx-auto min-h-dvh max-w-[100rem] px-4 py-6 sm:px-6 sm:py-8 lg:px-8 lg:py-10">
      <nav aria-label="Breadcrumb">
        <ol className="flex min-h-11 min-w-0 items-center gap-2 text-sm">
          <li className="shrink-0">
            <Link
              to="/"
              className="inline-flex min-h-11 items-center gap-2 whitespace-nowrap font-semibold text-blue-300 transition-[color,transform] duration-150 hover:text-blue-200 active:translate-y-px"
            >
              <ArrowLeft className="h-4 w-4" aria-hidden="true" />
              Clusters
            </Link>
          </li>
          <li aria-hidden="true" className="shrink-0 text-muted">
            <ChevronRight className="h-4 w-4" />
          </li>
          <li className="min-w-0 truncate text-muted" aria-current="page">
            {detail.cluster?.name ?? (detail.loading ? 'Loading…' : 'Cluster detail')}
          </li>
        </ol>
      </nav>

      {detail.statusError && (
        <section
          className="mt-4 flex flex-col gap-3 rounded-lg border border-danger/40 bg-danger-soft p-4 text-danger sm:flex-row sm:items-center sm:justify-between"
          role="alert"
        >
          <div className="flex min-w-0 items-start gap-3">
            <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
            <div className="min-w-0">
              <p className="font-semibold">Cluster status could not be loaded</p>
              <p className="mt-1 break-words text-sm">{detail.statusError}</p>
            </div>
          </div>
          <Button
            variant="outline"
            size="sm"
            className="self-start border-danger/40 sm:self-auto"
            onClick={() => window.location.reload()}
          >
            <RefreshCw className="h-4 w-4" aria-hidden="true" />
            Retry
          </Button>
        </section>
      )}

      <section className="mt-5" aria-live="polite">
        {detail.cluster ? (
          <ClusterOverview cluster={detail.cluster} nodes={detail.nodes} />
        ) : detail.loading ? (
          <ClusterOverviewSkeleton />
        ) : (
          <Card className="grid min-h-64 place-items-center border border-border px-6 text-center">
            <div>
              <AlertTriangle className="mx-auto h-7 w-7 text-degraded" aria-hidden="true" />
              <h1 className="mt-4 font-display text-xl font-bold">Cluster overview unavailable</h1>
              <p className="mt-2 max-w-lg text-muted">
                No cluster status was returned. Resource tabs remain available below when their data can be loaded.
              </p>
            </div>
          </Card>
        )}
      </section>

      <Tabs value={tab} onValueChange={setTab} className="mt-8">
        <div className="flex flex-col gap-4 border-b border-border pb-4 lg:flex-row lg:items-end lg:justify-between">
          <div className="min-w-0">
            <h2 className="font-display text-xl font-bold">Cluster resources</h2>
            <TabsList className="mt-3 flex-wrap overflow-visible border-b-0">
              <TabsTrigger
                value="pods"
                className={tabTriggerClass(tab === 'pods')}
              >
                Pods{' '}
                <TabCount
                  count={detail.pods.data.length}
                  unavailable={detail.pods.loading || Boolean(detail.pods.error)}
                />
              </TabsTrigger>
              <TabsTrigger
                value="services"
                className={tabTriggerClass(tab === 'services')}
              >
                Services{' '}
                <TabCount
                  count={detail.services.data.length}
                  unavailable={detail.services.loading || Boolean(detail.services.error)}
                />
              </TabsTrigger>
              <TabsTrigger
                value="deployments"
                className={tabTriggerClass(tab === 'deployments')}
              >
                Deployments{' '}
                <TabCount
                  count={detail.deployments.data.length}
                  unavailable={detail.deployments.loading || Boolean(detail.deployments.error)}
                />
              </TabsTrigger>
              <TabsTrigger
                value="events"
                className={tabTriggerClass(tab === 'events')}
              >
                Events{' '}
                <TabCount
                  count={detail.events.data.length}
                  unavailable={detail.events.loading || Boolean(detail.events.error)}
                />
              </TabsTrigger>
              <TabsTrigger
                value="logs"
                className={tabTriggerClass(tab === 'logs')}
              >
                Logs
              </TabsTrigger>
            </TabsList>
          </div>
          {tab !== 'logs' && (
            <div
              className="flex min-w-0 flex-wrap items-center gap-2 [&_button]:h-11"
              aria-label="Resource filters"
            >
              <NamespaceSelector
                namespaces={detail.namespaces}
                value={detail.namespace}
                onChange={detail.setNamespace}
              />
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
    </main>
  );
}

function tabTriggerClass(active: boolean): string {
  const states = 'active:translate-y-px disabled:cursor-not-allowed disabled:opacity-50';
  return active
    ? `!border-blue-400 bg-blue-500/10 !text-blue-200 ${states}`
    : `hover:text-blue-200 ${states}`;
}

function TabCount({ count, unavailable }: { count: number; unavailable: boolean }) {
  const label = unavailable ? 'Count unavailable' : `${count} items`;
  return (
    <span className="rounded bg-elevated px-1.5 py-0.5 font-mono text-xs tabular-nums text-muted" aria-label={label}>
      {unavailable ? '—' : count.toLocaleString()}
    </span>
  );
}
