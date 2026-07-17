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
  labels: Record<string, string>
}
