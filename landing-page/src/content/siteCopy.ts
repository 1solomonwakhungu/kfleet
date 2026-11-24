export interface SiteMeta {
  readonly name: string
  readonly version: string
  readonly tagline: string
  readonly description: string
  readonly license: string
}

export interface SiteLinks {
  readonly repoUrl: string
  readonly releaseUrl: string
  readonly releasesUrl: string
  readonly issuesUrl: string
  readonly liveDemoUrl: string
  readonly docsUrl: string
  readonly discussionsUrl: string
  readonly contributingUrl: string
  readonly licenseUrl: string
}

export interface NavItem {
  readonly label: string
  readonly href: string
}

export interface ActionLink {
  readonly label: string
  readonly href: string
  readonly external: boolean
}

export interface NavCopy {
  readonly items: readonly NavItem[]
  readonly primaryCta: ActionLink
  readonly secondaryCta: ActionLink
}

export interface HeroAnnouncement {
  readonly badge: string
  readonly text: string
  readonly href: string
}

export interface QuickCommand {
  readonly label: string
  readonly command: string
}

export interface DemoFrameCopy {
  readonly urlLabel: string
  readonly statusLabel: string
  readonly gifSrc: string
  readonly gifAlt: string
  readonly caption: string
}

export interface HeroCopy {
  readonly announcement: HeroAnnouncement
  readonly headline: string
  readonly headlineAccent: string
  readonly subheadline: string
  readonly primaryCta: ActionLink
  readonly secondaryCta: ActionLink
  readonly tertiaryCta: ActionLink
  readonly quickCommand: QuickCommand
  readonly demoFrame: DemoFrameCopy
}

export interface MetricStat {
  readonly value: string
  readonly label: string
  readonly detail: string
}

export interface SecurityCardCopy {
  readonly id: string
  readonly title: string
  readonly description: string
  readonly guarantees: readonly string[]
  readonly manifestPath: string
  readonly clusterRoleYaml: string
}

export type SampleEventStatus = 'unreachable' | 'degraded' | 'healthy'

export interface SampleTimelineEvent {
  readonly cluster: string
  readonly status: SampleEventStatus
  readonly statusLabel: string
  readonly summary: string
  readonly timestampLabel: string
}

export interface OfflineHistoryCardCopy {
  readonly title: string
  readonly description: string
  readonly sampleEvents: readonly SampleTimelineEvent[]
}

export type PolicySeverity = 'low' | 'medium' | 'high' | 'critical'
export type PolicyScope = 'fleet' | 'cluster' | 'namespace' | 'workload'
export type PolicyStatus = 'pass' | 'fail' | 'unknown' | 'stale'

export interface PolicyCheckItem {
  readonly id: string
  readonly label: string
  readonly severity: PolicySeverity
  readonly scope: PolicyScope
}

export interface PolicyCardCopy {
  readonly title: string
  readonly description: string
  readonly statuses: readonly PolicyStatus[]
  readonly checks: readonly PolicyCheckItem[]
}

export interface AlertsLogsCardCopy {
  readonly title: string
  readonly description: string
  readonly highlights: readonly string[]
}

export interface FeaturesCopy {
  readonly id: string
  readonly eyebrow: string
  readonly title: string
  readonly subtitle: string
  readonly securityCard: SecurityCardCopy
  readonly offlineHistoryCard: OfflineHistoryCardCopy
  readonly policyCard: PolicyCardCopy
  readonly alertsLogsCard: AlertsLogsCardCopy
}

export interface McpToolItem {
  readonly name: string
  readonly category: string
  readonly summary: string
}

export interface McpCopy {
  readonly id: string
  readonly eyebrow: string
  readonly title: string
  readonly subtitle: string
  readonly demoCommand: string
  readonly claudeCodeCommand: string
  readonly demoConfigJson: string
  readonly tools: readonly McpToolItem[]
  readonly samplePrompts: readonly string[]
}

export interface ArchitecturePillar {
  readonly index: string
  readonly title: string
  readonly description: string
}

export interface ArchitectureCopy {
  readonly id: string
  readonly eyebrow: string
  readonly title: string
  readonly subtitle: string
  readonly pillars: readonly ArchitecturePillar[]
  readonly topologyCaption: string
}

export interface ComparisonRow {
  readonly capability: string
  readonly values: readonly [string, string, string, string, string, string]
}

export interface ComparisonCopy {
  readonly id: string
  readonly eyebrow: string
  readonly title: string
  readonly subtitle: string
  readonly footnote: string
  readonly columns: readonly [string, string, string, string, string, string, string]
  readonly rows: readonly ComparisonRow[]
}

export type QuickstartTabId = 'live' | 'docker' | 'brew' | 'helm'

export interface QuickstartStep {
  readonly label: string
  readonly command: string
  readonly note?: string
}

export interface QuickstartTab {
  readonly id: QuickstartTabId
  readonly label: string
  readonly badge: string
  readonly title: string
  readonly description: string
  readonly steps: readonly QuickstartStep[]
}

export interface QuickstartCopy {
  readonly id: string
  readonly eyebrow: string
  readonly title: string
  readonly subtitle: string
  readonly tabs: readonly QuickstartTab[]
}

export interface TradeoffItem {
  readonly scenario: string
  readonly recommendation: string
  readonly reason: string
}

export interface TradeoffsCopy {
  readonly eyebrow: string
  readonly title: string
  readonly subtitle: string
  readonly items: readonly TradeoffItem[]
}

export interface FaqItem {
  readonly id: string
  readonly question: string
  readonly answer: string
}

export interface FaqCopy {
  readonly id: string
  readonly eyebrow: string
  readonly title: string
  readonly subtitle: string
  readonly items: readonly FaqItem[]
}

export interface CtaBannerCopy {
  readonly headline: string
  readonly description: string
  readonly primaryCta: ActionLink
  readonly secondaryCta: ActionLink
}

export interface FooterLink {
  readonly label: string
  readonly href: string
  readonly external: boolean
}

export interface FooterColumn {
  readonly title: string
  readonly links: readonly FooterLink[]
}

export interface FooterCopy {
  readonly tagline: string
  readonly copyright: string
  readonly columns: readonly FooterColumn[]
}

export interface SiteCopy {
  readonly meta: SiteMeta
  readonly links: SiteLinks
  readonly nav: NavCopy
  readonly hero: HeroCopy
  readonly metrics: readonly MetricStat[]
  readonly features: FeaturesCopy
  readonly mcp: McpCopy
  readonly architecture: ArchitectureCopy
  readonly comparison: ComparisonCopy
  readonly quickstart: QuickstartCopy
  readonly tradeoffs: TradeoffsCopy
  readonly faq: FaqCopy
  readonly cta: CtaBannerCopy
  readonly footer: FooterCopy
}

// Bumped automatically by release-please (see release-please-config.json extra-files).
const version = '0.5.0' // x-release-please-version
const releaseTag = `v${version}`

const repoUrl = 'https://github.com/1solomonwakhungu/kfleet'
const releaseUrl = `${repoUrl}/releases/tag/${releaseTag}`
const liveDemoUrl = 'https://kfleet.sullylab.de'
const docsBaseUrl = `${repoUrl}/blob/main/docs`

export const siteCopy: SiteCopy = {
  meta: {
    name: 'kfleet',
    version: releaseTag,
    tagline: 'Multi-cluster Kubernetes visibility. Zero kubeconfigs.',
    description:
      'A 14 MB Go + SQLite control plane with outbound 3-rule read-only agents, live reverse-WebSocket pod logs, cross-cluster drift checks, and a built-in 9-tool MCP server for AI diagnostics.',
    license: 'Apache License 2.0',
  },

  links: {
    repoUrl,
    releaseUrl,
    releasesUrl: `${repoUrl}/releases`,
    issuesUrl: `${repoUrl}/issues`,
    liveDemoUrl,
    docsUrl: `${docsBaseUrl}/README.md`,
    discussionsUrl: `${repoUrl}/discussions`,
    contributingUrl: `${repoUrl}/blob/main/CONTRIBUTING.md`,
    licenseUrl: `${repoUrl}/blob/main/LICENSE`,
  },

  nav: {
    items: [
      { label: 'Features', href: '#features' },
      { label: 'Security', href: '#security' },
      { label: 'MCP', href: '#mcp' },
      { label: 'Architecture', href: '#architecture' },
      { label: 'Compare', href: '#compare' },
      { label: 'Quickstart', href: '#quickstart' },
      { label: 'FAQ', href: '#faq' },
    ],
    primaryCta: {
      label: 'Try live demo',
      href: liveDemoUrl,
      external: true,
    },
    secondaryCta: {
      label: 'GitHub',
      href: repoUrl,
      external: true,
    },
  },

  hero: {
    announcement: {
      badge: releaseTag,
      text: 'Read-only MCP server & Homebrew CLI now live',
      href: releaseUrl,
    },
    headline: 'Multi-cluster Kubernetes visibility.',
    headlineAccent: 'Zero kubeconfigs.',
    subheadline:
      'A 14 MB Go + SQLite control plane with outbound 3-rule read-only agents, live reverse-WebSocket pod logs, cross-cluster drift checks, and a built-in 9-tool MCP server for AI diagnostics.',
    primaryCta: {
      label: 'Try live demo',
      href: liveDemoUrl,
      external: true,
    },
    secondaryCta: {
      label: 'Star on GitHub',
      href: repoUrl,
      external: true,
    },
    tertiaryCta: {
      label: 'Run locally in 10s',
      href: '#quickstart',
      external: false,
    },
    quickCommand: {
      label: '10-second Docker demo',
      command:
        `docker run --rm -p 8080:8080 -e KFLEET_DEMO_MODE=true ghcr.io/1solomonwakhungu/kfleet/hub:${releaseTag}`,
    },
    demoFrame: {
      urlLabel: 'kfleet.sullylab.de',
      statusLabel: 'Read-only synthetic demo · HTTP 405 on writes · No login',
      gifSrc: '/demo.gif',
      gifAlt:
        'Tour of the kfleet web UI showing three synthetic clusters (demo-us-central, demo-eu-west, demo-edge-lab), cluster detail view, live pod logs, policy and drift results, and alert history',
      caption:
        'Running live with three synthetic clusters (demo-us-central, demo-eu-west, demo-edge-lab). No real cluster data or credentials.',
    },
  },

  metrics: [
    {
      value: '14 MB',
      label: 'Static Go binary',
      detail: 'CGO_ENABLED=0, embedded UI & SQLite WAL',
    },
    {
      value: '~20 MB',
      label: 'Idle hub memory',
      detail: '3 clusters loaded, zero CRDs or operators',
    },
    {
      value: '3 rules',
      label: 'Agent ClusterRole',
      detail: 'Read-only inventory + pods/log, no Secrets',
    },
    {
      value: '9 tools',
      label: 'Built-in MCP server',
      detail: 'Read-only stdio diagnostics via kfleet mcp',
    },
  ],

  features: {
    id: 'features',
    eyebrow: 'Capabilities',
    title: 'Built for fleet clarity, not cluster sprawl',
    subtitle:
      'Everything an engineer or AI assistant needs to inspect cluster health, trace drift, and read pod logs — without widening write or secret blast radius.',
    securityCard: {
      id: 'security',
      title: 'Zero central kubeconfigs. 3-rule read-only agent RBAC.',
      description:
        'The hub never stores kubeconfigs, service account tokens, or cluster certificates. Approved agents dial out over HTTPS and WebSocket using a minimal ClusterRole that cannot read Secrets, ConfigMaps, or mutate workloads.',
      guarantees: [
        'No Secrets access',
        'No ConfigMaps access',
        'No write verbs',
        'Outbound-only agent',
      ],
      manifestPath: 'charts/kfleet-agent/templates/clusterrole.yaml',
      clusterRoleYaml: `rules:
  - apiGroups: [""]
    resources: ["pods", "services", "nodes", "events", "namespaces"]
    verbs: ["list"]
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["list"]
  - apiGroups: [""]
    resources: ["pods/log"]
    verbs: ["get"]`,
    },
    offlineHistoryCard: {
      title: 'Remembers when clusters go dark',
      description:
        'Live-pull tools go blind the second a cluster API server drops offline. kfleet keeps each cluster’s last snapshot plus a 90-day append-only SQLite timeline of registrations, disconnects, reconnects, version changes, and policy findings.',
      sampleEvents: [
        {
          cluster: 'demo-edge-lab',
          status: 'unreachable',
          statusLabel: 'Unreachable',
          summary: 'Last snapshot preserved (22m ago)',
          timestampLabel: '22m ago',
        },
        {
          cluster: 'demo-eu-west',
          status: 'degraded',
          statusLabel: 'Degraded',
          summary: 'ledger-api-6f75-demo CrashLoopBackOff (7 restarts)',
          timestampLabel: '4m ago',
        },
        {
          cluster: 'demo-us-central',
          status: 'healthy',
          statusLabel: 'Healthy',
          summary: 'v1.31.4 · 3 nodes ready',
          timestampLabel: '12s ago',
        },
      ],
    },
    policyCard: {
      title: '7 built-in policy & drift checks',
      description:
        'Continuously evaluates version skew, required labels, workload drift, and pod security baselines — reporting pass, fail, unknown, or stale without auto-mutating anything.',
      statuses: ['pass', 'fail', 'unknown', 'stale'],
      checks: [
        {
          id: 'kubernetes-version-consistency',
          label: 'Kubernetes version consistency',
          severity: 'high',
          scope: 'fleet',
        },
        {
          id: 'required-cluster-labels',
          label: 'Required cluster labels',
          severity: 'medium',
          scope: 'cluster',
        },
        {
          id: 'namespace-consistency',
          label: 'Namespace consistency',
          severity: 'medium',
          scope: 'namespace',
        },
        {
          id: 'workload-configuration-consistency',
          label: 'Workload config consistency',
          severity: 'high',
          scope: 'workload',
        },
        {
          id: 'workload-availability',
          label: 'Workload availability',
          severity: 'high',
          scope: 'workload',
        },
        {
          id: 'pod-security-baseline',
          label: 'Pod security baseline',
          severity: 'critical',
          scope: 'workload',
        },
        {
          id: 'namespace-pod-security',
          label: 'Namespace pod security',
          severity: 'high',
          scope: 'namespace',
        },
      ],
    },
    alertsLogsCard: {
      title: 'Reverse-WebSocket pod logs & signed alerts',
      description:
        'Stream live or previous container logs over SSE through the agent’s outbound WebSocket — even from NAT/edge clusters — and route HMAC-SHA256-signed health webhooks with durable retries.',
      highlights: [
        '512-line backpressure buffer',
        'previous=true crashloop logs',
        'HMAC-SHA256 signed webhooks',
        'Prometheus /metrics',
      ],
    },
  },

  mcp: {
    id: 'mcp',
    eyebrow: 'Model Context Protocol',
    title: 'Give AI clients fleet visibility without kubeconfig sprawl',
    subtitle:
      'Ship read-only Kubernetes diagnostics directly inside kfleet mcp (Homebrew) and hub mcp (Releases / Docker). One audited hub boundary — or try it right now with zero credentials against our public demo.',
    demoCommand: 'KFLEET_HUB_URL=https://kfleet.sullylab.de kfleet mcp',
    claudeCodeCommand:
      'claude mcp add kfleet-demo -e KFLEET_HUB_URL=https://kfleet.sullylab.de -- kfleet mcp',
    demoConfigJson: `{
  "mcpServers": {
    "kfleet-demo": {
      "command": "kfleet",
      "args": ["mcp"],
      "env": {
        "KFLEET_HUB_URL": "https://kfleet.sullylab.de"
      }
    }
  }
}`,
    tools: [
      {
        name: 'list_clusters',
        category: 'Fleet',
        summary: 'Every cluster with health, Kubernetes version, node count, and pod count.',
      },
      {
        name: 'get_cluster_status',
        category: 'Cluster',
        summary: 'Latest full status snapshot and node details for a single cluster.',
      },
      {
        name: 'list_namespaces',
        category: 'Inventory',
        summary: 'Namespaces in a cluster’s latest snapshot.',
      },
      {
        name: 'get_pods',
        category: 'Inventory',
        summary: 'Pods in a cluster, optionally filtered by namespace.',
      },
      {
        name: 'get_events',
        category: 'Events',
        summary: 'Kubernetes events from the latest snapshot (filterable by namespace and Normal/Warning).',
      },
      {
        name: 'get_recent_events',
        category: 'Timeline',
        summary:
          'Durable operational timeline of registrations, disconnects, reconnects, version changes, and policy findings.',
      },
      {
        name: 'find_crashloop_pods',
        category: 'Diagnostics',
        summary:
          'Pods in CrashLoopBackOff or with >= 5 container restarts across one cluster or the whole fleet.',
      },
      {
        name: 'get_pod_logs',
        category: 'Logs',
        summary:
          'Recent container log lines including previous=true for crashlooping containers (default 100 lines, max 1000, capped at 256 KiB).',
      },
      {
        name: 'diagnose_cluster',
        category: 'Diagnostics',
        summary: 'Composite report combining cluster status, crashlooping pods, and Warning events.',
      },
    ],
    samplePrompts: [
      'List my clusters and summarize anything that is not healthy.',
      'Find crashlooping pods across the fleet and show me the previous container logs from the worst one.',
      'Diagnose demo-eu-west and explain its recent Warning events.',
      'Show recent disconnects, reconnects, and policy findings across the fleet.',
    ],
  },

  architecture: {
    id: 'architecture',
    eyebrow: 'Architecture',
    title: 'One 14 MB boundary between your fleet and every client',
    subtitle:
      'No management cluster, no CRDs, no cert-manager, and no inbound firewall rules to your cluster API servers.',
    pillars: [
      {
        index: '01',
        title: 'The hub is the boundary',
        description:
          'One static Go binary serves the embedded React UI, REST API, live dashboard WebSocket, SSE log relay, policy engine, signed alert worker, Prometheus /metrics, and MCP clients behind role-based auth (admin, operator, read_only) and an append-only audit log.',
      },
      {
        index: '02',
        title: 'Agents stay minimal & outbound',
        description:
          'One lightweight Deployment per cluster (50m CPU / 64Mi RAM default requests) registers with a shared token, waits for operator or admin approval, then initiates outbound HTTPS heartbeats (15s), snapshots (30s), and a reverse WebSocket channel for logs.',
      },
      {
        index: '03',
        title: 'Operational history survives outages',
        description:
          'SQLite in WAL mode persists each cluster’s latest snapshot alongside a 90-day operational timeline that survives disconnects and cluster deletion, marking clusters stale after 3 missed heartbeat intervals.',
      },
    ],
    topologyCaption:
      'One 14 MB Go hub + SQLite. One 3-rule read-only outbound agent per cluster.',
  },

  comparison: {
    id: 'compare',
    eyebrow: 'Comparison',
    title: 'A deliberately smaller surface',
    subtitle:
      'Built for read-only fleet visibility and safe MCP diagnostics — not cluster provisioning or day-2 mutations.',
    footnote:
      'Competitor details checked November 2026. If a cell is out of date, please open an issue on GitHub.',
    columns: [
      'Capability',
      'kfleet',
      'Rancher',
      'Lens',
      'Headlamp',
      'OCM',
      'Kubeconfig MCP servers',
    ],
    rows: [
      {
        capability: 'Central footprint',
        values: [
          '1 Go binary + SQLite',
          'Dedicated K8s cluster',
          'None',
          'None, or in-cluster app',
          'Operator + controllers',
          'Local/in-cluster process',
        ],
      },
      {
        capability: 'In-cluster footprint',
        values: [
          '1 agent (3 read rules)',
          'cluster-admin agent',
          'None',
          'None',
          'Klusterlet agents',
          'None',
        ],
      },
      {
        capability: 'Central/laptop kubeconfigs',
        values: [
          'None',
          'Central proxy',
          'Per-user kubeconfig',
          'Per-user kubeconfig',
          'Agents pull work',
          'Required on client/host',
        ],
      },
      {
        capability: 'Can mutate clusters',
        values: ['No', 'Yes', 'Yes', 'Yes', 'Yes', 'Depends on RBAC'],
      },
      {
        capability: 'Offline cluster history',
        values: [
          'Last snapshot + 90d timeline',
          'Last status',
          'None',
          'None',
          'Last status',
          'None',
        ],
      },
      {
        capability: 'AI / MCP interface',
        values: [
          'Read-only MCP built in (9 tools)',
          'Liz + MCP (Prime)',
          'Prism + MCP (proprietary)',
          'AI Assistant plugin',
          'Community MCP servers',
          'MCP only (no UI/alerts/drift)',
        ],
      },
      {
        capability: 'License',
        values: [
          'Apache 2.0',
          'Apache 2.0',
          'Proprietary / free tier',
          'Apache 2.0',
          'Apache 2.0',
          'Varies',
        ],
      },
    ],
  },

  quickstart: {
    id: 'quickstart',
    eyebrow: 'Quickstart',
    title: 'Try it in seconds',
    subtitle:
      'Explore the hosted demo immediately, spin up a local Docker container with no Kubernetes, launch three kind clusters via Homebrew, or deploy with Helm.',
    tabs: [
      {
        id: 'live',
        label: 'Live demo',
        badge: '0s',
        title: 'Explore the hosted synthetic fleet',
        description:
          'Open the read-only demo in your browser or connect your MCP client right now — zero logins, zero kubeconfigs, HTTP 405 on all writes.',
        steps: [
          {
            label: 'Open the web UI',
            command: 'https://kfleet.sullylab.de',
            note: 'Three synthetic clusters (demo-us-central, demo-eu-west, demo-edge-lab) with live log simulation, policy drift, and timeline history.',
          },
          {
            label: 'Connect an MCP client (no credentials)',
            command: 'KFLEET_HUB_URL=https://kfleet.sullylab.de kfleet mcp',
            note: 'Or register it directly in Claude Code with: claude mcp add kfleet-demo -e KFLEET_HUB_URL=https://kfleet.sullylab.de -- kfleet mcp',
          },
        ],
      },
      {
        id: 'docker',
        label: 'Docker',
        badge: '10s',
        title: 'Run locally with no Kubernetes required',
        description:
          'Starts the embedded React UI and REST API against an in-memory SQLite database seeded with three synthetic clusters.',
        steps: [
          {
            label: `Launch hub ${releaseTag} in demo mode`,
            command:
              `docker run --rm -p 8080:8080 \\\n  -e KFLEET_DEMO_MODE=true \\\n  ghcr.io/1solomonwakhungu/kfleet/hub:${releaseTag}`,
            note: 'Then open http://localhost:8080 in your browser.',
          },
        ],
      },
      {
        id: 'brew',
        label: 'Homebrew + kind',
        badge: '60s',
        title: 'Boot three real local kind clusters',
        description:
          'Installs the kfleet CLI (quickstart + read-only MCP server) with kind, kubectl, and helm, then deploys the released hub and three agents locally.',
        steps: [
          {
            label: 'Install the Homebrew cask',
            command: 'brew install --cask 1solomonwakhungu/tap/kfleet',
          },
          {
            label: 'Create three kind clusters and start the stack',
            command: 'kfleet quickstart',
            note: 'Prints your local hub URL and generated admin credentials. Run kfleet cleanup to tear everything down.',
          },
          {
            label: 'Start the read-only stdio MCP server',
            command: 'kfleet mcp',
          },
        ],
      },
      {
        id: 'helm',
        label: 'Helm',
        badge: 'Prod',
        title: 'Deploy hub + outbound agents via OCI Helm charts',
        description:
          'Install the hub once with SQLite persistence enabled, then deploy the 3-rule read-only agent into every cluster you want to observe.',
        steps: [
          {
            label: '1. Install kfleet-hub (once)',
            command: `export KFLEET_REGISTRATION_TOKEN="$(openssl rand -hex 32)"
export KFLEET_ADMIN_PASSWORD="$(openssl rand -base64 24)"

helm upgrade --install kfleet-hub \\
  oci://ghcr.io/1solomonwakhungu/charts/kfleet-hub \\
  --version ${version} \\
  --namespace kfleet-system \\
  --create-namespace \\
  --set registration.token="$KFLEET_REGISTRATION_TOKEN" \\
  --set-string auth.bootstrapAdmin.username=admin \\
  --set-string auth.bootstrapAdmin.email=admin@example.com \\
  --set-string auth.bootstrapAdmin.password="$KFLEET_ADMIN_PASSWORD" \\
  --set persistence.enabled=true \\
  --set persistence.size=1Gi`,
          },
          {
            label: '2. Install kfleet-agent in each cluster',
            command: `helm upgrade --install kfleet-agent \\
  oci://ghcr.io/1solomonwakhungu/charts/kfleet-agent \\
  --version ${version} \\
  --kube-context my-cluster \\
  --namespace kfleet-system \\
  --create-namespace \\
  --set hub.url=https://kfleet.example.com \\
  --set hub.token="$KFLEET_REGISTRATION_TOKEN" \\
  --set cluster.name=my-cluster`,
            note: 'Sign in to the hub UI and approve the pending agent on the Agents page.',
          },
        ],
      },
    ],
  },

  tradeoffs: {
    eyebrow: 'Honest scope',
    title: 'When to pick something else',
    subtitle:
      'kfleet is intentionally narrow. Use the right tool when you need mutations or active-active HA.',
    items: [
      {
        scenario: 'Cluster provisioning, GitOps, or workload deployment',
        recommendation: 'Use Rancher, ArgoCD, Flux, or OCM.',
        reason: 'kfleet never writes to Kubernetes clusters.',
      },
      {
        scenario: 'Interactive debugging (exec, port-forward, editing YAML)',
        recommendation: 'Use k9s, Lens/Freelens, Headlamp, or kubectl.',
        reason:
          'kfleet focuses on fleet visibility, drift checks, and log inspection rather than interactive cluster mutation.',
      },
      {
        scenario: 'Multi-replica HA control plane',
        recommendation: 'Use a clustered management plane with external storage.',
        reason:
          'The hub is a single replica on SQLite. A hub outage interrupts visibility (never your workloads), but it is not active-active HA.',
      },
      {
        scenario: 'Broad CRD or custom resource inventory',
        recommendation: 'Use a full resource indexer or cluster console.',
        reason:
          'The agent collects 6 core resource types (pods, deployments, services, nodes, namespaces, events) plus pods/log to keep RBAC minimal.',
      },
    ],
  },

  faq: {
    id: 'faq',
    eyebrow: 'FAQ',
    title: 'Engineering questions, answered directly',
    subtitle:
      'Security boundaries, storage trade-offs, NAT log streaming, and production deployment details.',
    items: [
      {
        id: 'why-no-kubeconfigs',
        question: 'Why does kfleet never store kubeconfigs?',
        answer:
          'Every cluster runs an outbound-only agent that initiates HTTPS and WebSocket connections to the hub after operator or admin approval. Because the hub never dials into Kubernetes API servers, it holds zero kubeconfigs, service account tokens, or cluster certificates — and private or edge clusters never need inbound firewall rules.',
      },
      {
        id: 'can-it-mutate-or-read-secrets',
        question: 'Can kfleet or an AI client modify my clusters or read Secrets?',
        answer:
          'No. The agent ClusterRole grants exactly three read-only rules: list on pods, services, nodes, events, namespaces, and deployments, plus get on pods/log. It cannot read Secrets or ConfigMaps and has no write verbs. MCP clients only talk to the hub through 9 read-only tools as a read_only user (or anonymously against demo mode).',
      },
      {
        id: 'why-sqlite',
        question: 'Why SQLite instead of Postgres or a management cluster?',
        answer:
          'Shipping a single static Go binary (CGO_ENABLED=0) with SQLite in WAL mode eliminates management clusters, CRDs, operators, and external database maintenance while idling around ~20 MB of memory. One file stores latest snapshots, a 90-day append-only operational timeline, policy findings, alert state, and the audit log.',
      },
      {
        id: 'how-pod-logs-work',
        question: 'How does live pod log streaming work for private or NAT clusters?',
        answer:
          'Each approved agent keeps an outbound WebSocket connection open to the hub. When a browser or MCP client requests container logs (including previous=true for crashlooping containers), the hub sends the log request down that reverse channel and streams lines back over Server-Sent Events with a 512-line per-stream backpressure buffer.',
      },
      {
        id: 'how-to-test-mcp',
        question: 'How do I test MCP without setting up a cluster first?',
        answer:
          'Point your MCP client at the hosted read-only demo at https://kfleet.sullylab.de with no username or password using KFLEET_HUB_URL=https://kfleet.sullylab.de kfleet mcp (or claude mcp add). You can also run the hub locally in demo mode with Docker in 10 seconds.',
      },
      {
        id: 'license-and-production',
        question: 'What license is kfleet under, and how is it deployed in production?',
        answer:
          'kfleet is open source under the Apache License 2.0. In production, install the hub once via the kfleet-hub OCI Helm chart (or run the static binary/container on a VM) with SQLite persistence enabled, then deploy kfleet-agent in each cluster using existing Kubernetes Secrets and custom CA bundles as needed.',
      },
    ],
  },

  cta: {
    headline: 'See every cluster. Keep every kubeconfig off laptops.',
    description:
      'Open the live synthetic demo in one click, or run kfleet locally in 10 seconds with Docker or Homebrew.',
    primaryCta: {
      label: 'Try live demo',
      href: liveDemoUrl,
      external: true,
    },
    secondaryCta: {
      label: 'Star on GitHub',
      href: repoUrl,
      external: true,
    },
  },

  footer: {
    tagline:
      'Read-only multi-cluster Kubernetes visibility & MCP diagnostics in a single 14 MB Go + SQLite binary.',
    copyright: 'kfleet · Apache License 2.0 · 2026',
    columns: [
      {
        title: 'Product',
        links: [
          { label: 'Live demo', href: liveDemoUrl, external: true },
          { label: 'Features', href: '#features', external: false },
          { label: 'MCP server', href: '#mcp', external: false },
          { label: 'Quickstart', href: '#quickstart', external: false },
          { label: `Release ${releaseTag}`, href: releaseUrl, external: true },
        ],
      },
      {
        title: 'Documentation',
        links: [
          { label: 'Docs index', href: `${docsBaseUrl}/README.md`, external: true },
          { label: 'Operations runbook', href: `${docsBaseUrl}/operations.md`, external: true },
          { label: 'Pod log streaming', href: `${docsBaseUrl}/pod-logs.md`, external: true },
          { label: 'Policy & drift', href: `${docsBaseUrl}/policy-drift.md`, external: true },
          { label: 'Fleet health alerts', href: `${docsBaseUrl}/alerts.md`, external: true },
          { label: 'REST API', href: `${docsBaseUrl}/api.md`, external: true },
        ],
      },
      {
        title: 'Project',
        links: [
          { label: 'GitHub', href: repoUrl, external: true },
          { label: 'Discussions', href: `${repoUrl}/discussions`, external: true },
          { label: 'Issues', href: `${repoUrl}/issues`, external: true },
          { label: 'Contributing', href: `${repoUrl}/blob/main/CONTRIBUTING.md`, external: true },
          { label: 'Apache 2.0 License', href: `${repoUrl}/blob/main/LICENSE`, external: true },
        ],
      },
    ],
  },
}
