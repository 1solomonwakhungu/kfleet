import { normalizeCluster, type WireCluster } from '@/lib/api';
import type { ClusterUpdate, ClusterUpdateType } from '@/types/cluster';

const MIN_BACKOFF_MS = 1000;
const MAX_BACKOFF_MS = 30000;
const HEARTBEAT_INTERVAL_MS = 30000;

export type WSStatus = 'connecting' | 'open' | 'closed';

type Listener = (msg: ClusterUpdate) => void;
type StatusListener = (status: WSStatus) => void;

interface WireClusterUpdate {
  type: ClusterUpdateType;
  cluster: WireCluster;
}

function isClusterUpdate(value: unknown): value is WireClusterUpdate {
  if (typeof value !== 'object' || value === null) return false;
  const update = value as { type?: unknown; cluster?: unknown };
  if (
    update.type !== 'registered' &&
    update.type !== 'health_changed' &&
    update.type !== 'snapshot' &&
    update.type !== 'deleted' &&
    update.type !== 'added' &&
    update.type !== 'updated'
  ) {
    return false;
  }
  if (typeof update.cluster !== 'object' || update.cluster === null) return false;
  return typeof (update.cluster as { id?: unknown }).id === 'string';
}

// Connects to GET /ws/clusters (internal/server/handlers_ws.go). The hub
// pings at the raw WebSocket protocol level, which browsers answer
// automatically and never surface to JS, so liveness here is tracked with
// an application-level heartbeat send instead of a ping/pong exchange.
export class WSManager {
  private ws: WebSocket | null = null;
  private listeners = new Set<Listener>();
  private statusListeners = new Set<StatusListener>();
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  private manuallyClosed = false;
  private status: WSStatus = 'closed';

  connect(): void {
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return;
    }
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.manuallyClosed = false;
    this.open();
  }

  private open(): void {
    this.setStatus('connecting');
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const url = `${protocol}//${window.location.host}/ws/clusters`;
    const ws = new WebSocket(url);
    this.ws = ws;

    ws.onopen = () => {
      this.reconnectAttempts = 0;
      this.setStatus('open');
      this.startHeartbeat();
    };

    ws.onmessage = (event) => {
      try {
        const msg: unknown = JSON.parse(event.data);
        if (isClusterUpdate(msg)) {
          const update: ClusterUpdate = { ...msg, cluster: normalizeCluster(msg.cluster) };
          this.listeners.forEach((cb) => cb(update));
        }
      } catch {
        // ignore malformed frames
      }
    };

    ws.onclose = () => {
      this.stopHeartbeat();
      this.setStatus('closed');
      if (!this.manuallyClosed) this.scheduleReconnect();
    };

    ws.onerror = () => {
      ws.close();
    };
  }

  private startHeartbeat(): void {
    this.stopHeartbeat();
    this.heartbeatTimer = setInterval(() => {
      try {
        this.ws?.send(JSON.stringify({ type: 'ping' }));
      } catch {
        this.ws?.close();
      }
    }, HEARTBEAT_INTERVAL_MS);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) return;
    const backoff = Math.min(MIN_BACKOFF_MS * 2 ** this.reconnectAttempts, MAX_BACKOFF_MS);
    const jitter = backoff * (0.5 + Math.random() * 0.5);
    this.reconnectAttempts += 1;
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      if (!this.manuallyClosed) this.open();
    }, jitter);
  }

  private setStatus(status: WSStatus): void {
    this.status = status;
    this.statusListeners.forEach((cb) => cb(status));
  }

  getStatus(): WSStatus {
    return this.status;
  }

  subscribe(cb: Listener): () => void {
    this.listeners.add(cb);
    return () => this.listeners.delete(cb);
  }

  subscribeStatus(cb: StatusListener): () => void {
    this.statusListeners.add(cb);
    return () => this.statusListeners.delete(cb);
  }

  close(): void {
    this.manuallyClosed = true;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.stopHeartbeat();
    this.ws?.close();
    this.ws = null;
    this.setStatus('closed');
  }
}

export const wsManager = new WSManager();
