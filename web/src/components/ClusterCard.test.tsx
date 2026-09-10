import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'

import { ClusterCard } from './ClusterCard'
import type { Cluster } from '../types/cluster'

const cluster: Cluster = {
  id: 'cluster-1',
  name: 'production',
  health: 'healthy',
  nodeCount: 12,
  podCount: 240,
  k8sVersion: '1.31.2',
  agentVersion: '0.4.1',
  lastHeartbeat: new Date().toISOString(),
  registeredAt: '2026-08-01T00:00:00Z',
  labels: { env: 'prod', region: 'eu-west' },
}

function renderCard() {
  return render(
    <MemoryRouter>
      <ClusterCard cluster={cluster} to="/clusters/cluster-1" />
    </MemoryRouter>,
  )
}

describe('ClusterCard', () => {
  it('renders as a link to the cluster detail route', () => {
    renderCard()

    const link = screen.getByRole('link', { name: 'Open production cluster, health healthy' })
    expect(link.tagName).toBe('A')
    expect(link.getAttribute('href')).toBe('/clusters/cluster-1')
  })

  it('keeps the metric list inside the link', () => {
    renderCard()

    const link = screen.getByRole('link', { name: 'Open production cluster, health healthy' })
    const metrics = link.querySelector('dl')
    expect(metrics).not.toBeNull()
    expect(metrics?.querySelectorAll('dt')).toHaveLength(4)
  })
})