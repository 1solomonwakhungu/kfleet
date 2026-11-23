import { Check } from 'lucide-react'
import { CodeBlock } from '@/components/code-block'
import { Section, SectionHeader } from '@/components/section'
import { Badge, type BadgeProps } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  siteCopy,
  type PolicySeverity,
  type PolicyStatus,
  type SampleEventStatus,
} from '@/content/siteCopy'
import { cn } from '@/lib/utils'

type BadgeVariant = NonNullable<BadgeProps['variant']>

const EVENT_TONE: Record<SampleEventStatus, { dot: string; badge: BadgeVariant }> = {
  unreachable: { dot: 'bg-destructive', badge: 'destructive' },
  degraded: { dot: 'bg-warning', badge: 'warning' },
  healthy: { dot: 'bg-success', badge: 'success' },
}

const SEVERITY_BADGE: Record<PolicySeverity, BadgeVariant> = {
  critical: 'destructive',
  high: 'warning',
  medium: 'default',
  low: 'outline',
}

const STATUS_DOT: Record<PolicyStatus, string> = {
  pass: 'bg-success',
  fail: 'bg-destructive',
  unknown: 'bg-muted-foreground/60',
  stale: 'bg-warning',
}

// Illustrative wire shapes for the logs-and-alerts card. Route, SSE frames,
// and webhook headers are the documented ones (docs/pod-logs.md, docs/alerts.md).
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
    <Card id={card.id} className="lg:col-span-7">
      <CardHeader>
        <CardTitle className="text-xl">{card.title}</CardTitle>
        <CardDescription>{card.description}</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col gap-5">
        <ul className="flex flex-wrap gap-2">
          {card.guarantees.map((guarantee) => (
            <li key={guarantee}>
              <Badge variant="outline" className="gap-1.5 text-foreground">
                <Check aria-hidden="true" className="text-success" />
                {guarantee}
              </Badge>
            </li>
          ))}
        </ul>
        <CodeBlock
          code={card.clusterRoleYaml}
          language="yaml"
          title={<span className="block py-1 font-mono break-all whitespace-normal">{card.manifestPath}</span>}
          copyLabel="agent ClusterRole rules"
          copyEvent="security_yaml_copied"
        />
      </CardContent>
    </Card>
  )
}

function OfflineHistoryCard() {
  const card = siteCopy.features.offlineHistoryCard

  return (
    <Card className="lg:col-span-5">
      <CardHeader>
        <CardTitle className="text-xl">{card.title}</CardTitle>
        <CardDescription>{card.description}</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col">
        <ol className="flex flex-1 flex-col rounded-lg border border-border/60 bg-surface p-4">
          {card.sampleEvents.map((event, index) => {
            const tone = EVENT_TONE[event.status]
            const isLast = index === card.sampleEvents.length - 1
            return (
              <li
                key={event.cluster}
                className={cn(
                  'relative pl-7',
                  // Each entry draws its own connector down to the next dot, so
                  // the rail stays joined when the card stretches the list.
                  !isLast &&
                    'flex-1 pb-5 after:absolute after:top-5 after:-bottom-1 after:left-[5px] after:w-px after:bg-border',
                )}
              >
                <span
                  aria-hidden="true"
                  className={cn(
                    'absolute top-1.5 left-0 size-[11px] rounded-full',
                    tone.dot,
                  )}
                />
                <div className="flex flex-wrap items-center gap-x-2.5 gap-y-1">
                  <span className="font-mono text-[13px] font-medium text-foreground">{event.cluster}</span>
                  <Badge variant={tone.badge} size="sm">
                    {event.statusLabel}
                  </Badge>
                  <span className="basis-full text-xs text-muted-foreground sm:ml-auto sm:basis-auto">
                    {event.timestampLabel}
                  </span>
                </div>
                <p className="mt-1.5 text-[13px] leading-snug text-muted-foreground">{event.summary}</p>
              </li>
            )
          })}
        </ol>
      </CardContent>
    </Card>
  )
}

function PolicyCard() {
  const card = siteCopy.features.policyCard

  return (
    <Card className="lg:col-span-6">
      <CardHeader>
        <CardTitle className="text-xl">{card.title}</CardTitle>
        <CardDescription>{card.description}</CardDescription>
      </CardHeader>
      <CardContent className="flex-1 px-2">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="h-9">Check</TableHead>
              <TableHead className="h-9">Scope</TableHead>
              <TableHead className="h-9">Severity</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {card.checks.map((check) => (
              <TableRow key={check.id}>
                <TableCell className="py-2.5 text-foreground">{check.label}</TableCell>
                <TableCell className="py-2.5 font-mono text-xs text-muted-foreground">
                  {check.scope}
                </TableCell>
                <TableCell className="py-2.5">
                  <Badge variant={SEVERITY_BADGE[check.severity]} size="sm">
                    {check.severity}
                  </Badge>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
      <CardFooter>
        <ul aria-label="Result states" className="flex flex-wrap gap-x-5 gap-y-2">
          {card.statuses.map((status) => (
            <li key={status} className="flex items-center gap-2 font-mono text-xs text-muted-foreground">
              <span aria-hidden="true" className={cn('size-2 rounded-full', STATUS_DOT[status])} />
              {status}
            </li>
          ))}
        </ul>
      </CardFooter>
    </Card>
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
      <p className="border-b border-border/60 px-3.5 py-2 text-xs text-muted-foreground">{title}</p>
      <dl className="scrollbar-subtle flex flex-col gap-1.5 overflow-x-auto p-3.5 font-mono text-xs whitespace-nowrap">
        {lines.map((line) => (
          <div key={`${line.label}-${line.text}`} className="flex gap-3">
            <dt className="text-brand">{line.label}</dt>
            <dd className="text-foreground/80">{line.text}</dd>
          </div>
        ))}
      </dl>
    </div>
  )
}

function AlertsLogsCard() {
  const card = siteCopy.features.alertsLogsCard

  return (
    <Card className="lg:col-span-6">
      <CardHeader>
        <CardTitle className="text-xl">{card.title}</CardTitle>
        <CardDescription>{card.description}</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col gap-5">
        <ul className="grid gap-x-4 gap-y-2.5 sm:grid-cols-2">
          {card.highlights.map((highlight) => (
            <li key={highlight} className="flex items-start gap-2 text-sm text-foreground">
              <Check aria-hidden="true" className="mt-0.5 size-4 shrink-0 text-success" />
              {highlight}
            </li>
          ))}
        </ul>
        <div className="mt-auto divide-y divide-border/60 overflow-hidden rounded-lg border border-border/60 bg-surface">
          <WireLines title="Pod logs over Server-Sent Events" lines={LOG_STREAM_LINES} />
          <WireLines title="Signed alert webhook headers" lines={WEBHOOK_LINES} />
        </div>
      </CardContent>
    </Card>
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
      <div className="mt-12 grid grid-cols-1 gap-4 lg:grid-cols-12">
        <SecurityCard />
        <OfflineHistoryCard />
        <PolicyCard />
        <AlertsLogsCard />
      </div>
    </Section>
  )
}
