import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { MemoryRouter } from 'react-router-dom'

import { api } from '../lib/api'
import { useAuth } from '../auth/AuthContext'
import type { Alert } from '../types/alert'
import AlertsPage from './Alerts'

vi.mock('../lib/api', () => ({
  api: {
    listAlerts: vi.fn(),
    acknowledgeAlert: vi.fn(),
  },
}))
vi.mock('../auth/AuthContext', () => ({
  useAuth: vi.fn(),
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
    vi.clearAllMocks()
    vi.mocked(useAuth).mockReturnValue({
      user: {
        id: 'operator-1',
        username: 'operator',
        email: 'operator@example.com',
        role: 'operator',
        disabled: false,
        createdAt: '2026-07-23T00:00:00Z',
        updatedAt: '2026-07-23T00:00:00Z',
      },
      loading: false,
      login: vi.fn(),
      logout: vi.fn(),
    })
    vi.mocked(api.listAlerts).mockResolvedValue([alert])
    vi.mocked(api.acknowledgeAlert).mockResolvedValue({
      ...alert,
      status: 'acknowledged',
      acknowledgedBy: 'operator',
      acknowledgedAt: '2026-07-23T12:05:00Z',
    })
  })

  it('renders alert history and acknowledges a firing alert', async () => {
    render(<AlertsPage />, { wrapper: MemoryRouter })

    expect(await screen.findByText('production is degraded')).toBeTruthy()
    expect(screen.getByText('Retrying')).toBeTruthy()
    expect(screen.getByText('receiver returned 503')).toBeTruthy()
    expect(screen.getByText('1', { selector: 'p' })).toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: 'Acknowledge' }))
    await waitFor(() => expect(api.acknowledgeAlert).toHaveBeenCalledWith('alert-1'))
    expect(await screen.findAllByText('Acknowledged')).toHaveLength(3)
  })

  it('renders the affected cluster as a link to the cluster detail page', async () => {
    render(<AlertsPage />, { wrapper: MemoryRouter })

    const link = await screen.findByRole('link', { name: 'production' })
    expect(link.getAttribute('href')).toBe('/clusters/cluster-a')
  })

  it('polls for alert updates every 15 seconds and stops after unmount', async () => {
    vi.useFakeTimers({ toFake: ['setInterval', 'clearInterval'] })
    try {
      const teammateAcknowledged: Alert = {
        ...alert,
        status: 'acknowledged',
        acknowledgedBy: 'teammate',
      }
      vi.mocked(api.listAlerts)
        .mockClear()
        .mockResolvedValueOnce([alert])
        .mockResolvedValueOnce([teammateAcknowledged])

      const { unmount } = render(<AlertsPage />, { wrapper: MemoryRouter })

      expect(await screen.findByText('production is degraded')).toBeTruthy()
      expect(screen.getByRole('link', { name: 'production' })).toBeTruthy()

      await act(async () => {
        vi.advanceTimersByTime(15_000)
      })

      expect(api.listAlerts).toHaveBeenCalledTimes(2)
      expect(await screen.findByText('by teammate')).toBeTruthy()

      unmount()
      await act(async () => {
        vi.advanceTimersByTime(30_000)
      })

      expect(api.listAlerts).toHaveBeenCalledTimes(2)
    } finally {
      vi.useRealTimers()
    }
  })

  it('keeps alert acknowledgements read-only for viewers', async () => {
    vi.mocked(useAuth).mockReturnValue({
      user: {
        id: 'demo-read-only',
        username: 'public-demo',
        email: 'public-demo@invalid',
        role: 'read_only',
        disabled: false,
        createdAt: '2026-07-23T00:00:00Z',
        updatedAt: '2026-07-23T00:00:00Z',
      },
      loading: false,
      login: vi.fn(),
      logout: vi.fn(),
    })

    render(<AlertsPage />, { wrapper: MemoryRouter })

    const button = await screen.findByRole('button', { name: 'View only' })
    expect((button as HTMLButtonElement).disabled).toBe(true)
    fireEvent.click(button)
    expect(api.acknowledgeAlert).not.toHaveBeenCalled()
  })
})
