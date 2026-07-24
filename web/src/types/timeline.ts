// Timeline types mirror pkg/types.OperationalEvent and pkg/api.ListTimelineEventsResponse.

export type OperationalEventKind =
  | 'cluster_registered'
  | 'agent_approved'
  | 'heartbeat_state_change'
  | 'version_changed'
  | 'agent_reconnected'
  | 'agent_disconnected'
  | 'policy_finding';

export interface OperationalEvent {
  id: number;
  clusterId: string;
  kind: OperationalEventKind;
  message: string;
  details?: Record<string, string>;
  occurredAt: string;
}

export interface TimelinePage {
  events: OperationalEvent[];
  nextCursor?: number;
}
