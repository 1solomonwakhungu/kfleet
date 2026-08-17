import { useCallback, useEffect, useRef, useState } from 'react';

interface UsePodLogsOptions {
  clusterId: string;
  namespace: string;
  pod: string;
  container?: string;
  follow?: boolean;
}

export type PodLogStatus =
  | 'idle'
  | 'connecting'
  | 'streaming'
  | 'reconnecting'
  | 'ended'
  | 'unavailable';

export interface UsePodLogsResult {
  lines: string[];
  status: PodLogStatus;
  error: string | null;
  clear: () => void;
  retry: () => void;
}

const UNAVAILABLE_FALLBACK = 'Live logs are unavailable for this cluster right now.';

function buildLogUrl(
  clusterId: string,
  namespace: string,
  pod: string,
  container: string | undefined,
  follow: boolean,
): string {
  const params = new URLSearchParams();
  if (container) params.set('container', container);
  params.set('follow', String(follow));
  return `/api/v1/clusters/${encodeURIComponent(clusterId)}/pods/${encodeURIComponent(
    namespace,
  )}/${encodeURIComponent(pod)}/logs?${params.toString()}`;
}

// Streams GET /api/v1/clusters/:id/pods/:ns/:pod/logs as SSE. The hub relays
// the stream from the cluster's agent, so a cluster without a connected agent
// answers with a plain JSON error instead of an event stream. EventSource
// treats that as a permanent failure, which lets the UI surface a real error
// rather than retrying forever.
export function usePodLogs({
  clusterId,
  namespace,
  pod,
  container,
  follow = true,
}: UsePodLogsOptions): UsePodLogsResult {
  const [lines, setLines] = useState<string[]>([]);
  const [status, setStatus] = useState<PodLogStatus>('idle');
  const [error, setError] = useState<string | null>(null);
  const [attempt, setAttempt] = useState(0);
  const abortRef = useRef<AbortController | null>(null);

  const clear = useCallback(() => setLines([]), []);
  const retry = useCallback(() => {
    setError(null);
    setAttempt((current) => current + 1);
  }, []);

  useEffect(() => {
    if (!clusterId || !namespace || !pod) {
      setStatus('idle');
      setError(null);
      return;
    }

    const url = buildLogUrl(clusterId, namespace, pod, container, follow);
    let cancelled = false;
    const source = new EventSource(url);
    const abort = new AbortController();
    abortRef.current = abort;

    setStatus('connecting');
    setError(null);

    const fail = (message: string) => {
      if (cancelled) return;
      source.close();
      setStatus('unavailable');
      setError(message);
    };

    // EventSource cannot expose the body of a failed response, so the reason
    // for a permanent failure is fetched separately and reported verbatim.
    const explainFailure = async () => {
      try {
        const response = await fetch(buildLogUrl(clusterId, namespace, pod, container, false), {
          signal: abort.signal,
          headers: { Accept: 'application/json' },
        });
        if (response.ok) {
          fail(UNAVAILABLE_FALLBACK);
          return;
        }
        const payload = (await response.json()) as { error?: string };
        fail(payload.error || UNAVAILABLE_FALLBACK);
      } catch {
        if (!abort.signal.aborted) fail(UNAVAILABLE_FALLBACK);
      }
    };

    source.onopen = () => {
      if (cancelled) return;
      setStatus('streaming');
      setError(null);
    };

    source.onmessage = (event) => {
      if (cancelled) return;
      setLines((previous) => [...previous, event.data]);
    };

    source.addEventListener('end', () => {
      if (cancelled) return;
      source.close();
      setStatus('ended');
      setError(null);
    });

    source.addEventListener('stream-error', (event) => {
      if (cancelled) return;
      let message = UNAVAILABLE_FALLBACK;
      try {
        const payload = JSON.parse((event as MessageEvent<string>).data) as { message?: string };
        if (payload.message) message = payload.message;
      } catch {
        // Keep the fallback message when the payload is not valid JSON.
      }
      fail(message);
    });

    source.onerror = () => {
      if (cancelled) return;
      if (source.readyState === EventSource.CLOSED) {
        // The browser gave up, which means the hub rejected the request
        // outright. Reconnecting would only spin.
        void explainFailure();
        return;
      }
      setStatus('reconnecting');
      setError('Log stream interrupted, reconnecting…');
    };

    return () => {
      cancelled = true;
      abort.abort();
      abortRef.current = null;
      source.close();
      setStatus('idle');
    };
  }, [clusterId, namespace, pod, container, follow, attempt]);

  useEffect(() => () => abortRef.current?.abort(), []);

  return { lines, status, error, clear, retry };
}
