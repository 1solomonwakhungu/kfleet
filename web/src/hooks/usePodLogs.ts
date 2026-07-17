import { useCallback, useEffect, useState } from 'react';

interface UsePodLogsOptions {
  clusterId: string;
  namespace: string;
  pod: string;
  container?: string;
  follow?: boolean;
}

// Streams GET /api/v1/clusters/:id/pods/:ns/:pod/logs via SSE. This endpoint
// does not exist in the hub yet; EventSource retries automatically on error,
// so the UI just reflects connection state rather than tearing down.
export function usePodLogs({ clusterId, namespace, pod, container, follow = true }: UsePodLogsOptions) {
  const [lines, setLines] = useState<string[]>([]);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const clear = useCallback(() => setLines([]), []);

  useEffect(() => {
    if (!clusterId || !namespace || !pod) return;

    const params = new URLSearchParams();
    if (container) params.set('container', container);
    params.set('follow', String(follow));
    const url = `/api/v1/clusters/${clusterId}/pods/${namespace}/${pod}/logs?${params.toString()}`;

    let cancelled = false;
    const source = new EventSource(url);

    source.onopen = () => {
      if (cancelled) return;
      setConnected(true);
      setError(null);
    };

    source.onmessage = (event) => {
      if (cancelled) return;
      setLines((prev) => [...prev, event.data]);
    };

    source.onerror = () => {
      if (cancelled) return;
      setConnected(false);
      setError('Log stream disconnected, retrying…');
    };

    return () => {
      cancelled = true;
      source.close();
      setConnected(false);
    };
  }, [clusterId, namespace, pod, container, follow]);

  return { lines, connected, error, clear };
}
