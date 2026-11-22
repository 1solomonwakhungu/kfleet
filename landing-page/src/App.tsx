import { useRef, useState, type KeyboardEvent } from 'react'
import { Button } from '@/components/ui/button'
import {
  ActivityHeartbeatIcon,
  ArrowRightIcon,
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
    'No',
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

type QuickstartTabId = 'live' | 'docker' | 'brew'

interface QuickstartStep {
  step: string
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
      'Explore three synthetic clusters immediately in your browser — no login, all write requests blocked with HTTP 405. You can also connect an MCP client anonymously.',
    steps: [
      {
        step: 'UI',
        code: 'https://kfleet.sullylab.de',
        copyValue: liveDemoUrl,
      },
      {
        step: 'MCP',
        code: 'KFLEET_HUB_URL=https://kfleet.sullylab.de kfleet mcp',
        copyValue: 'KFLEET_HUB_URL=https://kfleet.sullylab.de kfleet mcp',
      },
    ],
    resultText:
      'Synthetic data only (demo-us-central, demo-eu-west, demo-edge-lab). No signup or credentials required.',
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
      'Run the embedded web UI and REST API locally against an in-memory SQLite database seeded with three synthetic clusters. No Kubernetes required.',
    steps: [
      {
        step: '01',
        code: 'docker run --rm -p 8080:8080 -e KFLEET_DEMO_MODE=true ghcr.io/1solomonwakhungu/kfleet/hub:v0.2.1',
        copyValue:
          'docker run --rm -p 8080:8080 -e KFLEET_DEMO_MODE=true ghcr.io/1solomonwakhungu/kfleet/hub:v0.2.1',
      },
    ],
    resultText:
      'Starts the 14 MB Go hub in demo mode with live UI, policy drift results, and synthetic pod logs.',
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
      'Spin up three real local kind clusters with the released OCI Helm charts, port-forwarded hub UI, and the kfleet mcp CLI included. Requires Docker running.',
    steps: [
      {
        step: '01',
        code: 'brew install --cask 1solomonwakhungu/tap/kfleet',
        copyValue: 'brew install --cask 1solomonwakhungu/tap/kfleet',
      },
      {
        step: '02',
        code: 'kfleet quickstart',
        copyValue: 'kfleet quickstart',
      },
    ],
    resultText:
      'Creates kfleet-1..3 clusters, installs hub and agents, prints generated credentials, and ships kfleet mcp.',
    resultLinkLabel: 'Open localhost:8080',
    resultLinkHref: 'http://localhost:8080',
    resultExternal: true,
    resultEvent: 'local_dashboard_opened',
  },
]

function FleetMap() {
  return (
    <figure className="fleet-map">
      <svg viewBox="0 0 1180 650" role="img" aria-labelledby="map-title map-description">
        <title id="map-title">kfleet system topology</title>
        <desc id="map-description">
          Operators and AI clients connect to one 14 MB kfleet hub. Approved outbound agents with
          3-rule read-only RBAC report state from production, staging, and development Kubernetes
          clusters.
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
        </defs>
        <rect width="1180" height="650" className="map-field" />
        <rect width="1180" height="650" fill="url(#map-grid)" />

        <g className="map-route">
          <path d="M235 157 H480 Q510 157 510 187 V252" markerEnd="url(#map-arrow)" />
          <path d="M945 157 H700 Q670 157 670 187 V252" markerEnd="url(#map-arrow)" />
          <path d="M590 386 V447 H215 V491" markerEnd="url(#map-arrow)" />
          <path d="M590 386 V491" markerEnd="url(#map-arrow)" />
          <path d="M590 447 H965 V491" markerEnd="url(#map-arrow)" />
          <path d="M690 319 H895" markerEnd="url(#map-arrow)" />
        </g>

        <g className="map-node client-node" transform="translate(95 100)">
          <rect width="280" height="114" rx="10" />
          <circle cx="30" cy="30" r="6" className="node-signal" />
          <text x="52" y="36" className="node-title">
            OPERATORS
          </text>
          <text x="30" y="72" className="node-copy">
            Web app · REST API
          </text>
          <text x="30" y="94" className="node-meta">
            zero kubeconfigs · role scoped
          </text>
        </g>

        <g className="map-node client-node" transform="translate(805 100)">
          <rect width="280" height="114" rx="10" />
          <circle cx="30" cy="30" r="6" className="node-signal" />
          <text x="52" y="36" className="node-title">
            AI CLIENTS
          </text>
          <text x="30" y="72" className="node-copy">
            Model Context Protocol
          </text>
          <text x="30" y="94" className="node-meta">
            9 read-only diagnosis tools
          </text>
        </g>

        <g className="map-node hub-node" transform="translate(440 252)">
          <rect width="300" height="134" rx="10" />
          <rect x="1" y="1" width="298" height="7" rx="4" className="hub-accent" />
          <text x="28" y="48" className="node-kicker">
            14 MB GO BINARY
          </text>
          <text x="28" y="80" className="hub-title">
            kfleet hub
          </text>
          <text x="28" y="108" className="node-meta">
            inventory · logs · policy · alerts
          </text>
          <circle cx="266" cy="68" r="8" className="node-healthy" />
        </g>

        <g className="map-node store-node" transform="translate(895 270)">
          <rect width="190" height="98" rx="10" />
          <text x="24" y="38" className="node-title">
            SQLITE
          </text>
          <text x="24" y="66" className="node-copy">
            Durable history
          </text>
          <text x="24" y="84" className="node-meta">
            90-day default retention
          </text>
        </g>

        {[
          { x: 80, name: 'PRODUCTION', detail: '3-rule read-only agent' },
          { x: 455, name: 'STAGING', detail: '3-rule read-only agent' },
          { x: 830, name: 'EDGE / NAT', detail: '3-rule read-only agent' },
        ].map(({ x, name, detail }) => (
          <g className="map-node cluster-node" transform={`translate(${x} 491)`} key={name}>
            <rect width="270" height="102" rx="10" />
            <circle cx="28" cy="28" r="7" className="node-healthy" />
            <text x="48" y="34" className="node-title">
              {name}
            </text>
            <text x="28" y="68" className="node-copy">
              Kubernetes cluster
            </text>
            <text x="28" y="88" className="node-meta">
              {detail} · outbound
            </text>
          </g>
        ))}

        <text x="590" y="233" textAnchor="middle" className="route-label">
          AUTHENTICATED REQUESTS
        </text>
        <text x="590" y="435" textAnchor="middle" className="route-label">
          OUTBOUND HTTPS · REVERSE WEBSOCKET LOGS
        </text>
      </svg>
      <figcaption>
        <span>
          <i className="legend-dot healthy" /> approved and reporting
        </span>
        <span>
          <i className="legend-dot route" /> outbound-only connection
        </span>
        <span>One 14 MB hub. One 3-rule read-only agent per cluster.</span>
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
  })
  const [activeTab, setActiveTab] = useState<QuickstartTabId>('live')
  const [copiedCode, setCopiedCode] = useState<string | null>(null)

  const currentTab = quickstartTabs.find((tab) => tab.id === activeTab) ?? quickstartTabs[0]

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
          </summary>
          <nav aria-label="Page navigation">
            <a href="#demo" onClick={closeNav}>
              Live UI <span>01</span>
            </a>
            <a href="#system" onClick={closeNav}>
              System map <span>02</span>
            </a>
            <a href="#capabilities" onClick={closeNav}>
              Capabilities <span>03</span>
            </a>
            <a href="#compare" onClick={closeNav}>
              Compare <span>04</span>
            </a>
            <a href="#quickstart" onClick={closeNav}>
              Quickstart <span>05</span>
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
            14 MB Go + SQLite · Zero central kubeconfigs · 3-rule read-only agent RBAC
          </p>
          <h1>
            One control plane.
            <br />
            <span>Your whole fleet.</span>
          </h1>
          <div className="orientation-meta">
            <p>
              Read-only multi-cluster Kubernetes visibility and MCP diagnostics in one 14 MB Go +
              SQLite binary — zero kubeconfigs centrally, no Secrets or write RBAC.
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
                Run locally in 10s <ArrowRightIcon aria-hidden="true" />
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
              <span className="demo-chrome-meta">HTTP 405 on writes · no login</span>
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

        <section className="capability-ledger" id="capabilities">
          <div className="ledger-intro">
            <h2>What moves through the map</h2>
            <p>
              kfleet observes and explains. Agents hold three read-only RBAC rules — never Secrets,
              never ConfigMaps, never writes.
            </p>
          </div>
          <dl>
            <div>
              <dt>
                <ActivityHeartbeatIcon aria-hidden="true" /> Fleet state &amp; logs
              </dt>
              <dd>
                Cluster health, workloads, Kubernetes events, live reverse-WebSocket pod logs, and
                an append-only operational timeline.
              </dd>
              <small>Live updates over WebSocket &amp; SSE</small>
            </div>
            <div>
              <dt>
                <ShieldCheckIcon aria-hidden="true" /> Policy &amp; drift
              </dt>
              <dd>
                Seven built-in read-only checks report pass, fail, unknown, or stale results across
                every cluster.
              </dd>
              <small>Zero automated mutations</small>
            </div>
            <div>
              <dt>
                <Terminal2Icon aria-hidden="true" /> Read-only MCP
              </dt>
              <dd>
                Nine built-in MCP tools list clusters, find crashlooping pods, stream previous
                container logs, and diagnose events without kubeconfigs.
              </dd>
              <small>Model Context Protocol · stdio</small>
            </div>
            <div>
              <dt>
                <ServerIcon aria-hidden="true" /> Durable SQLite history
              </dt>
              <dd>
                Retains last snapshots, disconnects, version changes, policy findings, and signed
                health alerts even after a cluster goes offline.
              </dd>
              <small>90-day retention · Admin · operator · read-only</small>
            </div>
          </dl>
        </section>

        <section className="architecture-notes">
          <div className="note-statement">
            <span>01</span>
            <h2>The hub is the boundary.</h2>
            <p>
              Operators and AI clients talk to one 14 MB Go service with role-scoped sessions and an
              audit log — never directly to cluster API servers.
            </p>
          </div>
          <div className="note-statement">
            <span>02</span>
            <h2>Agents stay minimal.</h2>
            <p>
              One approved outbound agent per cluster runs with a 3-rule read-only ClusterRole. No
              Secrets, no ConfigMaps, no write verbs.
            </p>
          </div>
          <div className="note-statement">
            <span>03</span>
            <h2>History survives outages.</h2>
            <p>
              Snapshots and timeline events remain queryable in SQLite when a private or edge
              cluster drops offline.
            </p>
          </div>
        </section>

        <section className="comparison" id="compare">
          <header>
            <h2>A deliberately smaller surface</h2>
            <p>
              Built for read-only fleet visibility and safe MCP diagnostics — not cluster
              provisioning or day-2 mutations. Competitor details checked November 2026.
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
          <p className="comparison-note">Competitor details checked November 2026.</p>
        </section>

        <section className="quickstart" id="quickstart">
          <header>
            <h2>
              Try it in
              <br />
              seconds.
            </h2>
            <p>
              Start in the browser with zero setup, spin up demo mode in Docker with no Kubernetes,
              or launch three local kind clusters via Homebrew.
            </p>
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
                    <code>{item.code}</code>
                    {item.copyValue ? (
                      <button
                        type="button"
                        className="copy-button"
                        onClick={() => handleCopy(item.copyValue!)}
                        aria-label={`Copy ${item.code}`}
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
          kfleet · Apache 2.0 · 2026
        </span>
        <div>
          <a
            href={liveDemoUrl}
            target="_blank"
            rel="noopener noreferrer"
            onClick={() => capture('footer_live_demo_opened')}
          >
            Live demo
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
