import { Fragment, useMemo, useState } from 'react'
import {
  ArrowRight,
  ChevronDown,
  Clock,
  Crosshair,
  Database,
  Download,
  ExternalLink,
  Eye,
  EyeOff,
  Flame,
  Globe2,
  ListFilter,
  ListIcon,
  MapPin,
  MoreHorizontal,
  Network,
  Radar,
  RefreshCw,
  Search,
  Server,
  Shield,
  ShieldAlert,
  Sparkles,
  Target,
  UserX,
  X,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type ThreatLevel = 'critical' | 'high' | 'medium' | 'low'
type AdversaryStatus = 'active' | 'dormant' | 'contained' | 'historical'
type Origin = 'external' | 'internal'

interface AdversaryAlert {
  id: string
  name: string
  severity: ThreatLevel
  technique: { id: string; name: string }
  ts: string
  echoes: number
}

interface Target {
  ip?: string
  host?: string
  user?: string
  hits: number
}

interface Adversary {
  id: string
  identifier: string
  kind: 'ip' | 'host' | 'user' | 'actor'
  origin: Origin
  status: AdversaryStatus
  threatLevel: ThreatLevel
  riskScore: number
  country?: string
  countryCode?: string
  city?: string
  asn?: number
  aso?: string
  firstSeen: string
  lastSeen: string
  alertsCount: number
  techniques: { id: string; name: string; tactic: string; count: number }[]
  targets: Target[]
  alerts: AdversaryAlert[]
  attribution?: { actor: string; confidence: 'high' | 'medium' | 'low'; source: string }
  tags: string[]
  iocMatches: { feed: string; type: string }[]
  activity: number[]
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const NOW = new Date('2026-05-01T16:42:00Z').getTime()
const min = (m: number) => new Date(NOW - m * 60_000).toISOString()
const hr = (h: number) => new Date(NOW - h * 3_600_000).toISOString()
const days = (d: number) => new Date(NOW - d * 86_400_000).toISOString()

const sineActivity = (seed: number, len = 24) =>
  Array.from({ length: len }, (_, i) =>
    Math.max(0, Math.round(Math.abs(Math.sin((i + seed) / 2.4)) * 10 + (i > seed && i < seed + 3 ? 14 : 0)))
  )

const ADVERSARIES: Adversary[] = [
  {
    id: 'adv-198.51.100.14',
    identifier: '198.51.100.14',
    kind: 'ip',
    origin: 'external',
    status: 'active',
    threatLevel: 'critical',
    riskScore: 96,
    country: 'Russia',
    countryCode: 'RU',
    city: 'Saint Petersburg',
    asn: 12389,
    aso: 'Rostelecom',
    firstSeen: hr(46),
    lastSeen: min(4),
    alertsCount: 23,
    techniques: [
      { id: 'T1558.003', name: 'Kerberoasting', tactic: 'Credential Access', count: 12 },
      { id: 'T1071.001', name: 'Web Protocols', tactic: 'C2', count: 6 },
      { id: 'T1059.001', name: 'PowerShell', tactic: 'Execution', count: 3 },
      { id: 'T1003', name: 'OS Credential Dumping', tactic: 'Credential Access', count: 2 },
    ],
    targets: [
      { host: 'WIN-7F3K2L9Q8X', user: 'svc_backup', hits: 14 },
      { host: 'fin-app-01', hits: 6 },
      { host: 'dc-01', hits: 3 },
    ],
    alerts: [
      { id: 'a-9f04c2', name: 'Suspected Kerberoasting against domain controller', severity: 'critical', technique: { id: 'T1558.003', name: 'Kerberoasting' }, ts: min(4), echoes: 142 },
      { id: 'a-c014f0', name: 'Outbound connection to known C2 IP', severity: 'critical', technique: { id: 'T1071.001', name: 'Web Protocols' }, ts: min(46), echoes: 8 },
      { id: 'a-44a811', name: 'Antivirus signature match — Mimikatz variant', severity: 'high', technique: { id: 'T1003', name: 'OS Credential Dumping' }, ts: min(62), echoes: 1 },
    ],
    attribution: { actor: 'Scattered Spider', confidence: 'high', source: 'AlienVault OTX, Mandiant' },
    tags: ['Scattered Spider', 'campaign-44', 'C2'],
    iocMatches: [
      { feed: 'AlienVault OTX', type: 'C2' },
      { feed: 'Mandiant', type: 'APT' },
      { feed: 'AbuseCH', type: 'malicious' },
    ],
    activity: sineActivity(2),
  },
  {
    id: 'adv-203.0.113.7',
    identifier: '203.0.113.7',
    kind: 'ip',
    origin: 'external',
    status: 'active',
    threatLevel: 'high',
    riskScore: 84,
    country: 'Brazil',
    countryCode: 'BR',
    city: 'São Paulo',
    asn: 27699,
    aso: 'Telefônica',
    firstSeen: hr(2),
    lastSeen: min(23),
    alertsCount: 312,
    techniques: [
      { id: 'T1110', name: 'Brute Force', tactic: 'Credential Access', count: 312 },
    ],
    targets: [
      { host: 'mail-srv-02', user: 'mthomas', hits: 312 },
    ],
    alerts: [
      { id: 'a-7c2d11', name: 'Multiple failed logins followed by success', severity: 'high', technique: { id: 'T1110', name: 'Brute Force' }, ts: min(23), echoes: 312 },
    ],
    tags: ['brute-force', 'external'],
    iocMatches: [{ feed: 'Internal blocklist', type: 'recent abuse' }],
    activity: sineActivity(7),
  },
  {
    id: 'adv-svc_backup',
    identifier: 'svc_backup',
    kind: 'user',
    origin: 'internal',
    status: 'contained',
    threatLevel: 'critical',
    riskScore: 91,
    firstSeen: hr(46),
    lastSeen: min(2),
    alertsCount: 18,
    techniques: [
      { id: 'T1558.003', name: 'Kerberoasting', tactic: 'Credential Access', count: 12 },
      { id: 'T1135', name: 'Network Share Discovery', tactic: 'Discovery', count: 4 },
      { id: 'T1059.001', name: 'PowerShell', tactic: 'Execution', count: 2 },
    ],
    targets: [
      { host: 'dc-01', hits: 8 },
      { host: 'fileshare-01', hits: 6 },
      { host: 'sql-prod-01', hits: 4 },
    ],
    alerts: [
      { id: 'a-3e88a1', name: 'Suspicious SMB share enumeration', severity: 'medium', technique: { id: 'T1135', name: 'Network Share Discovery' }, ts: min(38), echoes: 12 },
      { id: 'a-1ab32e', name: 'PowerShell execution with encoded command', severity: 'high', technique: { id: 'T1059.001', name: 'PowerShell' }, ts: min(11), echoes: 47 },
    ],
    attribution: { actor: 'Compromised account · likely Scattered Spider', confidence: 'high', source: 'Internal correlation' },
    tags: ['compromised', 'campaign-44', 'service-account'],
    iocMatches: [],
    activity: sineActivity(3),
  },
  {
    id: 'adv-45.83.66.12',
    identifier: '45.83.66.12',
    kind: 'ip',
    origin: 'external',
    status: 'active',
    threatLevel: 'high',
    riskScore: 71,
    country: 'Netherlands',
    countryCode: 'NL',
    city: 'Amsterdam',
    asn: 60781,
    aso: 'LeaseWeb',
    firstSeen: hr(8),
    lastSeen: min(74),
    alertsCount: 4,
    techniques: [
      { id: 'T1041', name: 'C2 Exfiltration', tactic: 'Exfiltration', count: 4 },
    ],
    targets: [{ host: 'fin-app-01', hits: 4 }],
    alerts: [
      { id: 'a-22f3cd', name: 'Unusual data transfer volume to external address', severity: 'medium', technique: { id: 'T1041', name: 'C2 Exfiltration' }, ts: min(74), echoes: 2 },
    ],
    tags: ['exfiltration', 'cloud-host'],
    iocMatches: [{ feed: 'Spamhaus', type: 'bulletproof hoster' }],
    activity: sineActivity(15),
  },
  {
    id: 'adv-185.220.101.42',
    identifier: '185.220.101.42',
    kind: 'ip',
    origin: 'external',
    status: 'dormant',
    threatLevel: 'medium',
    riskScore: 58,
    country: 'Germany',
    countryCode: 'DE',
    city: 'Frankfurt',
    asn: 208294,
    aso: 'F3 Netze e.V. (Tor)',
    firstSeen: days(3),
    lastSeen: hr(36),
    alertsCount: 7,
    techniques: [
      { id: 'T1090.003', name: 'Multi-hop Proxy', tactic: 'C2', count: 7 },
    ],
    targets: [{ host: 'webapp-01', hits: 7 }],
    alerts: [],
    tags: ['tor-exit', 'reconnaissance'],
    iocMatches: [{ feed: 'Tor Project', type: 'exit node' }],
    activity: sineActivity(0),
  },
  {
    id: 'adv-tferguson',
    identifier: 'tferguson',
    kind: 'user',
    origin: 'internal',
    status: 'historical',
    threatLevel: 'medium',
    riskScore: 45,
    firstSeen: hr(7),
    lastSeen: hr(7),
    alertsCount: 1,
    techniques: [{ id: 'T1484', name: 'Domain Policy Modification', tactic: 'Persistence', count: 1 }],
    targets: [{ host: 'dc-01', hits: 1 }],
    alerts: [
      { id: 'a-08f3a4', name: 'GPO modification by non-administrator account', severity: 'high', technique: { id: 'T1484', name: 'Domain Policy Modification' }, ts: hr(7), echoes: 1 },
    ],
    tags: ['investigated', 'benign'],
    iocMatches: [],
    activity: Array(24).fill(0).map((_, i) => (i === 17 ? 4 : 0)),
  },
]

/* ─── Style maps ───────────────────────────────────────────────────────── */

const THREAT_STYLE: Record<ThreatLevel, { bar: string; pill: string; label: string; ring: string }> = {
  critical: {
    bar: 'bg-red-500',
    pill: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300',
    label: 'Critical',
    ring: 'ring-red-500/40',
  },
  high: {
    bar: 'bg-orange-500',
    pill: 'bg-orange-500/15 text-orange-600 ring-orange-500/30 dark:text-orange-300',
    label: 'High',
    ring: 'ring-orange-500/40',
  },
  medium: {
    bar: 'bg-amber-500',
    pill: 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300',
    label: 'Medium',
    ring: 'ring-amber-500/40',
  },
  low: {
    bar: 'bg-sky-500',
    pill: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300',
    label: 'Low',
    ring: 'ring-sky-500/40',
  },
}

const STATUS_META: Record<AdversaryStatus, { label: string; pill: string; dot: string }> = {
  active: { label: 'Active now', pill: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300', dot: 'bg-red-500 animate-pulse' },
  dormant: { label: 'Dormant', pill: 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300', dot: 'bg-amber-500' },
  contained: { label: 'Contained', pill: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300', dot: 'bg-sky-500' },
  historical: { label: 'Historical', pill: 'bg-muted text-muted-foreground ring-border', dot: 'bg-muted-foreground' },
}

/* ─── Saved views ──────────────────────────────────────────────────────── */

interface SavedView {
  id: string
  label: string
  count: number
  predicate: (a: Adversary) => boolean
}

const SAVED_VIEWS: SavedView[] = [
  { id: 'all', label: 'All adversaries', count: ADVERSARIES.length, predicate: () => true },
  { id: 'active', label: 'Active now', count: ADVERSARIES.filter((a) => a.status === 'active').length, predicate: (a) => a.status === 'active' },
  { id: 'attributed', label: 'Known actors', count: ADVERSARIES.filter((a) => !!a.attribution).length, predicate: (a) => !!a.attribution },
  { id: 'external', label: 'External', count: ADVERSARIES.filter((a) => a.origin === 'external').length, predicate: (a) => a.origin === 'external' },
  { id: 'internal', label: 'Internal', count: ADVERSARIES.filter((a) => a.origin === 'internal').length, predicate: (a) => a.origin === 'internal' },
  { id: 'critical', label: 'Critical', count: ADVERSARIES.filter((a) => a.threatLevel === 'critical').length, predicate: (a) => a.threatLevel === 'critical' },
]

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function AdversariesPage() {
  const [view, setView] = useState<string>('all')
  const [search, setSearch] = useState('')
  const [layout, setLayout] = useState<'cards' | 'list'>('list')
  const [openAdv, setOpenAdv] = useState<Adversary | null>(null)

  const filtered = useMemo(() => {
    const v = SAVED_VIEWS.find((s) => s.id === view) ?? SAVED_VIEWS[0]
    return ADVERSARIES.filter(v.predicate).filter((a) =>
      search
        ? (a.identifier + a.country + a.aso + a.tags.join(' ') + (a.attribution?.actor ?? '')).toLowerCase().includes(search.toLowerCase())
        : true
    )
  }, [view, search])

  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      <Header total={filtered.length} />

      <div className="mt-5">
        <ActivityOverviewCard adversaries={ADVERSARIES} />
      </div>

      <div className="mt-5">
        <SavedViewTabs current={view} onChange={setView} />
      </div>

      <Toolbar search={search} onSearch={setSearch} layout={layout} onLayoutChange={setLayout} />

      {layout === 'cards' ? (
        <div className="mt-3 grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
          {filtered.map((a) => (
            <AdversaryCard key={a.id} adv={a} onOpen={() => setOpenAdv(a)} />
          ))}
          {filtered.length === 0 && (
            <div className="col-span-full rounded-xl border border-border bg-card px-6 py-16 text-center text-sm text-muted-foreground">
              No adversaries match this view.
            </div>
          )}
        </div>
      ) : (
        <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
          <ListHeader />
          {filtered.map((a) => (
            <AdversaryListRow key={a.id} adv={a} onOpen={() => setOpenAdv(a)} />
          ))}
          {filtered.length === 0 && (
            <div className="px-6 py-16 text-center text-sm text-muted-foreground">
              No adversaries match this view.
            </div>
          )}
        </div>
      )}

      {openAdv && <AdversaryDrawer adv={openAdv} onClose={() => setOpenAdv(null)} />}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ total }: { total: number }) {
  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold tracking-tight">Adversary view</h1>
          <span className="rounded-md bg-muted px-2 py-0.5 font-mono text-[11px] tabular-nums text-muted-foreground">
            {total.toLocaleString()} tracked
          </span>
        </div>
        <p className="mt-1 text-xs text-muted-foreground">
          Every IP, host, or user that has touched your environment hostilely. Pivot from any
          adversary to its TTPs, targets, and timeline.
        </p>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm">
          <Radar size={14} className="mr-2" />
          Threat intel
        </Button>
        <Button variant="default" size="sm">
          <Download size={14} className="mr-2" />
          Export
        </Button>
      </div>
    </header>
  )
}

/* ─── Activity overview ────────────────────────────────────────────────── */

function ActivityOverviewCard({ adversaries }: { adversaries: Adversary[] }) {
  const active = adversaries.filter((a) => a.status === 'active').length
  const countries = new Set(adversaries.filter((a) => a.countryCode && a.origin === 'external').map((a) => a.countryCode)).size
  const tracked = adversaries.length

  // Build a smooth area chart from sum of all adversaries' activity
  const buckets = Math.max(...adversaries.map((a) => a.activity.length), 24)
  const summed = Array.from({ length: buckets }, (_, i) =>
    adversaries.reduce((s, a) => s + (a.activity[i] ?? 0), 0)
  )
  const w = 1000
  const h = 100
  const max = Math.max(...summed) * 1.15
  const xs = summed.map((_, i) => (i * w) / (summed.length - 1))
  const ys = summed.map((v) => h - (v / max) * h)
  const linePath = summed.reduce((acc, _, i) => {
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
          <div className="text-[11px] uppercase tracking-wider text-muted-foreground">Adversary activity · last 24 hours</div>
          <div className="mt-1 flex items-baseline gap-3">
            <span className="text-3xl font-semibold tabular-nums">{active}</span>
            <span className="text-sm text-muted-foreground">
              <span className="text-foreground">active now</span>
              <span className="mx-2 text-border">·</span>
              <span className="text-foreground">{countries}</span> countries
              <span className="mx-2 text-border">·</span>
              <span className="text-foreground">{tracked}</span> tracked total
            </span>
          </div>
        </div>
      </div>

      <svg viewBox={`0 0 ${w} ${h}`} className="mt-4 h-24 w-full" preserveAspectRatio="none">
        <defs>
          <linearGradient id="advGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="rgb(244 63 94)" stopOpacity="0.28" />
            <stop offset="100%" stopColor="rgb(244 63 94)" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d={areaPath} fill="url(#advGrad)" />
        <path
          d={linePath}
          fill="none"
          stroke="rgb(244 63 94)"
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
            {active && (
              <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />
            )}
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
  layout: 'cards' | 'list'
  onLayoutChange: (l: 'cards' | 'list') => void
}) {
  return (
    <div className="mt-3 flex flex-wrap items-center gap-2">
      <div className="relative min-w-[280px] flex-1">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search by IP, host, user, country, ASN, actor…"
          value={search}
          onChange={(e) => onSearch(e.target.value)}
          className="h-9 pl-9 pr-24"
        />
      </div>
      <Button variant="outline" size="sm">
        <Clock size={14} className="mr-2" />
        Last 24 hours
        <ChevronDown size={12} className="ml-1.5 opacity-60" />
      </Button>
      <Button variant="outline" size="sm">
        <ListFilter size={14} className="mr-2" />
        Filters
        <span className="ml-2 rounded-md bg-primary/15 px-1.5 py-0.5 font-mono text-[10px] text-primary">2</span>
      </Button>
      <div className="ml-auto flex items-center gap-1 rounded-md border border-border p-0.5">
        <button
          onClick={() => onLayoutChange('cards')}
          title="Card view"
          className={cn(
            'flex h-7 w-7 items-center justify-center rounded transition-colors',
            layout === 'cards' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'
          )}
        >
          <Database size={13} />
        </button>
        <button
          onClick={() => onLayoutChange('list')}
          title="List view"
          className={cn(
            'flex h-7 w-7 items-center justify-center rounded transition-colors',
            layout === 'list' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'
          )}
        >
          <ListIcon size={13} />
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

/* ─── Adversary card (default view) ────────────────────────────────────── */

function AdversaryCard({ adv, onOpen }: { adv: Adversary; onOpen: () => void }) {
  const tl = THREAT_STYLE[adv.threatLevel]
  const st = STATUS_META[adv.status]
  return (
    <button
      onClick={onOpen}
      className={cn(
        'group flex flex-col gap-3 rounded-xl border bg-card p-4 text-left transition-all hover:shadow-md',
        'border-border'
      )}
    >
      {/* Top row: avatar + identity + status */}
      <div className="flex items-start gap-3">
        <AdversaryAvatar adv={adv} />
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
            <span className="capitalize">{adv.kind}</span>
            <span>·</span>
            <span className="capitalize">{adv.origin}</span>
          </div>
          <div className="mt-0.5 flex items-center gap-2">
            <span className="truncate font-mono text-sm font-semibold">{adv.identifier}</span>
          </div>
          {(adv.country || adv.aso) && (
            <div className="mt-0.5 truncate text-[11px] text-muted-foreground">
              {adv.countryCode && (
                <span className="mr-1.5 rounded-sm bg-muted px-1 py-0.5 font-mono text-[9px] uppercase">
                  {adv.countryCode}
                </span>
              )}
              {adv.country && <span>{adv.country}</span>}
              {adv.aso && <span> · {adv.aso}</span>}
            </div>
          )}
        </div>
        <span className={cn('shrink-0 inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset', st.pill)}>
          <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
          {st.label}
        </span>
      </div>

      {/* Threat level + risk */}
      <div className="flex items-center gap-2">
        <span className={cn('inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset', tl.pill)}>
          <ShieldAlert size={10} />
          {tl.label}
        </span>
        <RiskRing value={adv.riskScore} />
        {adv.attribution && (
          <span className="inline-flex items-center gap-1 rounded-md border border-fuchsia-500/30 bg-fuchsia-500/10 px-1.5 py-0.5 text-[10px] font-medium text-fuchsia-500">
            <UserX size={10} />
            {adv.attribution.actor}
          </span>
        )}
      </div>

      {/* Stats grid */}
      <div className="grid grid-cols-3 gap-2 border-t border-border pt-3 text-[11px]">
        <CardStat label="Alerts" value={adv.alertsCount.toString()} />
        <CardStat label="Techniques" value={adv.techniques.length.toString()} />
        <CardStat label="Targets" value={adv.targets.length.toString()} />
      </div>

      {/* Activity sparkline */}
      <div>
        <div className="mb-1 flex items-center justify-between text-[10px] text-muted-foreground">
          <span>Activity · 24h</span>
          <span className="font-mono">last seen {relativeTime(adv.lastSeen)}</span>
        </div>
        <ActivitySparkline data={adv.activity} tone={tl.bar.replace('bg-', 'fill-')} />
      </div>

      {/* MITRE techniques */}
      <div className="flex flex-wrap gap-1">
        {adv.techniques.slice(0, 3).map((t) => (
          <span
            key={t.id}
            className="inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5 text-[10px]"
            title={`${t.tactic} · ${t.count} hits`}
          >
            <span className="font-mono">{t.id}</span>
            <span className="text-muted-foreground">{t.name}</span>
          </span>
        ))}
        {adv.techniques.length > 3 && (
          <span className="rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
            +{adv.techniques.length - 3}
          </span>
        )}
      </div>
    </button>
  )
}

function AdversaryAvatar({ adv }: { adv: Adversary }) {
  const tl = THREAT_STYLE[adv.threatLevel]
  const Icon =
    adv.kind === 'user'
      ? UserX
      : adv.kind === 'host'
        ? Server
        : adv.kind === 'actor'
          ? ShieldAlert
          : Globe2
  return (
    <div className="relative">
      <div
        className={cn(
          'flex h-11 w-11 shrink-0 items-center justify-center rounded-lg ring-2',
          adv.kind === 'user' ? 'bg-violet-500/15 text-violet-600 dark:text-violet-300' : 'bg-muted text-foreground/80',
          tl.ring
        )}
      >
        <Icon size={18} />
      </div>
      {adv.status === 'active' && (
        <span className="absolute -right-0.5 -top-0.5 h-2.5 w-2.5 rounded-full bg-red-500 ring-2 ring-card animate-pulse" />
      )}
    </div>
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

function ActivitySparkline({ data, tone }: { data: number[]; tone: string }) {
  const w = 240
  const h = 28
  const max = Math.max(...data, 1)
  const bw = w / data.length
  return (
    <svg viewBox={`0 0 ${w} ${h}`} className="h-7 w-full" preserveAspectRatio="none">
      {data.map((v, i) => {
        const bh = (v / max) * (h - 2)
        return <rect key={i} x={i * bw + 0.5} y={h - bh} width={bw - 1} height={bh} rx={0.5} className={cn(tone, 'opacity-80')} />
      })}
    </svg>
  )
}

function RiskRing({ value }: { value: number }) {
  const tone =
    value >= 90
      ? 'stroke-red-500'
      : value >= 70
        ? 'stroke-orange-500'
        : value >= 50
          ? 'stroke-amber-500'
          : 'stroke-emerald-500'
  const r = 9
  const c = 2 * Math.PI * r
  const off = c - (value / 100) * c
  return (
    <div className="relative inline-flex h-6 w-6 items-center justify-center">
      <svg viewBox="0 0 24 24" className="h-6 w-6 -rotate-90">
        <circle cx="12" cy="12" r={r} className="fill-none stroke-muted" strokeWidth="2.5" />
        <circle
          cx="12"
          cy="12"
          r={r}
          className={cn('fill-none', tone)}
          strokeWidth="2.5"
          strokeDasharray={c}
          strokeDashoffset={off}
          strokeLinecap="round"
        />
      </svg>
      <span className="absolute inset-0 flex items-center justify-center font-mono text-[9px] font-semibold tabular-nums">
        {value}
      </span>
    </div>
  )
}

/* ─── List view ────────────────────────────────────────────────────────── */

const LIST_COLS = '40px 1fr 100px 120px 70px 70px 100px 110px 36px'

function ListHeader() {
  return (
    <div
      className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: LIST_COLS }}
    >
      <div />
      <div>Adversary</div>
      <div>Threat</div>
      <div>Origin</div>
      <div className="text-right">Risk</div>
      <div className="text-right">Alerts</div>
      <div>Status</div>
      <div>Last seen</div>
      <div />
    </div>
  )
}

function AdversaryListRow({ adv, onOpen }: { adv: Adversary; onOpen: () => void }) {
  const tl = THREAT_STYLE[adv.threatLevel]
  const st = STATUS_META[adv.status]
  return (
    <div
      onClick={onOpen}
      className="group grid cursor-pointer items-center gap-3 border-b border-border px-4 py-2.5 text-xs hover:bg-muted/40 last:border-b-0"
      style={{ gridTemplateColumns: LIST_COLS }}
    >
      <AdversaryAvatar adv={adv} />
      <div className="min-w-0">
        <div className="truncate font-mono text-sm font-medium">{adv.identifier}</div>
        <div className="truncate text-[11px] text-muted-foreground">
          {adv.countryCode && (
            <span className="mr-1 rounded-sm bg-muted px-1 py-0.5 font-mono text-[9px] uppercase">
              {adv.countryCode}
            </span>
          )}
          {adv.country ?? <span className="capitalize">{adv.kind} · {adv.origin}</span>}
          {adv.aso && <span> · {adv.aso}</span>}
          {adv.attribution && (
            <span className="ml-2 text-fuchsia-500">→ {adv.attribution.actor}</span>
          )}
        </div>
      </div>
      <div>
        <span className={cn('rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset', tl.pill)}>
          {tl.label}
        </span>
      </div>
      <div className="text-[11px] capitalize text-muted-foreground">
        <span className="inline-flex items-center gap-1">
          {adv.origin === 'external' ? <Globe2 size={11} /> : <Network size={11} />}
          {adv.origin}
        </span>
      </div>
      <div className="flex justify-end">
        <RiskRing value={adv.riskScore} />
      </div>
      <div className="text-right tabular-nums">{adv.alertsCount.toLocaleString()}</div>
      <div>
        <span className={cn('inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset', st.pill)}>
          <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
          {st.label}
        </span>
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">{relativeTime(adv.lastSeen)}</div>
      <div className="flex justify-end opacity-0 group-hover:opacity-100">
        <button onClick={(e) => e.stopPropagation()} className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-background hover:text-foreground">
          <MoreHorizontal size={14} />
        </button>
      </div>
    </div>
  )
}

/* ─── Drawer ───────────────────────────────────────────────────────────── */

type Tab = 'profile' | 'attack' | 'targets' | 'timeline'

function AdversaryDrawer({ adv, onClose }: { adv: Adversary; onClose: () => void }) {
  const [tab, setTab] = useState<Tab>('profile')
  const tl = THREAT_STYLE[adv.threatLevel]
  const st = STATUS_META[adv.status]

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[1000px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="flex min-w-0 flex-1 items-start gap-3">
              <AdversaryAvatar adv={adv} />
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                  <span className="capitalize">{adv.kind}</span>
                  <span>·</span>
                  <span className="capitalize">{adv.origin}</span>
                  <span>·</span>
                  <span>first seen {relativeTime(adv.firstSeen)}</span>
                </div>
                <h2 className="mt-0.5 truncate font-mono text-xl font-semibold">{adv.identifier}</h2>
                <div className="mt-1.5 flex flex-wrap items-center gap-2 text-[11px]">
                  <span className={cn('rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset', tl.pill)}>
                    {tl.label}
                  </span>
                  <span className={cn('inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset', st.pill)}>
                    <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
                    {st.label}
                  </span>
                  {adv.attribution && (
                    <span className="inline-flex items-center gap-1 rounded-md border border-fuchsia-500/30 bg-fuchsia-500/10 px-1.5 py-0.5 font-medium text-fuchsia-500">
                      <UserX size={10} />
                      {adv.attribution.actor}
                      <span className="text-fuchsia-500/70">· {adv.attribution.confidence}</span>
                    </span>
                  )}
                  {adv.country && (
                    <span className="inline-flex items-center gap-1 rounded-md border border-border bg-background/40 px-1.5 py-0.5">
                      <MapPin size={10} />
                      {adv.country}
                    </span>
                  )}
                  {adv.tags.map((t) => (
                    <span key={t} className="rounded-md bg-muted px-1.5 py-0.5">
                      {t}
                    </span>
                  ))}
                </div>
              </div>
              <RiskRing value={adv.riskScore} />
            </div>
            <button
              onClick={onClose}
              className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <X size={16} />
            </button>
          </div>

          <div className="mt-4 flex flex-wrap items-center gap-2">
            <Button size="sm">
              <Shield size={14} className="mr-1.5" />
              Block at perimeter
            </Button>
            <Button size="sm" variant="outline">
              <Eye size={14} className="mr-1.5" />
              Watch
            </Button>
            <Button size="sm" variant="outline">
              <EyeOff size={14} className="mr-1.5" />
              Suppress
            </Button>
            <Button size="sm" variant="outline">
              <Flame size={14} className="mr-1.5 text-amber-500" />
              Open incident
            </Button>
            <Button size="sm" variant="outline">
              <ExternalLink size={14} className="mr-1.5" />
              Lookup IOC
            </Button>
            <Button size="sm" variant="ghost" className="ml-auto">
              <MoreHorizontal size={14} />
            </Button>
          </div>
        </header>

        <nav className="flex items-center gap-1 border-b border-border px-6">
          <DrawerTab id="profile" current={tab} onChange={setTab} icon={UserX}>
            Profile
          </DrawerTab>
          <DrawerTab id="attack" current={tab} onChange={setTab} icon={Crosshair}>
            Attack chain
          </DrawerTab>
          <DrawerTab id="targets" current={tab} onChange={setTab} icon={Target}>
            Targets
            <span className="ml-1.5 rounded bg-muted px-1.5 py-0.5 font-mono text-[10px]">
              {adv.targets.length}
            </span>
          </DrawerTab>
          <DrawerTab id="timeline" current={tab} onChange={setTab} icon={Clock}>
            Timeline
          </DrawerTab>
        </nav>

        <div className="flex-1 overflow-y-auto bg-muted/20">
          {tab === 'profile' && <ProfileTab adv={adv} />}
          {tab === 'attack' && <AttackTab adv={adv} />}
          {tab === 'targets' && <TargetsTab adv={adv} />}
          {tab === 'timeline' && <TimelineTab adv={adv} />}
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

/* ─── Drawer: Profile ──────────────────────────────────────────────────── */

function ProfileTab({ adv }: { adv: Adversary }) {
  return (
    <div className="grid grid-cols-3 gap-4 p-6">
      <div className="col-span-2 space-y-4">
        <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-4">
          <div className="mb-2 flex items-center gap-2 text-xs font-medium">
            <Sparkles size={14} className="text-fuchsia-500" />
            Behavioral fingerprint
          </div>
          <p className="text-sm leading-relaxed">
            This adversary's TTP profile shows{' '}
            <strong>credential extraction → C2 over web protocols</strong> with high consistency. Time
            between actions averages <span className="font-mono">42s</span> (faster than 87% of
            tracked adversaries). Activity bursts cluster around{' '}
            <span className="font-mono">15:00–17:00 UTC</span>.
          </p>
          <div className="mt-3 flex flex-wrap gap-2 text-[11px]">
            <button className="rounded-md border border-border bg-background/60 px-2 py-1 hover:bg-background">
              Find similar adversaries
            </button>
            <button className="rounded-md border border-border bg-background/60 px-2 py-1 hover:bg-background">
              Predict next target
            </button>
          </div>
        </div>

        <Section title="Activity over 24 hours">
          <ActivityChart data={adv.activity} />
        </Section>

        <Section title="Top techniques used">
          <ul className="space-y-2.5">
            {adv.techniques.map((t) => {
              const max = Math.max(...adv.techniques.map((x) => x.count))
              return (
                <li key={t.id} className="grid grid-cols-[1fr_auto] items-center gap-3 text-xs">
                  <div>
                    <div className="flex items-baseline gap-2">
                      <span className="font-mono text-[10px] text-muted-foreground">{t.id}</span>
                      <span className="font-medium">{t.name}</span>
                      <span className="text-[10px] text-muted-foreground">{t.tactic}</span>
                    </div>
                    <div className="mt-1 h-1.5 overflow-hidden rounded-full bg-muted">
                      <div
                        className="h-full rounded-full bg-gradient-to-r from-fuchsia-500 to-rose-500"
                        style={{ width: `${(t.count / max) * 100}%` }}
                      />
                    </div>
                  </div>
                  <span className="font-mono tabular-nums">{t.count}</span>
                </li>
              )
            })}
          </ul>
        </Section>
      </div>

      <aside className="col-span-1 space-y-4">
        <KeyValueCard
          title="Identification"
          rows={[
            ['Identifier', <span className="font-mono">{adv.identifier}</span>],
            ['Kind', <span className="capitalize">{adv.kind}</span>],
            ['Origin', <span className="capitalize">{adv.origin}</span>],
            ['First seen', absTimestamp(adv.firstSeen)],
            ['Last seen', absTimestamp(adv.lastSeen)],
          ]}
        />
        {(adv.country || adv.asn) && (
          <KeyValueCard
            title="Geolocation"
            rows={[
              ...(adv.country ? [['Country', `${adv.countryCode} · ${adv.country}`] as [string, React.ReactNode]] : []),
              ...(adv.city ? [['City', adv.city] as [string, React.ReactNode]] : []),
              ...(adv.asn ? [['ASN', <span className="font-mono">AS{adv.asn}</span>] as [string, React.ReactNode]] : []),
              ...(adv.aso ? [['Provider', adv.aso] as [string, React.ReactNode]] : []),
            ]}
          />
        )}
        <div className="rounded-lg border border-border bg-card p-4">
          <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">Attribution</div>
          {adv.attribution ? (
            <div className="space-y-1.5 text-xs">
              <div className="flex items-center gap-2">
                <UserX size={13} className="text-fuchsia-500" />
                <span className="font-medium">{adv.attribution.actor}</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-muted-foreground">Confidence:</span>
                <span
                  className={cn(
                    'rounded px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset',
                    adv.attribution.confidence === 'high'
                      ? 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300'
                      : adv.attribution.confidence === 'medium'
                        ? 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300'
                        : 'bg-muted text-muted-foreground ring-border'
                  )}
                >
                  {adv.attribution.confidence}
                </span>
              </div>
              <div className="text-muted-foreground">Source: {adv.attribution.source}</div>
            </div>
          ) : (
            <div className="text-xs italic text-muted-foreground">
              No known actor attribution. Run pattern matcher to compare against actor TTP database.
            </div>
          )}
        </div>
      </aside>
    </div>
  )
}

function ActivityChart({ data }: { data: number[] }) {
  const w = 600
  const h = 100
  const max = Math.max(...data, 1)
  const bw = w / data.length
  return (
    <svg viewBox={`0 0 ${w} ${h}`} className="h-24 w-full">
      {[0.5, 1].map((p) => (
        <line
          key={p}
          x1="0"
          x2={w}
          y1={h - h * p * 0.92 - 4}
          y2={h - h * p * 0.92 - 4}
          className="stroke-border/40"
          strokeDasharray="2 4"
        />
      ))}
      {data.map((v, i) => {
        const bh = (v / max) * (h - 16)
        return (
          <rect
            key={i}
            x={i * bw + 1}
            y={h - bh - 12}
            width={bw - 2}
            height={bh}
            rx={1}
            className={v > max * 0.7 ? 'fill-red-500/80' : 'fill-violet-500/60'}
          />
        )
      })}
      <g className="fill-muted-foreground" fontSize="9">
        {[0, 6, 12, 18, 23].map((i) => (
          <text key={i} x={i * bw + bw / 2} y={h - 2} textAnchor="middle">
            {`${String((i - 23 + 24) % 24).padStart(2, '0')}h`}
          </text>
        ))}
      </g>
    </svg>
  )
}

/* ─── Drawer: Attack chain ─────────────────────────────────────────────── */

const TACTICS = [
  'Reconnaissance',
  'Initial Access',
  'Execution',
  'Persistence',
  'Defense Evasion',
  'Credential Access',
  'Discovery',
  'Lateral Movement',
  'C2',
  'Exfiltration',
  'Impact',
]

function AttackTab({ adv }: { adv: Adversary }) {
  const used = new Set(adv.techniques.map((t) => t.tactic))
  return (
    <div className="space-y-4 p-6">
      <div className="rounded-lg border border-border bg-card p-4">
        <h4 className="mb-3 text-sm font-semibold">MITRE ATT&CK kill chain</h4>
        <p className="mb-4 text-xs text-muted-foreground">
          Tactics this adversary has progressed through. Highlighted cells show techniques observed.
        </p>
        <div className="grid grid-cols-1 gap-1 md:grid-cols-3 lg:grid-cols-4">
          {TACTICS.map((tactic) => {
            const techs = adv.techniques.filter((t) => t.tactic === tactic)
            const active = used.has(tactic)
            return (
              <div
                key={tactic}
                className={cn(
                  'rounded-md border p-2',
                  active ? 'border-rose-500/40 bg-rose-500/5' : 'border-border bg-background/30'
                )}
              >
                <div className="flex items-center justify-between text-[10px] uppercase tracking-wider">
                  <span className={active ? 'font-semibold text-rose-500' : 'text-muted-foreground'}>
                    {tactic}
                  </span>
                  {active && <span className="font-mono text-rose-500">{techs.length}</span>}
                </div>
                {active ? (
                  <ul className="mt-1.5 space-y-1">
                    {techs.map((t) => (
                      <li key={t.id} className="rounded bg-card px-1.5 py-1 text-[10px]">
                        <span className="font-mono text-rose-500">{t.id}</span>{' '}
                        <span className="text-foreground/80">{t.name}</span>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <div className="mt-1.5 text-[10px] text-muted-foreground italic">no activity</div>
                )}
              </div>
            )
          })}
        </div>
      </div>

      <div className="rounded-lg border border-border bg-card p-4">
        <h4 className="mb-3 text-sm font-semibold">Technique chain (chronological)</h4>
        <div className="flex flex-wrap items-center gap-2">
          {adv.techniques.map((t, i) => (
            <Fragment key={t.id}>
              <span className="inline-flex items-center gap-1.5 rounded-md border border-border bg-background/40 px-2 py-1 text-xs">
                <Crosshair size={11} className="text-rose-500" />
                <span className="font-mono">{t.id}</span>
                <span>{t.name}</span>
              </span>
              {i < adv.techniques.length - 1 && <ArrowRight size={12} className="text-muted-foreground" />}
            </Fragment>
          ))}
        </div>
      </div>
    </div>
  )
}

/* ─── Drawer: Targets ──────────────────────────────────────────────────── */

function TargetsTab({ adv }: { adv: Adversary }) {
  const totalHits = adv.targets.reduce((s, t) => s + t.hits, 0)
  return (
    <div className="space-y-4 p-6">
      <div className="rounded-lg border border-border bg-card p-4">
        <h4 className="mb-3 text-sm font-semibold">Attack flow</h4>
        <div className="grid grid-cols-[1fr_auto_1fr_auto_1fr] items-center gap-3">
          {/* Adversary */}
          <div className="rounded-lg border border-rose-500/30 bg-rose-500/5 p-3">
            <div className="flex items-center gap-2 text-xs font-medium">
              <UserX size={13} className="text-rose-500" />
              Adversary
            </div>
            <div className="mt-1 truncate font-mono text-sm">{adv.identifier}</div>
          </div>
          <ArrowRight size={16} className="text-muted-foreground" />
          {/* Techniques bucket */}
          <div className="rounded-lg border border-amber-500/30 bg-amber-500/5 p-3">
            <div className="flex items-center gap-2 text-xs font-medium">
              <Crosshair size={13} className="text-amber-500" />
              {adv.techniques.length} techniques
            </div>
            <div className="mt-1 text-[11px] text-muted-foreground">{totalHits.toLocaleString()} total hits</div>
          </div>
          <ArrowRight size={16} className="text-muted-foreground" />
          {/* Targets */}
          <div className="rounded-lg border border-sky-500/30 bg-sky-500/5 p-3">
            <div className="flex items-center gap-2 text-xs font-medium">
              <Target size={13} className="text-sky-500" />
              {adv.targets.length} targets hit
            </div>
            <div className="mt-1 text-[11px] text-muted-foreground">
              {adv.targets.reduce((s, t) => s + (t.user ? 1 : 0), 0)} users · {adv.targets.length}{' '}
              hosts
            </div>
          </div>
        </div>
      </div>

      <div className="rounded-lg border border-border bg-card">
        <div className="grid grid-cols-[1fr_120px_120px_80px] gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
          <div>Target</div>
          <div>Type</div>
          <div className="text-right">Hits</div>
          <div className="text-right">Share</div>
        </div>
        {adv.targets.map((t, i) => {
          const share = (t.hits / totalHits) * 100
          return (
            <div
              key={i}
              className="grid grid-cols-[1fr_120px_120px_80px] items-center gap-3 border-b border-border px-4 py-2.5 text-xs last:border-b-0 hover:bg-muted/40"
            >
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <Server size={12} className="text-muted-foreground" />
                  <span className="font-mono">{t.host ?? t.ip}</span>
                  {t.user && <span className="text-muted-foreground">· {t.user}</span>}
                </div>
              </div>
              <div className="text-muted-foreground">{t.host ? 'host' : t.user ? 'user' : 'asset'}</div>
              <div className="text-right tabular-nums">{t.hits.toLocaleString()}</div>
              <div className="flex items-center justify-end gap-1.5">
                <div className="h-1 w-12 overflow-hidden rounded-full bg-muted">
                  <div className="h-full rounded-full bg-rose-500" style={{ width: `${share}%` }} />
                </div>
                <span className="font-mono text-[10px] tabular-nums">{Math.round(share)}%</span>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

/* ─── Drawer: Timeline ─────────────────────────────────────────────────── */

function TimelineTab({ adv }: { adv: Adversary }) {
  return (
    <div className="p-6">
      <div className="mb-4 rounded-lg border border-border bg-card p-4">
        <h4 className="mb-3 text-sm font-semibold">Recent alerts triggered</h4>
        <ul className="space-y-2.5">
          {adv.alerts.map((a) => {
            const sev = THREAT_STYLE[a.severity]
            return (
              <li key={a.id} className="flex items-start gap-3 text-xs">
                <span className={cn('mt-1 h-2 w-2 shrink-0 rounded-full', sev.bar)} />
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="truncate font-medium">{a.name}</span>
                  </div>
                  <div className="mt-0.5 flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
                    <span className="font-mono">{a.id}</span>
                    <span>·</span>
                    <span className="font-mono">{a.technique.id}</span>
                    <span>·</span>
                    <span>{a.echoes} echoes</span>
                  </div>
                </div>
                <span className="shrink-0 font-mono text-[10px] text-muted-foreground">{relativeTime(a.ts)}</span>
              </li>
            )
          })}
        </ul>
      </div>

      <ol className="relative space-y-4 border-l border-border pl-6">
        <li className="relative">
          <span className="absolute -left-[27px] top-1 h-3 w-3 rounded-full border-2 border-card bg-rose-500" />
          <div className="text-xs">
            <span className="font-medium">Latest activity</span> · {relativeTime(adv.lastSeen)}
          </div>
          <div className="text-[10px] text-muted-foreground">{absTimestamp(adv.lastSeen)}</div>
        </li>
        <li className="relative">
          <span className="absolute -left-[27px] top-1 h-3 w-3 rounded-full border-2 border-card bg-violet-500" />
          <div className="text-xs">Auto-correlated to {adv.alertsCount} alerts</div>
        </li>
        {adv.attribution && (
          <li className="relative">
            <span className="absolute -left-[27px] top-1 h-3 w-3 rounded-full border-2 border-card bg-fuchsia-500" />
            <div className="text-xs">
              Attribution match: <span className="font-medium">{adv.attribution.actor}</span>
            </div>
            <div className="text-[10px] text-muted-foreground">
              {adv.attribution.confidence} confidence · {adv.attribution.source}
            </div>
          </li>
        )}
        <li className="relative">
          <span className="absolute -left-[27px] top-1 h-3 w-3 rounded-full border-2 border-card bg-sky-500" />
          <div className="text-xs">
            <span className="font-medium">First seen</span> · {relativeTime(adv.firstSeen)}
          </div>
          <div className="text-[10px] text-muted-foreground">{absTimestamp(adv.firstSeen)}</div>
        </li>
      </ol>
    </div>
  )
}

/* ─── Small parts ──────────────────────────────────────────────────────── */

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <h4 className="mb-2 text-sm font-semibold">{title}</h4>
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
          <div key={k} className="contents">
            <dt className="text-muted-foreground">{k}</dt>
            <dd className="break-words">{v}</dd>
          </div>
        ))}
      </dl>
    </div>
  )
}

/* ─── Helpers ──────────────────────────────────────────────────────────── */

function relativeTime(iso: string) {
  const diff = NOW - new Date(iso).getTime()
  const m = Math.round(diff / 60_000)
  if (m < 1) return 'just now'
  if (m < 60) return `${m}m ago`
  const h = Math.round(m / 60)
  if (h < 24) return `${h}h ago`
  return `${Math.round(h / 24)}d ago`
}

function absTimestamp(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

