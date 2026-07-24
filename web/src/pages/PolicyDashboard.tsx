import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  AlertCircle,
  AlertTriangle,
  CheckCircle2,
  CircleHelp,
  Clock3,
  RefreshCw,
  ShieldCheck,
} from 'lucide-react'

import { Badge } from '../components/ui/badge'
import { Button } from '../components/ui/button'
import { Card } from '../components/ui/card'
import { api } from '../lib/api'
import { cn } from '../lib/utils'
import type { PolicyResult, PolicyResultsResponse, PolicySeverity, PolicyStatus } from '../types/policy'

type StatusFilter = 'all' | PolicyStatus
type SeverityFilter = 'all' | PolicySeverity

const severityPriority: Record<PolicySeverity, number> = {
  critical: 0,
  high: 1,
  medium: 2,
  low: 3,
}

const statusPriority: Record<PolicyStatus, number> = {
  fail: 0,
  stale: 1,
  unknown: 2,
  pass: 3,
}

export default function PolicyDashboard() {
  const [data, setData] = useState<PolicyResultsResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [severityFilter, setSeverityFilter] = useState<SeverityFilter>('all')

  const refresh = useCallback(async (signal?: AbortSignal) => {
    setLoading(true)
    setError(null)
    try {
      setData(await api.getPolicyResults(signal))
    } catch (reason) {
      if (signal?.aborted) return
      setError(reason instanceof Error ? reason : new Error('Policy results could not be loaded'))
    } finally {
      if (!signal?.aborted) setLoading(false)
    }
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    void refresh(controller.signal)
    return () => controller.abort()
  }, [refresh])

  const visibleResults = useMemo(() => {
    return [...(data?.results ?? [])]
      .filter((result) => statusFilter === 'all' || result.status === statusFilter)
      .filter((result) => severityFilter === 'all' || result.severity === severityFilter)
      .sort((left, right) =>
        statusPriority[left.status] - statusPriority[right.status]
        || severityPriority[left.severity] - severityPriority[right.severity]
        || left.policyName.localeCompare(right.policyName)
        || (left.subject.clusterName ?? '').localeCompare(right.subject.clusterName ?? ''),
      )
  }, [data, severityFilter, statusFilter])

  return (
    <main className="mx-auto min-h-dvh max-w-[100rem] px-4 py-6 sm:px-6 sm:py-8 lg:px-8">
      <header className="flex flex-col gap-5 border-b border-border pb-6 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-sm font-semibold text-blue-400">READ-ONLY ASSURANCE</p>
          <h1 className="mt-2 font-display text-3xl font-bold tracking-tight sm:text-4xl">Policy and drift</h1>
          <p className="mt-2 max-w-3xl text-sm leading-6 text-muted sm:text-base">
            Built-in configuration checks across the latest tenant-scoped cluster snapshots. kFLEET reports drift but never changes cluster state.
          </p>
        </div>
        <Button
          className="shrink-0 bg-blue-600 text-white hover:bg-blue-500 hover:brightness-100"
          onClick={() => void refresh()}
          disabled={loading}
        >
          <RefreshCw className={cn('h-4 w-4', loading && 'animate-spin')} aria-hidden="true" />
          {loading ? 'Evaluating…' : 'Evaluate now'}
        </Button>
      </header>

      {error ? (
        <Card className="mt-6 border border-unreachable/40 bg-danger-soft p-5" role="alert">
          <div className="flex items-start gap-3">
            <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0 text-unreachable" aria-hidden="true" />
            <div>
              <h2 className="font-display text-lg font-bold">Policy evaluation unavailable</h2>
              <p className="mt-1 text-sm text-muted">{error.message}</p>
              <Button variant="outline" size="sm" className="mt-4" onClick={() => void refresh()}>
                <RefreshCw className="h-4 w-4" aria-hidden="true" />
                Retry
              </Button>
            </div>
          </div>
        </Card>
      ) : (
        <>
          <PolicySummaryStrip data={data} loading={loading} />

          <section className="mt-6" aria-labelledby="policy-findings-heading">
            <div className="flex flex-col gap-4 border-b border-border pb-4 md:flex-row md:items-end md:justify-between">
              <div>
                <h2 id="policy-findings-heading" className="font-display text-xl font-bold tracking-tight">Evaluation results</h2>
                <p className="mt-1 text-sm text-muted" role="status">
                  {loading ? 'Reading durable snapshots' : `${visibleResults.length} of ${data?.summary.total ?? 0} results`}
                </p>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <FilterSelect
                  label="Status"
                  value={statusFilter}
                  onChange={(value) => setStatusFilter(value as StatusFilter)}
                  options={['all', 'fail', 'stale', 'unknown', 'pass']}
                />
                <FilterSelect
                  label="Severity"
                  value={severityFilter}
                  onChange={(value) => setSeverityFilter(value as SeverityFilter)}
                  options={['all', 'critical', 'high', 'medium', 'low']}
                />
              </div>
            </div>

            {loading && !data ? (
              <div className="grid gap-3 py-5">
                {[0, 1, 2, 3].map((item) => (
                  <div key={item} className="h-28 animate-pulse rounded-lg border border-border bg-surface" />
                ))}
              </div>
            ) : visibleResults.length === 0 ? (
              <Card className="mt-5 grid min-h-56 place-items-center border-dashed p-8 text-center">
                <div>
                  <ShieldCheck className="mx-auto h-8 w-8 text-blue-400" aria-hidden="true" />
                  <h3 className="mt-4 font-display text-lg font-bold">No matching results</h3>
                  <p className="mt-2 text-sm text-muted">Broaden the status or severity filters to see more checks.</p>
                </div>
              </Card>
            ) : (
              <div className="grid gap-3 py-5">
                {visibleResults.map((result, index) => (
                  <PolicyResultCard key={`${result.policyId}:${subjectKey(result)}:${index}`} result={result} />
                ))}
              </div>
            )}
          </section>
        </>
      )}
    </main>
  )
}

function PolicySummaryStrip({ data, loading }: { data: PolicyResultsResponse | null; loading: boolean }) {
  const values: Array<{ status: PolicyStatus; label: string; icon: typeof CheckCircle2 }> = [
    { status: 'fail', label: 'Failing', icon: AlertCircle },
    { status: 'stale', label: 'Stale', icon: Clock3 },
    { status: 'unknown', label: 'Unknown', icon: CircleHelp },
    { status: 'pass', label: 'Passing', icon: CheckCircle2 },
  ]
  return (
    <section className="mt-6 grid grid-cols-2 gap-3 lg:grid-cols-4" aria-label="Policy evaluation summary">
      {values.map(({ status, label, icon: Icon }) => (
        <Card key={status} className="border border-border p-4">
          <div className="flex items-center justify-between gap-3">
            <p className="text-sm font-semibold text-muted">{label}</p>
            <StatusIcon status={status} icon={Icon} />
          </div>
          <p className="mt-3 font-display text-3xl font-bold" aria-label={`${label}: ${data?.summary.byStatus[status] ?? 0}`}>
            {loading && !data ? '·' : data?.summary.byStatus[status] ?? 0}
          </p>
        </Card>
      ))}
    </section>
  )
}

function PolicyResultCard({ result }: { result: PolicyResult }) {
  const evidence = [
    ...Object.entries(result.expected ?? {}).map(([key, value]) => [`Expected ${key}`, value] as const),
    ...Object.entries(result.actual ?? {}).map(([key, value]) => [`Observed ${key}`, value] as const),
  ].filter(([, value]) => value !== '')
  return (
    <Card className="border border-border p-4 sm:p-5">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div className="flex min-w-0 items-start gap-3">
          <StatusIcon status={result.status} />
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="font-display text-base font-bold">{result.policyName}</h3>
              <SeverityBadge severity={result.severity} />
              <Badge variant="outline">{result.category}</Badge>
            </div>
            <p className="mt-2 text-sm text-muted">{result.message}</p>
            <p className="mt-3 font-mono text-xs text-muted">{subjectLabel(result)}</p>
          </div>
        </div>
        <Badge className={statusClass(result.status)}>{result.status.toUpperCase()}</Badge>
      </div>
      {evidence.length > 0 && (
        <dl className="mt-4 grid gap-2 border-t border-border pt-4 text-xs sm:grid-cols-2 xl:grid-cols-3">
          {evidence.map(([key, value]) => (
            <div key={key} className="min-w-0">
              <dt className="font-mono uppercase tracking-wide text-muted">{splitLabel(key)}</dt>
              <dd className="mt-1 break-all text-foreground">{value}</dd>
            </div>
          ))}
        </dl>
      )}
    </Card>
  )
}

function FilterSelect({
  label,
  value,
  onChange,
  options,
}: {
  label: string
  value: string
  onChange: (value: string) => void
  options: string[]
}) {
  return (
    <label className="text-xs font-semibold text-muted">
      {label}
      <select
        className="mt-1 block min-h-10 w-full rounded-md border border-border bg-surface px-3 text-sm text-foreground"
        value={value}
        onChange={(event) => onChange(event.target.value)}
      >
        {options.map((option) => (
          <option key={option} value={option}>{splitLabel(option)}</option>
        ))}
      </select>
    </label>
  )
}

function StatusIcon({ status, icon: IconOverride }: { status: PolicyStatus; icon?: typeof CheckCircle2 }) {
  const Icon = IconOverride ?? {
    pass: CheckCircle2,
    fail: AlertCircle,
    unknown: CircleHelp,
    stale: Clock3,
  }[status]
  return (
    <span className={cn('grid h-9 w-9 shrink-0 place-items-center rounded-md', statusIconClass(status))}>
      <Icon className="h-5 w-5" aria-hidden="true" />
    </span>
  )
}

function SeverityBadge({ severity }: { severity: PolicySeverity }) {
  return (
    <Badge className={cn(
      severity === 'critical' && 'border-unreachable/50 bg-danger-soft text-unreachable',
      severity === 'high' && 'border-degraded/50 bg-degraded-soft text-degraded',
      (severity === 'medium' || severity === 'low') && 'border-border bg-elevated text-muted',
    )}>
      {severity}
    </Badge>
  )
}

function statusClass(status: PolicyStatus) {
  return cn(
    status === 'pass' && 'border-healthy/40 bg-healthy-soft text-healthy',
    status === 'fail' && 'border-unreachable/40 bg-danger-soft text-unreachable',
    status === 'stale' && 'border-degraded/40 bg-degraded-soft text-degraded',
    status === 'unknown' && 'border-border bg-elevated text-muted',
  )
}

function statusIconClass(status: PolicyStatus) {
  return cn(
    status === 'pass' && 'bg-healthy-soft text-healthy',
    status === 'fail' && 'bg-danger-soft text-unreachable',
    status === 'stale' && 'bg-degraded-soft text-degraded',
    status === 'unknown' && 'bg-elevated text-muted',
  )
}

function subjectKey(result: PolicyResult) {
  const subject = result.subject
  return [subject.clusterId, subject.namespace, subject.kind, subject.name].filter(Boolean).join('/')
}

function subjectLabel(result: PolicyResult) {
  const subject = result.subject
  return [
    subject.clusterName || subject.clusterId || 'fleet',
    subject.namespace,
    subject.kind && subject.name ? `${subject.kind}/${subject.name}` : subject.name,
  ].filter(Boolean).join(' / ')
}

function splitLabel(value: string) {
  return value.replace(/([a-z])([A-Z])/g, '$1 $2').replace(/[-_]/g, ' ').replace(/^./, (character) => character.toUpperCase())
}
