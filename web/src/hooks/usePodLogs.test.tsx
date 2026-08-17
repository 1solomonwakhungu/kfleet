import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { usePodLogs } from './usePodLogs';

type Listener = (event: Event) => void;

// FakeEventSource replaces the jsdom EventSource, which does not exist in the
// test environment, and lets tests drive the SSE lifecycle deterministically.
class FakeEventSource {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSED = 2;
  static instances: FakeEventSource[] = [];

  readyState = FakeEventSource.CONNECTING;
  closed = false;
  onopen: Listener | null = null;
  onmessage: ((event: MessageEvent<string>) => void) | null = null;
  onerror: Listener | null = null;
  private listeners = new Map<string, Listener[]>();

  constructor(readonly url: string) {
    FakeEventSource.instances.push(this);
  }

  addEventListener(name: string, listener: Listener) {
    this.listeners.set(name, [...(this.listeners.get(name) ?? []), listener]);
  }

  removeEventListener() {}

  close() {
    this.closed = true;
    this.readyState = FakeEventSource.CLOSED;
  }

  emitOpen() {
    this.readyState = FakeEventSource.OPEN;
    this.onopen?.(new Event('open'));
  }

  emitMessage(data: string) {
    this.onmessage?.(new MessageEvent('message', { data }));
  }

  emit(name: string, data: string) {
    for (const listener of this.listeners.get(name) ?? []) {
      listener(new MessageEvent(name, { data }));
    }
  }

  emitTerminalError() {
    this.readyState = FakeEventSource.CLOSED;
    this.onerror?.(new Event('error'));
  }

  emitTransientError() {
    this.readyState = FakeEventSource.CONNECTING;
    this.onerror?.(new Event('error'));
  }
}

function installFakeEventSource() {
  FakeEventSource.instances = [];
  vi.stubGlobal('EventSource', FakeEventSource);
  return FakeEventSource;
}

const options = { clusterId: 'cluster-1', namespace: 'apps', pod: 'api' };

describe('usePodLogs', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('requests the hub log route with the container and follow params', () => {
    const eventSource = installFakeEventSource();
    renderHook(() => usePodLogs({ ...options, container: 'sidecar', follow: false }));

    expect(eventSource.instances[0].url).toBe(
      '/api/v1/clusters/cluster-1/pods/apps/api/logs?container=sidecar&follow=false',
    );
  });

  it('collects streamed lines and ends cleanly on the end event', async () => {
    const eventSource = installFakeEventSource();
    const { result } = renderHook(() => usePodLogs(options));
    const source = eventSource.instances[0];

    act(() => source.emitOpen());
    expect(result.current.status).toBe('streaming');

    act(() => {
      source.emitMessage('alpha');
      source.emitMessage('bravo');
    });
    expect(result.current.lines).toEqual(['alpha', 'bravo']);

    act(() => source.emit('end', '{}'));
    await waitFor(() => expect(result.current.status).toBe('ended'));
    expect(source.closed).toBe(true);
    expect(result.current.error).toBeNull();
  });

  it('surfaces a stream-error event without reconnecting', async () => {
    const eventSource = installFakeEventSource();
    const { result } = renderHook(() => usePodLogs(options));
    const source = eventSource.instances[0];

    act(() => source.emitOpen());
    act(() => source.emit('stream-error', JSON.stringify({ message: 'pods "api" is forbidden' })));

    await waitFor(() => expect(result.current.status).toBe('unavailable'));
    expect(result.current.error).toBe('pods "api" is forbidden');
    expect(source.closed).toBe(true);
  });

  it('reports the hub error when the cluster has no connected agent', async () => {
    const eventSource = installFakeEventSource();
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        json: async () => ({ error: 'no agent is connected for this cluster', code: 503 }),
      }),
    );

    const { result } = renderHook(() => usePodLogs(options));
    act(() => eventSource.instances[0].emitTerminalError());

    await waitFor(() => expect(result.current.status).toBe('unavailable'));
    expect(result.current.error).toBe('no agent is connected for this cluster');
    // A permanently failed stream must not be reopened.
    expect(eventSource.instances).toHaveLength(1);
  });

  it('keeps retrying transparently for transient interruptions', async () => {
    const eventSource = installFakeEventSource();
    const { result } = renderHook(() => usePodLogs(options));

    act(() => eventSource.instances[0].emitTransientError());
    await waitFor(() => expect(result.current.status).toBe('reconnecting'));
    expect(result.current.error).toMatch(/reconnecting/i);
  });

  it('opens a new stream when the caller retries', async () => {
    const eventSource = installFakeEventSource();
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({ ok: false, json: async () => ({ error: 'no agent is connected for this cluster' }) }),
    );

    const { result } = renderHook(() => usePodLogs(options));
    act(() => eventSource.instances[0].emitTerminalError());
    await waitFor(() => expect(result.current.status).toBe('unavailable'));

    act(() => result.current.retry());
    await waitFor(() => expect(eventSource.instances).toHaveLength(2));
    expect(result.current.status).toBe('connecting');
  });

  it('closes the stream when the consumer unmounts', () => {
    const eventSource = installFakeEventSource();
    const { unmount } = renderHook(() => usePodLogs(options));

    unmount();

    expect(eventSource.instances[0].closed).toBe(true);
  });

  it('stays idle without a selected pod', () => {
    const eventSource = installFakeEventSource();
    const { result } = renderHook(() => usePodLogs({ ...options, pod: '' }));

    expect(eventSource.instances).toHaveLength(0);
    expect(result.current.status).toBe('idle');
  });
});
