export type PolicyStatus = 'pass' | 'fail' | 'unknown' | 'stale'
export type PolicySeverity = 'low' | 'medium' | 'high' | 'critical'
export type PolicyScope = 'fleet' | 'cluster' | 'namespace' | 'workload'

export interface PolicySubject {
  clusterId?: string
  clusterName?: string
  namespace?: string
  kind?: string
  name?: string
}

export interface PolicyResult {
  policyId: string
  policyName: string
  category: string
  severity: PolicySeverity
  scope: PolicyScope
  status: PolicyStatus
  subject: PolicySubject
  message: string
  expected?: Record<string, string>
  actual?: Record<string, string>
  observedAt?: string
  evaluatedAt: string
}

export interface PolicySummary {
  total: number
  byStatus: Record<PolicyStatus, number>
  bySeverity: Record<PolicySeverity, number>
  clusterCount: number
  evaluatedAt: string
}

export interface PolicyResultsResponse {
  results: PolicyResult[]
  summary: PolicySummary
}
