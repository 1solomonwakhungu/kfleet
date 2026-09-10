import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import ClusterDetail from './ClusterDetail'
import { AuthProvider } from '../auth/AuthContext'
import { api, ApiError } from '../lib/api'
import type { ClusterStatus } from '../types/cluster'
import type { PodInfo } from '../types/resources'

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
}

const pods: PodInfo[] = [
  {
    name: 'api-7d9f',
    namespace: 'payments',
    phase: 'Running',
    nodeName: 'node-1',
    restartCount: 0,
    ready: true,
    startTime: '2026-07-19T10:00:00Z',
  },
  {
    name: 'worker-1',
    namespace: 'default',
    phase: 'Running',
    nodeName: 'node-2',
    restartCount: 1,
    ready: false,
    startTime: '2026-07-19T09:00:00Z',
  },
]

let latestSearch = ''

function SearchProbe() {
  const location = useLocation()
  latestSearch = location.search
  return null
}

function renderDetail(id = 'cluster-a', search = '') {
  latestSearch = ''
  return render(
    <MemoryRouter initialEntries={[`/clusters/${id}${search}`]}>
      <AuthProvider>
        <Routes>
          <Route path="/clusters/:id" element={<ClusterDetail />} />
        </Routes>
        <SearchProbe />
      </AuthProvider>
    </MemoryRouter>,
  )
}

// The log viewer opens an EventSource as soon as a pod is selected; jsdom has
// no EventSource, so tests stub it with an inert stand-in.
class StubEventSource {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSED = 2
  readyState = StubEventSource.CONNECTING
  onopen: (() => void) | null = null
  onmessage: ((event: MessageEvent) => void) | null = null
  onerror: (() => void) | null = null
  close(): void {}
  addEventListener(): void {}
  removeEventListener(): void {}
}

function searchParam(name: string): string | null {
  return new URLSearchParams(latestSearch).get(name)
}

describe('ClusterDetail error states', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({
      id: 'admin-1',
      username: 'admin',
      email: 'admin@example.com',
      role: 'admin',
      disabled: false,
      createdAt: '2026-07-23T00:00:00Z',
      updatedAt: '2026-07-23T00:00:00Z',
    }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    })))
    vi.spyOn(api, 'getClusterStatus').mockResolvedValue(status)
    vi.spyOn(api, 'getPods').mockResolvedValue([])
    vi.spyOn(api, 'getServices').mockResolvedValue([])
    vi.spyOn(api, 'getDeployments').mockResolvedValue([])
    vi.spyOn(api, 'getEvents').mockResolvedValue([])
    vi.spyOn(api, 'getNamespaces').mockResolvedValue([])
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
    document.title = 'kfleet'
  })

  it('shows "Cluster not found" without the retry banner when the cluster does not exist', async () => {
    vi.mocked(api.getClusterStatus).mockRejectedValue(new ApiError(404, 'cluster not found'))

    renderDetail('gone')

    expect(await screen.findByRole('heading', { name: 'Cluster not found' })).toBeTruthy()
    expect(screen.queryByRole('button', { name: 'Retry' })).toBeNull()
    expect(document.title).toBe('Cluster not found · kfleet')
    expect(screen.getByRole('link', { name: 'Back to clusters' }).getAttribute('href')).toBe('/')
  })

  it('keeps the retry banner and overview-unavailable state for other failures', async () => {
    vi.mocked(api.getClusterStatus).mockRejectedValue(new ApiError(503, 'hub unavailable'))

    renderDetail()

    expect(await screen.findByRole('heading', { name: 'Cluster overview unavailable' })).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Retry' })).toBeTruthy()
    expect(document.title).toBe('Loading · kfleet')
  })
})

describe('ClusterDetail deep links', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({
      id: 'admin-1',
      username: 'admin',
      email: 'admin@example.com',
      role: 'admin',
      disabled: false,
      createdAt: '2026-07-23T00:00:00Z',
      updatedAt: '2026-07-23T00:00:00Z',
    }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    })))
    vi.stubGlobal('EventSource', StubEventSource as unknown as typeof EventSource)
    vi.spyOn(api, 'getClusterStatus').mockResolvedValue(status)
    vi.spyOn(api, 'getPods').mockResolvedValue(pods)
    vi.spyOn(api, 'getServices').mockResolvedValue([])
    vi.spyOn(api, 'getDeployments').mockResolvedValue([])
    vi.spyOn(api, 'getEvents').mockResolvedValue([])
    vi.spyOn(api, 'getNamespaces').mockResolvedValue([])
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
    document.title = 'kfleet'
  })

  it('opens directly on the deep-linked logs tab', async () => {
    renderDetail('cluster-a', '?tab=logs')

    await waitFor(() =>
      expect(screen.getByRole('link', { name: 'Logs' }).getAttribute('aria-current')).toBe('page'),
    )
    expect(screen.getByRole('region', { name: 'Pod log viewer' })).toBeTruthy()
    expect(screen.getByRole('link', { name: /^Pods/ }).getAttribute('aria-current')).toBeNull()
  })

  it('deep links into a namespace filter and fetches its resources', async () => {
    renderDetail('cluster-a', '?namespace=payments')

    await waitFor(() =>
      expect(api.getPods).toHaveBeenCalledWith('cluster-a', 'payments', expect.any(AbortSignal)),
    )
    const namespaceSelect = screen.getByLabelText('Namespace') as HTMLSelectElement
    expect(namespaceSelect.value).toBe('payments')
    expect(screen.getByRole('link', { name: /^Pods/ }).getAttribute('aria-current')).toBe('page')
  })

  it('deep links into logs with a selected pod', async () => {
    renderDetail('cluster-a', '?tab=logs&pod=api-7d9f')

    const podSelect = (await screen.findByLabelText('Pod for log stream')) as HTMLSelectElement
    await waitFor(() => expect(podSelect.value).toBe('payments/api-7d9f'))
  })

  it('carries tab and pod params when opening logs from the pods tab', async () => {
    renderDetail('cluster-a', '?namespace=payments')

    fireEvent.click(
      await screen.findByRole('button', { name: 'View logs for pod api-7d9f in namespace payments' }),
    )

    await waitFor(() => expect(searchParam('tab')).toBe('logs'))
    expect(searchParam('pod')).toBe('api-7d9f')
    expect(searchParam('namespace')).toBe('payments')
  })

  it('updates the URL on tab switches without a full navigation', async () => {
    renderDetail('cluster-a', '?namespace=payments')

    await waitFor(() => expect(api.getClusterStatus).toHaveBeenCalledTimes(1))
    fireEvent.click(screen.getByRole('link', { name: /^Services/ }))

    await waitFor(() => expect(searchParam('tab')).toBe('services'))
    expect(searchParam('namespace')).toBe('payments')
    expect(api.getClusterStatus).toHaveBeenCalledTimes(1)

    fireEvent.click(screen.getByRole('link', { name: /^Pods/ }))
    await waitFor(() => expect(latestSearch).toBe('?namespace=payments'))
    expect(api.getClusterStatus).toHaveBeenCalledTimes(1)
  })

  it('clears the pod param when leaving the logs tab', async () => {
    renderDetail('cluster-a', '?tab=logs&pod=api-7d9f')

    const podSelect = (await screen.findByLabelText('Pod for log stream')) as HTMLSelectElement
    await waitFor(() => expect(podSelect.value).toBe('payments/api-7d9f'))

    fireEvent.click(screen.getByRole('link', { name: /^Pods/ }))

    await waitFor(() => expect(latestSearch).toBe(''))
    expect(searchParam('pod')).toBeNull()
  })
})
