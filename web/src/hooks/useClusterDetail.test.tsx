import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { api, ApiError } from '@/lib/api';
import type { ClusterStatus } from '@/types/cluster';
import { useClusterDetail } from './useClusterDetail';

const status: ClusterStatus = {
  cluster: {
    id: 'cluster-a',
    name: 'Cluster A',
    health: 'healthy',
    nodeCount: 1,
    podCount: 2,
    k8sVersion: '1.31',
    agentVersion: '0.1',
    lastHeartbeat: '2026-07-19T12:00:00Z',
    registeredAt: '2026-07-19T11:00:00Z',
    labels: {},
  },
  nodes: [],
};

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

describe('useClusterDetail', () => {
  beforeEach(() => {
    vi.spyOn(api, 'getClusterStatus').mockResolvedValue(status);
    vi.spyOn(api, 'getPods').mockResolvedValue([]);
    vi.spyOn(api, 'getServices').mockResolvedValue([]);
    vi.spyOn(api, 'getDeployments').mockResolvedValue([]);
    vi.spyOn(api, 'getEvents').mockResolvedValue([]);
    vi.spyOn(api, 'getNamespaces').mockResolvedValue([]);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('settles resource loading state and exposes network errors', async () => {
    vi.mocked(api.getPods).mockRejectedValue(new Error('network unavailable'));
    vi.mocked(api.getServices).mockRejectedValue(new Error('network unavailable'));
    vi.mocked(api.getDeployments).mockRejectedValue(new Error('network unavailable'));
    vi.mocked(api.getEvents).mockRejectedValue(new Error('network unavailable'));

    const { result } = renderHook(() => useClusterDetail('cluster-a'));

    await waitFor(() => expect(result.current.pods.loading).toBe(false));
    for (const resource of [result.current.pods, result.current.services, result.current.deployments, result.current.events]) {
      expect(resource).toMatchObject({ data: [], loading: false, error: 'network unavailable' });
    }
  });

  it('settles aborted resources without surfacing user-facing errors', async () => {
    const abortError = new DOMException('The operation was aborted', 'AbortError');
    vi.mocked(api.getPods).mockRejectedValue(abortError);
    vi.mocked(api.getServices).mockRejectedValue(abortError);
    vi.mocked(api.getDeployments).mockRejectedValue(abortError);
    vi.mocked(api.getEvents).mockRejectedValue(abortError);

    const { result } = renderHook(() => useClusterDetail('cluster-a'));

    await waitFor(() => expect(result.current.pods.loading).toBe(false));
    for (const resource of [result.current.pods, result.current.services, result.current.deployments, result.current.events]) {
      expect(resource).toMatchObject({ data: [], loading: false, error: null });
    }
  });

  it('ignores stale errors and loading updates after a request is aborted', async () => {
    const firstStatus = deferred<ClusterStatus>();
    const secondStatus = deferred<ClusterStatus>();
    vi.mocked(api.getClusterStatus)
      .mockImplementationOnce(() => firstStatus.promise)
      .mockImplementationOnce(() => secondStatus.promise);

    const { result, rerender } = renderHook(({ id }) => useClusterDetail(id), {
      initialProps: { id: 'cluster-a' },
    });
    await waitFor(() => expect(api.getClusterStatus).toHaveBeenCalledTimes(1));

    rerender({ id: 'cluster-b' });
    await waitFor(() => expect(api.getClusterStatus).toHaveBeenCalledTimes(2));
    await act(async () => {
      firstStatus.reject(new ApiError(500, 'stale request failed'));
      await Promise.resolve();
    });

    expect(result.current.statusError).toBeNull();
    expect(result.current.loading).toBe(true);

    await act(async () => {
      secondStatus.resolve(status);
    });
  });
});
