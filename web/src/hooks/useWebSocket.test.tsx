import { renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { wsManager } from '@/lib/ws';
import { useWebSocket } from './useWebSocket';

describe('useWebSocket', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('cleans up both subscriptions when the consumer unmounts', () => {
    const unsubscribeStatus = vi.fn();
    const unsubscribeMessages = vi.fn();
    vi.spyOn(wsManager, 'getStatus').mockReturnValue('closed');
    vi.spyOn(wsManager, 'subscribeStatus').mockReturnValue(unsubscribeStatus);
    vi.spyOn(wsManager, 'subscribe').mockReturnValue(unsubscribeMessages);
    vi.spyOn(wsManager, 'connect').mockImplementation(() => undefined);

    const { unmount } = renderHook(() => useWebSocket());
    expect(wsManager.connect).toHaveBeenCalledOnce();

    unmount();

    expect(unsubscribeStatus).toHaveBeenCalledOnce();
    expect(unsubscribeMessages).toHaveBeenCalledOnce();
  });
});
