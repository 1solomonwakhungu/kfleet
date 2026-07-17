import { useCallback, useEffect, useRef, useState } from 'react'

import type { Cluster, ClusterHealth } from '../types/cluster'

interface APICluster {
  id: string
  name: string
  health?: string
  nodeCount?: number
  podCount?: number
  k8sVersion?: string
  agentVersion?: string
  version?: string
  lastHeartbeat?: string
  labels?: Record<string, string> | null
}

interface ListClustersResponse {
  clusters: APICluster[]
}

const healthValues = new Set<ClusterHealth>(['healthy', 'degraded', 'unreachable', 'unknown'])

function toCluster(cluster: APICluster): Cluster {
  const health = healthValues.has(cluster.health as ClusterHealth)
    ? (cluster.health as ClusterHealth)
    : 'unknown'

  return {
    id: cluster.id,
    name: cluster.name,
    health,
    nodeCount: cluster.nodeCount ?? 0,
    podCount: cluster.podCount ?? 0,
    k8sVersion: cluster.k8sVersion ?? cluster.version ?? '',
    agentVersion: cluster.agentVersion ?? '',
    lastHeartbeat: cluster.lastHeartbeat ?? '',
    labels: cluster.labels ?? {},
  }
}

export function useClusters() {
  const [clusters, setClusters] = useState<Cluster[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const controllerRef = useRef<AbortController | null>(null)

  const refresh = useCallback(async () => {
    controllerRef.current?.abort()
    const controller = new AbortController()
    controllerRef.current = controller

    try {
      const response = await fetch('/api/v1/clusters', { signal: controller.signal })
      if (!response.ok) throw new Error(`Cluster request failed with status ${response.status}`)
      const body = (await response.json()) as ListClustersResponse | APICluster[]
      const items = Array.isArray(body) ? body : body.clusters
      setClusters(items.map(toCluster))
      setError(null)
    } catch (requestError) {
      if (requestError instanceof DOMException && requestError.name === 'AbortError') return
      setError(requestError instanceof Error ? requestError : new Error('Could not load clusters'))
    } finally {
      if (!controller.signal.aborted) setLoading(false)
    }
  }, [])

  useEffect(() => {
    void refresh()
    const interval = window.setInterval(() => void refresh(), 5_000)
    return () => {
      window.clearInterval(interval)
      controllerRef.current?.abort()
    }
  }, [refresh])

  return { clusters, loading, error, refresh }
}
