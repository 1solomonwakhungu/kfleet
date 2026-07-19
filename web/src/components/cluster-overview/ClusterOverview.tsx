/* Hallmark · pre-emit critique: P4 H5 E4 S5 R5 V4 */
/* Hallmark · genre: atmospheric · macrostructure: Stat-Led · theme: Midnight · tone: austere technical · anchor hue: blue */
import {
  Activity,
  Boxes,
  CalendarClock,
  CheckCircle2,
  Cpu,
  Gauge,
  MemoryStick,
  Server,
  Tag,
} from 'lucide-react';
import { HealthBadge } from '@/components/HealthBadge';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { timeAgo } from '@/lib/utils';
import type { Cluster, ClusterNode } from '@/types/cluster';

interface ClusterOverviewProps {
  cluster: Cluster;
  nodes: ClusterNode[];
}

type RuntimeCluster = Cluster & {
  registeredAt?: string;
  version?: string;
};

type RuntimeNode = ClusterNode & {
  ready?: boolean;
  cpuCapacity?: string;
  memoryCapacity?: string;
};

interface TimestampValue {
  relative: string;
  exact: string;
  iso: string;
}

interface CapacityMetric {
  label: string;
  value: string;
  coverage: string;
  icon: typeof Cpu;
}

export function ClusterOverview({ cluster, nodes }: ClusterOverviewProps) {
  const runtimeCluster = cluster as RuntimeCluster;
  const version = cluster.k8sVersion || runtimeCluster.version;
  const registered = formatTimestamp(runtimeCluster.registeredAt);
  const heartbeat = formatTimestamp(cluster.lastHeartbeat);
  const labels = Object.entries(cluster.labels ?? {}).sort(([left], [right]) => left.localeCompare(right));
  const readiness = getReadiness(nodes);
  const capacity = getCapacityMetrics(nodes);

  return (
    <Card className="overflow-hidden border border-border" aria-label={`${cluster.name} overview`}>
      <CardHeader className="gap-5 border-b border-border p-5 sm:p-6 lg:flex-row lg:items-start lg:justify-between">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <HealthBadge health={cluster.health} />
            {version ? (
              <Badge className="border border-blue-400/30 bg-blue-500/10 font-mono text-blue-300">
                Kubernetes {version}
              </Badge>
            ) : (
              <Badge className="bg-elevated text-muted">Kubernetes version unavailable</Badge>
            )}
          </div>
          <h1 className="mt-4 min-w-0 font-display text-3xl font-bold tracking-tight [overflow-wrap:anywhere] sm:text-4xl">
            {cluster.name}
          </h1>
          <p className="mt-2 max-w-2xl text-base leading-relaxed text-muted">
            Current health, reported workload totals, node readiness, and cluster capacity.
          </p>
        </div>

        <dl className="grid min-w-0 gap-x-5 gap-y-4 border-t border-border pt-5 sm:grid-cols-2 lg:w-[28rem] lg:border-t-0 lg:pt-0">
          <TimestampRow icon={Activity} label="Last heartbeat" timestamp={heartbeat} />
          <TimestampRow icon={CalendarClock} label="Registered" timestamp={registered} />
        </dl>
      </CardHeader>

      <CardContent className="p-0">
        <dl className="grid grid-cols-2 sm:grid-cols-4">
          <OverviewMetric
            className="border-b border-r border-border sm:border-b-0"
            icon={Server}
            label="Nodes"
            value={formatCount(cluster.nodeCount)}
          />
          <OverviewMetric
            className="border-b border-border sm:border-b-0 sm:border-r"
            icon={Boxes}
            label="Pods"
            value={formatCount(cluster.podCount)}
          />
          <OverviewMetric
            className="border-r border-border"
            icon={CheckCircle2}
            label="Node readiness"
            value={readiness?.value ?? '—'}
            detail={readiness?.detail ?? 'Readiness unavailable'}
          />
          <OverviewMetric
            icon={Gauge}
            label="Node snapshot"
            value={nodes.length > 0 ? nodes.length.toLocaleString() : '—'}
            detail={nodes.length > 0 ? `${pluralize(nodes.length, 'node')} reporting details` : 'No node details reported'}
          />
        </dl>

        <div className="grid border-t border-border lg:grid-cols-[minmax(0,1.35fr)_minmax(18rem,0.65fr)]">
          <section className="min-w-0 p-5 sm:p-6" aria-labelledby="capacity-heading">
            <div className="flex items-center gap-2">
              <Gauge className="h-4 w-4 text-blue-300" aria-hidden="true" />
              <h2 id="capacity-heading" className="font-display text-lg font-bold">
                Node capacity
              </h2>
            </div>
            {capacity.length > 0 ? (
              <dl className="mt-4 grid gap-x-5 gap-y-4 sm:grid-cols-3">
                {capacity.map((metric) => (
                  <CapacityMetricItem key={metric.label} metric={metric} />
                ))}
              </dl>
            ) : (
              <div className="mt-4 rounded-md border border-dashed border-border px-4 py-5">
                <p className="font-medium">Capacity unavailable</p>
                <p className="mt-1 text-sm leading-relaxed text-muted">
                  This cluster has not reported CPU, memory, or pod capacity details.
                </p>
              </div>
            )}
          </section>

          <section
            className="min-w-0 border-t border-border p-5 sm:p-6 lg:border-l lg:border-t-0"
            aria-labelledby="labels-heading"
          >
            <div className="flex items-center gap-2">
              <Tag className="h-4 w-4 text-blue-300" aria-hidden="true" />
              <h2 id="labels-heading" className="font-display text-lg font-bold">
                Labels
              </h2>
            </div>
            {labels.length > 0 ? (
              <div className="mt-4 flex flex-wrap gap-2">
                {labels.map(([key, value]) => (
                  <Badge key={key} className="max-w-full border border-border bg-background font-mono text-foreground">
                    <span className="block max-w-full truncate" title={`${key}=${value}`}>
                      <span className="text-blue-300">{key}</span>={value}
                    </span>
                  </Badge>
                ))}
              </div>
            ) : (
              <p className="mt-4 text-sm leading-relaxed text-muted">No labels have been reported for this cluster.</p>
            )}
          </section>
        </div>
      </CardContent>
    </Card>
  );
}

export function ClusterOverviewSkeleton() {
  return (
    <Card className="overflow-hidden border border-border" aria-label="Loading cluster overview" aria-busy="true">
      <CardHeader className="gap-4 border-b border-border p-5 sm:p-6">
        <div className="h-6 w-40 animate-pulse rounded bg-elevated" />
        <div className="h-10 w-2/3 max-w-lg animate-pulse rounded bg-elevated" />
        <div className="h-5 w-full max-w-xl animate-pulse rounded bg-elevated" />
      </CardHeader>
      <CardContent className="p-0">
        <div className="grid grid-cols-2 divide-x divide-y divide-border sm:grid-cols-4 sm:divide-y-0">
          {Array.from({ length: 4 }, (_, index) => (
            <div key={index} className="p-5 sm:p-6">
              <div className="h-4 w-20 animate-pulse rounded bg-elevated" />
              <div className="mt-3 h-8 w-16 animate-pulse rounded bg-elevated" />
            </div>
          ))}
        </div>
        <div className="grid gap-4 border-t border-border p-5 sm:grid-cols-3 sm:p-6">
          {Array.from({ length: 3 }, (_, index) => (
            <div key={index} className="h-24 animate-pulse rounded-md bg-elevated" />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

interface TimestampRowProps {
  icon: typeof Activity;
  label: string;
  timestamp: TimestampValue | null;
}

function TimestampRow({ icon: Icon, label, timestamp }: TimestampRowProps) {
  return (
    <div className="min-w-0 border-t border-border pt-3 first:border-t-0 first:pt-0 sm:border-t-0 sm:border-l sm:pl-5 sm:first:border-l-0 sm:first:pl-0">
      <dt className="flex items-center gap-2 text-sm text-muted">
        <Icon className="h-4 w-4 shrink-0 text-blue-300" aria-hidden="true" />
        {label}
      </dt>
      <dd className="mt-2 font-mono text-sm font-semibold tabular-nums">
        {timestamp ? (
          <time dateTime={timestamp.iso} title={timestamp.exact}>
            {timestamp.relative}
          </time>
        ) : (
          'Unavailable'
        )}
      </dd>
      <p className="mt-1 truncate text-sm text-muted" title={timestamp?.exact}>
        {timestamp?.exact ?? 'No timestamp reported'}
      </p>
    </div>
  );
}

interface OverviewMetricProps {
  icon: typeof Server;
  label: string;
  value: string;
  detail?: string;
  className?: string;
}

function OverviewMetric({ icon: Icon, label, value, detail, className = '' }: OverviewMetricProps) {
  return (
    <div className={`min-w-0 p-5 sm:p-6 ${className}`}>
      <dt className="flex items-center gap-2 text-sm text-muted">
        <Icon className="h-4 w-4 shrink-0 text-blue-300" aria-hidden="true" />
        {label}
      </dt>
      <dd className="mt-2 font-display text-2xl font-bold tabular-nums">{value}</dd>
      {detail && <p className="mt-1 text-sm leading-relaxed text-muted">{detail}</p>}
    </div>
  );
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

function CapacityMetricItem({ metric }: { metric: CapacityMetric }) {
  const Icon = metric.icon;
  return (
    <div className="min-w-0 border-t border-border pt-3">
      <dt className="flex items-center gap-2 text-sm text-muted">
        <Icon className="h-4 w-4 shrink-0 text-blue-300" aria-hidden="true" />
        {metric.label}
      </dt>
      <dd className="mt-2 truncate font-mono text-lg font-semibold tabular-nums" title={metric.value}>
        {metric.value}
      </dd>
      <p className="mt-1 text-sm text-muted">{metric.coverage}</p>
    </div>
  );
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
      icon: Cpu,
    });
  }
  if (memory.length > 0) {
    metrics.push({
      label: 'Memory capacity',
      value: formatMemoryTotal(memory) ?? (memory.length === 1 ? memory[0] : 'Reported'),
      coverage: coverageLabel(memory.length, nodes.length),
      icon: MemoryStick,
    });
  }
  if (pods.length > 0) {
    metrics.push({
      label: 'Pod capacity',
      value: formatIntegerTotal(pods) ?? (pods.length === 1 ? pods[0] : 'Reported'),
      coverage: coverageLabel(pods.length, nodes.length),
      icon: Boxes,
    });
  }

  return metrics;
}

function getCapacity(node: ClusterNode, key: 'cpu' | 'memory' | 'pods'): string | null {
  const runtimeNode = node as RuntimeNode;
  if (key === 'cpu' && runtimeNode.cpuCapacity) return runtimeNode.cpuCapacity;
  if (key === 'memory' && runtimeNode.memoryCapacity) return runtimeNode.memoryCapacity;

  const value = node.capacity?.[key];
  return typeof value === 'string' && value.trim() ? value.trim() : null;
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
