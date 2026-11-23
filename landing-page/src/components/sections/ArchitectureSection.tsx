import { useId } from 'react'
import { Section, SectionHeader } from '@/components/section'
import { Card, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { siteCopy } from '@/content/siteCopy'
import { cn } from '@/lib/utils'

type LineKind = 'kicker' | 'title' | 'hubTitle' | 'copy' | 'meta'

interface NodeLine {
  readonly kind: LineKind
  readonly text: string
}

interface DiagramNode {
  readonly x: number
  readonly y: number
  readonly width: number
  readonly height: number
  readonly tone: 'default' | 'hub' | 'store'
  readonly lines: readonly NodeLine[]
}

interface DiagramEdge {
  /** Drawn from the side that opens the connection towards the hub. */
  readonly d: string
  readonly kind: 'client' | 'agent'
}

interface DiagramLayout {
  readonly width: number
  readonly height: number
  readonly nodes: readonly DiagramNode[]
  readonly edges: readonly DiagramEdge[]
}

// `advance` is the distance from the previous baseline (or the node's top
// padding for the first line) to this line's baseline, in viewBox units.
const LINE_STYLE: Record<LineKind, { size: number; advance: number; className: string }> = {
  kicker: { size: 11, advance: 16, className: 'fill-brand font-mono' },
  title: { size: 14, advance: 19, className: 'fill-foreground font-semibold' },
  hubTitle: { size: 22, advance: 29, className: 'fill-foreground font-semibold' },
  copy: { size: 12, advance: 18, className: 'fill-foreground/80' },
  meta: { size: 11.5, advance: 17, className: 'fill-muted-foreground' },
}

const NODE_TONE: Record<DiagramNode['tone'], string> = {
  default: 'fill-card stroke-border',
  hub: 'fill-card stroke-primary [filter:drop-shadow(0_0_26px_oklch(0.567_0.159_275.2/0.4))]',
  store: 'fill-surface stroke-border',
}

const NODE_PADDING_X = 14
const NODE_PADDING_TOP = 9

const HUB_LINES: readonly NodeLine[] = [
  { kind: 'kicker', text: '14 MB static Go binary' },
  { kind: 'hubTitle', text: 'kfleet hub' },
  { kind: 'copy', text: 'UI · API · Log relay · 7 drift checks' },
  { kind: 'meta', text: 'Holds zero Kubernetes credentials' },
]

const STORE_LINES: readonly NodeLine[] = [
  { kind: 'title', text: 'SQLite (WAL)' },
  { kind: 'meta', text: 'Snapshots · 90d timeline · audit' },
]

const CLUSTER_NAMES = ['Production', 'Staging', 'Edge (NAT)'] as const

// Desktop: clients on the left, hub in the middle, clusters on the right.
const WIDE_LAYOUT: DiagramLayout = {
  width: 960,
  height: 400,
  nodes: [
    {
      x: 1,
      y: 52,
      width: 226,
      height: 92,
      tone: 'default',
      lines: [
        { kind: 'title', text: 'Operators' },
        { kind: 'copy', text: 'Web UI · REST API · SSE logs' },
        { kind: 'meta', text: 'Role-scoped sessions' },
      ],
    },
    {
      x: 1,
      y: 256,
      width: 226,
      height: 92,
      tone: 'default',
      lines: [
        { kind: 'title', text: 'AI clients' },
        { kind: 'copy', text: 'kfleet mcp (stdio)' },
        { kind: 'meta', text: '9 read-only tools · read_only user' },
      ],
    },
    { x: 342, y: 96, width: 276, height: 208, tone: 'hub', lines: HUB_LINES },
    { x: 358, y: 216, width: 244, height: 72, tone: 'store', lines: STORE_LINES },
    ...CLUSTER_NAMES.map(
      (name, index): DiagramNode => ({
        x: 733,
        y: 12 + index * 142,
        width: 226,
        height: 92,
        tone: 'default',
        lines: [
          { kind: 'title', text: name },
          { kind: 'copy', text: 'kfleet agent · outbound only' },
          { kind: 'meta', text: '3 read-only RBAC rules · no Secrets' },
        ],
      }),
    ),
  ],
  edges: [
    { kind: 'client', d: 'M227 98 C 290 98, 280 160, 342 160' },
    { kind: 'client', d: 'M227 302 C 290 302, 280 240, 342 240' },
    { kind: 'agent', d: 'M733 58 C 670 58, 680 150, 618 150' },
    { kind: 'agent', d: 'M733 200 L 618 200' },
    { kind: 'agent', d: 'M733 342 C 670 342, 680 250, 618 250' },
  ],
}

// Narrow screens: clients on top, hub in the middle, clusters underneath, with
// shorter node text so every label stays readable at phone width.
const TALL_LAYOUT: DiagramLayout = {
  width: 344,
  height: 452,
  nodes: [
    {
      x: 1,
      y: 1,
      width: 165,
      height: 82,
      tone: 'default',
      lines: [
        { kind: 'title', text: 'Operators' },
        { kind: 'copy', text: 'Web UI · REST API' },
        { kind: 'meta', text: 'Role-scoped sessions' },
      ],
    },
    {
      x: 178,
      y: 1,
      width: 165,
      height: 82,
      tone: 'default',
      lines: [
        { kind: 'title', text: 'AI clients' },
        { kind: 'copy', text: 'kfleet mcp (stdio)' },
        { kind: 'meta', text: '9 read-only tools' },
      ],
    },
    { x: 32, y: 130, width: 280, height: 190, tone: 'hub', lines: HUB_LINES },
    { x: 46, y: 236, width: 252, height: 70, tone: 'store', lines: STORE_LINES },
    ...CLUSTER_NAMES.map(
      (name, index): DiagramNode => ({
        x: 1 + index * 117,
        y: 367,
        width: 108,
        height: 84,
        tone: 'default',
        lines: [
          { kind: 'title', text: name },
          { kind: 'copy', text: 'Outbound agent' },
          { kind: 'meta', text: '3 read rules' },
        ],
      }),
    ),
  ],
  edges: [
    { kind: 'client', d: 'M83 83 C 83 108, 122 104, 122 130' },
    { kind: 'client', d: 'M261 83 C 261 108, 222 104, 222 130' },
    { kind: 'agent', d: 'M55 367 C 55 340, 104 346, 104 320' },
    { kind: 'agent', d: 'M172 367 L 172 320' },
    { kind: 'agent', d: 'M289 367 C 289 340, 240 346, 240 320' },
  ],
}

const DIAGRAM_TITLE = 'kfleet zero-kubeconfig system topology'
const DIAGRAM_DESCRIPTION =
  'Operators and AI clients connect to one 14 MB kfleet hub over authenticated sessions and read-only MCP tools. Approved agents inside production, staging, and edge clusters open outbound HTTPS and WebSocket connections to the hub using only three read-only RBAC rules. The hub stores snapshots and a 90-day timeline in SQLite and holds zero Kubernetes credentials.'

function TopologyNode({ node }: { readonly node: DiagramNode }) {
  let baseline = node.y + NODE_PADDING_TOP

  return (
    <g>
      <rect
        x={node.x}
        y={node.y}
        width={node.width}
        height={node.height}
        rx={10}
        strokeWidth={node.tone === 'hub' ? 1.5 : 1}
        className={NODE_TONE[node.tone]}
      />
      {node.lines.map((line) => {
        const style = LINE_STYLE[line.kind]
        baseline += style.advance
        return (
          <text
            key={line.text}
            x={node.x + NODE_PADDING_X}
            y={baseline}
            fontSize={style.size}
            className={style.className}
          >
            {line.text}
          </text>
        )
      })}
    </g>
  )
}

function TopologyDiagram({
  layout,
  className,
}: {
  readonly layout: DiagramLayout
  readonly className?: string
}) {
  // Two diagrams share the page (one per breakpoint), so every id is unique.
  const id = useId()
  const titleId = `${id}-title`
  const descriptionId = `${id}-description`
  const clientArrowId = `${id}-client-arrow`
  const agentArrowId = `${id}-agent-arrow`

  return (
    <svg
      viewBox={`0 0 ${layout.width} ${layout.height}`}
      role="img"
      aria-labelledby={`${titleId} ${descriptionId}`}
      className={cn('h-auto w-full font-sans', className)}
    >
      <title id={titleId}>{DIAGRAM_TITLE}</title>
      <desc id={descriptionId}>{DIAGRAM_DESCRIPTION}</desc>
      <defs>
        <marker
          id={clientArrowId}
          viewBox="0 0 8 8"
          refX="7"
          refY="4"
          markerWidth="8"
          markerHeight="8"
          markerUnits="userSpaceOnUse"
          orient="auto"
        >
          <path d="M0 0.5 L8 4 L0 7.5 z" className="fill-muted-foreground" />
        </marker>
        <marker
          id={agentArrowId}
          viewBox="0 0 8 8"
          refX="7"
          refY="4"
          markerWidth="9"
          markerHeight="9"
          markerUnits="userSpaceOnUse"
          orient="auto"
        >
          <path d="M0 0.5 L8 4 L0 7.5 z" className="fill-brand" />
        </marker>
      </defs>

      {layout.edges.map((edge) =>
        edge.kind === 'agent' ? (
          <path
            key={edge.d}
            d={edge.d}
            fill="none"
            strokeWidth={1.5}
            strokeDasharray="6 6"
            markerEnd={`url(#${agentArrowId})`}
            className="animate-dash-flow stroke-brand motion-reduce:animate-none"
          />
        ) : (
          <path
            key={edge.d}
            d={edge.d}
            fill="none"
            strokeWidth={1.25}
            markerEnd={`url(#${clientArrowId})`}
            className="stroke-muted-foreground/70"
          />
        ),
      )}

      {layout.nodes.map((node) => (
        <TopologyNode key={`${node.x}-${node.y}`} node={node} />
      ))}
    </svg>
  )
}

function TopologyLegend() {
  return (
    <ul className="flex flex-col gap-x-8 gap-y-2 text-left text-[13px] text-muted-foreground sm:flex-row sm:flex-wrap sm:justify-center">
      <li className="flex items-center gap-2.5">
        <svg aria-hidden="true" viewBox="0 0 28 8" className="h-2 w-7 shrink-0">
          <path d="M0 4 H28" strokeWidth={1.25} className="stroke-muted-foreground/70" />
        </svg>
        Authenticated sessions and read-only MCP
      </li>
      <li className="flex items-center gap-2.5">
        <svg aria-hidden="true" viewBox="0 0 28 8" className="h-2 w-7 shrink-0">
          <path d="M0 4 H28" strokeWidth={1.5} strokeDasharray="6 5" className="stroke-brand" />
        </svg>
        Agents dial out: HTTPS snapshots (30s) and WebSocket logs
      </li>
    </ul>
  )
}

export function ArchitectureSection() {
  const { architecture } = siteCopy

  return (
    <Section id={architecture.id} aria-labelledby="architecture-heading">
      <SectionHeader
        headingId="architecture-heading"
        eyebrow={architecture.eyebrow}
        title={architecture.title}
        subtitle={architecture.subtitle}
      />

      <figure className="mt-12 rounded-xl border border-border/60 bg-surface px-4 py-6 sm:p-8 lg:p-10">
        <TopologyDiagram layout={WIDE_LAYOUT} className="hidden lg:block" />
        <TopologyDiagram layout={TALL_LAYOUT} className="mx-auto max-w-md lg:hidden" />
        <figcaption className="mt-8 flex flex-col items-center gap-4 border-t border-border/60 pt-6 text-center">
          <TopologyLegend />
          <p className="text-sm font-medium text-foreground">{architecture.topologyCaption}</p>
        </figcaption>
      </figure>

      <ol className="mt-4 grid gap-4 lg:grid-cols-3">
        {architecture.pillars.map((pillar) => (
          <li key={pillar.index} className="flex">
            <Card className="flex-1">
              <CardHeader className="gap-3">
                <span aria-hidden="true" className="font-mono text-sm text-brand">
                  {pillar.index}
                </span>
                <CardTitle>{pillar.title}</CardTitle>
                <CardDescription>{pillar.description}</CardDescription>
              </CardHeader>
            </Card>
          </li>
        ))}
      </ol>
    </Section>
  )
}
