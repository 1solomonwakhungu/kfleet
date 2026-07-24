import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { api } from '../lib/api'
import type { Alert } from '../types/alert'
import AlertsPage from './Alerts'

vi.mock('../lib/api', () => ({
  api: {
    listAlerts: vi.fn(),
    acknowledgeAlert: vi.fn(),
  },
}))

const alert: Alert = {
  id: 'alert-1',
  ruleId: 'fleet-health-degraded',
  ruleName: 'Cluster health degraded',
  clusterId: 'cluster-a',
  clusterName: 'production',
  dedupeKey: 'rule:cluster:degraded',
  health: 'degraded',
  severity: 'warning',
  summary: 'production is degraded',
  status: 'firing',
  triggeredAt: '2026-07-23T12:00:00Z',
  updatedAt: '2026-07-23T12:00:00Z',
  deliveryStatus: 'retrying',
  deliveryAttempts: 1,
  lastDeliveryError: 'receiver returned 503',
}

describe('AlertsPage', () => {
  beforeEach(() => {
    vi.mocked(api.listAlerts).mockResolvedValue([alert])
    vi.mocked(api.acknowledgeAlert).mockResolvedValue({
      ...alert,
      status: 'acknowledged',
      acknowledgedBy: 'operator',
      acknowledgedAt: '2026-07-23T12:05:00Z',
    })
  })

  it('renders alert history and acknowledges a firing alert', async () => {
    render(<AlertsPage />)

    expect(await screen.findByText('production is degraded')).toBeTruthy()
    expect(screen.getByText('Retrying')).toBeTruthy()
    expect(screen.getByText('receiver returned 503')).toBeTruthy()
    expect(screen.getByText('1', { selector: 'p' })).toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: 'Acknowledge' }))
    await waitFor(() => expect(api.acknowledgeAlert).toHaveBeenCalledWith('alert-1'))
    expect(await screen.findAllByText('Acknowledged')).toHaveLength(3)
  })
})
