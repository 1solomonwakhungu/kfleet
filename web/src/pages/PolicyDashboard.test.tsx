import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { MemoryRouter } from 'react-router-dom'

import PolicyDashboard from './PolicyDashboard'

const response = {
  summary: {
    total: 3,
    byStatus: { pass: 1, fail: 1, unknown: 1, stale: 0 },
    bySeverity: { low: 0, medium: 1, high: 1, critical: 1 },
    clusterCount: 1,
    evaluatedAt: '2026-07-23T12:00:00Z',
  },
  results: [
    {
      policyId: 'pod-security-baseline',
      policyName: 'Pod security baseline',
      category: 'Security',
      severity: 'critical',
      scope: 'workload',
      status: 'fail',
      subject: { clusterId: 'a', clusterName: 'alpha', namespace: 'apps', kind: 'Pod', name: 'api' },
      message: 'Pod violates the built-in restricted security profile',
      actual: { violations: 'privileged,runAsNonRoot' },
      evaluatedAt: '2026-07-23T12:00:00Z',
    },
    {
      policyId: 'kubernetes-version-consistency',
      policyName: 'Kubernetes version consistency',
      category: 'Kubernetes',
      severity: 'high',
      scope: 'fleet',
      status: 'pass',
      subject: { clusterId: 'b', clusterName: 'beta' },
      message: 'Kubernetes version matches the fleet baseline',
      evaluatedAt: '2026-07-23T12:00:00Z',
    },
    {
      policyId: 'agent-heartbeat-freshness',
      policyName: 'Agent heartbeat freshness',
      category: 'Agents',
      severity: 'medium',
      scope: 'fleet',
      status: 'unknown',
      subject: {},
      message: 'No snapshot available to evaluate this check',
      evaluatedAt: '2026-07-23T12:00:00Z',
    },
  ],
}

describe('PolicyDashboard', () => {
  const fetchMock = vi.fn<typeof fetch>()

  beforeEach(() => {
    fetchMock.mockReset()
    // A Response body can only be consumed once, so each call gets a fresh instance.
    fetchMock.mockImplementation(() => Promise.resolve(new Response(JSON.stringify(response), {
      headers: { 'Content-Type': 'application/json' },
    })))
    vi.stubGlobal('fetch', fetchMock)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('shows policy summary, evidence, filters results, and links cluster subjects', async () => {
    render(<PolicyDashboard />, { wrapper: MemoryRouter })

    expect(await screen.findByRole('heading', { name: 'Policy and drift' })).toBeTruthy()
    expect(await screen.findByText('Pod security baseline')).toBeTruthy()
    expect(screen.getByLabelText('Failing: 1')).toBeTruthy()
    expect(screen.getByText('privileged,runAsNonRoot')).toBeTruthy()

    const clusterLinks = screen.getAllByRole('link', { name: 'alpha' })
    expect(clusterLinks).toHaveLength(1)
    expect(clusterLinks[0].getAttribute('href')).toBe('/clusters/a')
    expect(screen.getByText((_, element) => element?.textContent === 'alpha / apps / Pod/api')).toBeTruthy()
    expect(screen.getByRole('link', { name: 'beta' }).getAttribute('href')).toBe('/clusters/b')

    const fleetSubject = screen.getByText('fleet')
    expect(fleetSubject.closest('a')).toBeNull()

    fireEvent.change(screen.getByLabelText('Status'), { target: { value: 'pass' } })

    expect(screen.queryByText('Pod security baseline')).toBeNull()
    expect(screen.getByText('Kubernetes version consistency')).toBeTruthy()
    expect(screen.getByText('1 of 3 results')).toBeTruthy()
  })

  it('polls for policy results every 15 seconds', async () => {
    vi.useFakeTimers({ toFake: ['setInterval', 'clearInterval'] })
    try {
      render(<PolicyDashboard />, { wrapper: MemoryRouter })

      expect(await screen.findByText('Pod security baseline')).toBeTruthy()
      expect(fetchMock).toHaveBeenCalledTimes(1)

      await act(async () => {
        vi.advanceTimersByTime(15_000)
      })

      expect(fetchMock).toHaveBeenCalledTimes(2)
      expect(await screen.findByText('Pod security baseline')).toBeTruthy()
    } finally {
      vi.useRealTimers()
    }
  })

  it('keeps previously loaded results when a poll fails and recovers on the next poll', async () => {
    vi.useFakeTimers({ toFake: ['setInterval', 'clearInterval'] })
    try {
      render(<PolicyDashboard />, { wrapper: MemoryRouter })

      expect(await screen.findByText('Pod security baseline')).toBeTruthy()
      expect(fetchMock).toHaveBeenCalledTimes(1)

      fetchMock.mockResolvedValueOnce(new Response(JSON.stringify({ error: 'evaluation failed' }), {
        status: 500,
        headers: { 'Content-Type': 'application/json' },
      }))
      await act(async () => {
        vi.advanceTimersByTime(15_000)
      })

      expect(fetchMock).toHaveBeenCalledTimes(2)
      expect(await screen.findByText('evaluation failed')).toBeTruthy()
      expect(screen.getByText('Pod security baseline')).toBeTruthy()

      await act(async () => {
        vi.advanceTimersByTime(15_000)
      })

      expect(fetchMock).toHaveBeenCalledTimes(3)
      await waitFor(() => expect(screen.queryByText('evaluation failed')).toBeNull())
      expect(screen.getByText('Pod security baseline')).toBeTruthy()
    } finally {
      vi.useRealTimers()
    }
  })

  it('surfaces API failures with a retry action', async () => {
    fetchMock.mockResolvedValueOnce(new Response(JSON.stringify({ error: 'evaluation failed' }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    }))

    render(<PolicyDashboard />, { wrapper: MemoryRouter })

    expect(await screen.findByRole('alert')).toBeTruthy()
    expect(screen.getByText('evaluation failed')).toBeTruthy()
    expect(screen.getByRole('button', { name: /Retry/ })).toBeTruthy()
  })
})
