import { Button } from '@/components/ui/button'
import {
  ActivityHeartbeatIcon,
  ArrowRightIcon,
  CheckIcon,
  SearchIcon,
  ServerIcon,
  ShieldCheckIcon,
  Terminal2Icon,
} from '@/components/icons'
import './App.css'

const repositoryUrl = 'https://github.com/1solomonwakhungu/kfleet'

const comparison = [
  ['Single-binary hub', 'Yes', 'No', 'No', 'Yes'],
  ['Built-in web app', 'Yes', 'Yes', 'Desktop', 'Yes'],
  ['Native AI interface', 'Yes', 'No', 'No', 'No'],
  ['Agent footprint', 'One agent', 'Multiple', 'None', 'None'],
  ['Open-source license', 'Apache 2.0', 'Open source', 'Partial', 'Apache 2.0'],
]

function FleetMap() {
  return (
    <figure className="fleet-map">
      <svg viewBox="0 0 1180 650" role="img" aria-labelledby="map-title map-description">
        <title id="map-title">kfleet system topology</title>
        <desc id="map-description">Operators and AI clients connect to one kfleet hub. Approved agents report state from production, staging, and development Kubernetes clusters.</desc>
        <defs>
          <pattern id="map-grid" width="36" height="36" patternUnits="userSpaceOnUse">
            <path d="M 36 0 L 0 0 0 36" className="map-grid-line" />
          </pattern>
          <marker id="map-arrow" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="5" markerHeight="5" orient="auto-start-reverse">
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
          <text x="52" y="36" className="node-title">OPERATORS</text>
          <text x="30" y="72" className="node-copy">Web app · REST API</text>
          <text x="30" y="94" className="node-meta">human access · role scoped</text>
        </g>

        <g className="map-node client-node" transform="translate(805 100)">
          <rect width="280" height="114" rx="10" />
          <circle cx="30" cy="30" r="6" className="node-signal" />
          <text x="52" y="36" className="node-title">AI CLIENTS</text>
          <text x="30" y="72" className="node-copy">Model Context Protocol</text>
          <text x="30" y="94" className="node-meta">read-only diagnosis tools</text>
        </g>

        <g className="map-node hub-node" transform="translate(440 252)">
          <rect width="300" height="134" rx="10" />
          <rect x="1" y="1" width="298" height="7" rx="4" className="hub-accent" />
          <text x="28" y="48" className="node-kicker">SINGLE GO BINARY</text>
          <text x="28" y="80" className="hub-title">kfleet hub</text>
          <text x="28" y="108" className="node-meta">inventory · events · policy · auth</text>
          <circle cx="266" cy="68" r="8" className="node-healthy" />
        </g>

        <g className="map-node store-node" transform="translate(895 270)">
          <rect width="190" height="98" rx="10" />
          <text x="24" y="38" className="node-title">SQLITE</text>
          <text x="24" y="66" className="node-copy">Durable history</text>
          <text x="24" y="84" className="node-meta">90-day default retention</text>
        </g>

        {[
          { x: 80, name: 'PRODUCTION', detail: 'approved agent' },
          { x: 455, name: 'STAGING', detail: 'approved agent' },
          { x: 830, name: 'DEVELOPMENT', detail: 'approved agent' },
        ].map(({ x, name, detail }) => (
          <g className="map-node cluster-node" transform={`translate(${x} 491)`} key={name}>
            <rect width="270" height="102" rx="10" />
            <circle cx="28" cy="28" r="7" className="node-healthy" />
            <text x="48" y="34" className="node-title">{name}</text>
            <text x="28" y="68" className="node-copy">Kubernetes cluster</text>
            <text x="28" y="88" className="node-meta">{detail} · state reports</text>
          </g>
        ))}

        <text x="590" y="233" textAnchor="middle" className="route-label">AUTHENTICATED REQUESTS</text>
        <text x="590" y="435" textAnchor="middle" className="route-label">REGISTRATION · HEARTBEATS · SNAPSHOTS</text>
      </svg>
      <figcaption>
        <span><i className="legend-dot healthy" /> approved and reporting</span>
        <span><i className="legend-dot route" /> authenticated connection</span>
        <span>One hub. One agent per cluster.</span>
      </figcaption>
    </figure>
  )
}

function App() {
  return (
    <div className="site-shell dark">
      <a className="skip-link" href="#main">Skip to content</a>
      <header className="site-header">
        <a className="brand" href="#top" aria-label="kfleet home"><img className="brand-mark" src="/brand/kfleet-mark.svg" width="28" height="28" alt="" aria-hidden="true" draggable={false} /><span>kfleet</span></a>
        <details className="command-nav">
          <summary><SearchIcon aria-hidden="true" /><span>Navigate kfleet</span><kbd>⌘ K</kbd></summary>
          <nav aria-label="Page navigation">
            <a href="#system">System map <span>01</span></a>
            <a href="#capabilities">Capabilities <span>02</span></a>
            <a href="#compare">Compare <span>03</span></a>
            <a href="#quickstart">Quickstart <span>04</span></a>
          </nav>
        </details>
      </header>

      <main id="main">
        <section className="orientation" id="top">
          <p className="orientation-line">Operators → hub → every cluster.</p>
          <h1>One control plane.<br /><span>Your whole fleet.</span></h1>
          <div className="orientation-meta">
            <p>See fleet health, find policy drift, and diagnose Kubernetes issues with AI.</p>
            <div className="orientation-actions"><Button asChild size="lg"><a href="#quickstart">Run locally <ArrowRightIcon aria-hidden="true" /></a></Button></div>
          </div>
        </section>

        <section className="system-section" id="system">
          <FleetMap />
        </section>

        <section className="capability-ledger" id="capabilities">
          <div className="ledger-intro">
            <h2>What moves through the map</h2>
            <p>kfleet observes and explains. It never auto-remediates your clusters.</p>
          </div>
          <dl>
            <div><dt><ActivityHeartbeatIcon aria-hidden="true" /> Fleet state</dt><dd>Cluster health, workloads, Kubernetes events, and an append-only operational timeline.</dd><small>Live updates over WebSocket</small></div>
            <div><dt><ShieldCheckIcon aria-hidden="true" /> Policy drift</dt><dd>Seven read-only checks report pass, fail, unknown, or stale results.</dd><small>No automated changes</small></div>
            <div><dt><Terminal2Icon aria-hidden="true" /> AI diagnosis</dt><dd>Built-in tools list clusters, locate crash loops, inspect events, and assemble a diagnosis.</dd><small>Model Context Protocol</small></div>
            <div><dt><ServerIcon aria-hidden="true" /> Durable control</dt><dd>SQLite retains registrations, disconnects, version changes, and policy findings.</dd><small>Admin · operator · read-only</small></div>
          </dl>
        </section>

        <section className="architecture-notes">
          <div className="note-statement"><span>01</span><h2>The hub is the boundary.</h2><p>Operators and AI clients connect to one authenticated service, not directly to every cluster.</p></div>
          <div className="note-statement"><span>02</span><h2>Agents stay small.</h2><p>One approved agent collects a normalized snapshot and reports it through a stable API.</p></div>
          <div className="note-statement"><span>03</span><h2>History survives.</h2><p>Events remain available across restarts and cluster deletion until retention removes them.</p></div>
        </section>

        <section className="comparison" id="compare">
          <header><h2>A deliberately smaller surface</h2><p>Native bundled capabilities only. Integrations may change each product’s scope.</p></header>
          <div className="table-wrap" tabIndex={0} aria-label="Scrollable product comparison">
            <table><thead><tr><th>Capability</th><th>kfleet</th><th>Rancher</th><th>Lens</th><th>Headlamp</th></tr></thead><tbody>{comparison.map(row => <tr key={row[0]}>{row.map((cell, index) => <td key={`${row[0]}-${cell}`} className={index === 1 ? 'kfleet-cell' : ''}>{cell}</td>)}</tr>)}</tbody></table>
          </div>
        </section>

        <section className="quickstart" id="quickstart">
          <header><h2>Three clusters.<br />One minute.</h2><p>Requires Docker, kind, kubectl, and Helm 3.</p></header>
          <div className="commands" aria-label="Quickstart commands">
            <div><span>01</span><code>git clone https://github.com/1solomonwakhungu/kfleet.git</code></div>
            <div><span>02</span><code>cd kfleet</code></div>
            <div><span>03</span><code>./hack/quickstart.sh</code></div>
          </div>
          <div className="quickstart-result"><CheckIcon aria-hidden="true" /><p>The script creates three local clusters, installs the hub and agents, then waits for registration.</p><a href="http://localhost:8080">Open localhost:8080</a></div>
        </section>
      </main>

      <footer>
        <span className="footer-brand"><img className="footer-mark" src="/brand/kfleet-mark.svg" width="24" height="24" alt="" aria-hidden="true" draggable={false} />kfleet · Apache 2.0 · 2026</span>
        <div><a href={`${repositoryUrl}/blob/main/CONTRIBUTING.md`}>Contribute</a><a href={`${repositoryUrl}/releases`}>Releases</a><a href={repositoryUrl}>GitHub</a></div>
      </footer>
    </div>
  )
}

export default App
