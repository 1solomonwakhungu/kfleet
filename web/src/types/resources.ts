// PodInfo and EventInfo mirror pkg/types.Pod and pkg/types.Event exactly.
// ServiceInfo and DeploymentInfo have no backing Go type or handler yet;
// they follow the same REST convention as the existing cluster routes so
// the UI can slot in once the hub grows those endpoints.

export interface PodInfo {
  name: string;
  namespace: string;
  phase: string;
  nodeName: string;
  restartCount: number;
  ready: boolean;
  startTime: string;
}

export interface EventInfo {
  clusterId: string;
  namespace: string;
  reason: string;
  message: string;
  type: string;
  count: number;
  lastTimestamp: string;
}

export interface ServicePort {
  name: string;
  port: number;
  targetPort: number;
  protocol: string;
}

export interface ServiceInfo {
  name: string;
  namespace: string;
  type: string;
  clusterIP: string;
  ports: ServicePort[];
  age: string;
}

export interface DeploymentInfo {
  name: string;
  namespace: string;
  readyReplicas: number;
  desiredReplicas: number;
  updatedReplicas: number;
  availableReplicas: number;
  age: string;
}
