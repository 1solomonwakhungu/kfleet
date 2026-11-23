import { Check } from 'lucide-react'
import { CodeBlock } from '@/components/code-block'
import { Section, SectionHeader } from '@/components/section'
import { TonalChip, type TonalChipTone } from '@/components/tonal/chip'
import { SurfaceTile } from '@/components/tonal/surface-tile'
import {
  siteCopy,
  type PolicySeverity,
  type PolicyStatus,
  type SampleEventStatus,
} from '@/content/siteCopy'
import { cn } from '@/lib/utils'

const EVENT_TONE: Record<SampleEventStatus, { dot: string; chip: TonalChipTone }> = {
  unreachable: { dot: 'bg-[var(--t-chip-danger-fg)]', chip: 'danger' },
  degraded: { dot: 'bg-[var(--t-chip-warning-fg)]', chip: 'warning' },
  healthy: { dot: 'bg-[var(--t-chip-mint-fg)]', chip: 'mint' },
}

const SEVERITY_CHIP: Record<PolicySeverity, TonalChipTone> = {
  critical: 'danger',
  high: 'warning',
  medium: 'blue',
  low: 'neutral',
}

const STATUS_DOT: Record<PolicyStatus, string> = {
  pass: 'bg-[var(--t-chip-mint-fg)]',
  fail: 'bg-[var(--t-chip-danger-fg)]',
  unknown: 'bg-[var(--t-ink-muted)]',
  stale: 'bg-[var(--t-chip-warning-fg)]',
}

const LOG_STREAM_LINES = [
  { label: 'GET', text: '/api/v1/clusters/{id}/pods/{namespace}/{pod}/logs' },
  { label: 'event', text: 'message · one raw log line per frame' },
  { label: 'event', text: 'end · the stream finished normally' },
  { label: 'event', text: 'stream-error · the agent could not stream' },
] as const

const WEBHOOK_LINES = [
  { label: 'X-Kfleet-Event', text: 'fleet.health.alert' },
  { label: 'X-Kfleet-Delivery', text: '<alert UUID>' },
  { label: 'X-Kfleet-Timestamp', text: '<Unix seconds>' },
  { label: 'X-Kfleet-Signature', text: 'v1=<hex HMAC-SHA256>' },
] as const

function SecurityCard() {
  const card = siteCopy.features.securityCard

  return (
    <SurfaceTile
      id={card.id}
      tone="cool"
      padding="lg"
      className="flex flex-col justify-between gap-6 lg:col-span-7"
    >
      <div className="flex flex-col gap-3">
        <h3 className="t-headline text-[24px] sm:text-[26px]">{card.title}</h3>
        <p className="t-body text-[14.5px]">{card.description}</p>
        <ul className="mt-2 flex flex-wrap gap-2">
          {card.guarantees.map((guarantee) => (
            <li key={guarantee}>
              <TonalChip
                tone="neutral"
                className="border border-[var(--t-border-subtle)] bg-[var(--t-elevated)] text-[var(--t-ink)]"
              >
                <Check aria-hidden="true" className="text-[var(--t-chip-mint-fg)]" />
                {guarantee}
              </TonalChip>
            </li>
          ))}
        </ul>
      </div>

      <CodeBlock
        code={card.clusterRoleYaml}
        language="yaml"
        title={
          <span className="block py-0.5 font-mono break-all whitespace-normal">
            {card.manifestPath}
          </span>
        }
        copyLabel="agent ClusterRole rules"
        copyEvent="security_yaml_copied"
      />
    </SurfaceTile>
  )
}

function OfflineHistoryCard() {
  const card = siteCopy.features.offlineHistoryCard

  return (
    <SurfaceTile
      tone="warm"
      padding="lg"
      className="flex flex-col justify-between gap-6 lg:col-span-5"
    >
      <div className="flex flex-col gap-3">
        <h3 className="t-headline text-[24px] sm:text-[26px]">{card.title}</h3>
        <p className="t-body text-[14.5px]">{card.description}</p>
      </div>

      <ol className="flex flex-1 flex-col rounded-[18px] border border-[var(--t-border-subtle)] bg-[var(--t-elevated)] p-5">
        {card.sampleEvents.map((event, index) => {
          const tone = EVENT_TONE[event.status]
          const isLast = index === card.sampleEvents.length - 1
          return (
            <li
              key={event.cluster}
              className={cn(
                'relative pl-7',
                !isLast &&
                  'flex-1 pb-5 after:absolute after:top-5 after:-bottom-1 after:left-[5px] after:w-px after:bg-[var(--t-border-subtle)]',
              )}
            >
              <span
                aria-hidden="true"
                className={cn('absolute top-1.5 left-0 h-[11px] w-[11px] rounded-full', tone.dot)}
              />
              <div className="flex flex-wrap items-center gap-x-2.5 gap-y-1">
                <span className="font-mono text-[13px] font-medium text-[var(--t-ink)]">
                  {event.cluster}
                </span>
                <TonalChip tone={tone.chip} className="px-2.5 py-0.5 text-[11px]">
                  {event.statusLabel}
                </TonalChip>
                <span className="basis-full text-[12px] text-[var(--t-ink-muted)] sm:ml-auto sm:basis-auto">
                  {event.timestampLabel}
                </span>
              </div>
              <p className="mt-1.5 text-[13px] leading-snug text-[var(--t-ink-secondary)]">
                {event.summary}
              </p>
            </li>
          )
        })}
      </ol>
    </SurfaceTile>
  )
}

function PolicyCard() {
  const card = siteCopy.features.policyCard

  return (
    <SurfaceTile
      tone="neutral"
      padding="lg"
      className="flex flex-col justify-between gap-6 lg:col-span-6"
    >
      <div className="flex flex-col gap-3">
        <h3 className="t-headline text-[24px] sm:text-[26px]">{card.title}</h3>
        <p className="t-body text-[14.5px]">{card.description}</p>
      </div>

      <div className="overflow-hidden rounded-[18px] border border-[var(--t-border-subtle)] bg-[var(--t-elevated)]">
        <div className="overflow-x-auto">
          <table className="w-full border-collapse text-left">
            <thead>
              <tr className="border-b border-[var(--t-border-subtle)] bg-[var(--t-surface)] text-[12px] font-medium text-[var(--t-ink-muted)]">
                <th scope="col" className="px-4 py-3 font-medium">
                  Check
                </th>
                <th scope="col" className="px-4 py-3 font-medium">
                  Scope
                </th>
                <th scope="col" className="px-4 py-3 font-medium">
                  Severity
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[var(--t-border-subtle)] text-[13.5px]">
              {card.checks.map((check) => (
                <tr
                  key={check.id}
                  className="transition-colors duration-150 hover:bg-[var(--t-surface)]/70"
                >
                  <td className="px-4 py-2.5 font-medium text-[var(--t-ink)]">{check.label}</td>
                  <td className="px-4 py-2.5 font-mono text-[12px] text-[var(--t-ink-muted)]">
                    {check.scope}
                  </td>
                  <td className="px-4 py-2.5">
                    <TonalChip
                      tone={SEVERITY_CHIP[check.severity]}
                      className="px-2.5 py-0.5 text-[11px]"
                    >
                      {check.severity}
                    </TonalChip>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <ul
          aria-label="Result states"
          className="flex flex-wrap gap-x-5 gap-y-2 border-t border-[var(--t-border-subtle)] bg-[var(--t-surface)]/70 px-4 py-3"
        >
          {card.statuses.map((status) => (
            <li
              key={status}
              className="flex items-center gap-2 font-mono text-[12px] text-[var(--t-ink-muted)]"
            >
              <span aria-hidden="true" className={cn('h-2 w-2 rounded-full', STATUS_DOT[status])} />
              {status}
            </li>
          ))}
        </ul>
      </div>
    </SurfaceTile>
  )
}

function WireLines({
  title,
  lines,
}: {
  readonly title: string
  readonly lines: readonly { readonly label: string; readonly text: string }[]
}) {
  return (
    <div className="min-w-0">
      <p className="border-b border-[var(--t-border-subtle)] bg-[var(--t-surface)]/70 px-4 py-2.5 text-[12px] font-medium text-[var(--t-ink-muted)]">
        {title}
      </p>
      <dl className="scrollbar-subtle flex flex-col gap-1.5 overflow-x-auto p-4 font-mono text-[12px] whitespace-nowrap">
        {lines.map((line) => (
          <div key={`${line.label}-${line.text}`} className="flex gap-3">
            <dt className="font-medium text-[var(--t-blue)]">{line.label}</dt>
            <dd className="text-[var(--t-ink-secondary)]">{line.text}</dd>
          </div>
        ))}
      </dl>
    </div>
  )
}

function AlertsLogsCard() {
  const card = siteCopy.features.alertsLogsCard

  return (
    <SurfaceTile
      tone="mint"
      padding="lg"
      className="flex flex-col justify-between gap-6 lg:col-span-6"
    >
      <div className="flex flex-col gap-3">
        <h3 className="t-headline text-[24px] sm:text-[26px]">{card.title}</h3>
        <p className="t-body text-[14.5px]">{card.description}</p>
        <ul className="mt-2 grid gap-x-4 gap-y-2.5 sm:grid-cols-2">
          {card.highlights.map((highlight) => (
            <li
              key={highlight}
              className="flex items-start gap-2 text-[14px] font-medium text-[var(--t-ink)]"
            >
              <Check
                aria-hidden="true"
                className="mt-0.5 h-4 w-4 shrink-0 text-[var(--t-chip-mint-fg)]"
              />
              <span>{highlight}</span>
            </li>
          ))}
        </ul>
      </div>

      <div className="divide-y divide-[var(--t-border-subtle)] overflow-hidden rounded-[18px] border border-[var(--t-border-subtle)] bg-[var(--t-elevated)]">
        <WireLines title="Pod logs over Server-Sent Events" lines={LOG_STREAM_LINES} />
        <WireLines title="Signed alert webhook headers" lines={WEBHOOK_LINES} />
      </div>
    </SurfaceTile>
  )
}

export function FeaturesBento() {
  const { features } = siteCopy

  return (
    <Section id={features.id} aria-labelledby="features-heading">
      <SectionHeader
        headingId="features-heading"
        eyebrow={features.eyebrow}
        title={features.title}
        subtitle={features.subtitle}
      />
      <div className="mt-10 grid grid-cols-1 gap-5 lg:grid-cols-12">
        <SecurityCard />
        <OfflineHistoryCard />
        <PolicyCard />
        <AlertsLogsCard />
      </div>
    </Section>
  )
}
