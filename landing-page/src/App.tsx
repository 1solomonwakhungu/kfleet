import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { ActivityHeartbeatIcon, ArrowRightIcon, BrandGithubIcon, CheckIcon, ServerIcon, ShieldCheckIcon, Terminal2Icon } from '@/components/icons'
import './App.css'

const repositoryUrl = 'https://github.com/1solomonwakhungu/kfleet'
const capabilities = [
  { icon: ActivityHeartbeatIcon, title: 'One live fleet view', body: 'Track health, workloads, Kubernetes events, and retained operational history across every connected cluster.' },
  { icon: ShieldCheckIcon, title: 'Drift without guesswork', body: 'Evaluate seven built-in policy checks with pass, fail, unknown, and stale results. kfleet never auto-remediates.' },
  { icon: Terminal2Icon, title: 'Ask your fleet directly', body: 'Use the built-in Model Context Protocol server to inspect clusters, find crash loops, and diagnose warnings in natural language.' },
]
const comparison = [
  ['Single-binary hub', 'Yes', 'No', 'No', 'Yes'],
  ['Built-in web UI', 'Yes', 'Yes', 'Desktop', 'Yes'],
  ['Native AI interface', 'Yes', 'No', 'No', 'No'],
  ['Agent footprint', 'One agent', 'Multiple', 'None', 'None'],
  ['Open source', 'Apache 2.0', 'Yes', 'Partial', 'Apache 2.0'],
]

function FleetPreview() {
  return <Card className="fleet-preview" aria-label="Example kfleet dashboard showing three healthy clusters">
    <div className="preview-toolbar"><div className="preview-title"><span className="brand-mark small" aria-hidden="true">k</span> Fleet overview</div><span className="live-state"><span /> Live</span></div>
    <div className="summary-strip"><div><strong>3</strong><span>Clusters</span></div><div><strong>3</strong><span>Healthy</span></div><div><strong>42</strong><span>Workloads</span></div><div><strong>0</strong><span>Open alerts</span></div></div>
    <div className="cluster-list">{[['production', 'v1.31.2', '18 workloads', '12s ago'], ['staging', 'v1.31.2', '14 workloads', '8s ago'], ['development', 'v1.30.6', '10 workloads', '4s ago']].map(([name, version, workloads, heartbeat]) => <div className="cluster-row" key={name}><span className="health-dot" aria-label="Healthy" /><div><strong>{name}</strong><span>{version}</span></div><span>{workloads}</span><span>{heartbeat}</span></div>)}</div>
    <div className="preview-footer"><span>Last fleet event</span><span>staging agent reconnected</span><time>09:41:18</time></div>
  </Card>
}

function App() {
  return <div className="site-shell dark">
    <a className="skip-link" href="#main">Skip to content</a>
    <header className="site-header"><a className="brand" href="#top" aria-label="kfleet home"><span className="brand-mark" aria-hidden="true">k</span><span>kfleet</span></a><nav aria-label="Primary navigation"><a href="#capabilities">Capabilities</a><a href="#architecture">Architecture</a><a href="#quickstart">Quickstart</a></nav><Button asChild variant="secondary" size="sm"><a href={repositoryUrl} target="_blank" rel="noreferrer"><BrandGithubIcon className="size-4" aria-hidden="true" /> GitHub</a></Button></header>
    <main id="main">
      <section className="hero" id="top"><div className="hero-copy"><h1>Operate every cluster from one place.</h1><Badge variant="secondary" size="lg">Open source · Apache 2.0</Badge><p>See fleet health, find policy drift, and diagnose issues with AI.</p><div className="hero-actions"><Button asChild size="lg"><a href={repositoryUrl} target="_blank" rel="noreferrer"><BrandGithubIcon className="size-4" aria-hidden="true" /> View on GitHub</a></Button><Button asChild variant="outline" size="lg"><a href="#quickstart">Run the quickstart <ArrowRightIcon className="size-4" aria-hidden="true" /></a></Button></div></div><FleetPreview /></section>
      <section className="trust-line" aria-label="Supported workflows"><span>Kubernetes</span><span>Helm</span><span>WebSocket</span><span>Model Context Protocol</span><span>Multi-tenant</span></section>
      <section className="section capabilities" id="capabilities"><div className="section-heading"><h2>See the signal. Keep control.</h2><p>One operational layer for people, automation, and AI clients.</p></div><div className="capability-list">{capabilities.map(({ icon: Icon, title, body }, index) => <article key={title}><span className="feature-number">0{index + 1}</span><Icon className="feature-icon" aria-hidden="true" /><div><h3>{title}</h3><p>{body}</p></div></article>)}</div></section>
      <section className="section architecture" id="architecture"><div className="section-heading"><h2>Small footprint. Clear boundaries.</h2><p>The hub is the only service your operators and AI clients connect to.</p></div><div className="architecture-grid"><div className="flow-diagram" aria-label="kfleet architecture flow"><div className="flow-node client"><Terminal2Icon aria-hidden="true" /><strong>Operators and AI clients</strong><span>Web UI · REST · MCP</span></div><div className="flow-line"><span>Authenticated requests</span></div><div className="flow-node hub"><ServerIcon aria-hidden="true" /><strong>kfleet hub</strong><span>Single Go binary · SQLite</span></div><div className="flow-line"><span>Registration and reports</span></div><div className="agent-row"><span>prod agent</span><span>stage agent</span><span>dev agent</span></div></div><div className="architecture-copy"><h3>Designed for understandable operations</h3><ul><li><CheckIcon aria-hidden="true" /><span><strong>Normalized state.</strong> Small agents collect cluster snapshots and report through one stable API.</span></li><li><CheckIcon aria-hidden="true" /><span><strong>Durable history.</strong> Registrations, disconnects, version changes, and policy findings survive restarts.</span></li><li><CheckIcon aria-hidden="true" /><span><strong>Explicit access.</strong> Admin, operator, and read-only roles protect human and machine workflows.</span></li></ul></div></div></section>
      <section className="section compare" id="compare"><div className="section-heading"><h2>Built for a lighter path.</h2><p>Native bundled capabilities, compared without counting third-party extensions.</p></div><div className="table-wrap" tabIndex={0} aria-label="Scrollable product comparison"><table><thead><tr><th>Capability</th><th>kfleet</th><th>Rancher</th><th>Lens</th><th>Headlamp</th></tr></thead><tbody>{comparison.map(row => <tr key={row[0]}>{row.map((cell, index) => <td key={`${row[0]}-${cell}`} className={index === 1 ? 'kfleet-cell' : ''}>{cell}</td>)}</tr>)}</tbody></table></div></section>
      <section className="section quickstart" id="quickstart"><div className="quickstart-copy"><h2>Try a three-cluster fleet.</h2><p>Docker, kind, kubectl, and Helm are all you need. The script builds the hub and agents, creates three local clusters, and waits for registration.</p><div className="requirements"><span>Docker</span><span>kind</span><span>kubectl</span><span>Helm 3</span></div></div><div className="code-panel" aria-label="Quickstart commands"><div className="code-title"><Terminal2Icon aria-hidden="true" /> Terminal</div><pre><code><span>git clone</span> https://github.com/1solomonwakhungu/kfleet.git{`\n`}<span>cd</span> kfleet{`\n`}<span>./hack/quickstart.sh</span></code></pre><div className="code-note"><span className="health-dot" /> Hub ready at localhost:8080</div></div></section>
      <section className="closing"><div><h2>Your fleet, without the weight.</h2><p>Inspect the code, run the local demo, and connect your first cluster.</p></div></section>
    </main>
    <footer><a className="brand" href="#top"><span className="brand-mark small" aria-hidden="true">k</span><span>kfleet</span></a><p>Lightweight multi-cluster Kubernetes management.</p><div><a href={`${repositoryUrl}/blob/main/CONTRIBUTING.md`}>Contribute</a><a href={`${repositoryUrl}/releases`}>Releases</a><a href={repositoryUrl}>GitHub</a></div></footer>
  </div>
}
export default App
