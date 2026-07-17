import { useCallback, useEffect, useState } from 'react';
import { api, ApiError } from '@/lib/api';
import type { Cluster, ClusterNode } from '@/types/cluster';
import type { PodInfo, ServiceInfo, DeploymentInfo, EventInfo } from '@/types/resources';

interface ResourceState<T> {
  data: T;
  loading: boolean;
  error: string | null;
}

function idle<T>(initial: T): ResourceState<T> {
  return { data: initial, loading: true, error: null };
}

export function useClusterDetail(clusterId: string | undefined) {
  const [cluster, setCluster] = useState<Cluster | null>(null);
  const [nodes, setNodes] = useState<ClusterNode[]>([]);
  const [pods, setPods] = useState<ResourceState<PodInfo[]>>(idle([]));
  const [services, setServices] = useState<ResourceState<ServiceInfo[]>>(idle([]));
  const [deployments, setDeployments] = useState<ResourceState<DeploymentInfo[]>>(idle([]));
  const [events, setEvents] = useState<ResourceState<EventInfo[]>>(idle([]));
  const [namespaces, setNamespaces] = useState<string[]>([]);
  const [namespace, setNamespace] = useState<string | undefined>(undefined);
  const [statusError, setStatusError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(
    async (signal: AbortSignal) => {
      if (!clusterId) return;
      setLoading(true);

      try {
        const status = await api.getClusterStatus(clusterId, signal);
        setCluster(status.cluster);
        setNodes(status.nodes);
        setStatusError(null);
      } catch (err) {
        if (err instanceof ApiError) setStatusError(err.message);
      } finally {
        setLoading(false);
      }

      async function fetchResource<T>(
        fn: (id: string, ns: string | undefined, signal: AbortSignal) => Promise<T>,
        set: (state: ResourceState<T>) => void,
        fallback: T,
      ) {
        set({ data: fallback, loading: true, error: null });
        try {
          const data = await fn(clusterId as string, namespace, signal);
          set({ data, loading: false, error: null });
        } catch (err) {
          if (err instanceof ApiError) {
            set({ data: fallback, loading: false, error: err.message });
          }
        }
      }

      await Promise.allSettled([
        fetchResource(api.getPods, setPods, []),
        fetchResource(api.getServices, setServices, []),
        fetchResource(api.getDeployments, setDeployments, []),
        fetchResource(api.getEvents, setEvents, []),
      ]);

      try {
        const ns = await api.getNamespaces(clusterId, signal);
        setNamespaces(ns);
      } catch {
        // Endpoint may not exist yet; namespaces are derived from pods below instead.
      }
    },
    [clusterId, namespace],
  );

  useEffect(() => {
    const controller = new AbortController();
    load(controller.signal);
    return () => controller.abort();
  }, [load]);

  const derivedNamespaces =
    namespaces.length > 0 ? namespaces : Array.from(new Set(pods.data.map((p) => p.namespace))).sort();

  return {
    cluster,
    nodes,
    pods,
    services,
    deployments,
    events,
    namespaces: derivedNamespaces,
    namespace,
    setNamespace,
    loading,
    statusError,
  };
}
