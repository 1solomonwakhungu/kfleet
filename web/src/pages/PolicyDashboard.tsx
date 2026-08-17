import { useCallback, useEffect, useMemo, useState } from 'react'
import { Button, Flash, FormControl, Heading, Label, Select, Text } from '@primer/react'
import { Blankslate, SkeletonText } from '@primer/react/experimental'
import {
  AlertIcon,
  CheckCircleIcon,
  ClockIcon,
  QuestionIcon,
  ShieldCheckIcon,
  SyncIcon,
  type Icon,
} from '@primer/octicons-react'

import { api } from '../lib/api'
import type { PolicyResult, PolicyResultsResponse, PolicySeverity, PolicyStatus } from '../types/policy'
import layout from '../styles/layout.module.css'
import styles from './PolicyDashboard.module.css'

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
    <main className={layout.page}>
      <header className={layout.pageHeader}>
        <div className={layout.pageHeaderText}>
          <Heading as="h1" variant="large">
            Policy and drift
          </Heading>
          <Text className={layout.pageDescription}>
            Built-in configuration checks across the latest tenant-scoped cluster snapshots. kfleet reports drift but
            never changes cluster state.
          </Text>
        </div>
        <Button variant="primary" leadingVisual={SyncIcon} disabled={loading} onClick={() => void refresh()}>
          {loading ? 'Evaluating…' : 'Evaluate now'}
        </Button>
      </header>

      {error ? (
        <Flash variant="danger" role="alert">
          <Text weight="semibold">Policy evaluation unavailable</Text>
          <Text className={layout.pageDescription}>{error.message}</Text>
          <div className={styles.flashAction}>
            <Button leadingVisual={SyncIcon} onClick={() => void refresh()}>
              Retry
            </Button>
          </div>
        </Flash>
      ) : (
        <>
          <PolicySummaryStrip data={data} loading={loading} />

          <section className={styles.results} aria-labelledby="policy-findings-heading">
            <div className={styles.resultsHeader}>
              <div>
                <Heading as="h2" variant="small" id="policy-findings-heading">
                  Evaluation results
                </Heading>
                <Text className={layout.pageDescription} role="status">
                  {loading
                    ? 'Reading durable snapshots'
                    : `${visibleResults.length} of ${data?.summary.total ?? 0} results`}
                </Text>
              </div>
              <div className={styles.filters}>
                <FormControl>
                  <FormControl.Label>Status</FormControl.Label>
                  <Select value={statusFilter} onChange={(event) => setStatusFilter(event.target.value as StatusFilter)}>
                    {(['all', 'fail', 'stale', 'unknown', 'pass'] as const).map((option) => (
                      <Select.Option key={option} value={option}>
                        {splitLabel(option)}
                      </Select.Option>
                    ))}
                  </Select>
                </FormControl>
                <FormControl>
                  <FormControl.Label>Severity</FormControl.Label>
                  <Select
                    value={severityFilter}
                    onChange={(event) => setSeverityFilter(event.target.value as SeverityFilter)}
                  >
                    {(['all', 'critical', 'high', 'medium', 'low'] as const).map((option) => (
                      <Select.Option key={option} value={option}>
                        {splitLabel(option)}
                      </Select.Option>
                    ))}
                  </Select>
                </FormControl>
              </div>
            </div>

            {loading && !data ? (
              <div className={styles.resultList}>
                {[0, 1, 2, 3].map((item) => (
                  <div key={item} className={`${layout.box} ${layout.boxBody}`}>
                    <SkeletonText size="titleSmall" maxWidth="16rem" />
                    <SkeletonText size="bodyMedium" maxWidth="80%" />
                  </div>
                ))}
              </div>
            ) : visibleResults.length === 0 ? (
              <div className={`${layout.box} ${styles.emptyState}`}>
                <Blankslate>
                  <Blankslate.Visual>
                    <ShieldCheckIcon size={24} />
                  </Blankslate.Visual>
                  <Blankslate.Heading as="h3">No matching results</Blankslate.Heading>
                  <Blankslate.Description>
                    Broaden the status or severity filters to see more checks.
                  </Blankslate.Description>
                </Blankslate>
              </div>
            ) : (
              <div className={styles.resultList}>
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

const STATUS_PRESENTATION: Record<PolicyStatus, { label: string; icon: Icon; tone: string }> = {
  fail: { label: 'Failing', icon: AlertIcon, tone: styles.toneDanger },
  stale: { label: 'Stale', icon: ClockIcon, tone: styles.toneAttention },
  unknown: { label: 'Unknown', icon: QuestionIcon, tone: styles.toneMuted },
  pass: { label: 'Passing', icon: CheckCircleIcon, tone: styles.toneSuccess },
}

const STATUS_VARIANTS: Record<PolicyStatus, 'danger' | 'attention' | 'secondary' | 'success'> = {
  fail: 'danger',
  stale: 'attention',
  unknown: 'secondary',
  pass: 'success',
}

const SEVERITY_VARIANTS: Record<PolicySeverity, 'danger' | 'attention' | 'secondary'> = {
  critical: 'danger',
  high: 'attention',
  medium: 'secondary',
  low: 'secondary',
}

function PolicySummaryStrip({ data, loading }: { data: PolicyResultsResponse | null; loading: boolean }) {
  return (
    <section className={`${layout.grid} ${layout.grid4}`} aria-label="Policy evaluation summary">
      {(['fail', 'stale', 'unknown', 'pass'] as const).map((status) => {
        const presentation = STATUS_PRESENTATION[status]
        const StatusGlyph = presentation.icon
        const count = data?.summary.byStatus[status] ?? 0

        return (
          <div key={status} className={`${layout.box} ${layout.boxBody}`}>
            <div className={styles.summaryHead}>
              <span className={layout.metricLabel}>{presentation.label}</span>
              <StatusGlyph size={16} className={presentation.tone} />
            </div>
            <p className={layout.metricValue} aria-label={`${presentation.label}: ${count}`}>
              {loading && !data ? '·' : count}
            </p>
          </div>
        )
      })}
    </section>
  )
}

function PolicyResultCard({ result }: { result: PolicyResult }) {
  const evidence = [
    ...Object.entries(result.expected ?? {}).map(([key, value]) => [`Expected ${key}`, value] as const),
    ...Object.entries(result.actual ?? {}).map(([key, value]) => [`Observed ${key}`, value] as const),
  ].filter(([, value]) => value !== '')
  const presentation = STATUS_PRESENTATION[result.status]
  const StatusGlyph = presentation.icon

  return (
    <article className={`${layout.box} ${layout.boxBody}`}>
      <div className={styles.resultHead}>
        <StatusGlyph size={16} className={presentation.tone} />
        <div className={styles.resultBody}>
          <div className={styles.resultTitle}>
            <Text weight="semibold">{result.policyName}</Text>
            <Label variant={SEVERITY_VARIANTS[result.severity]}>{result.severity}</Label>
            <Label variant="secondary">{result.category}</Label>
          </div>
          <Text className={layout.pageDescription}>{result.message}</Text>
          <span className={`${layout.mono} ${layout.muted} ${styles.subject}`}>{subjectLabel(result)}</span>
        </div>
        <Label variant={STATUS_VARIANTS[result.status]}>{result.status.toUpperCase()}</Label>
      </div>

      {evidence.length > 0 && (
        <dl className={styles.evidence}>
          {evidence.map(([key, value]) => (
            <div key={key} className={styles.evidenceItem}>
              <dt>{splitLabel(key)}</dt>
              <dd>{value}</dd>
            </div>
          ))}
        </dl>
      )}
    </article>
  )
}

function splitLabel(value: string) {
  return value
    .replace(/([a-z])([A-Z])/g, '$1 $2')
    .replace(/[-_]/g, ' ')
    .replace(/^./, (character) => character.toUpperCase())
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
