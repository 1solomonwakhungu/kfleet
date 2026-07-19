import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { api, ApiError } from './api';

describe('api', () => {
  const fetchMock = vi.fn<typeof fetch>();

  beforeEach(() => {
    vi.stubGlobal('fetch', fetchMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('URL-encodes cluster IDs in every cluster-scoped path', async () => {
    fetchMock.mockImplementation(async () =>
      new Response('[]', { headers: { 'Content-Type': 'application/json' } }),
    );

    const id = 'fleet/us central?#%';
    await api.getCluster(id);
    await api.getClusterStatus(id);
    await api.getEvents(id);
    await api.getPods(id);
    await api.getServices(id);
    await api.getDeployments(id);
    await api.getNamespaces(id);

    expect(fetchMock.mock.calls.map(([url]) => url)).toEqual([
      '/api/v1/clusters/fleet%2Fus%20central%3F%23%25',
      '/api/v1/clusters/fleet%2Fus%20central%3F%23%25/status',
      '/api/v1/clusters/fleet%2Fus%20central%3F%23%25/events',
      '/api/v1/clusters/fleet%2Fus%20central%3F%23%25/pods',
      '/api/v1/clusters/fleet%2Fus%20central%3F%23%25/services',
      '/api/v1/clusters/fleet%2Fus%20central%3F%23%25/deployments',
      '/api/v1/clusters/fleet%2Fus%20central%3F%23%25/namespaces',
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
});
