import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { api } from '@/lib/api';
import type { TimelinePage } from '@/types/timeline';
import { useTimeline } from './useTimeline';

const registered = {
  id: 3,
  clusterId: 'cluster-a',
  kind: 'cluster_registered' as const,
  message: 'registered',
  occurredAt: '2026-07-22T00:00:00Z',
};

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((res) => {
    resolve = res;
  });
  return { promise, resolve };
}

describe('useTimeline', () => {
  beforeEach(() => {
    vi.spyOn(api, 'getTimeline').mockResolvedValue({ events: [registered] });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('loads the recent range and appends distinct older pages', async () => {
    vi.mocked(api.getTimeline)
      .mockResolvedValueOnce({ events: [registered], nextCursor: 3 })
      .mockResolvedValueOnce({
        events: [
          registered,
          { ...registered, id: 2, kind: 'agent_approved', message: 'approved' },
        ],
      });

    const { result } = renderHook(() => useTimeline('cluster-a'));
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(api.getTimeline).toHaveBeenNthCalledWith(
      1,
      'cluster-a',
      expect.objectContaining({ limit: 40, since: expect.any(String) }),
      expect.any(AbortSignal),
    );

    await act(async () => {
      await result.current.loadMore();
    });

    expect(api.getTimeline).toHaveBeenNthCalledWith(
      2,
      'cluster-a',
      expect.objectContaining({ before: 3, limit: 40 }),
      expect.any(AbortSignal),
    );
    expect(result.current.events.map((event) => event.id)).toEqual([3, 2]);
    expect(result.current.hasMore).toBe(false);
  });

  it('ignores a stale response after the selected range changes', async () => {
    const first = deferred<TimelinePage>();
    vi.mocked(api.getTimeline)
      .mockImplementationOnce(() => first.promise)
      .mockResolvedValueOnce({ events: [{ ...registered, id: 7, message: 'all retained' }] });

    const { result } = renderHook(() => useTimeline('cluster-a'));
    await waitFor(() => expect(api.getTimeline).toHaveBeenCalledTimes(1));

    act(() => result.current.setRange('all'));
    await waitFor(() => expect(api.getTimeline).toHaveBeenCalledTimes(2));
    await waitFor(() => expect(result.current.events[0]?.id).toBe(7));

    await act(async () => {
      first.resolve({ events: [{ ...registered, id: 1, message: 'stale' }] });
      await Promise.resolve();
    });

    expect(result.current.events[0]?.id).toBe(7);
    expect(vi.mocked(api.getTimeline).mock.calls[1][1]).toEqual({ limit: 40, since: undefined });
  });
});
