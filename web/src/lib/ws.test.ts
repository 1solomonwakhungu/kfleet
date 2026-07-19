import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Cluster, ClusterUpdate } from '@/types/cluster';
import { WSManager } from './ws';

class FakeWebSocket {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;
  static instances: FakeWebSocket[] = [];

  readonly url: string;
  readyState = FakeWebSocket.CONNECTING;
  onopen: (() => void) | null = null;
  onmessage: ((event: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;
  send = vi.fn();

  constructor(url: string) {
    this.url = url;
    FakeWebSocket.instances.push(this);
  }

  open(): void {
    this.readyState = FakeWebSocket.OPEN;
    this.onopen?.();
  }

  receive(data: string): void {
    this.onmessage?.({ data });
  }

  close(): void {
    if (this.readyState === FakeWebSocket.CLOSED) return;
    this.readyState = FakeWebSocket.CLOSED;
    this.onclose?.();
  }
}

const cluster: Cluster = {
  id: 'cluster-a',
  name: 'Cluster A',
  health: 'healthy',
  nodeCount: 2,
  podCount: 10,
  k8sVersion: '1.31',
  agentVersion: '0.1',
  lastHeartbeat: '2026-07-19T12:00:00Z',
  registeredAt: '2026-07-19T11:00:00Z',
  labels: {},
};

describe('WSManager', () => {
  let manager: WSManager;

  beforeEach(() => {
    vi.useFakeTimers();
    vi.stubGlobal('WebSocket', FakeWebSocket);
    FakeWebSocket.instances = [];
    manager = new WSManager();
  });

  afterEach(() => {
    manager.close();
    vi.useRealTimers();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('ignores invalid JSON and structurally malformed frames', () => {
    const listener = vi.fn();
    manager.subscribe(listener);
    manager.connect();
    const socket = FakeWebSocket.instances[0];

    socket.receive('{invalid');
    socket.receive(JSON.stringify({ hello: 'world' }));
    socket.receive(JSON.stringify({ type: 'updated', cluster: { name: 'missing id' } }));

    expect(listener).not.toHaveBeenCalled();

    const update: ClusterUpdate = { type: 'updated', cluster };
    socket.receive(JSON.stringify(update));
    expect(listener).toHaveBeenCalledOnce();
    expect(listener).toHaveBeenCalledWith(update);
  });

  it.each(['registered', 'health_changed', 'snapshot'])('accepts the server %s update type', (type) => {
    const listener = vi.fn();
    manager.subscribe(listener);
    manager.connect();
    const update = { type, cluster };

    FakeWebSocket.instances[0].receive(JSON.stringify(update));

    expect(listener).toHaveBeenCalledWith(update);
  });

  it('normalizes the hub cluster wire shape before notifying listeners', () => {
    const listener = vi.fn();
    manager.subscribe(listener);
    manager.connect();

    FakeWebSocket.instances[0].receive(JSON.stringify({
      type: 'snapshot',
      cluster: {
        id: 'cluster-a',
        name: 'Cluster A',
        health: 'healthy',
        version: 'v1.32.3',
        agentVersion: '0.2.0',
        nodeCount: 5,
        podCount: 42,
        registeredAt: '2026-07-19T11:00:00Z',
        lastHeartbeat: '2026-07-19T12:00:00Z',
        labels: null,
      },
    }));

    expect(listener).toHaveBeenCalledWith(expect.objectContaining({
      type: 'snapshot',
      cluster: expect.objectContaining({
        k8sVersion: 'v1.32.3',
        agentVersion: '0.2.0',
        labels: {},
      }),
    }));
  });

  it('cancels a queued reconnect when connect is requested explicitly', () => {
    vi.spyOn(Math, 'random').mockReturnValue(0);
    manager.connect();
    FakeWebSocket.instances[0].close();

    manager.connect();
    expect(FakeWebSocket.instances).toHaveLength(2);

    vi.advanceTimersByTime(1_000);
    expect(FakeWebSocket.instances).toHaveLength(2);
  });

  it('cancels queued reconnects when closed', () => {
    manager.connect();
    FakeWebSocket.instances[0].close();

    manager.close();
    vi.advanceTimersByTime(30_000);

    expect(FakeWebSocket.instances).toHaveLength(1);
    expect(manager.getStatus()).toBe('closed');
  });

  it('cleanup functions remove message and status subscriptions', () => {
    const messageListener = vi.fn();
    const statusListener = vi.fn();
    const unsubscribeMessage = manager.subscribe(messageListener);
    const unsubscribeStatus = manager.subscribeStatus(statusListener);
    manager.connect();
    const socket = FakeWebSocket.instances[0];

    expect(statusListener).toHaveBeenCalledWith('connecting');
    unsubscribeMessage();
    unsubscribeStatus();
    socket.open();
    socket.receive(JSON.stringify({ type: 'added', cluster }));

    expect(messageListener).not.toHaveBeenCalled();
    expect(statusListener).toHaveBeenCalledTimes(1);
  });
});
