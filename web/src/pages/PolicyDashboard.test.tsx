import { fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import PolicyDashboard from './PolicyDashboard'

const response = {
  summary: {
    total: 2,
    byStatus: { pass: 1, fail: 1, unknown: 0, stale: 0 },
    bySeverity: { low: 0, medium: 0, high: 1, critical: 1 },
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
      subject: { clusterId: 'a', clusterName: 'alpha' },
      message: 'Kubernetes version matches the fleet baseline',
      evaluatedAt: '2026-07-23T12:00:00Z',
    },
  ],
}

describe('PolicyDashboard', () => {
  const fetchMock = vi.fn<typeof fetch>()

  beforeEach(() => {
    fetchMock.mockReset()
    fetchMock.mockResolvedValue(new Response(JSON.stringify(response), {
      headers: { 'Content-Type': 'application/json' },
    }))
    vi.stubGlobal('fetch', fetchMock)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('shows policy summary, evidence, and filters results', async () => {
    render(<PolicyDashboard />)

    expect(await screen.findByRole('heading', { name: 'Policy and drift' })).toBeTruthy()
    expect(await screen.findByText('Pod security baseline')).toBeTruthy()
    expect(screen.getByLabelText('Failing: 1')).toBeTruthy()
    expect(screen.getByText('privileged,runAsNonRoot')).toBeTruthy()
    expect(screen.getByText('alpha / apps / Pod/api')).toBeTruthy()

    fireEvent.change(screen.getByLabelText('Status'), { target: { value: 'pass' } })

    expect(screen.queryByText('Pod security baseline')).toBeNull()
    expect(screen.getByText('Kubernetes version consistency')).toBeTruthy()
    expect(screen.getByText('1 of 2 results')).toBeTruthy()
  })

  it('surfaces API failures with a retry action', async () => {
    fetchMock.mockResolvedValueOnce(new Response(JSON.stringify({ error: 'evaluation failed' }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    }))

    render(<PolicyDashboard />)

    expect(await screen.findByRole('alert')).toBeTruthy()
    expect(screen.getByText('evaluation failed')).toBeTruthy()
    expect(screen.getByRole('button', { name: /Retry/ })).toBeTruthy()
  })
})
