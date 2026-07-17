import type { Cluster, ClusterStatus } from '@/types/cluster';
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
  const res = await fetch(`${BASE}${path}`, {
    method: 'GET',
    headers: { Accept: 'application/json' },
    signal,
  });

  if (!res.ok) {
    let body: unknown;
    try {
      body = await res.json();
    } catch {
      body = await res.text().catch(() => undefined);
    }
    const message = isErrorBody(body) ? body.error : `request failed with status ${res.status}`;
    throw new ApiError(res.status, message, body);
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return (await res.json()) as T;
}

function qs(params: Record<string, string | undefined>): string {
  const entries = Object.entries(params).filter((entry): entry is [string, string] => Boolean(entry[1]));
  if (entries.length === 0) return '';
  return `?${new URLSearchParams(entries).toString()}`;
}

export const api = {
  // Backed by GET /api/v1/clusters (internal/server/handlers_clusters.go).
  listClusters: (signal?: AbortSignal) => get<{ clusters: Cluster[] }>('/clusters', signal).then((r) => r.clusters),
  // Backed by GET /api/v1/clusters/{id}.
  getCluster: (id: string, signal?: AbortSignal) => get<Cluster>(`/clusters/${id}`, signal),
  // Backed by GET /api/v1/clusters/{id}/status.
  getClusterStatus: (id: string, signal?: AbortSignal) => get<ClusterStatus>(`/clusters/${id}/status`, signal),
  // Backed by GET /api/v1/clusters/{id}/events (currently a stub that always returns []).
  getEvents: (id: string, ns?: string, signal?: AbortSignal) =>
    get<EventInfo[]>(`/clusters/${id}/events${qs({ namespace: ns })}`, signal),
  // Not implemented server-side yet; follows the same route convention as the
  // handlers above so the UI can slot in once the hub adds these endpoints.
  getPods: (id: string, ns?: string, signal?: AbortSignal) =>
    get<PodInfo[]>(`/clusters/${id}/pods${qs({ namespace: ns })}`, signal),
  getServices: (id: string, ns?: string, signal?: AbortSignal) =>
    get<ServiceInfo[]>(`/clusters/${id}/services${qs({ namespace: ns })}`, signal),
  getDeployments: (id: string, ns?: string, signal?: AbortSignal) =>
    get<DeploymentInfo[]>(`/clusters/${id}/deployments${qs({ namespace: ns })}`, signal),
  getNamespaces: (id: string, signal?: AbortSignal) => get<string[]>(`/clusters/${id}/namespaces`, signal),
};
