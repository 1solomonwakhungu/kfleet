import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import ClusterDetail from './ClusterDetail'
import { AuthProvider } from '../auth/AuthContext'
import { api, ApiError } from '../lib/api'
import type { ClusterStatus } from '../types/cluster'

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

function renderDetail(id = 'cluster-a') {
  return render(
    <MemoryRouter initialEntries={[`/clusters/${id}`]}>
      <AuthProvider>
        <Routes>
          <Route path="/clusters/:id" element={<ClusterDetail />} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
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
