import { useMemo, useState } from 'react'
import { Blankslate } from '@primer/react/experimental'
import { Button, Flash, Heading, Text } from '@primer/react'
import { AlertIcon, SearchIcon, ServerIcon, SyncIcon } from '@primer/octicons-react'
import { useNavigate } from 'react-router-dom'

import { ClusterCard } from '../components/ClusterCard'
import { DashboardSkeleton } from '../components/dashboard/DashboardSkeleton'
import { FleetControls, type FleetSort, type HealthFilter } from '../components/dashboard/FleetControls'
import { FleetSummary } from '../components/dashboard/FleetSummary'
import { useClusters } from '../hooks/useClusters'
import type { Cluster, ClusterHealth } from '../types/cluster'
import layout from '../styles/layout.module.css'
import styles from './Dashboard.module.css'

const healthPriority: Record<ClusterHealth, number> = {
  unreachable: 0,
  degraded: 1,
  unknown: 2,
  healthy: 3,
}

function compareText(left: string, right: string) {
  const normalizedLeft = left.toLowerCase()
  const normalizedRight = right.toLowerCase()
  if (normalizedLeft < normalizedRight) return -1
  if (normalizedLeft > normalizedRight) return 1
  if (left < right) return -1
  if (left > right) return 1
  return 0
}

function compareByName(left: Cluster, right: Cluster) {
  return compareText(left.name, right.name) || compareText(left.id, right.id)
}

function heartbeatTimestamp(value: string) {
  const timestamp = Date.parse(value)
  return Number.isNaN(timestamp) ? Number.NEGATIVE_INFINITY : timestamp
}

function sortClusters(left: Cluster, right: Cluster, sort: FleetSort) {
  if (sort === 'health') {
    return healthPriority[left.health] - healthPriority[right.health] || compareByName(left, right)
  }

  if (sort === 'heartbeat') {
    return heartbeatTimestamp(right.lastHeartbeat) - heartbeatTimestamp(left.lastHeartbeat) || compareByName(left, right)
  }

  return compareByName(left, right)
}

function matchesSearch(cluster: Cluster, query: string) {
  if (!query) return true

  const searchableLabels = Object.entries(cluster.labels).flatMap(([key, value]) => [
    key,
    value,
    `${key}=${value}`,
    `${key}:${value}`,
  ])

  return [cluster.name, ...searchableLabels].some((value) => value.toLowerCase().includes(query))
}

export function Dashboard() {
  const navigate = useNavigate()
  const { clusters, loading, error, refresh } = useClusters()
  const [search, setSearch] = useState('')
  const [healthFilter, setHealthFilter] = useState<HealthFilter>('all')
  const [sort, setSort] = useState<FleetSort>('health')

  const visibleClusters = useMemo(() => {
    const query = search.trim().toLowerCase()
    return clusters
      .filter((cluster) => healthFilter === 'all' || cluster.health === healthFilter)
      .filter((cluster) => matchesSearch(cluster, query))
      .sort((left, right) => sortClusters(left, right, sort))
  }, [clusters, healthFilter, search, sort])

  const hasActiveControls = search.trim().length > 0 || healthFilter !== 'all' || sort !== 'health'
  const resetControls = () => {
    setSearch('')
    setHealthFilter('all')
    setSort('health')
  }

  return (
    <main className={layout.page}>
      <header className={layout.pageHeader}>
        <div className={layout.pageHeaderText}>
          <Heading as="h1" variant="large">
            Fleet dashboard
          </Heading>
          <Text className={layout.pageDescription}>
            Operational health, capacity, and agent freshness across registered Kubernetes clusters.
          </Text>
        </div>
        <div className={styles.headerActions}>
          <Text size="small" className={`${layout.mono} ${layout.muted} ${styles.refreshNote}`}>
            Auto-refresh · 5s
          </Text>
          <Button leadingVisual={SyncIcon} variant="primary" disabled={loading} onClick={() => void refresh()}>
            {loading ? 'Refreshing…' : 'Refresh fleet'}
          </Button>
        </div>
      </header>

      {loading ? (
        <DashboardSkeleton />
      ) : (
        <>
          <FleetSummary clusters={clusters} />

          {error && clusters.length > 0 && (
            <Flash variant="danger" className={styles.flash}>
              <div className={styles.flashBody}>
                <AlertIcon size={16} />
                <div>
                  <Text weight="semibold">The latest fleet snapshot could not be loaded.</Text>
                  <Text className={layout.pageDescription}>Showing the last available data. {error.message}</Text>
                </div>
                <Button leadingVisual={SyncIcon} onClick={() => void refresh()}>
                  Retry
                </Button>
              </div>
            </Flash>
          )}

          {clusters.length > 0 && (
            <FleetControls
              search={search}
              onSearchChange={setSearch}
              health={healthFilter}
              onHealthChange={setHealthFilter}
              sort={sort}
              onSortChange={setSort}
              resultCount={visibleClusters.length}
              totalCount={clusters.length}
              hasActiveControls={hasActiveControls}
              onReset={resetControls}
            />
          )}

          <section className={styles.inventory} aria-labelledby="cluster-inventory-heading">
            <div className={styles.inventoryHeader}>
              <Heading as="h2" variant="small" id="cluster-inventory-heading">
                Cluster inventory
              </Heading>
              {clusters.length > 0 && (
                <Text
                  size="small"
                  className={`${layout.mono} ${layout.muted}`}
                  role="status"
                  aria-live="polite"
                  aria-atomic="true"
                >
                  {visibleClusters.length} of {clusters.length}
                </Text>
              )}
            </div>

            {error && clusters.length === 0 ? (
              <DashboardState
                icon={AlertIcon}
                title="Fleet data is unavailable"
                description={`The hub did not return a cluster snapshot. ${error.message}`}
                actionLabel="Retry connection"
                onAction={() => void refresh()}
              />
            ) : clusters.length === 0 ? (
              <DashboardState
                icon={ServerIcon}
                title="No clusters registered"
                description="Connect a kfleet agent to this hub to begin monitoring its Kubernetes cluster."
                actionLabel="Check again"
                onAction={() => void refresh()}
              />
            ) : visibleClusters.length === 0 ? (
              <DashboardState
                icon={SearchIcon}
                title="No clusters match these controls"
                description="Try a different cluster name or label, or broaden the health filter."
                actionLabel="Reset controls"
                onAction={resetControls}
              />
            ) : (
              <div className={`${layout.grid} ${layout.grid3}`}>
                {visibleClusters.map((cluster) => (
                  <ClusterCard
                    key={cluster.id}
                    cluster={cluster}
                    onClick={() => navigate(`/clusters/${encodeURIComponent(cluster.id)}`)}
                  />
                ))}
              </div>
            )}
          </section>
        </>
      )}
    </main>
  )
}

interface DashboardStateProps {
  icon: typeof ServerIcon
  title: string
  description: string
  actionLabel: string
  onAction: () => void
}

function DashboardState({ icon: Icon, title, description, actionLabel, onAction }: DashboardStateProps) {
  return (
    <div className={layout.box}>
      <Blankslate>
        <Blankslate.Visual>
          <Icon size={24} />
        </Blankslate.Visual>
        <Blankslate.Heading>{title}</Blankslate.Heading>
        <Blankslate.Description>{description}</Blankslate.Description>
        <Blankslate.PrimaryAction onClick={onAction}>{actionLabel}</Blankslate.PrimaryAction>
      </Blankslate>
    </div>
  )
}
