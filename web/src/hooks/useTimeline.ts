import { useCallback, useEffect, useRef, useState } from 'react';
import { api, ApiError } from '@/lib/api';
import type { OperationalEvent } from '@/types/timeline';

export type TimelineRange = '24h' | '7d' | '30d' | '90d' | 'all';

const PAGE_SIZE = 40;
const RANGE_MS: Record<Exclude<TimelineRange, 'all'>, number> = {
  '24h': 24 * 60 * 60 * 1000,
  '7d': 7 * 24 * 60 * 60 * 1000,
  '30d': 30 * 24 * 60 * 60 * 1000,
  '90d': 90 * 24 * 60 * 60 * 1000,
};

function errorMessage(error: unknown): string {
  if (error instanceof ApiError || error instanceof Error) return error.message;
  return 'Request failed';
}

function isAbortError(error: unknown): boolean {
  return error instanceof DOMException
    ? error.name === 'AbortError'
    : error instanceof Error && error.name === 'AbortError';
}

function sinceForRange(range: TimelineRange): string | undefined {
  if (range === 'all') return undefined;
  return new Date(Date.now() - RANGE_MS[range]).toISOString();
}

export function useTimeline(clusterId: string) {
  const [events, setEvents] = useState<OperationalEvent[]>([]);
  const [nextCursor, setNextCursor] = useState<number | undefined>();
  const [range, setRange] = useState<TimelineRange>('7d');
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const generation = useRef(0);
  const firstPageController = useRef<AbortController | null>(null);
  const moreController = useRef<AbortController | null>(null);

  const refresh = useCallback(async () => {
    const requestGeneration = ++generation.current;
    firstPageController.current?.abort();
    moreController.current?.abort();
    const controller = new AbortController();
    firstPageController.current = controller;
    setLoading(true);
    setLoadingMore(false);
    setError(null);

    try {
      const page = await api.getTimeline(
        clusterId,
        { since: sinceForRange(range), limit: PAGE_SIZE },
        controller.signal,
      );
      if (controller.signal.aborted || generation.current !== requestGeneration) return;
      setEvents(page.events ?? []);
      setNextCursor(page.nextCursor);
    } catch (requestError) {
      if (controller.signal.aborted || generation.current !== requestGeneration || isAbortError(requestError)) return;
      setEvents([]);
      setNextCursor(undefined);
      setError(errorMessage(requestError));
    } finally {
      if (!controller.signal.aborted && generation.current === requestGeneration) setLoading(false);
    }
  }, [clusterId, range]);

  useEffect(() => {
    void refresh();
    return () => {
      generation.current += 1;
      firstPageController.current?.abort();
      moreController.current?.abort();
    };
  }, [refresh]);

  const loadMore = useCallback(async () => {
    if (!nextCursor || loading || loadingMore) return;
    const requestGeneration = generation.current;
    const controller = new AbortController();
    moreController.current?.abort();
    moreController.current = controller;
    setLoadingMore(true);
    setError(null);

    try {
      const page = await api.getTimeline(
        clusterId,
        { since: sinceForRange(range), before: nextCursor, limit: PAGE_SIZE },
        controller.signal,
      );
      if (controller.signal.aborted || generation.current !== requestGeneration) return;
      setEvents((current) => {
        const existing = new Set(current.map((event) => event.id));
        return [...current, ...(page.events ?? []).filter((event) => !existing.has(event.id))];
      });
      setNextCursor(page.nextCursor);
    } catch (requestError) {
      if (controller.signal.aborted || generation.current !== requestGeneration || isAbortError(requestError)) return;
      setError(errorMessage(requestError));
    } finally {
      if (!controller.signal.aborted && generation.current === requestGeneration) setLoadingMore(false);
    }
  }, [clusterId, loading, loadingMore, nextCursor, range]);

  return {
    events,
    range,
    setRange,
    loading,
    loadingMore,
    error,
    hasMore: Boolean(nextCursor),
    refresh,
    loadMore,
  };
}
