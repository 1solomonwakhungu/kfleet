export type ClusterHealth = 'healthy' | 'degraded' | 'unreachable' | 'unknown'

export interface Cluster {
  id: string
  name: string
  health: ClusterHealth
  nodeCount: number
  podCount: number
  k8sVersion: string
  agentVersion: string
  lastHeartbeat: string
  registeredAt: string
  labels: Record<string, string>
}

export interface ClusterNode {
  name: string
  status: string
  roles: string[]
  version: string
  cpuCapacity: string
  memoryCapacity: string
  ready: boolean
}

export interface ClusterStatus {
  cluster: Cluster
  nodes: ClusterNode[]
}

export type ClusterUpdateType =
  | 'registered'
  | 'health_changed'
  | 'snapshot'
  | 'deleted'
  | 'added'
  | 'updated'

export interface ClusterUpdate {
  type: ClusterUpdateType
  cluster: Cluster
}
