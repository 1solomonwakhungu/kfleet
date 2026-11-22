import { useEffect, useRef, useState, type KeyboardEvent } from 'react'
import { Button } from '@/components/ui/button'
import {
  ActivityHeartbeatIcon,
  ArrowRightIcon,
  BellIcon,
  BrandGithubIcon,
  CheckIcon,
  ExternalLinkIcon,
  SearchIcon,
  ServerIcon,
  ShieldCheckIcon,
  Terminal2Icon,
} from '@/components/icons'
import { capture } from '@/lib/posthog'
import './App.css'

const repositoryUrl = 'https://github.com/1solomonwakhungu/kfleet'
const liveDemoUrl = 'https://kfleet.sullylab.de'

const comparisonColumns = [
  'Capability',
  'kfleet',
  'Rancher',
  'Lens',
  'Headlamp',
  'OCM',
  'Kubeconfig MCP servers',
] as const

const comparisonRows = [
  [
    'Central footprint',
    '1 Go binary + SQLite',
    'Dedicated K8s cluster',
    'None',
    'None, or in-cluster app',
    'Operator + controllers',
    'Local/in-cluster process',
  ],
  [
    'In-cluster footprint',
    '1 agent (3 read rules)',
    'cluster-admin agent',
    'None',
    'None',
    'Klusterlet agents',
    'None',
  ],
  [
    'Central/laptop kubeconfigs',
    'None',
    'Central proxy',
    'Per-user kubeconfig',
    'Per-user kubeconfig',
    'Agents pull work',
    'Required on client/host',
  ],
  [
    'Can mutate clusters',
    'No (read-only)',
    'Yes',
    'Yes',
    'Yes',
    'Yes',
    'Depends on RBAC',
  ],
  [
    'Offline cluster history',
    'Last snapshot + 90d timeline',
    'Last status',
    'None',
    'None',
    'Last status',
    'None',
  ],
  [
    'AI / MCP interface',
    'Read-only MCP built in (9 tools)',
    'Liz + MCP (Prime)',
    'Prism + MCP (proprietary)',
    'AI Assistant plugin',
    'Community MCP servers',
    'MCP only (no UI/alerts/drift)',
  ],
  [
    'License',
    'Apache 2.0',
    'Apache 2.0',
    'Proprietary / free tier',
    'Apache 2.0',
    'Apache 2.0',
    'Varies',
  ],
] as const

const agentClusterRoleYaml = `rules:
  - apiGroups: [""]
    resources: ["pods", "services", "nodes", "events", "namespaces"]
    verbs: ["list"]
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["list"]
  - apiGroups: [""]
    resources: ["pods/log"]
    verbs: ["get"]`

const mcpPrompts = [
  'List my clusters and summarize anything degraded, unreachable, or off our Kubernetes version.',
  'Find crashlooping pods across the fleet and show the previous container logs from the worst one.',
  'Diagnose demo-eu-west and explain its recent Warning events, reconnects, and policy findings.',
]

const mcpTools = [
  'list_clusters',
  'get_cluster_status',
  'list_namespaces',
  'get_pods',
  'get_events',
  'get_recent_events',
  'find_crashloop_pods',
  'get_pod_logs',
  'diagnose_cluster',
] as const

type QuickstartTabId = 'live' | 'docker' | 'brew' | 'helm'

interface QuickstartStep {
  step: string
  label?: string
  code: string
  copyValue?: string
}

interface QuickstartTab {
  id: QuickstartTabId
  label: string
  badge: string
  summary: string
  steps: QuickstartStep[]
  resultText: string
  resultLinkLabel: string
  resultLinkHref: string
  resultExternal: boolean
  resultEvent: string
}

const quickstartTabs: QuickstartTab[] = [
  {
    id: 'live',
    label: 'Live demo',
    badge: '0s',
    summary:
      'Explore three synthetic clusters in your browser right now—no signup, no login, every write rejected with HTTP 405. You can also connect an MCP client anonymously.',
    steps: [
      {
        step: 'UI',
        label: 'Open the hosted read-only fleet dashboard',
        code: 'https://kfleet.sullylab.de',
        copyValue: liveDemoUrl,
      },
      {
        step: 'MCP',
        label: 'Point any stdio MCP client at the public demo (no credentials required)',
        code: 'KFLEET_HUB_URL=https://kfleet.sullylab.de kfleet mcp',
        copyValue: 'KFLEET_HUB_URL=https://kfleet.sullylab.de kfleet mcp',
      },
    ],
    resultText:
      'Synthetic data only (demo-us-central, demo-eu-west, demo-edge-lab). Try find_crashloop_pods on demo-eu-west.',
    resultLinkLabel: 'Open live demo',
    resultLinkHref: liveDemoUrl,
    resultExternal: true,
    resultEvent: 'quickstart_live_demo_opened',
  },
  {
    id: 'docker',
    label: 'Docker',
    badge: '10s',
    summary:
      'Run the embedded web UI, REST API, and demo MCP target locally against an in-memory SQLite database. Needs only Docker—no Kubernetes required.',
    steps: [
      {
        step: '01',
        label: 'Start the 14 MB Go hub in synthetic demo mode',
        code: 'docker run --rm -p 8080:8080 -e KFLEET_DEMO_MODE=true ghcr.io/1solomonwakhungu/kfleet/hub:v0.3.0',
        copyValue:
          'docker run --rm -p 8080:8080 -e KFLEET_DEMO_MODE=true ghcr.io/1solomonwakhungu/kfleet/hub:v0.3.0',
      },
    ],
    resultText:
      'Serves the live UI, policy drift checks, operational timeline, and synthetic pod logs on port 8080.',
    resultLinkLabel: 'Open localhost:8080',
    resultLinkHref: 'http://localhost:8080',
    resultExternal: true,
    resultEvent: 'local_dashboard_opened',
  },
  {
    id: 'brew',
    label: 'Homebrew + kind',
    badge: '60s',
    summary:
      'Spin up three real local kind clusters with the released OCI Helm charts, port-forwarded hub UI, and the kfleet mcp CLI bundled. Requires only Docker running.',
    steps: [
      {
        step: '01',
        label: 'Install kfleet along with kind, kubectl, and Helm',
        code: 'brew install --cask 1solomonwakhungu/tap/kfleet',
        copyValue: 'brew install --cask 1solomonwakhungu/tap/kfleet',
      },
      {
        step: '02',
        label: 'Create 3 kind clusters, deploy hub + agents, and port-forward :8080',
        code: 'kfleet quickstart',
        copyValue: 'kfleet quickstart',
      },
      {
        step: '03',
        label: 'Remove all local clusters and port-forwards cleanly when finished',
        code: 'kfleet cleanup',
        copyValue: 'kfleet cleanup',
      },
    ],
    resultText:
      'Creates kfleet-1..3 clusters, prints generated admin credentials, and ships kfleet mcp in the same CLI.',
    resultLinkLabel: 'Open localhost:8080',
    resultLinkHref: 'http://localhost:8080',
    resultExternal: true,
    resultEvent: 'local_dashboard_opened',
  },
  {
    id: 'helm',
    label: 'Helm OCI (prod)',
    badge: '2m',
    summary:
      'Deploy one single-replica hub (100m CPU · 128Mi RAM · 1Gi disk) and one read-only agent per cluster (50m CPU · 64Mi RAM). No CRDs, operators, or cert-manager needed.',
    steps: [
      {
        step: '01',
        label: 'Install the hub once (persistence 1Gi enabled by default)',
        code: 'helm upgrade --install kfleet-hub oci://ghcr.io/1solomonwakhungu/charts/kfleet-hub -n kfleet-system --create-namespace --set registration.token="$KFLEET_TOKEN" --set-string auth.bootstrapAdmin.username=admin --set-string auth.bootstrapAdmin.email=admin@example.com --set-string auth.bootstrapAdmin.password="$KFLEET_ADMIN_PASSWORD"',
        copyValue:
          'helm upgrade --install kfleet-hub oci://ghcr.io/1solomonwakhungu/charts/kfleet-hub -n kfleet-system --create-namespace --set registration.token="$KFLEET_TOKEN" --set-string auth.bootstrapAdmin.username=admin --set-string auth.bootstrapAdmin.email=admin@example.com --set-string auth.bootstrapAdmin.password="$KFLEET_ADMIN_PASSWORD"',
      },
      {
        step: '02',
        label: 'Install the 3-rule read-only agent in each cluster you want to see',
        code: 'helm upgrade --install kfleet-agent oci://ghcr.io/1solomonwakhungu/charts/kfleet-agent --kube-context my-cluster -n kfleet-system --create-namespace --set hub.url=https://kfleet.example.com --set hub.token="$KFLEET_TOKEN" --set cluster.name=my-cluster',
        copyValue:
          'helm upgrade --install kfleet-agent oci://ghcr.io/1solomonwakhungu/charts/kfleet-agent --kube-context my-cluster -n kfleet-system --create-namespace --set hub.url=https://kfleet.example.com --set hub.token="$KFLEET_TOKEN" --set cluster.name=my-cluster',
      },
    ],
    resultText:
      'Sign in as admin and approve each agent on the Agents page. Supports existingSecret for GitOps and hub.caBundle for private CAs.',
    resultLinkLabel: 'Full Helm guide',
    resultLinkHref: `${repositoryUrl}#install-with-helm`,
    resultExternal: true,
    resultEvent: 'helm_guide_opened',
  },
]

function FleetMap() {
  return (
    <figure className="fleet-map">
      <svg viewBox="0 0 1180 660" role="img" aria-labelledby="map-title map-description">
        <title id="map-title">kfleet zero-kubeconfig system topology</title>
        <desc id="map-description">
          Operators and AI clients connect to one 14 MB kfleet hub over authenticated sessions and
          read-only MCP tools. Approved agents inside production, staging, and edge clusters open
          outbound HTTPS and WebSocket connections up to the hub using only three read-only RBAC
          rules. The hub stores snapshots and 90-day history in SQLite and holds zero Kubernetes
          credentials.
        </desc>
        <defs>
          <pattern id="map-grid" width="36" height="36" patternUnits="userSpaceOnUse">
            <path d="M 36 0 L 0 0 0 36" className="map-grid-line" />
          </pattern>
          <marker
            id="map-arrow"
            viewBox="0 0 10 10"
            refX="8"
            refY="5"
            markerWidth="5"
            markerHeight="5"
            orient="auto-start-reverse"
          >
            <path d="M 0 0 L 10 5 L 0 10 z" className="map-arrow" />
          </marker>
          <marker
            id="map-arrow-outbound"
            viewBox="0 0 10 10"
            refX="8"
            refY="5"
            markerWidth="5"
            markerHeight="5"
            orient="auto-start-reverse"
          >
            <path d="M 0 0 L 10 5 L 0 10 z" className="map-arrow-outbound" />
          </marker>
        </defs>
        <rect width="1180" height="660" className="map-field" />
        <rect width="1180" height="660" fill="url(#map-grid)" />

        <g className="map-route">
          <path d="M390 124 H502 Q522 124 522 144 V238" markerEnd="url(#map-arrow)" />
          <path d="M790 124 H678 Q658 124 658 144 V238" markerEnd="url(#map-arrow)" />
          <path d="M765 316 H863" markerEnd="url(#map-arrow)" />
        </g>

        <g className="map-route-outbound">
          <path d="M220 508 V456 H540 V392" markerEnd="url(#map-arrow-outbound)" />
          <path d="M590 508 V392" markerEnd="url(#map-arrow-outbound)" />
          <path d="M960 508 V456 H640 V392" markerEnd="url(#map-arrow-outbound)" />
        </g>

        <g className="map-node client-node" transform="translate(80 68)">
          <rect width="310" height="112" rx="10" />
          <circle cx="28" cy="30" r="6" className="node-signal" />
          <text x="48" y="35" className="node-title">
            OPERATORS
          </text>
          <text x="28" y="68" className="node-copy">
            Web UI · REST API · SSE logs
          </text>
          <text x="28" y="92" className="node-meta">
            zero kubeconfigs · role-scoped sessions
          </text>
        </g>

        <g className="map-node client-node" transform="translate(790 68)">
          <rect width="310" height="112" rx="10" />
          <circle cx="28" cy="30" r="6" className="node-signal" />
          <text x="48" y="35" className="node-title">
            AI CLIENTS
          </text>
          <text x="28" y="68" className="node-copy">
            kfleet mcp (stdio)
          </text>
          <text x="28" y="92" className="node-meta">
            9 read-only tools · read_only user
          </text>
        </g>

        <g className="map-node hub-node" transform="translate(415 240)">
          <rect width="350" height="150" rx="10" />
          <rect x="1" y="1" width="348" height="7" rx="4" className="hub-accent" />
          <text x="26" y="42" className="node-kicker">
            14 MB STATIC GO BINARY
          </text>
          <text x="26" y="76" className="hub-title">
            kfleet hub
          </text>
          <text x="26" y="104" className="node-copy">
            UI · API · Log relay · 7 drift checks
          </text>
          <text x="26" y="128" className="node-meta">
            holds zero Kubernetes credentials
          </text>
          <circle cx="322" cy="64" r="7" className="node-healthy" />
        </g>

        <g className="map-node store-node" transform="translate(865 264)">
          <rect width="235" height="104" rx="10" />
          <text x="22" y="36" className="node-title">
            SQLITE
          </text>
          <text x="22" y="64" className="node-copy">
            Durable local store
          </text>
          <text x="22" y="86" className="node-meta">
            snapshots · 90d timeline · audit
          </text>
        </g>

        {[
          { x: 70, name: 'PRODUCTION' },
          { x: 440, name: 'STAGING' },
          { x: 810, name: 'EDGE / NAT' },
        ].map(({ x, name }) => (
          <g className="map-node cluster-node" transform={`translate(${x} 508)`} key={name}>
            <rect width="300" height="106" rx="10" />
            <circle cx="26" cy="28" r="6" className="node-healthy" />
            <text x="44" y="33" className="node-title">
              {name}
            </text>
            <text x="26" y="66" className="node-copy">
              kfleet agent · outbound only
            </text>
            <text x="26" y="88" className="node-meta">
              3 read-only RBAC rules · no Secrets
            </text>
          </g>
        ))}

        <text x="590" y="106" textAnchor="middle" className="route-label">
          AUTHENTICATED SESSIONS · READ-ONLY MCP
        </text>

        <g transform="translate(362 442)">
          <rect width="456" height="28" rx="6" className="route-pill" />
          <text x="228" y="18" textAnchor="middle" className="route-label-outbound">
            AGENTS DIAL OUT · HTTPS SNAPSHOTS (30s) · WEBSOCKET LOGS
          </text>
        </g>
      </svg>
      <figcaption>
        <span>
          <i className="legend-dot healthy" /> approved agent (outbound only)
        </span>
        <span>
          <i className="legend-dot route" /> authenticated hub connection
        </span>
        <span>Nothing dials into your clusters. Zero kubeconfigs stored centrally.</span>
      </figcaption>
    </figure>
  )
}

function App() {
  const navRef = useRef<HTMLDetailsElement>(null)
  const tabRefs = useRef<Record<QuickstartTabId, HTMLButtonElement | null>>({
    live: null,
    docker: null,
    brew: null,
    helm: null,
  })
  const [activeTab, setActiveTab] = useState<QuickstartTabId>('live')
  const [copiedCode, setCopiedCode] = useState<string | null>(null)

  const currentTab = quickstartTabs.find((tab) => tab.id === activeTab) ?? quickstartTabs[0]

  useEffect(() => {
    const onKeyDown = (event: globalThis.KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault()
        if (navRef.current) {
          navRef.current.open = !navRef.current.open
        }
      } else if (event.key === 'Escape' && navRef.current?.open) {
        navRef.current.open = false
      }
    }
    const onMouseDown = (event: MouseEvent) => {
      if (navRef.current?.open && !navRef.current.contains(event.target as Node)) {
        navRef.current.open = false
      }
    }
    window.addEventListener('keydown', onKeyDown)
    window.addEventListener('mousedown', onMouseDown)
    return () => {
      window.removeEventListener('keydown', onKeyDown)
      window.removeEventListener('mousedown', onMouseDown)
    }
  }, [])

  const closeNav = () => {
    if (navRef.current) {
      navRef.current.open = false
    }
  }

  const selectTab = (tabId: QuickstartTabId) => {
    setActiveTab(tabId)
    capture(`quickstart_tab_${tabId}`)
  }

  const handleTabKeyDown = (event: KeyboardEvent<HTMLButtonElement>, index: number) => {
    let nextIndex: number | null = null
    if (event.key === 'ArrowRight') {
      nextIndex = (index + 1) % quickstartTabs.length
    } else if (event.key === 'ArrowLeft') {
      nextIndex = (index - 1 + quickstartTabs.length) % quickstartTabs.length
    } else if (event.key === 'Home') {
      nextIndex = 0
    } else if (event.key === 'End') {
      nextIndex = quickstartTabs.length - 1
    }

    if (nextIndex !== null) {
      event.preventDefault()
      const nextTab = quickstartTabs[nextIndex]
      selectTab(nextTab.id)
      tabRefs.current[nextTab.id]?.focus()
    }
  }

  const handleCopy = async (value: string) => {
    try {
      await navigator.clipboard.writeText(value)
      setCopiedCode(value)
      capture('quickstart_command_copied')
      window.setTimeout(() => {
        setCopiedCode((prev) => (prev === value ? null : prev))
      }, 1800)
    } catch {
      // Clipboard API may be unavailable in restricted contexts.
    }
  }

  return (
    <div className="site-shell dark">
      <a className="skip-link" href="#main">
        Skip to content
      </a>
      <header className="site-header">
        <a className="brand" href="#top" aria-label="kfleet home">
          <img
            className="brand-mark"
            src="/brand/kfleet-mark.svg"
            width="28"
            height="28"
            alt=""
            aria-hidden="true"
            draggable={false}
          />
          <span>kfleet</span>
        </a>
        <details className="command-nav" ref={navRef}>
          <summary>
            <SearchIcon aria-hidden="true" />
            <span>Navigate kfleet</span>
            <kbd>⌘ K</kbd>
          </summary>
          <nav aria-label="Page navigation">
            <a href="#demo" onClick={closeNav}>
              Live UI <span>01</span>
            </a>
            <a href="#system" onClick={closeNav}>
              Architecture <span>02</span>
            </a>
            <a href="#why" onClick={closeNav}>
              Why kfleet <span>03</span>
            </a>
            <a href="#capabilities" onClick={closeNav}>
              Capabilities <span>04</span>
            </a>
            <a href="#proof" onClick={closeNav}>
              RBAC &amp; AI tools <span>05</span>
            </a>
            <a href="#compare" onClick={closeNav}>
              Compare <span>06</span>
            </a>
            <a href="#quickstart" onClick={closeNav}>
              Quickstart <span>07</span>
            </a>
          </nav>
        </details>
        <a
          className="repo-link"
          href={repositoryUrl}
          target="_blank"
          rel="noopener noreferrer"
          onClick={() => capture('header_github_clicked')}
        >
          <BrandGithubIcon aria-hidden="true" />
          <span>GitHub</span>
        </a>
      </header>

      <main id="main">
        <section className="orientation" id="top">
          <p className="orientation-line">
            Read-only Kubernetes fleet control plane · 14 MB Go + SQLite · Zero kubeconfigs
          </p>
          <h1>
            See every cluster.
            <br />
            <span>Touch none of them.</span>
          </h1>
          <div className="orientation-meta">
            <p>
              kfleet is an open-source, read-only control plane for a fleet of Kubernetes clusters.
              One 14 MB Go binary with SQLite gives engineers a live web UI, streaming pod logs,
              and 9 read-only MCP diagnostic tools for AI clients—without storing a single
              kubeconfig or granting write access.
            </p>
            <div className="orientation-actions">
              <div className="orientation-cta-row">
                <Button asChild size="lg">
                  <a
                    href={liveDemoUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    onClick={() => capture('live_demo_cta_clicked')}
                  >
                    Try live demo <ExternalLinkIcon aria-hidden="true" />
                  </a>
                </Button>
                <Button asChild variant="secondary" size="lg">
                  <a
                    href={repositoryUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    onClick={() => capture('hero_github_cta_clicked')}
                  >
                    <BrandGithubIcon aria-hidden="true" /> GitHub
                  </a>
                </Button>
              </div>
              <a
                className="quickstart-jump"
                href="#quickstart"
                onClick={() => capture('quickstart_cta_clicked')}
              >
                See what you need &amp; run in 10s <ArrowRightIcon aria-hidden="true" />
              </a>
            </div>
          </div>
        </section>

        <section className="demo-section" id="demo" aria-labelledby="demo-heading">
          <div className="section-eyebrow">
            <h2 id="demo-heading">Live web UI</h2>
            <a
              href={liveDemoUrl}
              target="_blank"
              rel="noopener noreferrer"
              onClick={() => capture('demo_section_link_clicked')}
            >
              Open interactive demo <ExternalLinkIcon aria-hidden="true" />
            </a>
          </div>
          <figure className="demo-frame">
            <div className="demo-chrome" aria-hidden="true">
              <span className="demo-chrome-url">
                <i className="legend-dot healthy" /> kfleet.sullylab.de · read-only demo mode
              </span>
              <span className="demo-chrome-meta">HTTP 405 on writes · no login · anonymous MCP</span>
            </div>
            <a
              className="demo-media-link"
              href={liveDemoUrl}
              target="_blank"
              rel="noopener noreferrer"
              onClick={() => capture('demo_gif_clicked')}
              aria-label="Open the interactive kfleet live demo at kfleet.sullylab.de"
            >
              <img
                src="/demo.gif"
                alt="Tour of the kfleet web UI showing the fleet dashboard with three synthetic clusters (demo-us-central, demo-eu-west, demo-edge-lab), cluster detail view, live pod logs, policy and drift results, and alert history"
              />
            </a>
            <figcaption>
              <span>
                Synthetic demo clusters: <code>demo-us-central</code>, <code>demo-eu-west</code>,{' '}
                <code>demo-edge-lab</code>. Nothing shown comes from a real environment.
              </span>
              <a
                href={liveDemoUrl}
                target="_blank"
                rel="noopener noreferrer"
                onClick={() => capture('demo_caption_cta_clicked')}
              >
                Launch kfleet.sullylab.de <ExternalLinkIcon aria-hidden="true" />
              </a>
            </figcaption>
          </figure>
        </section>

        <section className="system-section" id="system" aria-labelledby="system-heading">
          <div className="section-eyebrow">
            <h2 id="system-heading">Architecture topology</h2>
            <span>Outbound-only agents · No exposed cluster APIs</span>
          </div>
          <FleetMap />
        </section>

        <section className="architecture-notes" id="why">
          <div className="note-statement">
            <span>01 · ZERO BLAST RADIUS</span>
            <h2>The hub cannot touch your clusters.</h2>
            <p>
              Most multi-cluster tools force a trade-off: hand out kubeconfigs to every laptop and
              AI client, or run a platform whose agent holds <code>cluster-admin</code>. kfleet
              stores zero cluster credentials and every agent dials out with just three read-only
              RBAC rules—never Secrets, never ConfigMaps, never writes.
            </p>
          </div>
          <div className="note-statement">
            <span>02 · ZERO PLATFORM TAX</span>
            <h2>One binary. One SQLite file.</h2>
            <p>
              Skip dedicated management clusters, operators, CRDs, and cert-manager. The hub
              compiles the React UI, REST API, log relay, alert engine, and MCP server into a single
              14 MB static Go binary that idles around 20 MB of RAM—deployable via Helm, Docker, or
              a plain VM.
            </p>
          </div>
          <div className="note-statement">
            <span>03 · BUILT TO REMEMBER</span>
            <h2>History outlives outages.</h2>
            <p>
              Kubeconfig tools go blank the second a private or edge cluster drops off the network.
              kfleet keeps each cluster’s latest snapshot and a 90-day append-only operational
              timeline in SQLite, so you still see last known state, disconnects, and what changed
              right before it fell over.
            </p>
          </div>
        </section>

        <section className="capability-ledger" id="capabilities">
          <div className="ledger-intro">
            <h2>Everything you need to see and triage a fleet</h2>
            <p>
              Built for platform teams and SREs who want immediate multi-cluster clarity—and safe AI
              diagnostics—without widening who can change production.
            </p>
          </div>
          <dl>
            <div>
              <dt>
                <ActivityHeartbeatIcon aria-hidden="true" /> Live fleet inventory
              </dt>
              <dd>
                Track cluster health, Kubernetes versions, nodes, pods, deployments, services, and
                events across every cluster with deep-linkable URLs and live WebSocket updates.
              </dd>
              <small>30s snapshots · marked unreachable after 3 missed intervals</small>
            </div>
            <div>
              <dt>
                <Terminal2Icon aria-hidden="true" /> Live pod log streaming
              </dt>
              <dd>
                Tail, follow, and search container logs in the browser without exposing any cluster
                API server. Approved agents relay lines over outbound WebSockets to SSE streams.
              </dd>
              <small>No kubeconfig · no inbound firewall rules</small>
            </div>
            <div>
              <dt>
                <SearchIcon aria-hidden="true" /> Built-in read-only AI (MCP)
              </dt>
              <dd>
                Connect Claude Desktop, Cursor, or any stdio MCP client with{' '}
                <code>kfleet mcp</code>. Nine read-only tools find crashlooping pods, pull previous
                container logs, inspect timeline history, and diagnose unhealthy clusters.
              </dd>
              <small>Authenticates as a read_only hub user · works with any model</small>
            </div>
            <div>
              <dt>
                <ShieldCheckIcon aria-hidden="true" /> Policy &amp; drift checks
              </dt>
              <dd>
                Seven built-in checks flag Kubernetes version skew, missing required labels,
                namespace and deployment spec drift, unavailable workloads, and pod security
                baseline gaps.
              </dd>
              <small>pass · fail · unknown · stale — reports only, never auto-remediates</small>
            </div>
            <div>
              <dt>
                <BellIcon aria-hidden="true" /> Health alerts &amp; webhooks
              </dt>
              <dd>
                Alert when clusters turn degraded or unreachable with cooldown deduplication,
                operator acknowledgement, and HMAC-SHA256-signed webhooks with durable SQLite
                retries.
              </dd>
              <small>Dead-letter tracking · Prometheus /metrics included</small>
            </div>
            <div>
              <dt>
                <ServerIcon aria-hidden="true" /> RBAC, audit &amp; tenants
              </dt>
              <dd>
                Enforce admin, operator, and read-only roles with explicit agent approval queues,
                token rotation, tenant-scoped isolation, and a SQLite-enforced append-only audit
                log.
              </dd>
              <small>Fail-closed agent registration · 90-day default timeline retention</small>
            </div>
          </dl>
        </section>

        <section className="proof-section" id="proof" aria-label="Security boundary and AI tools">
          <div className="proof-card">
            <div className="proof-header">
              <span>ENTIRE AGENT CLUSTERROLE · 3 RULES</span>
              <h3>Audit the entire blast radius in ten seconds.</h3>
              <p>
                This is the complete ClusterRole installed in each cluster. No Secrets, no
                ConfigMaps, no watch permissions, and zero write verbs.
              </p>
            </div>
            <pre className="proof-code">
              <code>{agentClusterRoleYaml}</code>
            </pre>
            <p className="proof-footer">
              Distroless non-root image · read-only root filesystem · 50m CPU / 64Mi RAM default
            </p>
          </div>

          <div className="proof-card">
            <div className="proof-header">
              <span>MODEL CONTEXT PROTOCOL · 9 READ-ONLY TOOLS</span>
              <h3>Ask your fleet in plain English.</h3>
              <p>
                Give an AI assistant real-time cluster visibility and crash logs without risking a
                hallucinated write—or test it anonymously against the live demo right now.
              </p>
            </div>
            <ul className="prompt-list">
              {mcpPrompts.map((prompt) => (
                <li key={prompt}>“{prompt}”</li>
              ))}
            </ul>
            <div className="tool-chips" aria-label="Included read-only MCP tools">
              {mcpTools.map((tool) => (
                <code key={tool}>{tool}</code>
              ))}
            </div>
            <p className="proof-footer">
              Ships in kfleet mcp &amp; kfleet-hub mcp · previous crash logs up to 256 KiB
            </p>
          </div>
        </section>

        <section className="comparison" id="compare">
          <header>
            <h2>A deliberately smaller surface</h2>
            <p>
              Rancher, OCM, Lens, and Headlamp do more than kfleet—they also mutate clusters. kfleet
              is built for read-only fleet visibility and safe MCP diagnostics without expanding
              cluster write access.
            </p>
          </header>
          <div className="table-wrap" tabIndex={0} aria-label="Scrollable product comparison">
            <table>
              <thead>
                <tr>
                  {comparisonColumns.map((column, index) => (
                    <th key={column} scope="col" className={index === 1 ? 'kfleet-col' : ''}>
                      {column}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {comparisonRows.map((row) => (
                  <tr key={row[0]}>
                    {row.map((cell, index) => (
                      <td
                        key={`${row[0]}-${comparisonColumns[index]}`}
                        className={index === 1 ? 'kfleet-cell' : ''}
                      >
                        {cell}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <p className="comparison-honesty">
            <strong>When to pick something else:</strong> need workload deployment, cluster
            provisioning, <code>kubectl exec</code>, or a multi-replica HA control plane? Pick
            Rancher, OCM, Lens, or Headlamp. kfleet stays intentionally read-only and single-replica
            on SQLite.
          </p>
          <p className="comparison-note">Competitor details checked November 2026.</p>
        </section>

        <section className="quickstart" id="quickstart">
          <header>
            <h2>
              What you need.
              <br />
              How to start.
            </h2>
            <p>
              Explore the hosted demo in one click, boot demo mode in Docker with no Kubernetes,
              launch three real kind clusters via Homebrew, or deploy directly to your clusters with
              Helm.
            </p>
            <ul className="requirements-list" aria-label="Requirements by setup path">
              <li>
                <strong>Live demo (0s)</strong>
                <span>Just a browser — or kfleet mcp with no credentials</span>
              </li>
              <li>
                <strong>Docker demo (10s)</strong>
                <span>Docker only — no Kubernetes cluster required</span>
              </li>
              <li>
                <strong>Local 3-cluster fleet (60s)</strong>
                <span>Docker running; Homebrew installs kfleet, kind, kubectl, and Helm</span>
              </li>
              <li>
                <strong>Production Helm (2m)</strong>
                <span>
                  1 hub (100m CPU · 128Mi · 1Gi disk) + 1 agent/cluster (50m · 64Mi)
                </span>
              </li>
            </ul>
          </header>
          <div className="quickstart-panel">
            <div className="quickstart-tabs" role="tablist" aria-label="Quickstart methods">
              {quickstartTabs.map((tab, index) => {
                const isSelected = tab.id === activeTab
                return (
                  <button
                    key={tab.id}
                    ref={(el) => {
                      tabRefs.current[tab.id] = el
                    }}
                    type="button"
                    role="tab"
                    id={`tab-${tab.id}`}
                    aria-selected={isSelected}
                    aria-controls={`panel-${tab.id}`}
                    tabIndex={isSelected ? 0 : -1}
                    className={`quickstart-tab ${isSelected ? 'is-active' : ''}`}
                    onClick={() => selectTab(tab.id)}
                    onKeyDown={(event) => handleTabKeyDown(event, index)}
                  >
                    <span>{tab.label}</span>
                    <span className="tab-badge">{tab.badge}</span>
                  </button>
                )
              })}
            </div>

            <div
              role="tabpanel"
              id={`panel-${currentTab.id}`}
              aria-labelledby={`tab-${currentTab.id}`}
              className="quickstart-tabpanel"
            >
              <p className="quickstart-summary">{currentTab.summary}</p>
              <div className="commands" aria-label={`${currentTab.label} commands`}>
                {currentTab.steps.map((item) => (
                  <div key={`${currentTab.id}-${item.step}`}>
                    <span className="command-step">{item.step}</span>
                    <div className="command-body">
                      {item.label ? <span className="command-label">{item.label}</span> : null}
                      <code>{item.code}</code>
                    </div>
                    {item.copyValue ? (
                      <button
                        type="button"
                        className="copy-button"
                        onClick={() => void handleCopy(item.copyValue!)}
                        aria-label={
                          copiedCode === item.copyValue
                            ? `Copied ${item.code}`
                            : `Copy ${item.code}`
                        }
                      >
                        {copiedCode === item.copyValue ? 'Copied' : 'Copy'}
                      </button>
                    ) : null}
                  </div>
                ))}
              </div>
              <div className="quickstart-result">
                <CheckIcon aria-hidden="true" />
                <p>{currentTab.resultText}</p>
                <a
                  href={currentTab.resultLinkHref}
                  target={currentTab.resultExternal ? '_blank' : undefined}
                  rel={currentTab.resultExternal ? 'noopener noreferrer' : undefined}
                  onClick={() => capture(currentTab.resultEvent)}
                >
                  {currentTab.resultLinkLabel} <ExternalLinkIcon aria-hidden="true" />
                </a>
              </div>
            </div>
          </div>
        </section>
      </main>

      <footer>
        <span className="footer-brand">
          <img
            className="footer-mark"
            src="/brand/kfleet-mark.svg"
            width="24"
            height="24"
            alt=""
            aria-hidden="true"
            draggable={false}
          />
          kfleet · Read-only Kubernetes fleet control plane · Apache 2.0 · 2026
        </span>
        <div>
          <a
            href={liveDemoUrl}
            target="_blank"
            rel="noopener noreferrer"
            onClick={() => capture('footer_live_demo_clicked')}
          >
            Live demo
          </a>
          <a
            href={`${repositoryUrl}/blob/main/docs/README.md`}
            target="_blank"
            rel="noopener noreferrer"
            onClick={() => capture('docs_opened')}
          >
            Docs
          </a>
          <a
            href={`${repositoryUrl}/blob/main/CONTRIBUTING.md`}
            target="_blank"
            rel="noopener noreferrer"
            onClick={() => capture('contribution_guide_opened')}
          >
            Contribute
          </a>
          <a
            href={`${repositoryUrl}/releases`}
            target="_blank"
            rel="noopener noreferrer"
            onClick={() => capture('release_notes_opened')}
          >
            Releases
          </a>
          <a
            href={repositoryUrl}
            target="_blank"
            rel="noopener noreferrer"
            onClick={() => capture('repository_opened')}
          >
            GitHub
          </a>
        </div>
      </footer>
    </div>
  )
}

export default App
