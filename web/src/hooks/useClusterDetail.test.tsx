import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { MemoryRouter, useLocation } from 'react-router-dom';
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

function makeRouterWrapper(initialEntry: string) {
  return function RouterWrapper({ children }: { children: ReactNode }) {
    return (
      <MemoryRouter initialEntries={[initialEntry]}>
        {children}
        <LocationProbe />
      </MemoryRouter>
    );
  };
}

function LocationProbe() {
  const location = useLocation();
  (window as unknown as { __testLocationSearch?: string }).__testLocationSearch = location.search;
  return null;
}

function currentSearch(): string {
  return (window as unknown as { __testLocationSearch?: string }).__testLocationSearch ?? '';
}

function renderDetailHook(initialEntry = '/clusters/cluster-a') {
  const wrapper = makeRouterWrapper(initialEntry);
  const rendered = renderHook(() => useClusterDetail('cluster-a'), { wrapper });
  return { ...rendered, currentSearch };
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
    delete (window as unknown as { __testLocationSearch?: string }).__testLocationSearch;
  });

  it('flags a 404 cluster status as not found', async () => {
    vi.mocked(api.getClusterStatus).mockRejectedValue(new ApiError(404, 'cluster not found'));

    const { result } = renderDetailHook();

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.cluster).toBeNull();
    expect(result.current.statusNotFound).toBe(true);
    expect(result.current.statusError).toBe('cluster not found');
  });

  it('does not flag other status errors as not found', async () => {
    vi.mocked(api.getClusterStatus).mockRejectedValue(new ApiError(503, 'hub unavailable'));

    const { result } = renderDetailHook();

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.statusNotFound).toBe(false);
    expect(result.current.statusError).toBe('hub unavailable');
  });

  it('settles resource loading state and exposes network errors', async () => {
    vi.mocked(api.getPods).mockRejectedValue(new Error('network unavailable'));
    vi.mocked(api.getServices).mockRejectedValue(new Error('network unavailable'));
    vi.mocked(api.getDeployments).mockRejectedValue(new Error('network unavailable'));
    vi.mocked(api.getEvents).mockRejectedValue(new Error('network unavailable'));

    const { result } = renderDetailHook();

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

    const { result } = renderDetailHook();

    await waitFor(() => expect(result.current.pods.loading).toBe(false));
    for (const resource of [result.current.pods, result.current.services, result.current.deployments, result.current.events]) {
      expect(resource).toMatchObject({ data: [], loading: false, error: null });
    }
  });

  it('deep links into a namespace from the URL and fetches its resources', async () => {
    const { result } = renderDetailHook('/clusters/cluster-a?namespace=payments');

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.namespace).toBe('payments');
    expect(api.getPods).toHaveBeenCalledWith('cluster-a', 'payments', expect.any(AbortSignal));
    expect(api.getServices).toHaveBeenCalledWith('cluster-a', 'payments', expect.any(AbortSignal));
  });

  it('writes a namespace change to the URL with replace and refetches resources', async () => {
    const { result, currentSearch: search } = renderDetailHook();

    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => result.current.setNamespace('payments'));

    await waitFor(() => expect(api.getPods).toHaveBeenCalledWith('cluster-a', 'payments', expect.any(AbortSignal)));
    expect(result.current.namespace).toBe('payments');
    expect(search()).toBe('?namespace=payments');
  });

  it('clears the namespace param when the filter is reset to all namespaces', async () => {
    const { result, currentSearch: search } = renderDetailHook('/clusters/cluster-a?namespace=payments');

    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => result.current.setNamespace(undefined));

    await waitFor(() => expect(api.getPods).toHaveBeenCalledWith('cluster-a', undefined, expect.any(AbortSignal)));
    expect(result.current.namespace).toBeUndefined();
    expect(search()).toBe('');
  });

  it('ignores stale errors and loading updates after a request is aborted', async () => {
    const firstStatus = deferred<ClusterStatus>();
    const secondStatus = deferred<ClusterStatus>();
    vi.mocked(api.getClusterStatus)
      .mockImplementationOnce(() => firstStatus.promise)
      .mockImplementationOnce(() => secondStatus.promise);

    const wrapper = makeRouterWrapper('/clusters/cluster-a');
    const { result, rerender } = renderHook(({ id }) => useClusterDetail(id), {
      initialProps: { id: 'cluster-a' },
      wrapper,
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
