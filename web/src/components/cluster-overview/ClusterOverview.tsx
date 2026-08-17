import { Label, Text } from '@primer/react'
import { SkeletonText } from '@primer/react/experimental'
import {
  CheckCircleIcon,
  ClockIcon,
  CpuIcon,
  DatabaseIcon,
  MeterIcon,
  PulseIcon,
  ServerIcon,
  StackIcon,
  TagIcon,
  type Icon,
} from '@primer/octicons-react'

import { HealthLabel } from '../HealthLabel'
import { timeAgo } from '../../lib/utils'
import type { Cluster, ClusterNode } from '../../types/cluster'
import styles from './ClusterOverview.module.css'

interface ClusterOverviewProps {
  cluster: Cluster
  nodes: ClusterNode[]
}

type RuntimeCluster = Cluster & {
  registeredAt?: string
  version?: string
}

type RuntimeNode = ClusterNode & {
  ready: boolean
  cpuCapacity: string
  memoryCapacity: string
}

interface TimestampValue {
  relative: string
  exact: string
  iso: string
}

interface CapacityMetric {
  label: string
  value: string
  coverage: string
  icon: Icon
}

export function ClusterOverview({ cluster, nodes }: ClusterOverviewProps) {
  const runtimeCluster = cluster as RuntimeCluster
  const version = cluster.k8sVersion || runtimeCluster.version
  const registered = formatTimestamp(runtimeCluster.registeredAt)
  const heartbeat = formatTimestamp(cluster.lastHeartbeat)
  const labels = Object.entries(cluster.labels ?? {}).sort(([left], [right]) => left.localeCompare(right))
  const readiness = getReadiness(nodes)
  const capacity = getCapacityMetrics(nodes)

  return (
    <section className={styles.overview} aria-label={`${cluster.name} overview`}>
      <div className={styles.header}>
        <div className={styles.headerMain}>
          <div className={styles.badges}>
            <HealthLabel health={cluster.health} />
            {version ? (
              <Label variant="accent">Kubernetes {version}</Label>
            ) : (
              <Label variant="secondary">Kubernetes version unavailable</Label>
            )}
          </div>
          <h1 className={styles.name}>{cluster.name}</h1>
          <Text className={styles.description}>
            Current health, reported workload totals, node readiness, and cluster capacity.
          </Text>
        </div>

        <dl className={styles.timestamps}>
          <TimestampRow icon={PulseIcon} label="Last heartbeat" timestamp={heartbeat} />
          <TimestampRow icon={ClockIcon} label="Registered" timestamp={registered} />
        </dl>
      </div>

      <dl className={styles.metrics}>
        <OverviewMetric icon={ServerIcon} label="Nodes" value={formatCount(cluster.nodeCount)} />
        <OverviewMetric icon={StackIcon} label="Pods" value={formatCount(cluster.podCount)} />
        <OverviewMetric
          icon={CheckCircleIcon}
          label="Node readiness"
          value={readiness?.value ?? '—'}
          detail={readiness?.detail ?? 'Readiness unavailable'}
        />
        <OverviewMetric
          icon={MeterIcon}
          label="Node snapshot"
          value={nodes.length > 0 ? nodes.length.toLocaleString() : '—'}
          detail={
            nodes.length > 0 ? `${pluralize(nodes.length, 'node')} reporting details` : 'No node details reported'
          }
        />
      </dl>

      <div className={styles.panels}>
        <section className={styles.panel} aria-labelledby="capacity-heading">
          <h2 id="capacity-heading" className={styles.panelTitle}>
            <MeterIcon size={16} className={styles.panelIcon} />
            Node capacity
          </h2>
          {capacity.length > 0 ? (
            <dl className={styles.capacity}>
              {capacity.map((metric) => (
                <CapacityMetricItem key={metric.label} metric={metric} />
              ))}
            </dl>
          ) : (
            <div className={styles.unavailable}>
              <Text weight="semibold">Capacity unavailable</Text>
              <Text className={styles.description}>
                This cluster has not reported CPU, memory, or pod capacity details.
              </Text>
            </div>
          )}
        </section>

        <section className={styles.panel} aria-labelledby="labels-heading">
          <h2 id="labels-heading" className={styles.panelTitle}>
            <TagIcon size={16} className={styles.panelIcon} />
            Labels
          </h2>
          {labels.length > 0 ? (
            <div className={styles.labels}>
              {labels.map(([key, value]) => (
                <Label key={key} variant="secondary" title={`${key}=${value}`}>
                  {key}={value}
                </Label>
              ))}
            </div>
          ) : (
            <Text className={styles.description}>No labels have been reported for this cluster.</Text>
          )}
        </section>
      </div>
    </section>
  )
}

export function ClusterOverviewSkeleton() {
  return (
    <section className={styles.overview} aria-label="Loading cluster overview" aria-busy="true">
      <div className={styles.header}>
        <div className={styles.headerMain}>
          <SkeletonText size="bodySmall" maxWidth="10rem" />
          <SkeletonText size="titleLarge" maxWidth="20rem" />
          <SkeletonText size="bodyMedium" maxWidth="28rem" />
        </div>
      </div>
      <dl className={styles.metrics}>
        {Array.from({ length: 4 }, (_, index) => (
          <div key={index} className={styles.metric}>
            <SkeletonText size="bodySmall" maxWidth="6rem" />
            <SkeletonText size="titleMedium" maxWidth="4rem" />
          </div>
        ))}
      </dl>
    </section>
  )
}

interface TimestampRowProps {
  icon: Icon
  label: string
  timestamp: TimestampValue | null
}

function TimestampRow({ icon: RowIcon, label, timestamp }: TimestampRowProps) {
  return (
    <div className={styles.timestamp}>
      <dt className={styles.metricLabel}>
        <RowIcon size={16} className={styles.panelIcon} />
        {label}
      </dt>
      <dd className={styles.timestampValue}>
        {timestamp ? (
          <time dateTime={timestamp.iso} title={timestamp.exact}>
            {timestamp.relative}
          </time>
        ) : (
          'Unavailable'
        )}
      </dd>
      <span className={styles.timestampExact} title={timestamp?.exact}>
        {timestamp?.exact ?? 'No timestamp reported'}
      </span>
    </div>
  )
}

interface OverviewMetricProps {
  icon: Icon
  label: string
  value: string
  detail?: string
}

function OverviewMetric({ icon: MetricIcon, label, value, detail }: OverviewMetricProps) {
  return (
    <div className={styles.metric}>
      <dt className={styles.metricLabel}>
        <MetricIcon size={16} className={styles.panelIcon} />
        {label}
      </dt>
      <dd className={styles.metricValue}>{value}</dd>
      {detail && <span className={styles.metricDetail}>{detail}</span>}
    </div>
  )
}

function CapacityMetricItem({ metric }: { metric: CapacityMetric }) {
  const MetricIcon = metric.icon

  return (
    <div className={styles.capacityItem}>
      <dt className={styles.metricLabel}>
        <MetricIcon size={16} className={styles.panelIcon} />
        {metric.label}
      </dt>
      <dd className={styles.capacityValue} title={metric.value}>
        {metric.value}
      </dd>
      <span className={styles.metricDetail}>{metric.coverage}</span>
    </div>
  )
}

function formatCount(value: number): string {
  return Number.isFinite(value) ? value.toLocaleString() : '—';
}

function formatTimestamp(value: string | undefined): TimestampValue | null {
  if (!value) return null;
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return null;

  return {
    relative: timeAgo(value),
    iso: parsed.toISOString(),
    exact: new Intl.DateTimeFormat(undefined, {
      dateStyle: 'medium',
      timeStyle: 'short',
    }).format(parsed),
  };
}

function getReadiness(nodes: ClusterNode[]): { value: string; detail: string } | null {
  const readiness = nodes.map((node) => {
    const runtimeNode = node as RuntimeNode;
    if (typeof runtimeNode.ready === 'boolean') return runtimeNode.ready;
    if (!node.status) return null;
    const status = node.status.trim().toLowerCase();
    if (!status) return null;
    return status === 'ready';
  });
  const known = readiness.filter((value): value is boolean => value !== null);
  if (known.length === 0) return null;

  const ready = known.filter(Boolean).length;
  return {
    value: `${ready}/${known.length}`,
    detail: ready === known.length ? 'All reporting nodes ready' : `${known.length - ready} not ready`,
  };
}

function getCapacityMetrics(nodes: ClusterNode[]): CapacityMetric[] {
  if (nodes.length === 0) return [];

  const cpu = nodes.map((node) => getCapacity(node, 'cpu')).filter(isString);
  const memory = nodes.map((node) => getCapacity(node, 'memory')).filter(isString);
  const pods = nodes.map((node) => getCapacity(node, 'pods')).filter(isString);
  const metrics: CapacityMetric[] = [];

  if (cpu.length > 0) {
    metrics.push({
      label: 'CPU capacity',
      value: formatCpuTotal(cpu) ?? (cpu.length === 1 ? cpu[0] : 'Reported'),
      coverage: coverageLabel(cpu.length, nodes.length),
      icon: CpuIcon,
    });
  }
  if (memory.length > 0) {
    metrics.push({
      label: 'Memory capacity',
      value: formatMemoryTotal(memory) ?? (memory.length === 1 ? memory[0] : 'Reported'),
      coverage: coverageLabel(memory.length, nodes.length),
      icon: DatabaseIcon,
    });
  }
  if (pods.length > 0) {
    metrics.push({
      label: 'Pod capacity',
      value: formatIntegerTotal(pods) ?? (pods.length === 1 ? pods[0] : 'Reported'),
      coverage: coverageLabel(pods.length, nodes.length),
      icon: StackIcon,
    });
  }

  return metrics;
}

function getCapacity(node: ClusterNode, key: 'cpu' | 'memory' | 'pods'): string | null {
  const runtimeNode = node as RuntimeNode;
  if (key === 'cpu' && runtimeNode.cpuCapacity) return runtimeNode.cpuCapacity;
  if (key === 'memory' && runtimeNode.memoryCapacity) return runtimeNode.memoryCapacity;

  return null;
}

function formatCpuTotal(values: string[]): string | null {
  const milliCores = values.map((value) => {
    const match = value.trim().match(/^(\d+(?:\.\d+)?)(m)?$/);
    if (!match) return null;
    const amount = Number(match[1]);
    return match[2] ? amount : amount * 1_000;
  });
  if (milliCores.some((value) => value === null)) return null;

  const total = milliCores.reduce<number>((sum, value) => sum + (value ?? 0), 0);
  return `${new Intl.NumberFormat(undefined, { maximumFractionDigits: 3 }).format(total / 1_000)} cores`;
}

function formatMemoryTotal(values: string[]): string | null {
  const factors: Record<string, number> = {
    '': 1,
    Ki: 1024,
    Mi: 1024 ** 2,
    Gi: 1024 ** 3,
    Ti: 1024 ** 4,
    Pi: 1024 ** 5,
    K: 1000,
    M: 1000 ** 2,
    G: 1000 ** 3,
    T: 1000 ** 4,
    P: 1000 ** 5,
  };
  const bytes = values.map((value) => {
    const match = value.trim().match(/^(\d+(?:\.\d+)?)(Ki|Mi|Gi|Ti|Pi|K|M|G|T|P)?$/);
    if (!match) return null;
    return Number(match[1]) * factors[match[2] ?? ''];
  });
  if (bytes.some((value) => value === null)) return null;

  const total = bytes.reduce<number>((sum, value) => sum + (value ?? 0), 0);
  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB'];
  let unitIndex = 0;
  let displayValue = total;
  while (displayValue >= 1024 && unitIndex < units.length - 1) {
    displayValue /= 1024;
    unitIndex += 1;
  }
  return `${new Intl.NumberFormat(undefined, { maximumFractionDigits: 1 }).format(displayValue)} ${units[unitIndex]}`;
}

function formatIntegerTotal(values: string[]): string | null {
  if (values.some((value) => !/^\d+$/.test(value.trim()))) return null;
  const total = values.reduce((sum, value) => sum + Number(value), 0);
  return total.toLocaleString();
}

function coverageLabel(reported: number, total: number): string {
  return reported === total ? `Across ${pluralize(total, 'node')}` : `${reported} of ${total} nodes reported`;
}

function pluralize(value: number, noun: string): string {
  return `${value.toLocaleString()} ${noun}${value === 1 ? '' : 's'}`;
}

function isString(value: string | null): value is string {
  return value !== null;
}
