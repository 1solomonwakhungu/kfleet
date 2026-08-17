import { useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Breadcrumbs, Button, Flash, Heading, UnderlineNav } from '@primer/react'
import { Blankslate } from '@primer/react/experimental'
import { AlertIcon, SyncIcon } from '@primer/octicons-react'

import { ClusterOverview, ClusterOverviewSkeleton } from '../components/cluster-overview/ClusterOverview'
import { NamespaceSelector } from '../components/detail/NamespaceSelector'
import { SearchFilter } from '../components/detail/SearchFilter'
import { PodsTab } from '../components/detail/PodsTab'
import { ServicesTab } from '../components/detail/ServicesTab'
import { DeploymentsTab } from '../components/detail/DeploymentsTab'
import { EventsTab } from '../components/detail/EventsTab'
import { LogsTab } from '../components/detail/LogsTab'
import { OperationalTimeline } from '../components/detail/OperationalTimeline'
import { RemoveClusterCard } from '../components/admin/RemoveClusterCard'
import { useClusterDetail } from '../hooks/useClusterDetail'
import type { PodInfo } from '../types/resources'
import layout from '../styles/layout.module.css'
import styles from './ClusterDetail.module.css'

type TabKey = 'pods' | 'services' | 'deployments' | 'events' | 'logs' | 'timeline'

export default function ClusterDetail() {
  const { id } = useParams<{ id: string }>()
  const detail = useClusterDetail(id)
  const [search, setSearch] = useState('')
  const [tab, setTab] = useState<TabKey>('pods')
  const [logsPod, setLogsPod] = useState<PodInfo | undefined>(undefined)

  const namespaceFilteredPods = useMemo(
    () => (detail.namespace ? detail.pods.data.filter((p) => p.namespace === detail.namespace) : detail.pods.data),
    [detail.pods.data, detail.namespace],
  )
  const namespaceFilteredServices = useMemo(
    () =>
      detail.namespace ? detail.services.data.filter((s) => s.namespace === detail.namespace) : detail.services.data,
    [detail.services.data, detail.namespace],
  )
  const namespaceFilteredDeployments = useMemo(
    () =>
      detail.namespace
        ? detail.deployments.data.filter((d) => d.namespace === detail.namespace)
        : detail.deployments.data,
    [detail.deployments.data, detail.namespace],
  )

  if (!id) {
    return (
      <main className={layout.page}>
        <div className={layout.box}>
          <Blankslate>
            <Blankslate.Visual>
              <AlertIcon size={24} />
            </Blankslate.Visual>
            <Blankslate.Heading as="h1">No cluster selected</Blankslate.Heading>
            <Blankslate.Description>
              Choose a cluster from the fleet to view its resources and status.
            </Blankslate.Description>
            <Blankslate.PrimaryAction href="/">Back to clusters</Blankslate.PrimaryAction>
          </Blankslate>
        </div>
      </main>
    )
  }

  const showResourceFilters = tab !== 'logs' && tab !== 'timeline'

  return (
    <main className={layout.page}>
      <Breadcrumbs>
        <Breadcrumbs.Item as={Link} to="/">
          Clusters
        </Breadcrumbs.Item>
        <Breadcrumbs.Item selected>
          {detail.cluster?.name ?? (detail.loading ? 'Loading…' : 'Cluster detail')}
        </Breadcrumbs.Item>
      </Breadcrumbs>

      {detail.statusError && (
        <Flash variant="danger" className={styles.flash}>
          <div className={styles.flashBody}>
            <AlertIcon size={16} />
            <div>
              <strong>Cluster status could not be loaded</strong>
              <p className={styles.flashDetail}>{detail.statusError}</p>
            </div>
            <Button leadingVisual={SyncIcon} onClick={() => window.location.reload()}>
              Retry
            </Button>
          </div>
        </Flash>
      )}

      <section className={styles.overview} aria-live="polite">
        {detail.cluster ? (
          <ClusterOverview cluster={detail.cluster} nodes={detail.nodes} />
        ) : detail.loading ? (
          <ClusterOverviewSkeleton />
        ) : (
          <div className={layout.box}>
            <Blankslate>
              <Blankslate.Visual>
                <AlertIcon size={24} />
              </Blankslate.Visual>
              <Blankslate.Heading as="h1">Cluster overview unavailable</Blankslate.Heading>
              <Blankslate.Description>
                No cluster status was returned. Resource tabs remain available below when their data can be loaded.
              </Blankslate.Description>
            </Blankslate>
          </div>
        )}
      </section>

      <section className={styles.resources}>
        <Heading as="h2" variant="small" className={styles.resourcesHeading}>
          Cluster resources
        </Heading>

        <div className={styles.tabsRow}>
          <div className={styles.tabs}>
            <UnderlineNav aria-label="Cluster resources">
              <UnderlineNav.Item
                aria-current={tab === 'pods' ? 'page' : undefined}
                counter={tabCounter(detail.pods.data.length, detail.pods.loading, detail.pods.error)}
                onSelect={selectTab(() => setTab('pods'))}
              >
                Pods
              </UnderlineNav.Item>
              <UnderlineNav.Item
                aria-current={tab === 'services' ? 'page' : undefined}
                counter={tabCounter(detail.services.data.length, detail.services.loading, detail.services.error)}
                onSelect={selectTab(() => setTab('services'))}
              >
                Services
              </UnderlineNav.Item>
              <UnderlineNav.Item
                aria-current={tab === 'deployments' ? 'page' : undefined}
                counter={tabCounter(detail.deployments.data.length, detail.deployments.loading, detail.deployments.error)}
                onSelect={selectTab(() => setTab('deployments'))}
              >
                Deployments
              </UnderlineNav.Item>
              <UnderlineNav.Item
                aria-current={tab === 'events' ? 'page' : undefined}
                counter={tabCounter(detail.events.data.length, detail.events.loading, detail.events.error)}
                onSelect={selectTab(() => setTab('events'))}
              >
                Events
              </UnderlineNav.Item>
              <UnderlineNav.Item
                aria-current={tab === 'logs' ? 'page' : undefined}
                onSelect={selectTab(() => setTab('logs'))}
              >
                Logs
              </UnderlineNav.Item>
              <UnderlineNav.Item
                aria-current={tab === 'timeline' ? 'page' : undefined}
                onSelect={selectTab(() => setTab('timeline'))}
              >
                Timeline
              </UnderlineNav.Item>
            </UnderlineNav>
          </div>

          {showResourceFilters && (
            <div className={styles.filters} aria-label="Resource filters">
              <NamespaceSelector
                namespaces={detail.namespaces}
                value={detail.namespace}
                onChange={detail.setNamespace}
              />
              <SearchFilter value={search} onChange={setSearch} />
            </div>
          )}
        </div>

        <div className={styles.tabPanel}>
          {tab === 'pods' && (
            <PodsTab
              pods={namespaceFilteredPods}
              loading={detail.pods.loading}
              error={detail.pods.error}
              search={search}
              onSelectPod={(pod) => {
                setLogsPod(pod)
                setTab('logs')
              }}
            />
          )}
          {tab === 'services' && (
            <ServicesTab
              services={namespaceFilteredServices}
              loading={detail.services.loading}
              error={detail.services.error}
              search={search}
            />
          )}
          {tab === 'deployments' && (
            <DeploymentsTab
              deployments={namespaceFilteredDeployments}
              loading={detail.deployments.loading}
              error={detail.deployments.error}
              search={search}
            />
          )}
          {tab === 'events' && (
            <EventsTab
              events={detail.events.data}
              loading={detail.events.loading}
              error={detail.events.error}
              search={search}
            />
          )}
          {tab === 'logs' && (
            <LogsTab clusterId={id} pods={detail.pods.data} selectedPod={logsPod} onSelectPod={setLogsPod} />
          )}
          {tab === 'timeline' && <OperationalTimeline clusterId={id} />}
        </div>
      </section>

      <section className={styles.danger}>
        <RemoveClusterCard clusterId={id} clusterName={detail.cluster?.name ?? id} />
      </section>
    </main>
  )
}

function tabCounter(count: number, loading: boolean, error: string | null): number | undefined {
  return loading || error ? undefined : count
}

function selectTab(activate: () => void) {
  return (event: React.MouseEvent<HTMLAnchorElement> | React.KeyboardEvent<HTMLAnchorElement>) => {
    event.preventDefault()
    activate()
  }
}
