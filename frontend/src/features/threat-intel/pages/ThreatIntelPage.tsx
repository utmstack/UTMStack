import { useMemo, useState } from 'react'
import {
  AlertTriangle,
  Bug,
  ChevronDown,
  Clock,
  Crosshair,
  Download,
  ExternalLink,
  FileText,
  Fingerprint,
  Globe2,
  Hash,
  Link as LinkIcon,
  ListFilter,
  MoreHorizontal,
  Network,
  RefreshCw,
  Search,
  Shield,
  Sparkles,
  UserX,
  X,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type IocType = 'ipv4' | 'domain' | 'sha256' | 'md5' | 'url' | 'cve'
type Severity = 'critical' | 'high' | 'medium' | 'low'
type FeedKind = 'commercial' | 'open' | 'internal'
type FeedStatus = 'healthy' | 'stale' | 'error' | 'paused'

interface IOC {
  id: string
  type: IocType
  value: string
  severity: Severity
  feeds: string[]
  matchesInEnv: number
  firstSeenInEnv?: string
  lastSeenInEnv?: string
  tags: string[]
  attributedActor?: string
}

interface ThreatActor {
  id: string
  name: string
  aliases: string[]
  techniques: string[]
  sectors: string[]
  origin: string[]
  matchedAdversaries: number
  lastActivity: string
  iocCount: number
  description: string
}

interface ThreatFeed {
  id: string
  name: string
  kind: FeedKind
  status: FeedStatus
  itemsTotal: number
  itemsAdded24h: number
  lastSync: string
  syncIntervalMin: number
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const NOW = new Date('2026-05-01T16:42:00Z').getTime()
const min = (m: number) => new Date(NOW - m * 60_000).toISOString()
const hr = (h: number) => new Date(NOW - h * 3_600_000).toISOString()
const days = (d: number) => new Date(NOW - d * 86_400_000).toISOString()

const IOCS: IOC[] = [
  {
    id: 'ioc-1',
    type: 'ipv4',
    value: '198.51.100.14',
    severity: 'critical',
    feeds: ['AlienVault OTX', 'Mandiant', 'AbuseIPDB'],
    matchesInEnv: 23,
    firstSeenInEnv: hr(46),
    lastSeenInEnv: min(4),
    tags: ['C2', 'Scattered Spider', 'campaign-44'],
    attributedActor: 'Scattered Spider',
  },
  {
    id: 'ioc-2',
    type: 'sha256',
    value: '9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08',
    severity: 'critical',
    feeds: ['VirusTotal', 'Mandiant'],
    matchesInEnv: 1,
    firstSeenInEnv: min(62),
    lastSeenInEnv: min(62),
    tags: ['Mimikatz', 'credential-dumping'],
    attributedActor: 'Scattered Spider',
  },
  {
    id: 'ioc-3',
    type: 'domain',
    value: 'malicious-update[.]com',
    severity: 'high',
    feeds: ['AbuseCH', 'AlienVault OTX'],
    matchesInEnv: 8,
    firstSeenInEnv: hr(2),
    lastSeenInEnv: min(14),
    tags: ['phishing', 'credential-harvest'],
  },
  {
    id: 'ioc-4',
    type: 'ipv4',
    value: '45.83.66.12',
    severity: 'high',
    feeds: ['Spamhaus', 'AlienVault OTX'],
    matchesInEnv: 4,
    firstSeenInEnv: hr(8),
    lastSeenInEnv: min(74),
    tags: ['exfiltration', 'bulletproof-hoster'],
  },
  {
    id: 'ioc-5',
    type: 'cve',
    value: 'CVE-2024-38077',
    severity: 'critical',
    feeds: ['NVD', 'CISA KEV'],
    matchesInEnv: 0,
    tags: ['RCE', 'Windows', 'KEV'],
  },
  {
    id: 'ioc-6',
    type: 'ipv4',
    value: '203.0.113.7',
    severity: 'medium',
    feeds: ['Internal blocklist'],
    matchesInEnv: 312,
    firstSeenInEnv: hr(2),
    lastSeenInEnv: min(23),
    tags: ['brute-force', 'recent-abuse'],
  },
  {
    id: 'ioc-7',
    type: 'url',
    value: 'hxxp://exfil[.]bad/upload',
    severity: 'high',
    feeds: ['Internal IOC db'],
    matchesInEnv: 2,
    firstSeenInEnv: hr(4),
    lastSeenInEnv: hr(4),
    tags: ['exfiltration'],
  },
  {
    id: 'ioc-8',
    type: 'md5',
    value: 'd41d8cd98f00b204e9800998ecf8427e',
    severity: 'medium',
    feeds: ['VirusTotal'],
    matchesInEnv: 0,
    tags: ['adware'],
  },
  {
    id: 'ioc-9',
    type: 'domain',
    value: 'cdn-update-service[.]net',
    severity: 'low',
    feeds: ['Spamhaus'],
    matchesInEnv: 0,
    tags: ['suspicious-tld'],
  },
  {
    id: 'ioc-10',
    type: 'cve',
    value: 'CVE-2024-21412',
    severity: 'high',
    feeds: ['NVD', 'CISA KEV'],
    matchesInEnv: 0,
    tags: ['SmartScreen-bypass', 'KEV'],
  },
]

const ACTORS: ThreatActor[] = [
  {
    id: 'actor-scattered-spider',
    name: 'Scattered Spider',
    aliases: ['UNC3944', 'Scatter Swine', 'Octo Tempest'],
    techniques: ['T1558.003', 'T1059.001', 'T1071.001', 'T1003'],
    sectors: ['Telecom', 'Hospitality', 'Tech', 'Finance'],
    origin: ['Unknown'],
    matchedAdversaries: 3,
    lastActivity: min(4),
    iocCount: 47,
    description:
      'Financially motivated threat group known for SIM swapping, social engineering, and Kerberoasting attacks against cloud and on-prem infrastructure.',
  },
  {
    id: 'actor-fin7',
    name: 'FIN7',
    aliases: ['Carbon Spider', 'Sangria Tempest'],
    techniques: ['T1566', 'T1059.001', 'T1027'],
    sectors: ['Retail', 'Hospitality', 'Finance'],
    origin: ['Russia (suspected)'],
    matchedAdversaries: 1,
    lastActivity: days(8),
    iocCount: 312,
    description:
      'Long-running cybercrime group targeting POS systems and financial data through targeted phishing.',
  },
  {
    id: 'actor-apt29',
    name: 'APT29',
    aliases: ['Cozy Bear', 'Midnight Blizzard', 'NOBELIUM'],
    techniques: ['T1078', 'T1098', 'T1606'],
    sectors: ['Government', 'Tech', 'Diplomatic'],
    origin: ['Russia (SVR)'],
    matchedAdversaries: 0,
    lastActivity: days(32),
    iocCount: 184,
    description:
      'Russian state-sponsored APT focused on intelligence collection. Known for SolarWinds and Microsoft 365 token theft campaigns.',
  },
  {
    id: 'actor-lazarus',
    name: 'Lazarus Group',
    aliases: ['Hidden Cobra', 'APT38'],
    techniques: ['T1190', 'T1059', 'T1486'],
    sectors: ['Finance', 'Crypto', 'Defense'],
    origin: ['North Korea (DPRK)'],
    matchedAdversaries: 0,
    lastActivity: days(60),
    iocCount: 521,
    description: 'DPRK state actor focused on financial theft and cryptocurrency targeting.',
  },
]

const FEEDS: ThreatFeed[] = [
  { id: 'f-otx', name: 'AlienVault OTX', kind: 'open', status: 'healthy', itemsTotal: 4_120_044, itemsAdded24h: 8_421, lastSync: min(8), syncIntervalMin: 15 },
  { id: 'f-mandiant', name: 'Mandiant Advantage', kind: 'commercial', status: 'healthy', itemsTotal: 218_410, itemsAdded24h: 412, lastSync: min(22), syncIntervalMin: 60 },
  { id: 'f-abusech', name: 'AbuseCH (URLhaus + ThreatFox)', kind: 'open', status: 'healthy', itemsTotal: 891_201, itemsAdded24h: 2_104, lastSync: min(11), syncIntervalMin: 30 },
  { id: 'f-virustotal', name: 'VirusTotal', kind: 'commercial', status: 'healthy', itemsTotal: 3_420_115, itemsAdded24h: 14_872, lastSync: min(4), syncIntervalMin: 10 },
  { id: 'f-spamhaus', name: 'Spamhaus DROP/EDROP', kind: 'open', status: 'healthy', itemsTotal: 18_421, itemsAdded24h: 22, lastSync: hr(2), syncIntervalMin: 240 },
  { id: 'f-cisa', name: 'CISA Known Exploited Vulns (KEV)', kind: 'open', status: 'healthy', itemsTotal: 1_142, itemsAdded24h: 0, lastSync: hr(8), syncIntervalMin: 1440 },
  { id: 'f-nvd', name: 'NVD CVE feed', kind: 'open', status: 'stale', itemsTotal: 218_001, itemsAdded24h: 47, lastSync: hr(36), syncIntervalMin: 360 },
  { id: 'f-misp', name: 'Internal MISP', kind: 'internal', status: 'healthy', itemsTotal: 4_812, itemsAdded24h: 18, lastSync: min(2), syncIntervalMin: 5 },
  { id: 'f-twinds', name: 'ThreatWinds Galaxy', kind: 'commercial', status: 'healthy', itemsTotal: 12_491_201, itemsAdded24h: 34_120, lastSync: min(1), syncIntervalMin: 5 },
  { id: 'f-recorded', name: 'Recorded Future', kind: 'commercial', status: 'error', itemsTotal: 0, itemsAdded24h: 0, lastSync: hr(12), syncIntervalMin: 60 },
]

/* ─── Style maps ───────────────────────────────────────────────────────── */

const SEVERITY_STYLE: Record<Severity, { bar: string; label: string; tone: string }> = {
  critical: { bar: 'bg-red-500', label: 'Critical', tone: 'text-red-500' },
  high: { bar: 'bg-orange-500', label: 'High', tone: 'text-orange-500' },
  medium: { bar: 'bg-amber-500', label: 'Medium', tone: 'text-amber-500' },
  low: { bar: 'bg-sky-500', label: 'Low', tone: 'text-sky-500' },
}

const TYPE_META: Record<IocType, { icon: LucideIcon; label: string; tone: string }> = {
  ipv4: { icon: Network, label: 'IPv4', tone: 'text-sky-500' },
  domain: { icon: Globe2, label: 'Domain', tone: 'text-violet-500' },
  sha256: { icon: Hash, label: 'SHA-256', tone: 'text-fuchsia-500' },
  md5: { icon: Fingerprint, label: 'MD5', tone: 'text-pink-500' },
  url: { icon: LinkIcon, label: 'URL', tone: 'text-amber-500' },
  cve: { icon: Bug, label: 'CVE', tone: 'text-red-500' },
}

const FEED_KIND_TONE: Record<FeedKind, string> = {
  commercial: 'text-violet-500',
  open: 'text-sky-500',
  internal: 'text-emerald-500',
}

const FEED_STATUS_META: Record<FeedStatus, { dot: string; label: string }> = {
  healthy: { dot: 'bg-emerald-500', label: 'Healthy' },
  stale: { dot: 'bg-amber-500', label: 'Stale' },
  error: { dot: 'bg-red-500 animate-pulse', label: 'Error' },
  paused: { dot: 'bg-muted-foreground', label: 'Paused' },
}

/* ─── Page ─────────────────────────────────────────────────────────────── */

type Tab = 'matched' | 'all' | 'actors' | 'feeds'

export function ThreatIntelPage() {
  const [tab, setTab] = useState<Tab>('matched')
  const [search, setSearch] = useState('')
  const [openIoc, setOpenIoc] = useState<IOC | null>(null)
  const [openActor, setOpenActor] = useState<ThreatActor | null>(null)

  const matchedIocs = IOCS.filter((i) => i.matchesInEnv > 0)
  const totalMatches = matchedIocs.reduce((s, i) => s + i.matchesInEnv, 0)

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <Header total={matchedIocs.length} />

      <div className="mt-5 grid grid-cols-12 gap-4">
        <div className="col-span-12 lg:col-span-8">
          <MatchOverviewCard total={totalMatches} matchedCount={matchedIocs.length} />
        </div>
        <div className="col-span-12 lg:col-span-4">
          <IntelInsightsCard />
        </div>
      </div>

      <div className="mt-5">
        <LookupBar />
      </div>

      <div className="mt-5">
        <TabsRow current={tab} onChange={setTab} />
      </div>

      <div className="mt-3 flex flex-wrap items-center gap-2">
        <div className="relative min-w-[280px] flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder={
              tab === 'feeds'
                ? 'Search feeds…'
                : tab === 'actors'
                  ? 'Search actors, aliases, techniques…'
                  : 'Search IOC value, tag, source…'
            }
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="h-9 pl-9"
          />
        </div>
        {(tab === 'matched' || tab === 'all') && (
          <>
            <Button variant="outline" size="sm">
              <Clock size={14} className="mr-2" />
              Last 24 hours
              <ChevronDown size={12} className="ml-1.5 opacity-60" />
            </Button>
            <Button variant="outline" size="sm">
              <ListFilter size={14} className="mr-2" />
              Filters
            </Button>
          </>
        )}
        <button
          title="Refresh"
          className="flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <RefreshCw size={14} />
        </button>
      </div>

      <div className="mt-3">
        {tab === 'matched' && <IocTable iocs={matchedIocs} search={search} onOpen={setOpenIoc} matched />}
        {tab === 'all' && <IocTable iocs={IOCS} search={search} onOpen={setOpenIoc} />}
        {tab === 'actors' && <ActorsList search={search} onOpen={setOpenActor} />}
        {tab === 'feeds' && <FeedsList search={search} />}
      </div>

      {openIoc && <IocDrawer ioc={openIoc} onClose={() => setOpenIoc(null)} />}
      {openActor && <ActorDrawer actor={openActor} onClose={() => setOpenActor(null)} />}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ total }: { total: number }) {
  return (
    <header className="flex flex-wrap items-center justify-between gap-3">
      <div className="text-xs text-muted-foreground">
        <span className="font-medium text-foreground">{total.toLocaleString()}</span> matched in your env
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm">
          <FileText size={14} className="mr-2" />
          Manage feeds
        </Button>
        <Button variant="default" size="sm">
          <Download size={14} className="mr-2" />
          Export IOCs
        </Button>
      </div>
    </header>
  )
}

/* ─── Match overview ───────────────────────────────────────────────────── */

function MatchOverviewCard({ total, matchedCount }: { total: number; matchedCount: number }) {
  const buckets = 48
  const data = useMemo(
    () =>
      Array.from({ length: buckets }, (_, i) =>
        Math.max(
          0,
          Math.round(Math.abs(Math.sin(i / 5)) * 18 + Math.abs(Math.sin(i / 11)) * 8 + (i === 38 ? 12 : 0) + 3)
        )
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
            IOC matches · last 24 hours
          </div>
          <div className="mt-1 flex items-baseline gap-3">
            <span className="text-3xl font-semibold tabular-nums">{total.toLocaleString()}</span>
            <span className="text-sm text-muted-foreground">
              <span className="text-foreground">{matchedCount}</span> distinct indicators
            </span>
          </div>
        </div>
      </div>

      <svg viewBox={`0 0 ${w} ${h}`} className="mt-4 h-24 w-full" preserveAspectRatio="none">
        <defs>
          <linearGradient id="iocGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="rgb(168 85 247)" stopOpacity="0.28" />
            <stop offset="100%" stopColor="rgb(168 85 247)" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d={areaPath} fill="url(#iocGrad)" />
        <path
          d={linePath}
          fill="none"
          stroke="rgb(168 85 247)"
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

/* ─── AI insights ──────────────────────────────────────────────────────── */

function IntelInsightsCard() {
  return (
    <div className="flex h-full flex-col gap-3 rounded-xl border border-border bg-card p-5">
      <div className="flex items-center gap-2">
        <Sparkles size={14} className="text-fuchsia-500" />
        <h3 className="text-sm font-semibold">SOC AI insights</h3>
      </div>

      <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">3 adversaries</span> in your env attributed to{' '}
        <span className="font-medium">Scattered Spider</span> — campaign appears active.{' '}
        <button className="text-primary hover:underline">View campaign</button>
      </div>

      <div className="rounded-lg border border-amber-500/30 bg-amber-500/5 p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">2 KEV CVEs</span> applicable to your stack — no
        patches deployed yet. <button className="text-primary hover:underline">Review</button>
      </div>

      <div className="rounded-lg border border-red-500/30 bg-red-500/5 p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">1 feed errored</span> · Recorded Future auth
        token expired 12h ago. <button className="text-primary hover:underline">Reconnect</button>
      </div>
    </div>
  )
}

/* ─── Lookup bar ───────────────────────────────────────────────────────── */

function LookupBar() {
  const [q, setQ] = useState('')
  return (
    <div className="rounded-xl border border-border bg-card p-4">
      <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <Crosshair size={12} className="text-fuchsia-500" />
        Lookup any indicator
      </div>
      <div className="mt-2 flex items-center gap-2">
        <div className="relative flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Paste an IP, domain, hash, URL, or CVE — we'll check feeds and your environment"
            className="h-10 pl-9 font-mono text-sm"
          />
        </div>
        <Button>
          <Crosshair size={13} className="mr-1.5" />
          Lookup
        </Button>
      </div>
    </div>
  )
}

/* ─── Tabs row ─────────────────────────────────────────────────────────── */

const TABS: { id: Tab; label: string; count: number }[] = [
  { id: 'matched', label: 'Matched in your env', count: IOCS.filter((i) => i.matchesInEnv > 0).length },
  { id: 'all', label: 'All IOCs', count: IOCS.length },
  { id: 'actors', label: 'Threat actors', count: ACTORS.length },
  { id: 'feeds', label: 'Feeds', count: FEEDS.length },
]

function TabsRow({ current, onChange }: { current: Tab; onChange: (t: Tab) => void }) {
  return (
    <div className="flex flex-wrap items-center gap-1 border-b border-border">
      {TABS.map((t) => {
        const active = current === t.id
        return (
          <button
            key={t.id}
            onClick={() => onChange(t.id)}
            className={cn(
              'group relative flex items-center gap-2 px-3 py-2 text-xs transition-colors',
              active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
            )}
          >
            <span>{t.label}</span>
            <span
              className={cn(
                'rounded-md px-1.5 py-0.5 font-mono text-[10px] tabular-nums',
                active ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground'
              )}
            >
              {t.count}
            </span>
            {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
          </button>
        )
      })}
    </div>
  )
}

/* ─── IOC table ────────────────────────────────────────────────────────── */

const IOC_COLS = '4px 90px 1fr 130px 70px 110px 100px 36px'

function IocTable({
  iocs,
  search,
  onOpen,
  matched,
}: {
  iocs: IOC[]
  search: string
  onOpen: (i: IOC) => void
  matched?: boolean
}) {
  const filtered = iocs.filter((i) =>
    search
      ? (i.value + i.tags.join(' ') + i.feeds.join(' ') + (i.attributedActor ?? ''))
          .toLowerCase()
          .includes(search.toLowerCase())
      : true
  )

  return (
    <div className="overflow-hidden rounded-xl border border-border bg-card">
      <div
        className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
        style={{ gridTemplateColumns: IOC_COLS }}
      >
        <div />
        <div>Type</div>
        <div>Indicator</div>
        <div>Source</div>
        <div className="text-right">{matched ? 'Matches' : 'Severity'}</div>
        <div>Tags</div>
        <div>Last seen</div>
        <div />
      </div>
      {filtered.map((i) => (
        <IocRow key={i.id} ioc={i} onOpen={() => onOpen(i)} matched={matched} />
      ))}
      {filtered.length === 0 && (
        <div className="px-6 py-16 text-center text-sm text-muted-foreground">No IOCs match.</div>
      )}
    </div>
  )
}

function IocRow({ ioc, onOpen, matched }: { ioc: IOC; onOpen: () => void; matched?: boolean }) {
  const sev = SEVERITY_STYLE[ioc.severity]
  const t = TYPE_META[ioc.type]
  const TIcon = t.icon
  return (
    <div
      onClick={onOpen}
      className="group grid cursor-pointer items-center gap-3 border-b border-border/60 px-4 py-2.5 text-xs hover:bg-muted/40 last:border-b-0"
      style={{ gridTemplateColumns: IOC_COLS }}
    >
      <span className={cn('h-3 w-1 rounded-full', sev.bar)} />
      <div className="flex items-center gap-1.5">
        <TIcon size={11} className={t.tone} />
        <span className="font-mono text-[10px] uppercase tracking-wider text-muted-foreground">
          {t.label}
        </span>
      </div>
      <div className="min-w-0">
        <div className="truncate font-mono">{ioc.value}</div>
        {ioc.attributedActor && (
          <div className="mt-0.5 flex items-center gap-1 text-[10px] text-fuchsia-500">
            <UserX size={9} />
            {ioc.attributedActor}
          </div>
        )}
      </div>
      <div className="truncate text-[11px] text-muted-foreground">
        {ioc.feeds[0]}
        {ioc.feeds.length > 1 && (
          <span className="ml-1 text-[10px] text-muted-foreground/70">+{ioc.feeds.length - 1}</span>
        )}
      </div>
      <div className="text-right">
        {matched ? (
          <span className="font-mono tabular-nums">{ioc.matchesInEnv.toLocaleString()}</span>
        ) : (
          <span className={cn('text-[11px] font-medium', sev.tone)}>{sev.label}</span>
        )}
      </div>
      <div className="flex flex-wrap gap-1">
        {ioc.tags.slice(0, 2).map((tag) => (
          <span key={tag} className="rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
            {tag}
          </span>
        ))}
        {ioc.tags.length > 2 && (
          <span className="text-[10px] text-muted-foreground">+{ioc.tags.length - 2}</span>
        )}
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">
        {ioc.lastSeenInEnv ? relativeTime(ioc.lastSeenInEnv) : '—'}
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

/* ─── Actors list ──────────────────────────────────────────────────────── */

function ActorsList({
  search,
  onOpen,
}: {
  search: string
  onOpen: (a: ThreatActor) => void
}) {
  const filtered = ACTORS.filter((a) =>
    search
      ? (a.name + a.aliases.join(' ') + a.techniques.join(' ') + a.sectors.join(' '))
          .toLowerCase()
          .includes(search.toLowerCase())
      : true
  )
  return (
    <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
      {filtered.map((a) => (
        <ActorCard key={a.id} actor={a} onOpen={() => onOpen(a)} />
      ))}
    </div>
  )
}

function ActorCard({ actor, onOpen }: { actor: ThreatActor; onOpen: () => void }) {
  const inEnv = actor.matchedAdversaries > 0
  return (
    <button
      onClick={onOpen}
      className={cn(
        'group flex flex-col gap-3 rounded-xl border bg-card p-4 text-left transition-all hover:shadow-md',
        inEnv ? 'border-fuchsia-500/30' : 'border-border'
      )}
    >
      <div className="flex items-start gap-3">
        <div
          className={cn(
            'flex h-10 w-10 shrink-0 items-center justify-center rounded-lg ring-2',
            inEnv
              ? 'bg-fuchsia-500/15 text-fuchsia-500 ring-fuchsia-500/40'
              : 'bg-muted text-foreground/80 ring-border'
          )}
        >
          <UserX size={18} />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="truncate text-sm font-semibold">{actor.name}</span>
            {inEnv && (
              <span className="rounded bg-fuchsia-500/15 px-1.5 py-0.5 text-[9px] font-medium uppercase text-fuchsia-500 ring-1 ring-fuchsia-500/30">
                in your env
              </span>
            )}
          </div>
          <div className="mt-0.5 truncate text-[11px] text-muted-foreground">
            aka {actor.aliases.join(', ')}
          </div>
        </div>
      </div>

      <p className="line-clamp-2 text-xs text-foreground/80">{actor.description}</p>

      <div className="grid grid-cols-3 gap-2 text-[11px]">
        <Stat label="Techniques" value={actor.techniques.length.toString()} />
        <Stat label="Adversaries" value={actor.matchedAdversaries.toString()} tone={inEnv ? 'text-fuchsia-500' : ''} />
        <Stat label="IOCs" value={actor.iocCount.toString()} />
      </div>

      <div className="flex flex-wrap items-center gap-1.5 border-t border-border pt-3 text-[10px] text-muted-foreground">
        <span className="text-[10px] uppercase tracking-wider">Sectors:</span>
        {actor.sectors.slice(0, 4).map((s) => (
          <span key={s} className="rounded bg-muted px-1.5 py-0.5">{s}</span>
        ))}
      </div>

      <div className="text-[10px] text-muted-foreground">
        Last activity in env <span className="font-mono">{relativeTime(actor.lastActivity)}</span>
      </div>
    </button>
  )
}

function Stat({ label, value, tone }: { label: string; value: string; tone?: string }) {
  return (
    <div className="rounded-md border border-border bg-background/40 px-2 py-1.5">
      <div className="text-[9px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className={cn('mt-0.5 font-semibold tabular-nums', tone)}>{value}</div>
    </div>
  )
}

/* ─── Feeds list ───────────────────────────────────────────────────────── */

const FEED_COLS = '12px 1fr 110px 110px 100px 110px 36px'

function FeedsList({ search }: { search: string }) {
  const filtered = FEEDS.filter((f) =>
    search ? f.name.toLowerCase().includes(search.toLowerCase()) : true
  )
  return (
    <div className="overflow-hidden rounded-xl border border-border bg-card">
      <div
        className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
        style={{ gridTemplateColumns: FEED_COLS }}
      >
        <div />
        <div>Feed</div>
        <div>Kind</div>
        <div className="text-right">Total</div>
        <div className="text-right">+24h</div>
        <div>Last sync</div>
        <div />
      </div>
      {filtered.map((f) => (
        <FeedRow key={f.id} feed={f} />
      ))}
      {filtered.length === 0 && (
        <div className="px-6 py-16 text-center text-sm text-muted-foreground">No feeds match.</div>
      )}
    </div>
  )
}

function FeedRow({ feed }: { feed: ThreatFeed }) {
  const st = FEED_STATUS_META[feed.status]
  return (
    <div
      className="group grid cursor-pointer items-center gap-3 border-b border-border/60 px-4 py-2.5 text-xs hover:bg-muted/40 last:border-b-0"
      style={{ gridTemplateColumns: FEED_COLS }}
    >
      <span className={cn('h-2 w-2 rounded-full', st.dot)} title={st.label} />
      <div className="min-w-0">
        <div className="truncate font-medium">{feed.name}</div>
        <div className="truncate text-[10px] text-muted-foreground">
          syncs every{' '}
          {feed.syncIntervalMin < 60
            ? `${feed.syncIntervalMin}m`
            : feed.syncIntervalMin < 1440
              ? `${Math.round(feed.syncIntervalMin / 60)}h`
              : `${Math.round(feed.syncIntervalMin / 1440)}d`}
        </div>
      </div>
      <div className={cn('text-[11px] capitalize', FEED_KIND_TONE[feed.kind])}>
        {feed.kind}
      </div>
      <div className="text-right font-mono tabular-nums text-muted-foreground">
        {feed.itemsTotal.toLocaleString()}
      </div>
      <div
        className={cn(
          'text-right font-mono tabular-nums',
          feed.itemsAdded24h > 0 ? 'text-emerald-500' : 'text-muted-foreground'
        )}
      >
        {feed.itemsAdded24h > 0 ? `+${feed.itemsAdded24h.toLocaleString()}` : '—'}
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">
        {relativeTime(feed.lastSync)}
      </div>
      <div className="flex justify-end opacity-0 group-hover:opacity-100">
        <button className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-background hover:text-foreground">
          <MoreHorizontal size={14} />
        </button>
      </div>
    </div>
  )
}

/* ─── IOC drawer ───────────────────────────────────────────────────────── */

function IocDrawer({ ioc, onClose }: { ioc: IOC; onClose: () => void }) {
  const sev = SEVERITY_STYLE[ioc.severity]
  const t = TYPE_META[ioc.type]
  const TIcon = t.icon
  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[820px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                <TIcon size={11} className={t.tone} />
                <span>{t.label}</span>
                <span>·</span>
                <span className={cn('font-medium', sev.tone)}>{sev.label}</span>
              </div>
              <h2 className="mt-1.5 break-all font-mono text-base font-semibold leading-snug">{ioc.value}</h2>
              <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                {ioc.attributedActor && (
                  <span className="inline-flex items-center gap-1 rounded-md border border-fuchsia-500/30 bg-fuchsia-500/10 px-1.5 py-0.5 text-fuchsia-500">
                    <UserX size={10} />
                    {ioc.attributedActor}
                  </span>
                )}
                {ioc.tags.map((tag) => (
                  <span key={tag} className="rounded-md bg-muted px-1.5 py-0.5">{tag}</span>
                ))}
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
            <Button size="sm">
              <Shield size={13} className="mr-1.5" />
              Block at perimeter
            </Button>
            <Button size="sm" variant="outline">
              <Search size={13} className="mr-1.5" />
              View matched events
            </Button>
            <Button size="sm" variant="outline">
              <ExternalLink size={13} className="mr-1.5" />
              External lookup
            </Button>
            <Button size="sm" variant="ghost" className="ml-auto">
              <MoreHorizontal size={13} />
            </Button>
          </div>
        </header>

        <div className="flex-1 overflow-y-auto bg-muted/20">
          <div className="grid grid-cols-3 gap-4 p-6">
            <div className="col-span-2 space-y-4">
              <Section title="Environment context">
                {ioc.matchesInEnv > 0 ? (
                  <dl className="grid grid-cols-[140px_1fr] gap-y-2 text-xs">
                    <KV k="Matches">
                      <span className="font-mono">{ioc.matchesInEnv.toLocaleString()}</span> events
                    </KV>
                    {ioc.firstSeenInEnv && (
                      <KV k="First seen">
                        <span className="font-mono">{absTimestamp(ioc.firstSeenInEnv)}</span>
                      </KV>
                    )}
                    {ioc.lastSeenInEnv && (
                      <KV k="Last seen">
                        <span className="font-mono">{absTimestamp(ioc.lastSeenInEnv)}</span>
                      </KV>
                    )}
                  </dl>
                ) : (
                  <p className="text-xs text-muted-foreground">
                    Not seen in your environment yet. Pre-block this IOC to prevent matches.
                  </p>
                )}
              </Section>

              <Section title="Threat intelligence sources">
                <ul className="space-y-1.5">
                  {ioc.feeds.map((f) => (
                    <li
                      key={f}
                      className="flex items-center justify-between rounded border border-border bg-background/40 px-3 py-2 text-xs"
                    >
                      <span className="font-medium">{f}</span>
                      <button className="inline-flex items-center gap-1 text-[11px] text-primary hover:underline">
                        Open <ExternalLink size={10} />
                      </button>
                    </li>
                  ))}
                </ul>
              </Section>

              <Section title="External lookups">
                <div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
                  {['VirusTotal', 'AbuseIPDB', 'Shodan', 'AlienVault OTX', 'GreyNoise', 'Censys'].map(
                    (name) => (
                      <button
                        key={name}
                        className="flex items-center justify-between rounded-md border border-border bg-background/40 px-3 py-2 text-xs hover:bg-muted"
                      >
                        <span>{name}</span>
                        <ExternalLink size={11} className="text-muted-foreground" />
                      </button>
                    )
                  )}
                </div>
              </Section>
            </div>

            <aside className="col-span-1 space-y-4">
              <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-4">
                <div className="mb-2 flex items-center gap-2 text-xs font-medium">
                  <Sparkles size={13} className="text-fuchsia-500" />
                  AI brief
                </div>
                <p className="text-xs leading-relaxed">
                  {ioc.matchesInEnv > 0
                    ? `This indicator has been observed ${ioc.matchesInEnv} times in your environment. ${ioc.attributedActor ? `Attribution: ${ioc.attributedActor}.` : 'Recommend immediate review of impacted assets.'}`
                    : 'Not yet observed in your environment, but listed as active in threat feeds. Pre-blocking is recommended.'}
                </p>
              </div>

              <div className="rounded-lg border border-border bg-card p-4">
                <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">
                  Reputation
                </div>
                <div className="space-y-2 text-xs">
                  <Metric label="VT detections" value="22 / 89" />
                  <Metric label="Abuse confidence" value="87%" />
                  <Metric label="Reports (90d)" value="214" />
                </div>
              </div>
            </aside>
          </div>
        </div>
      </div>
    </div>
  )
}

/* ─── Actor drawer ─────────────────────────────────────────────────────── */

function ActorDrawer({ actor, onClose }: { actor: ThreatActor; onClose: () => void }) {
  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[860px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="flex min-w-0 flex-1 items-start gap-3">
              <div
                className={cn(
                  'flex h-11 w-11 shrink-0 items-center justify-center rounded-lg ring-2',
                  actor.matchedAdversaries > 0
                    ? 'bg-fuchsia-500/15 text-fuchsia-500 ring-fuchsia-500/40'
                    : 'bg-muted text-foreground/80 ring-border'
                )}
              >
                <UserX size={20} />
              </div>
              <div className="min-w-0 flex-1">
                <h2 className="truncate text-xl font-semibold">{actor.name}</h2>
                <div className="mt-0.5 truncate text-[11px] text-muted-foreground">
                  aka {actor.aliases.join(', ')}
                </div>
                <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                  {actor.matchedAdversaries > 0 && (
                    <span className="rounded-md bg-fuchsia-500/15 px-1.5 py-0.5 font-medium text-fuchsia-500 ring-1 ring-fuchsia-500/30">
                      {actor.matchedAdversaries} adversaries in your env
                    </span>
                  )}
                  <span className="rounded-md border border-border bg-background/40 px-1.5 py-0.5">
                    Origin · {actor.origin.join(', ')}
                  </span>
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

          <div className="mt-4 flex items-center gap-2">
            <Button size="sm">
              <Crosshair size={13} className="mr-1.5" />
              Track campaign
            </Button>
            <Button size="sm" variant="outline">
              <Search size={13} className="mr-1.5" />
              View matched IOCs
            </Button>
            <Button size="sm" variant="outline">
              <ExternalLink size={13} className="mr-1.5" />
              MITRE ATT&CK
            </Button>
          </div>
        </header>

        <div className="flex-1 overflow-y-auto bg-muted/20 p-6">
          <div className="space-y-4">
            <Section title="About">
              <p className="text-sm leading-relaxed">{actor.description}</p>
            </Section>

            <div className="grid grid-cols-2 gap-4">
              <Section title="MITRE techniques">
                <div className="flex flex-wrap gap-1.5">
                  {actor.techniques.map((t) => (
                    <span
                      key={t}
                      className="inline-flex items-center gap-1 rounded-md border border-border bg-background/40 px-2 py-1 font-mono text-xs"
                    >
                      <Crosshair size={10} className="text-rose-500" />
                      {t}
                    </span>
                  ))}
                </div>
              </Section>

              <Section title="Targeted sectors">
                <div className="flex flex-wrap gap-1.5">
                  {actor.sectors.map((s) => (
                    <span key={s} className="rounded-md bg-muted px-2 py-1 text-xs">{s}</span>
                  ))}
                </div>
              </Section>
            </div>

            <Section title="Activity">
              <div className="grid grid-cols-3 gap-3 text-xs">
                <Stat label="Adversaries" value={actor.matchedAdversaries.toString()} tone={actor.matchedAdversaries > 0 ? 'text-fuchsia-500' : ''} />
                <Stat label="IOCs" value={actor.iocCount.toString()} />
                <Stat label="Last activity" value={relativeTime(actor.lastActivity)} />
              </div>
            </Section>
          </div>
        </div>
      </div>
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

function KV({ k, children }: { k: string; children: React.ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="break-words">{children}</dd>
    </>
  )
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-mono tabular-nums">{value}</span>
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

// Suppress potential unused-import warnings
void AlertTriangle
