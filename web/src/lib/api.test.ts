import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { api, ApiError } from './api';

describe('api', () => {
  const fetchMock = vi.fn<typeof fetch>();

  beforeEach(() => {
    fetchMock.mockReset();
    vi.stubGlobal('fetch', fetchMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('URL-encodes cluster IDs in every cluster-scoped path', async () => {
    fetchMock.mockImplementation(async (input) => {
      const url = String(input);
      const cluster = {
        id: 'cluster', name: 'Cluster', health: 'unknown', version: '',
        nodeCount: 0, podCount: 0, lastHeartbeat: '', registeredAt: '', labels: {},
      };
      const body = url.endsWith('/status')
        ? { cluster, nodes: [] }
        : url.includes('/timeline')
          ? { events: [] }
        : /\/clusters\/[^/]+$/.test(url)
          ? cluster
          : [];
      return new Response(JSON.stringify(body), { headers: { 'Content-Type': 'application/json' } });
    });

    const id = 'fleet/us central?#%';
    await api.getCluster(id);
    await api.getClusterStatus(id);
    await api.getEvents(id);
    await api.getPods(id);
    await api.getServices(id);
    await api.getDeployments(id);
    await api.getNamespaces(id);
    await api.getTimeline(id);
    await api.getClusterPolicyResults(id);

    expect(fetchMock.mock.calls.map(([url]) => url)).toEqual([
      '/api/v1/clusters/fleet%2Fus%20central%3F%23%25',
      '/api/v1/clusters/fleet%2Fus%20central%3F%23%25/status',
      '/api/v1/clusters/fleet%2Fus%20central%3F%23%25/events',
      '/api/v1/clusters/fleet%2Fus%20central%3F%23%25/pods',
      '/api/v1/clusters/fleet%2Fus%20central%3F%23%25/services',
      '/api/v1/clusters/fleet%2Fus%20central%3F%23%25/deployments',
      '/api/v1/clusters/fleet%2Fus%20central%3F%23%25/namespaces',
      '/api/v1/clusters/fleet%2Fus%20central%3F%23%25/timeline',
      '/api/v1/clusters/fleet%2Fus%20central%3F%23%25/policy-results',
    ]);
  });

  it('constructs encoded namespace queries and omits empty values', async () => {
    fetchMock.mockImplementation(async () =>
      new Response('[]', { headers: { 'Content-Type': 'application/json' } }),
    );

    await api.getPods('cluster', 'team a&b/?');
    await api.getEvents('cluster', '');
    await api.getServices('cluster', undefined);

    expect(fetchMock.mock.calls.map(([url]) => url)).toEqual([
      '/api/v1/clusters/cluster/pods?namespace=team+a%26b%2F%3F',
      '/api/v1/clusters/cluster/events',
      '/api/v1/clusters/cluster/services',
    ]);
  });

  it('constructs paginated and time-filtered timeline queries', async () => {
    fetchMock.mockResolvedValue(
      new Response(JSON.stringify({ events: [], nextCursor: 40 }), {
        headers: { 'Content-Type': 'application/json' },
      }),
    );

    await expect(api.getTimeline('cluster/a', {
      since: '2026-07-20T00:00:00Z',
      until: '2026-07-21T00:00:00Z',
      before: 81,
      limit: 40,
    })).resolves.toEqual({ events: [], nextCursor: 40 });

    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/clusters/cluster%2Fa/timeline?since=2026-07-20T00%3A00%3A00Z&until=2026-07-21T00%3A00%3A00Z&before=81&limit=40',
      expect.objectContaining({ method: 'GET' }),
    );
  });

  it('returns undefined for an empty successful response', async () => {
    fetchMock.mockResolvedValue(new Response(null, { status: 200 }));

    await expect(api.getNamespaces('cluster')).resolves.toBeUndefined();
  });

  it('reports a non-JSON successful response as an API error', async () => {
    fetchMock.mockResolvedValue(
      new Response('not json', { status: 200, headers: { 'Content-Type': 'text/plain' } }),
    );

    await expect(api.getNamespaces('cluster')).rejects.toMatchObject({
      name: 'ApiError',
      status: 200,
      message: 'response was not valid JSON',
      body: 'not json',
    });
  });

  it('preserves JSON API errors', async () => {
    fetchMock.mockResolvedValue(
      new Response(JSON.stringify({ error: 'cluster unavailable', detail: 'timed out' }), {
        status: 503,
        headers: { 'Content-Type': 'application/json' },
      }),
    );

    const request = api.getCluster('cluster');
    await expect(request).rejects.toBeInstanceOf(ApiError);
    await expect(request).rejects.toMatchObject({
      status: 503,
      message: 'cluster unavailable',
      body: { error: 'cluster unavailable', detail: 'timed out' },
    });
  });

  it('preserves text error bodies instead of losing them after JSON parsing', async () => {
    fetchMock.mockResolvedValue(new Response('upstream unavailable', { status: 502 }));

    await expect(api.getCluster('cluster')).rejects.toMatchObject({
      name: 'ApiError',
      status: 502,
      message: 'upstream unavailable',
      body: 'upstream unavailable',
    });
  });

  it('passes abort signals through and preserves abort failures', async () => {
    fetchMock.mockImplementation((_url, init) =>
      new Promise((_resolve, reject) => {
        init?.signal?.addEventListener('abort', () => reject(new DOMException('Aborted', 'AbortError')));
      }),
    );
    const controller = new AbortController();

    const request = api.getCluster('cluster', controller.signal);
    controller.abort();

    await expect(request).rejects.toMatchObject({ name: 'AbortError' });
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/clusters/cluster',
      expect.objectContaining({ signal: controller.signal }),
    );
  });

  it('normalizes hub cluster and node wire shapes for the UI', async () => {
    fetchMock.mockResolvedValueOnce(new Response(JSON.stringify({
      cluster: {
        id: 'cluster-a',
        name: 'Cluster A',
        health: 'healthy',
        version: 'v1.32.3',
        agentVersion: '0.1.0',
        nodeCount: 1,
        podCount: 12,
        registeredAt: '2026-07-19T11:00:00Z',
        lastHeartbeat: '2026-07-19T12:00:00Z',
        labels: null,
      },
      nodes: [{
        name: 'node-a',
        status: 'Ready',
        roles: ['control-plane'],
        version: 'v1.32.3',
        cpuCapacity: '8',
        memoryCapacity: '16Gi',
        ready: true,
      }],
    })));

    await expect(api.getClusterStatus('cluster-a')).resolves.toEqual({
      cluster: expect.objectContaining({
        k8sVersion: 'v1.32.3',
        agentVersion: '0.1.0',
        labels: {},
      }),
      nodes: [{
        name: 'node-a',
        status: 'Ready',
        roles: ['control-plane'],
        version: 'v1.32.3',
        cpuCapacity: '8',
        memoryCapacity: '16Gi',
        ready: true,
      }],
    });
  });
});
