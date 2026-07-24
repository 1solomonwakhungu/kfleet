import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useTimeline } from '@/hooks/useTimeline';
import { OperationalTimeline } from './OperationalTimeline';

vi.mock('@/hooks/useTimeline', () => ({ useTimeline: vi.fn() }));

describe('OperationalTimeline', () => {
  const loadMore = vi.fn();
  const refresh = vi.fn();

  beforeEach(() => {
    loadMore.mockReset();
    refresh.mockReset();
    vi.mocked(useTimeline).mockReturnValue({
      events: [{
        id: 9,
        clusterId: 'cluster-a',
        kind: 'policy_finding',
        message: 'Privileged container detected',
        details: {
          severity: 'high',
          ruleId: 'no-privileged',
          resource: 'pod/default/api',
        },
        occurredAt: new Date().toISOString(),
      }],
      range: '7d',
      setRange: vi.fn(),
      loading: false,
      loadingMore: false,
      error: null,
      hasMore: true,
      refresh,
      loadMore,
    });
  });

  it('renders event metadata and paginates older history', () => {
    render(<OperationalTimeline clusterId="cluster-a" />);

    expect(screen.getByRole('heading', { name: 'Operational timeline' })).toBeTruthy();
    expect(screen.getByText('Policy finding')).toBeTruthy();
    expect(screen.getByText('Privileged container detected')).toBeTruthy();
    expect(screen.getByText('no-privileged')).toBeTruthy();
    expect(screen.getByText('pod/default/api')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Load older events' }));
    expect(loadMore).toHaveBeenCalledOnce();
  });
});
