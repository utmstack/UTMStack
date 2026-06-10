import { useMemo, useState } from 'react'
import {
  Activity,
  AlertTriangle,
  Apple,
  Cloud,
  Clock,
  Cog,
  Copy,
  Download,
  ExternalLink,
  ListFilter,
  ListIcon,
  LayoutGrid,
  MoreHorizontal,
  Network,
  Pause,
  Play,
  Plus,
  RefreshCw,
  Search,
  Server,
  Sparkles,
  Terminal,
  Webhook,
  X,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type SourceKind = 'agent' | 'cloud' | 'network' | 'saas'
type SourceStatus = 'connected' | 'degraded' | 'disconnected' | 'paused'
type SubKind =
  | 'windows'
  | 'linux'
  | 'macos'
  | 'aws'
  | 'azure'
  | 'gcp'
  | 'o365'
  | 'okta'
  | 'firewall'
  | 'ids'
  | 'syslog'
  | 'salesforce'

interface DataSource {
  id: string
  name: string
  kind: SourceKind
  sub: SubKind
  status: SourceStatus
  ip?: string
  os?: string
  osVersion?: string
  version?: string
  lastSeen: string
  installedAt: string
  events24h: number
  eventsTotal: number
  rate: number
  tags: string[]
  health?: { cpu: number; memory: number; disk: number }
  lastError?: string
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const NOW = new Date('2026-05-01T16:42:00Z').getTime()
const min = (m: number) => new Date(NOW - m * 60_000).toISOString()
const hr = (h: number) => new Date(NOW - h * 3_600_000).toISOString()
const days = (d: number) => new Date(NOW - d * 86_400_000).toISOString()

const SOURCES: DataSource[] = [
  {
    id: 'ds-win-1',
    name: 'WIN-7F3K2L9Q8X',
    kind: 'agent',
    sub: 'windows',
    status: 'connected',
    ip: '10.4.18.22',
    os: 'Windows Server',
    osVersion: '2022',
    version: '11.2.7',
    lastSeen: min(0),
    installedAt: days(120),
    events24h: 142_004,
    eventsTotal: 18_220_412,
    rate: 142,
    tags: ['Finance', 'critical-asset'],
    health: { cpu: 22, memory: 41, disk: 68 },
  },
  {
    id: 'ds-win-dc',
    name: 'dc-01',
    kind: 'agent',
    sub: 'windows',
    status: 'connected',
    ip: '10.4.18.5',
    os: 'Windows Server',
    osVersion: '2022',
    version: '11.2.7',
    lastSeen: min(0),
    installedAt: days(220),
    events24h: 88_103,
    eventsTotal: 14_880_002,
    rate: 88,
    tags: ['IT', 'domain-controller'],
    health: { cpu: 18, memory: 52, disk: 41 },
  },
  {
    id: 'ds-lin-1',
    name: 'fin-app-01',
    kind: 'agent',
    sub: 'linux',
    status: 'connected',
    ip: '10.0.4.91',
    os: 'Ubuntu',
    osVersion: '22.04 LTS',
    version: '11.2.7',
    lastSeen: min(0),
    installedAt: days(45),
    events24h: 312_000,
    eventsTotal: 28_104_882,
    rate: 248,
    tags: ['Finance'],
    health: { cpu: 38, memory: 64, disk: 72 },
  },
  {
    id: 'ds-lin-2',
    name: 'mail-srv-02',
    kind: 'agent',
    sub: 'linux',
    status: 'connected',
    ip: '10.2.1.18',
    os: 'RHEL',
    osVersion: '9',
    version: '11.2.7',
    lastSeen: min(0),
    installedAt: days(60),
    events24h: 64_120,
    eventsTotal: 9_220_004,
    rate: 64,
    tags: ['IT'],
    health: { cpu: 12, memory: 28, disk: 34 },
  },
  {
    id: 'ds-mac-1',
    name: 'macbook-jsmith',
    kind: 'agent',
    sub: 'macos',
    status: 'degraded',
    ip: '10.6.4.18',
    os: 'macOS',
    osVersion: 'Sonoma 14.4',
    version: '11.2.5',
    lastSeen: min(8),
    installedAt: days(15),
    events24h: 4_812,
    eventsTotal: 142_204,
    rate: 5,
    tags: ['Workstation'],
    health: { cpu: 8, memory: 18, disk: 52 },
    lastError: 'Heartbeat lag 8 minutes — possible network instability',
  },
  {
    id: 'ds-cloud-aws',
    name: 'AWS CloudTrail',
    kind: 'cloud',
    sub: 'aws',
    status: 'connected',
    version: 'v2',
    lastSeen: min(0),
    installedAt: days(180),
    events24h: 481_022,
    eventsTotal: 84_220_044,
    rate: 482,
    tags: ['production'],
  },
  {
    id: 'ds-cloud-azure',
    name: 'Azure AD Sign-in Logs',
    kind: 'cloud',
    sub: 'azure',
    status: 'connected',
    version: 'v3',
    lastSeen: min(0),
    installedAt: days(140),
    events24h: 124_001,
    eventsTotal: 18_204_008,
    rate: 124,
    tags: ['identity'],
  },
  {
    id: 'ds-cloud-o365',
    name: 'Microsoft 365 Audit',
    kind: 'cloud',
    sub: 'o365',
    status: 'connected',
    version: 'v2',
    lastSeen: min(2),
    installedAt: days(140),
    events24h: 218_204,
    eventsTotal: 31_004_888,
    rate: 218,
    tags: ['identity', 'collaboration'],
  },
  {
    id: 'ds-cloud-gcp',
    name: 'GCP Audit Logs',
    kind: 'cloud',
    sub: 'gcp',
    status: 'paused',
    version: 'v1',
    lastSeen: hr(8),
    installedAt: days(78),
    events24h: 0,
    eventsTotal: 4_220_008,
    rate: 0,
    tags: [],
  },
  {
    id: 'ds-network-fw',
    name: 'edge-fw-bsa-01',
    kind: 'network',
    sub: 'firewall',
    status: 'connected',
    ip: '10.0.0.1',
    os: 'pfSense',
    osVersion: '2.7',
    version: 'syslog-fwd-1.4',
    lastSeen: min(0),
    installedAt: days(360),
    events24h: 1_204_882,
    eventsTotal: 380_220_044,
    rate: 1208,
    tags: ['perimeter'],
  },
  {
    id: 'ds-ids',
    name: 'Suricata IDS',
    kind: 'network',
    sub: 'ids',
    status: 'connected',
    ip: '10.0.0.5',
    os: 'Suricata',
    osVersion: '7.0',
    version: 'eve-out',
    lastSeen: min(0),
    installedAt: days(220),
    events24h: 88_201,
    eventsTotal: 12_004_882,
    rate: 88,
    tags: ['perimeter'],
  },
  {
    id: 'ds-syslog',
    name: 'Cisco ASA syslog',
    kind: 'network',
    sub: 'syslog',
    status: 'disconnected',
    ip: '10.0.0.20',
    os: 'Cisco ASA',
    osVersion: '9.18',
    version: 'rsyslog',
    lastSeen: hr(36),
    installedAt: days(420),
    events24h: 0,
    eventsTotal: 220_004_882,
    rate: 0,
    tags: ['perimeter', 'critical'],
    lastError: 'No events received in 36 hours — connector may be stopped',
  },
  {
    id: 'ds-saas-okta',
    name: 'Okta',
    kind: 'saas',
    sub: 'okta',
    status: 'connected',
    version: 'system-log v1',
    lastSeen: min(0),
    installedAt: days(180),
    events24h: 24_882,
    eventsTotal: 4_220_004,
    rate: 25,
    tags: ['identity'],
  },
  {
    id: 'ds-saas-sf',
    name: 'Salesforce Event Monitoring',
    kind: 'saas',
    sub: 'salesforce',
    status: 'connected',
    version: 'event-log-file',
    lastSeen: min(0),
    installedAt: days(95),
    events24h: 12_402,
    eventsTotal: 1_220_044,
    rate: 12,
    tags: ['CRM'],
  },
]

/* ─── Style maps ───────────────────────────────────────────────────────── */

const STATUS_META: Record<SourceStatus, { label: string; dot: string; tone: string }> = {
  connected: { label: 'Connected', dot: 'bg-emerald-500', tone: 'text-emerald-500' },
  degraded: { label: 'Degraded', dot: 'bg-amber-500', tone: 'text-amber-500' },
  disconnected: { label: 'Disconnected', dot: 'bg-red-500 animate-pulse', tone: 'text-red-500' },
  paused: { label: 'Paused', dot: 'bg-muted-foreground', tone: 'text-muted-foreground' },
}

const KIND_META: Record<SourceKind, { label: string; tone: string }> = {
  agent: { label: 'Agent', tone: 'text-sky-500' },
  cloud: { label: 'Cloud', tone: 'text-violet-500' },
  network: { label: 'Network', tone: 'text-fuchsia-500' },
  saas: { label: 'SaaS', tone: 'text-emerald-500' },
}

const SUB_META: Record<SubKind, { label: string; icon: LucideIcon | null; tone: string }> = {
  windows: { label: 'Windows', icon: Server, tone: 'text-sky-500' },
  linux: { label: 'Linux', icon: Terminal, tone: 'text-amber-500' },
  macos: { label: 'macOS', icon: Apple, tone: 'text-violet-500' },
  aws: { label: 'AWS', icon: Cloud, tone: 'text-orange-500' },
  azure: { label: 'Azure', icon: Cloud, tone: 'text-sky-500' },
  gcp: { label: 'GCP', icon: Cloud, tone: 'text-amber-500' },
  o365: { label: 'Microsoft 365', icon: Cloud, tone: 'text-fuchsia-500' },
  okta: { label: 'Okta', icon: Webhook, tone: 'text-sky-500' },
  firewall: { label: 'Firewall', icon: Network, tone: 'text-rose-500' },
  ids: { label: 'IDS/IPS', icon: Network, tone: 'text-violet-500' },
  syslog: { label: 'Syslog', icon: Network, tone: 'text-muted-foreground' },
  salesforce: { label: 'Salesforce', icon: Webhook, tone: 'text-sky-500' },
}

/* ─── Saved views ──────────────────────────────────────────────────────── */

interface SavedView {
  id: string
  label: string
  count: number
  predicate: (s: DataSource) => boolean
}

const SAVED_VIEWS: SavedView[] = [
  { id: 'all', label: 'All', count: SOURCES.length, predicate: () => true },
  { id: 'agents', label: 'Agents', count: SOURCES.filter((s) => s.kind === 'agent').length, predicate: (s) => s.kind === 'agent' },
  { id: 'cloud', label: 'Cloud', count: SOURCES.filter((s) => s.kind === 'cloud').length, predicate: (s) => s.kind === 'cloud' },
  { id: 'network', label: 'Network', count: SOURCES.filter((s) => s.kind === 'network').length, predicate: (s) => s.kind === 'network' },
  { id: 'saas', label: 'SaaS', count: SOURCES.filter((s) => s.kind === 'saas').length, predicate: (s) => s.kind === 'saas' },
  { id: 'issues', label: 'Issues', count: SOURCES.filter((s) => s.status === 'degraded' || s.status === 'disconnected').length, predicate: (s) => s.status === 'degraded' || s.status === 'disconnected' },
]

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function DataSourcesPage() {
  const [view, setView] = useState<string>('all')
  const [search, setSearch] = useState('')
  const [layout, setLayout] = useState<'list' | 'cards'>('list')
  const [openSrc, setOpenSrc] = useState<DataSource | null>(null)

  const filtered = useMemo(() => {
    const v = SAVED_VIEWS.find((s) => s.id === view) ?? SAVED_VIEWS[0]
    return SOURCES.filter(v.predicate).filter((s) =>
      search
        ? (s.name + (s.ip ?? '') + (s.os ?? '') + s.tags.join(' ') + s.sub)
            .toLowerCase()
            .includes(search.toLowerCase())
        : true
    )
  }, [view, search])

  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 py-6">
      <Header total={filtered.length} />

      <div className="mt-5 grid grid-cols-12 gap-4">
        <div className="col-span-12 lg:col-span-8">
          <IngestionOverviewCard sources={SOURCES} />
        </div>
        <div className="col-span-12 lg:col-span-4">
          <HealthInsightsCard />
        </div>
      </div>

      <div className="mt-5">
        <SavedViewTabs current={view} onChange={setView} />
      </div>

      <Toolbar search={search} onSearch={setSearch} layout={layout} onLayoutChange={setLayout} />

      {layout === 'list' ? (
        <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
          <ListHeader />
          {filtered.map((s) => (
            <SourceListRow key={s.id} source={s} onOpen={() => setOpenSrc(s)} />
          ))}
          {filtered.length === 0 && (
            <div className="px-6 py-16 text-center text-sm text-muted-foreground">
              No sources match this view.
            </div>
          )}
        </div>
      ) : (
        <div className="mt-3 grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
          {filtered.map((s) => (
            <SourceCard key={s.id} source={s} onOpen={() => setOpenSrc(s)} />
          ))}
          {filtered.length === 0 && (
            <div className="col-span-full rounded-xl border border-border bg-card px-6 py-16 text-center text-sm text-muted-foreground">
              No sources match this view.
            </div>
          )}
        </div>
      )}

      {openSrc && <SourceDrawer source={openSrc} onClose={() => setOpenSrc(null)} />}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ total }: { total: number }) {
  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold tracking-tight">Data sources</h1>
          <span className="rounded-md bg-muted px-2 py-0.5 font-mono text-[11px] tabular-nums text-muted-foreground">
            {total.toLocaleString()} connected
          </span>
        </div>
        <p className="mt-1 text-xs text-muted-foreground">
          Agents, cloud integrations, network devices, and SaaS apps feeding events.
        </p>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm">
          <Download size={14} className="mr-2" />
          Export
        </Button>
        <Button size="sm">
          <Plus size={14} className="mr-2" />
          Connect source
        </Button>
      </div>
    </header>
  )
}

/* ─── Ingestion overview ───────────────────────────────────────────────── */

function IngestionOverviewCard({ sources }: { sources: DataSource[] }) {
  const totalRate = sources.reduce((s, x) => s + x.rate, 0)
  const total24h = sources.reduce((s, x) => s + x.events24h, 0)
  const connected = sources.filter((s) => s.status === 'connected').length

  // Smooth area chart of total ingestion rate
  const buckets = 60
  const data = useMemo(
    () =>
      Array.from({ length: buckets }, (_, i) =>
        Math.round(2200 + Math.sin(i / 5) * 600 + Math.sin(i / 11) * 200 + (i > 50 ? 400 : 0))
      ),
    []
  )
  const w = 1000
  const h = 100
  const max = Math.max(...data) * 1.15
  const xs = data.map((_, i) => (i * w) / (data.length - 1))
  const ys = data.map((v) => h - (v / max) * h)
  const linePath = data.reduce((acc, _, i) => {
    if (i === 0) return `M ${xs[i]} ${ys[i]}`
    const prevX = xs[i - 1]
    const prevY = ys[i - 1]
    const cx1 = prevX + (xs[i] - prevX) / 2
    const cx2 = xs[i] - (xs[i] - prevX) / 2
    return `${acc} C ${cx1} ${prevY}, ${cx2} ${ys[i]}, ${xs[i]} ${ys[i]}`
  }, '')
  const areaPath = `${linePath} L ${xs[xs.length - 1]} ${h} L ${xs[0]} ${h} Z`

  return (
    <div className="rounded-xl border border-border bg-card p-6">
      <div className="flex items-baseline justify-between">
        <div>
          <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
            Ingestion · last 24 hours
          </div>
          <div className="mt-1 flex items-baseline gap-3">
            <span className="text-3xl font-semibold tabular-nums">
              {formatCount(total24h)}
            </span>
            <span className="text-sm text-muted-foreground">
              events
              <span className="mx-2 text-border">·</span>
              <span className="text-foreground">{totalRate.toLocaleString()} e/s</span> avg
              <span className="mx-2 text-border">·</span>
              <span className="text-foreground">{connected}</span> connected
            </span>
          </div>
        </div>
      </div>

      <svg viewBox={`0 0 ${w} ${h}`} className="mt-4 h-24 w-full" preserveAspectRatio="none">
        <defs>
          <linearGradient id="ingestionGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="rgb(34 211 238)" stopOpacity="0.28" />
            <stop offset="100%" stopColor="rgb(34 211 238)" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d={areaPath} fill="url(#ingestionGrad)" />
        <path
          d={linePath}
          fill="none"
          stroke="rgb(34 211 238)"
          strokeOpacity="0.85"
          strokeWidth="1.75"
          strokeLinejoin="round"
          strokeLinecap="round"
        />
      </svg>

      <div className="mt-2 flex justify-between text-[10px] text-muted-foreground">
        <span>24h ago</span>
        <span>12h ago</span>
        <span>now</span>
      </div>
    </div>
  )
}

/* ─── Health insights ──────────────────────────────────────────────────── */

function HealthInsightsCard() {
  const disconnected = SOURCES.filter((s) => s.status === 'disconnected')
  const degraded = SOURCES.filter((s) => s.status === 'degraded')
  return (
    <div className="flex h-full flex-col gap-3 rounded-xl border border-border bg-card p-5">
      <div className="flex items-center gap-2">
        <Sparkles size={14} className="text-fuchsia-500" />
        <h3 className="text-sm font-semibold">Health insights</h3>
      </div>

      {disconnected.length > 0 && (
        <div className="rounded-lg border border-red-500/30 bg-red-500/5 p-3 text-xs leading-relaxed">
          <span className="font-medium text-foreground">
            {disconnected.length} source{disconnected.length === 1 ? '' : 's'} disconnected
          </span>{' '}
          — {disconnected[0].name}
          {disconnected.length > 1 && ` and ${disconnected.length - 1} more`}.{' '}
          <button className="text-primary hover:underline">Investigate</button>
        </div>
      )}

      {degraded.length > 0 && (
        <div className="rounded-lg border border-amber-500/30 bg-amber-500/5 p-3 text-xs leading-relaxed">
          <span className="font-medium text-foreground">
            {degraded.length} source{degraded.length === 1 ? '' : 's'} degraded
          </span>{' '}
          — possible network or agent issues.{' '}
          <button className="text-primary hover:underline">View</button>
        </div>
      )}

      <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">3 hosts</span> in your env have no agent
        installed but are talking to monitored systems.{' '}
        <button className="text-primary hover:underline">Review</button>
      </div>
    </div>
  )
}

/* ─── Saved view tabs ──────────────────────────────────────────────────── */

function SavedViewTabs({ current, onChange }: { current: string; onChange: (id: string) => void }) {
  return (
    <div className="flex flex-wrap items-center gap-1 border-b border-border">
      {SAVED_VIEWS.map((v) => {
        const active = current === v.id
        return (
          <button
            key={v.id}
            onClick={() => onChange(v.id)}
            className={cn(
              'group relative flex items-center gap-2 px-3 py-2 text-xs transition-colors',
              active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
            )}
          >
            <span>{v.label}</span>
            <span
              className={cn(
                'rounded-md px-1.5 py-0.5 font-mono text-[10px] tabular-nums',
                active ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground'
              )}
            >
              {v.count}
            </span>
            {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
          </button>
        )
      })}
    </div>
  )
}

/* ─── Toolbar ──────────────────────────────────────────────────────────── */

function Toolbar({
  search,
  onSearch,
  layout,
  onLayoutChange,
}: {
  search: string
  onSearch: (s: string) => void
  layout: 'list' | 'cards'
  onLayoutChange: (l: 'list' | 'cards') => void
}) {
  return (
    <div className="mt-3 flex flex-wrap items-center gap-2">
      <div className="relative min-w-[280px] flex-1">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search by name, IP, OS, tag…"
          value={search}
          onChange={(e) => onSearch(e.target.value)}
          className="h-9 pl-9"
        />
      </div>
      <Button variant="outline" size="sm">
        <ListFilter size={14} className="mr-2" />
        Filters
      </Button>
      <div className="ml-auto flex items-center gap-1 rounded-md border border-border p-0.5">
        <button
          onClick={() => onLayoutChange('list')}
          title="List"
          className={cn(
            'flex h-7 w-7 items-center justify-center rounded transition-colors',
            layout === 'list' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'
          )}
        >
          <ListIcon size={13} />
        </button>
        <button
          onClick={() => onLayoutChange('cards')}
          title="Cards"
          className={cn(
            'flex h-7 w-7 items-center justify-center rounded transition-colors',
            layout === 'cards' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'
          )}
        >
          <LayoutGrid size={13} />
        </button>
      </div>
      <button
        title="Refresh"
        className="flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
      >
        <RefreshCw size={14} />
      </button>
    </div>
  )
}

/* ─── List ─────────────────────────────────────────────────────────────── */

const LIST_COLS = '36px 1fr 130px 120px 110px 110px 36px'

function ListHeader() {
  return (
    <div
      className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: LIST_COLS }}
    >
      <div />
      <div>Source</div>
      <div>Type</div>
      <div>Status</div>
      <div className="text-right">Rate</div>
      <div>Last seen</div>
      <div />
    </div>
  )
}

function SourceListRow({ source: s, onOpen }: { source: DataSource; onOpen: () => void }) {
  const st = STATUS_META[s.status]
  const sub = SUB_META[s.sub]
  const SubIcon = sub.icon
  return (
    <div
      onClick={onOpen}
      className="group grid cursor-pointer items-center gap-3 border-b border-border px-4 py-2.5 text-xs hover:bg-muted/40 last:border-b-0"
      style={{ gridTemplateColumns: LIST_COLS }}
    >
      <SourceIconBox sub={s.sub} />
      <div className="min-w-0">
        <div className="truncate text-sm font-medium">{s.name}</div>
        <div className="truncate text-[11px] text-muted-foreground">
          {s.ip && <span className="font-mono">{s.ip}</span>}
          {s.os && (
            <>
              {s.ip && <span> · </span>}
              <span>{s.os} {s.osVersion}</span>
            </>
          )}
          {!s.ip && !s.os && s.version && <span className="font-mono">{s.version}</span>}
          {s.tags.length > 0 && (
            <>
              <span> · </span>
              {s.tags.slice(0, 2).map((t, i) => (
                <span key={t} className="font-mono">
                  {i > 0 && ', '}
                  {t}
                </span>
              ))}
            </>
          )}
        </div>
      </div>
      <div className="flex items-center gap-1.5 text-[11px]">
        {SubIcon && <SubIcon size={11} className={sub.tone} />}
        <span className="text-muted-foreground">{sub.label}</span>
      </div>
      <div>
        <span className="inline-flex items-center gap-1.5 text-[11px] text-muted-foreground">
          <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
          {st.label}
        </span>
      </div>
      <div className="text-right font-mono tabular-nums">
        {s.rate > 0 ? (
          <span>
            {s.rate.toLocaleString()}
            <span className="text-[10px] text-muted-foreground"> e/s</span>
          </span>
        ) : (
          <span className="text-muted-foreground">—</span>
        )}
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">
        {relativeTime(s.lastSeen)}
      </div>
      <div className="flex justify-end opacity-0 group-hover:opacity-100">
        <button
          onClick={(e) => e.stopPropagation()}
          className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-background hover:text-foreground"
        >
          <MoreHorizontal size={14} />
        </button>
      </div>
    </div>
  )
}

function SourceIconBox({ sub }: { sub: SubKind }) {
  const meta = SUB_META[sub]
  const Icon = meta.icon
  return (
    <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-muted">
      {Icon && <Icon size={13} className={meta.tone} />}
    </div>
  )
}

/* ─── Cards ────────────────────────────────────────────────────────────── */

function SourceCard({ source: s, onOpen }: { source: DataSource; onOpen: () => void }) {
  const st = STATUS_META[s.status]
  const sub = SUB_META[s.sub]
  return (
    <button
      onClick={onOpen}
      className="group flex flex-col gap-3 rounded-xl border border-border bg-card p-4 text-left transition-all hover:shadow-md"
    >
      <div className="flex items-start gap-3">
        <SourceIconBox sub={s.sub} />
        <div className="min-w-0 flex-1">
          <div className="truncate text-sm font-semibold">{s.name}</div>
          <div className="truncate text-[11px] text-muted-foreground">
            {sub.label}
            {s.os && ` · ${s.os} ${s.osVersion}`}
          </div>
        </div>
        <span className="inline-flex items-center gap-1.5 text-[11px] text-muted-foreground">
          <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
          {st.label}
        </span>
      </div>

      <div className="grid grid-cols-3 gap-2 border-t border-border pt-3 text-[11px]">
        <CardStat label="Rate" value={s.rate > 0 ? `${s.rate} e/s` : '—'} />
        <CardStat label="24h" value={formatCount(s.events24h)} />
        <CardStat label="Total" value={formatCount(s.eventsTotal)} />
      </div>

      <div className="text-[11px] text-muted-foreground">
        Last seen <span className="font-mono">{relativeTime(s.lastSeen)}</span>
      </div>
    </button>
  )
}

function CardStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border border-border bg-background/40 px-2 py-1.5">
      <div className="text-[9px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className="mt-0.5 font-semibold tabular-nums">{value}</div>
    </div>
  )
}

/* ─── Drawer ───────────────────────────────────────────────────────────── */

type Tab = 'overview' | 'health' | 'config'

function SourceDrawer({ source: s, onClose }: { source: DataSource; onClose: () => void }) {
  const [tab, setTab] = useState<Tab>('overview')
  const st = STATUS_META[s.status]
  const sub = SUB_META[s.sub]
  const SubIcon = sub.icon

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[920px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="flex min-w-0 flex-1 items-start gap-3">
              <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-muted">
                {SubIcon && <SubIcon size={20} className={sub.tone} />}
              </div>
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                  <span>{KIND_META[s.kind].label}</span>
                  <span>·</span>
                  <span>{sub.label}</span>
                  <span>·</span>
                  <span>installed {relativeTime(s.installedAt)}</span>
                </div>
                <h2 className="mt-0.5 truncate text-xl font-semibold">{s.name}</h2>
                {s.ip && (
                  <div className="truncate font-mono text-xs text-muted-foreground">{s.ip}</div>
                )}
                <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                  <span className="inline-flex items-center gap-1.5 text-muted-foreground">
                    <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
                    {st.label}
                  </span>
                  {s.version && (
                    <>
                      <span className="text-border">·</span>
                      <span className="font-mono">v{s.version}</span>
                    </>
                  )}
                  {s.tags.map((t) => (
                    <span key={t} className="rounded-md bg-muted px-1.5 py-0.5">{t}</span>
                  ))}
                </div>
              </div>
            </div>
            <button
              onClick={onClose}
              className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <X size={16} />
            </button>
          </div>

          <div className="mt-4 flex flex-wrap items-center gap-2">
            {s.status === 'paused' || s.status === 'disconnected' ? (
              <Button size="sm">
                <Play size={13} className="mr-1.5" />
                Resume
              </Button>
            ) : (
              <Button size="sm" variant="outline">
                <Pause size={13} className="mr-1.5" />
                Pause
              </Button>
            )}
            <Button size="sm" variant="outline">
              <RefreshCw size={13} className="mr-1.5" />
              Reconnect
            </Button>
            <Button size="sm" variant="outline">
              <Activity size={13} className="mr-1.5" />
              View events
            </Button>
            <Button size="sm" variant="ghost" className="ml-auto">
              <MoreHorizontal size={13} />
            </Button>
          </div>
        </header>

        <nav className="flex items-center gap-1 border-b border-border px-6">
          <DrawerTab id="overview" current={tab} onChange={setTab} icon={Sparkles}>
            Overview
          </DrawerTab>
          <DrawerTab id="health" current={tab} onChange={setTab} icon={Activity}>
            Health
          </DrawerTab>
          <DrawerTab id="config" current={tab} onChange={setTab} icon={Cog}>
            Configuration
          </DrawerTab>
        </nav>

        <div className="flex-1 overflow-y-auto bg-muted/20">
          {tab === 'overview' && <OverviewTab source={s} />}
          {tab === 'health' && <HealthTab source={s} />}
          {tab === 'config' && <ConfigTab source={s} />}
        </div>
      </div>
    </div>
  )
}

function DrawerTab({
  id,
  current,
  onChange,
  icon: Icon,
  children,
}: {
  id: Tab
  current: Tab
  onChange: (t: Tab) => void
  icon: LucideIcon
  children: React.ReactNode
}) {
  const active = id === current
  return (
    <button
      onClick={() => onChange(id)}
      className={cn(
        'relative flex items-center gap-1.5 px-3 py-2.5 text-xs transition-colors',
        active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
      )}
    >
      <Icon size={13} />
      {children}
      {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
    </button>
  )
}

/* ─── Drawer: Overview ─────────────────────────────────────────────────── */

function OverviewTab({ source: s }: { source: DataSource }) {
  return (
    <div className="grid grid-cols-3 gap-4 p-6">
      <div className="col-span-2 space-y-4">
        {s.lastError && (
          <div className="rounded-lg border border-red-500/30 bg-red-500/5 p-3 text-xs">
            <div className="flex items-center gap-2 font-medium text-red-600 dark:text-red-300">
              <AlertTriangle size={13} />
              Issue detected
            </div>
            <p className="mt-1 text-muted-foreground">{s.lastError}</p>
          </div>
        )}

        <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-4">
          <div className="mb-2 flex items-center gap-2 text-xs font-medium">
            <Sparkles size={14} className="text-fuchsia-500" />
            AI brief
          </div>
          <p className="text-sm leading-relaxed">
            {s.kind === 'agent'
              ? `${s.name} is feeding ${formatCount(s.events24h)} events in the last 24h at ${s.rate} e/s. ${s.status === 'connected' ? 'Healthy and within baseline.' : 'Activity has degraded — see Health tab.'}`
              : s.kind === 'cloud'
                ? `${s.name} integration pulled ${formatCount(s.events24h)} events in the last 24h. ${s.status === 'paused' ? 'Currently paused.' : 'Operating normally.'}`
                : `${s.name} forwards ${formatCount(s.events24h)} events / 24h. ${s.status === 'disconnected' ? 'No traffic in 36h — check the connector.' : 'Healthy.'}`}
          </p>
        </div>

        <Section title="Ingestion · last 24 hours">
          <IngestionMiniChart sub={s.sub} />
        </Section>

        <Section title="Recent events">
          <p className="text-xs text-muted-foreground">
            <span className="font-mono text-foreground">{s.events24h.toLocaleString()}</span> events
            in the last 24h.{' '}
            <button className="text-primary hover:underline">Open in log explorer</button>
          </p>
        </Section>
      </div>

      <aside className="col-span-1 space-y-4">
        <KeyValueCard
          title="Stats"
          rows={[
            ['Rate', s.rate > 0 ? `${s.rate.toLocaleString()} events/sec` : '—'],
            ['24h events', s.events24h.toLocaleString()],
            ['Total events', s.eventsTotal.toLocaleString()],
            ['Installed', relativeTime(s.installedAt)],
            ['Last seen', relativeTime(s.lastSeen)],
          ]}
        />

        <KeyValueCard
          title="Identification"
          rows={[
            ['Name', s.name],
            ['ID', <span className="font-mono break-all text-[11px]">{s.id}</span>],
            ['Kind', KIND_META[s.kind].label],
            ['Subtype', SUB_META[s.sub].label],
            ['Version', s.version ? `v${s.version}` : '—'],
            ...(s.os ? [['OS', `${s.os} ${s.osVersion ?? ''}`] as [string, React.ReactNode]] : []),
            ...(s.ip ? [['IP', <span className="font-mono">{s.ip}</span>] as [string, React.ReactNode]] : []),
          ]}
        />
      </aside>
    </div>
  )
}

function IngestionMiniChart({ sub: _sub }: { sub: SubKind }) {
  const buckets = 48
  const data = useMemo(
    () =>
      Array.from({ length: buckets }, (_, i) =>
        Math.round(80 + Math.sin(i / 5) * 20 + Math.sin(i / 9) * 8 + (i > 40 ? 24 : 0))
      ),
    []
  )
  const w = 600
  const h = 80
  const max = Math.max(...data) * 1.15
  const xs = data.map((_, i) => (i * w) / (data.length - 1))
  const ys = data.map((v) => h - (v / max) * h)
  const linePath = data.reduce((acc, _, i) => {
    if (i === 0) return `M ${xs[i]} ${ys[i]}`
    const prevX = xs[i - 1]
    const prevY = ys[i - 1]
    const cx1 = prevX + (xs[i] - prevX) / 2
    const cx2 = xs[i] - (xs[i] - prevX) / 2
    return `${acc} C ${cx1} ${prevY}, ${cx2} ${ys[i]}, ${xs[i]} ${ys[i]}`
  }, '')
  const areaPath = `${linePath} L ${xs[xs.length - 1]} ${h} L ${xs[0]} ${h} Z`
  return (
    <svg viewBox={`0 0 ${w} ${h}`} className="h-20 w-full" preserveAspectRatio="none">
      <defs>
        <linearGradient id="dsMiniGrad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="rgb(34 211 238)" stopOpacity="0.28" />
          <stop offset="100%" stopColor="rgb(34 211 238)" stopOpacity="0" />
        </linearGradient>
      </defs>
      <path d={areaPath} fill="url(#dsMiniGrad)" />
      <path
        d={linePath}
        fill="none"
        stroke="rgb(34 211 238)"
        strokeOpacity="0.85"
        strokeWidth="1.5"
        strokeLinejoin="round"
        strokeLinecap="round"
      />
    </svg>
  )
}

/* ─── Drawer: Health ───────────────────────────────────────────────────── */

function HealthTab({ source: s }: { source: DataSource }) {
  if (!s.health && s.kind !== 'agent') {
    return (
      <div className="p-6">
        <Section title="Connector health">
          <div className="grid grid-cols-3 gap-3">
            <Stat label="Auth" value="OK" tone="text-emerald-500" />
            <Stat label="Last sync" value={relativeTime(s.lastSeen)} />
            <Stat label="Errors (24h)" value="0" />
          </div>
        </Section>
      </div>
    )
  }
  if (!s.health) {
    return <div className="p-6 text-xs text-muted-foreground">No health metrics for this source.</div>
  }
  return (
    <div className="space-y-4 p-6">
      <Section title="Resource utilization (host)">
        <div className="space-y-3">
          <Bar label="CPU" value={s.health.cpu} />
          <Bar label="Memory" value={s.health.memory} />
          <Bar label="Disk" value={s.health.disk} />
        </div>
      </Section>
      <Section title="Heartbeat">
        <div className="grid grid-cols-3 gap-3">
          <Stat label="Last heartbeat" value={relativeTime(s.lastSeen)} />
          <Stat label="Interval" value="60s" />
          <Stat label="Missed (24h)" value={s.status === 'degraded' ? '4' : '0'} tone={s.status === 'degraded' ? 'text-amber-500' : 'text-emerald-500'} />
        </div>
      </Section>
    </div>
  )
}

function Bar({ label, value }: { label: string; value: number }) {
  const tone =
    value >= 85 ? 'bg-red-500' : value >= 70 ? 'bg-amber-500' : 'bg-emerald-500'
  return (
    <div>
      <div className="mb-1 flex items-center justify-between text-[11px]">
        <span>{label}</span>
        <span className="font-mono tabular-nums">{value}%</span>
      </div>
      <div className="h-1.5 overflow-hidden rounded-full bg-muted">
        <div className={cn('h-full rounded-full', tone)} style={{ width: `${value}%` }} />
      </div>
    </div>
  )
}

function Stat({ label, value, tone }: { label: string; value: string; tone?: string }) {
  return (
    <div className="rounded-md border border-border bg-background/40 px-3 py-2">
      <div className="text-[10px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className={cn('mt-0.5 font-mono tabular-nums', tone)}>{value}</div>
    </div>
  )
}

/* ─── Drawer: Config ───────────────────────────────────────────────────── */

function ConfigTab({ source: s }: { source: DataSource }) {
  return (
    <div className="space-y-4 p-6">
      {s.kind === 'agent' && (
        <Section title="Agent install command">
          <pre className="overflow-x-auto rounded-md bg-muted/40 p-3 font-mono text-[11px] leading-relaxed">{`utmstack-agent install \\
  --token=eyJ...workspace_id=acme... \\
  --tags="${s.tags.join(',')}" \\
  --auto-update`}</pre>
          <div className="mt-2 flex items-center gap-2">
            <button className="inline-flex items-center gap-1 text-[11px] text-primary hover:underline">
              <Copy size={11} /> Copy command
            </button>
            <button className="inline-flex items-center gap-1 text-[11px] text-primary hover:underline">
              <ExternalLink size={11} /> View install guide
            </button>
          </div>
        </Section>
      )}

      {s.kind === 'cloud' && (
        <Section title="Connection">
          <dl className="grid grid-cols-[160px_1fr] gap-y-2 text-xs">
            <Row k="Auth method"><span>OAuth 2.0 · Service principal</span></Row>
            <Row k="Region"><span className="font-mono">us-east-1</span></Row>
            <Row k="Pull frequency"><span>every 5 minutes</span></Row>
            <Row k="Last token rotation"><span className="font-mono">{relativeTime(days(45))}</span></Row>
          </dl>
        </Section>
      )}

      {s.kind === 'network' && (
        <Section title="Connection">
          <dl className="grid grid-cols-[160px_1fr] gap-y-2 text-xs">
            <Row k="Listen port"><span className="font-mono">514/udp</span></Row>
            <Row k="Format"><span className="font-mono">RFC 5424</span></Row>
            <Row k="Allowed IPs"><span className="font-mono">{s.ip}/32</span></Row>
          </dl>
        </Section>
      )}

      <Section title="Tags">
        <div className="flex flex-wrap gap-1.5">
          {s.tags.length > 0 ? (
            s.tags.map((t) => (
              <span key={t} className="rounded-md bg-muted px-2 py-1 text-xs">{t}</span>
            ))
          ) : (
            <span className="text-xs text-muted-foreground italic">No tags assigned.</span>
          )}
          <button className="rounded-md border border-dashed border-border px-2 py-1 text-xs text-muted-foreground hover:bg-muted hover:text-foreground">
            + Add tag
          </button>
        </div>
      </Section>

      <Section title="Danger zone">
        <p className="text-xs text-muted-foreground">
          Disconnecting this source will stop ingestion immediately. Historical events stay searchable.
        </p>
        <button className="mt-3 rounded-md border border-red-500/30 bg-red-500/5 px-3 py-1.5 text-xs font-medium text-red-600 hover:bg-red-500/10 dark:text-red-300">
          Disconnect source
        </button>
      </Section>
    </div>
  )
}

/* ─── Small parts ──────────────────────────────────────────────────────── */

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <h4 className="mb-3 text-sm font-semibold">{title}</h4>
      {children}
    </div>
  )
}

function KeyValueCard({
  title,
  rows,
}: {
  title: string
  rows: [string, React.ReactNode][]
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">{title}</div>
      <dl className="grid grid-cols-[100px_1fr] gap-y-2 text-xs">
        {rows.map(([k, v]) => (
          <Row key={k} k={k}>
            {v}
          </Row>
        ))}
      </dl>
    </div>
  )
}

function Row({ k, children }: { k: string; children: React.ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="break-words">{children}</dd>
    </>
  )
}

/* ─── Helpers ──────────────────────────────────────────────────────────── */

function relativeTime(iso: string) {
  if (!iso) return '—'
  const diff = NOW - new Date(iso).getTime()
  const m = Math.round(diff / 60_000)
  if (m < 1) return 'just now'
  if (m < 60) return `${m}m ago`
  const h = Math.round(m / 60)
  if (h < 24) return `${h}h ago`
  return `${Math.round(h / 24)}d ago`
}

function formatCount(n: number) {
  if (n >= 1_000_000_000) return `${(n / 1_000_000_000).toFixed(1)}B`
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`
  return n.toString()
}

// Suppress potential unused-import warnings
void Clock
