import type { Cluster, ClusterStatus } from '@/types/cluster';
import type { Alert, AlertStatus } from '@/types/alert';
import type { PodInfo, EventInfo, ServiceInfo, DeploymentInfo } from '@/types/resources';

const BASE = '/api/v1';

export class ApiError extends Error {
  status: number;
  body: unknown;

  constructor(status: number, message: string, body?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.body = body;
  }
}

function isErrorBody(body: unknown): body is { error: string } {
  return typeof body === 'object' && body !== null && typeof (body as { error?: unknown }).error === 'string';
}

async function get<T>(path: string, signal?: AbortSignal): Promise<T> {
  return request<T>('GET', path, undefined, signal);
}

async function request<T>(method: string, path: string, payload?: unknown, signal?: AbortSignal): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method,
    headers: {
      Accept: 'application/json',
      ...(payload === undefined ? {} : { 'Content-Type': 'application/json' }),
    },
    body: payload === undefined ? undefined : JSON.stringify(payload),
    signal,
  });

  const text = await res.text();
  let body: unknown;
  if (text) {
    try {
      body = JSON.parse(text) as unknown;
    } catch {
      if (res.ok) {
        throw new ApiError(res.status, 'response was not valid JSON', text);
      }
      body = text;
    }
  }

  if (!res.ok) {
    const message = isErrorBody(body)
      ? body.error
      : typeof body === 'string'
        ? body
        : `request failed with status ${res.status}`;
    throw new ApiError(res.status, message, body);
  }

  if (body === undefined) {
    return undefined as T;
  }

  return body as T;
}

function qs(params: Record<string, string | undefined>): string {
  const entries = Object.entries(params).filter((entry): entry is [string, string] => Boolean(entry[1]));
  if (entries.length === 0) return '';
  return `?${new URLSearchParams(entries).toString()}`;
}

function clusterPath(id: string, suffix = ''): string {
  return `/clusters/${encodeURIComponent(id)}${suffix}`;
}

export interface WireCluster extends Omit<Cluster, 'k8sVersion' | 'agentVersion' | 'registeredAt' | 'labels'> {
  version?: string;
  k8sVersion?: string;
  agentVersion?: string;
  registeredAt?: string;
  labels?: Record<string, string> | null;
}

export function normalizeCluster(cluster: WireCluster): Cluster {
  return {
    ...cluster,
    k8sVersion: cluster.k8sVersion ?? cluster.version ?? '',
    agentVersion: cluster.agentVersion ?? '',
    registeredAt: cluster.registeredAt ?? '',
    labels: cluster.labels ?? {},
  };
}

export const api = {
  // Backed by GET /api/v1/clusters (internal/server/handlers_clusters.go).
  listClusters: (signal?: AbortSignal) => get<{ clusters: WireCluster[] }>('/clusters', signal).then((r) => r.clusters.map(normalizeCluster)),
  // Backed by GET /api/v1/clusters/{id}.
  getCluster: (id: string, signal?: AbortSignal) => get<WireCluster>(clusterPath(id), signal).then(normalizeCluster),
  // Backed by GET /api/v1/clusters/{id}/status.
  getClusterStatus: (id: string, signal?: AbortSignal) =>
    get<Omit<ClusterStatus, 'cluster'> & { cluster: WireCluster }>(clusterPath(id, '/status'), signal)
      .then((status) => ({ ...status, cluster: normalizeCluster(status.cluster) })),
  // Backed by persisted snapshot resource endpoints in handlers_clusters.go.
  getEvents: (id: string, ns?: string, signal?: AbortSignal) =>
    get<EventInfo[]>(`${clusterPath(id, '/events')}${qs({ namespace: ns })}`, signal),
  getPods: (id: string, ns?: string, signal?: AbortSignal) =>
    get<PodInfo[]>(`${clusterPath(id, '/pods')}${qs({ namespace: ns })}`, signal),
  getServices: (id: string, ns?: string, signal?: AbortSignal) =>
    get<ServiceInfo[]>(`${clusterPath(id, '/services')}${qs({ namespace: ns })}`, signal),
  getDeployments: (id: string, ns?: string, signal?: AbortSignal) =>
    get<DeploymentInfo[]>(`${clusterPath(id, '/deployments')}${qs({ namespace: ns })}`, signal),
  getNamespaces: (id: string, signal?: AbortSignal) => get<string[]>(clusterPath(id, '/namespaces'), signal),
  listAlerts: (status?: AlertStatus, signal?: AbortSignal) =>
    get<{ alerts: Alert[] }>(`/alerts${qs({ status })}`, signal).then((response) => response.alerts),
  acknowledgeAlert: (id: string, acknowledgedBy = 'operator', signal?: AbortSignal) =>
    request<Alert>(
      'POST',
      `/alerts/${encodeURIComponent(id)}/acknowledge`,
      { acknowledgedBy },
      signal,
    ),
};
