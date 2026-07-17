import { useEffect, useState } from 'react';
import { wsManager, type WSStatus } from '@/lib/ws';
import { useClusterStore } from '@/store/clusterStore';

// wsManager is a module-level singleton so the connection survives route
// navigation; this hook only manages listener subscriptions, not the
// connection lifecycle itself.
export function useWebSocket() {
  const [status, setStatus] = useState<WSStatus>(wsManager.getStatus());
  const applyUpdate = useClusterStore((s) => s.applyUpdate);

  useEffect(() => {
    const unsubscribeStatus = wsManager.subscribeStatus(setStatus);
    const unsubscribeMessages = wsManager.subscribe(applyUpdate);
    wsManager.connect();

    return () => {
      unsubscribeStatus();
      unsubscribeMessages();
    };
  }, [applyUpdate]);

  return { status };
}
