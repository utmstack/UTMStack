import { useMemo, useState } from 'react'
import {
  ArrowRight,
  BadgeCheck,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Clock,
  Columns3,
  Crosshair,
  Database,
  Download,
  Flame,
  History,
  ListFilter,
  MoreHorizontal,
  Network,
  RefreshCw,
  Search,
  Server,
  Sparkles,
  Tag as TagIcon,
  Target,
  UserX,
  X,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type Severity = 'critical' | 'high' | 'medium' | 'low'
type Status = 'open' | 'in_review' | 'ignored' | 'completed' | 'auto_review' | 'false_positive'

interface Endpoint {
  ip: string
  host?: string
  user?: string
  country?: string
  countryCode?: string
}

interface Alert {
  id: string
  name: string
  severity: Severity
  status: Status
  timestamp: string
  category: string
  technique: { id: string; name: string }
  sensor: string
  echoes: number
  risk: number
  source: Endpoint
  adversary: Endpoint
  tags: string[]
  isIncident?: boolean
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const NOW = new Date('2026-05-01T16:42:00Z').getTime()
const min = (m: number) => new Date(NOW - m * 60_000).toISOString()

const ALERTS: Alert[] = [
  {
    id: 'a-9f04c2',
    name: 'Suspected Kerberoasting against domain controller',
    severity: 'critical',
    status: 'open',
    timestamp: min(4),
    category: 'Credential Access',
    technique: { id: 'T1558.003', name: 'Kerberoasting' },
    sensor: 'Windows Defender',
    echoes: 142,
    risk: 96,
    source: { ip: '10.4.18.22', host: 'WIN-7F3K2L9Q8X', user: 'svc_backup', countryCode: 'US' },
    adversary: { ip: '198.51.100.14', country: 'Russia', countryCode: 'RU' },
    tags: ['Scattered Spider'],
    isIncident: true,
  },
  {
    id: 'a-1ab32e',
    name: 'PowerShell execution with encoded command',
    severity: 'high',
    status: 'in_review',
    timestamp: min(11),
    category: 'Execution',
    technique: { id: 'T1059.001', name: 'PowerShell' },
    sensor: 'Sysmon',
    echoes: 47,
    risk: 88,
    source: { ip: '10.0.4.91', host: 'fin-app-01', user: 'jsmith', countryCode: 'US' },
    adversary: { ip: '10.0.4.91', host: 'fin-app-01', countryCode: 'US' },
    tags: [],
  },
  {
    id: 'a-7c2d11',
    name: 'Multiple failed logins followed by success',
    severity: 'high',
    status: 'open',
    timestamp: min(23),
    category: 'Credential Access',
    technique: { id: 'T1110', name: 'Brute Force' },
    sensor: 'AD Audit',
    echoes: 312,
    risk: 84,
    source: { ip: '203.0.113.7', country: 'Brazil', countryCode: 'BR' },
    adversary: { ip: '10.2.1.18', host: 'mail-srv-02', user: 'mthomas', countryCode: 'US' },
    tags: [],
  },
  {
    id: 'a-3e88a1',
    name: 'Suspicious SMB share enumeration',
    severity: 'medium',
    status: 'open',
    timestamp: min(38),
    category: 'Discovery',
    technique: { id: 'T1135', name: 'Network Share Discovery' },
    sensor: 'Suricata',
    echoes: 12,
    risk: 67,
    source: { ip: '10.4.18.22', host: 'WIN-7F3K2L9Q8X', user: 'svc_backup', countryCode: 'US' },
    adversary: { ip: '10.4.18.5', host: 'fileshare-01', countryCode: 'US' },
    tags: [],
  },
  {
    id: 'a-c014f0',
    name: 'Outbound connection to known C2 IP',
    severity: 'critical',
    status: 'in_review',
    timestamp: min(46),
    category: 'Command and Control',
    technique: { id: 'T1071.001', name: 'Web Protocols' },
    sensor: 'Firewall',
    echoes: 8,
    risk: 94,
    source: { ip: '10.0.4.91', host: 'fin-app-01', countryCode: 'US' },
    adversary: { ip: '198.51.100.14', country: 'Russia', countryCode: 'RU' },
    tags: ['Scattered Spider'],
  },
  {
    id: 'a-5b9072',
    name: 'New scheduled task on critical asset',
    severity: 'medium',
    status: 'open',
    timestamp: min(58),
    category: 'Persistence',
    technique: { id: 'T1053.005', name: 'Scheduled Task' },
    sensor: 'Sysmon',
    echoes: 3,
    risk: 62,
    source: { ip: '10.0.4.91', host: 'fin-app-01', user: 'SYSTEM', countryCode: 'US' },
    adversary: { ip: '10.0.4.91', host: 'fin-app-01', countryCode: 'US' },
    tags: [],
  },
  {
    id: 'a-44a811',
    name: 'AV signature match — Mimikatz variant',
    severity: 'high',
    status: 'open',
    timestamp: min(62),
    category: 'Credential Access',
    technique: { id: 'T1003', name: 'OS Credential Dumping' },
    sensor: 'Windows Defender',
    echoes: 1,
    risk: 89,
    source: { ip: '10.4.18.22', host: 'WIN-7F3K2L9Q8X', user: 'svc_backup', countryCode: 'US' },
    adversary: { ip: '10.4.18.22', host: 'WIN-7F3K2L9Q8X', countryCode: 'US' },
    tags: [],
  },
  {
    id: 'a-22f3cd',
    name: 'Unusual data transfer to external address',
    severity: 'medium',
    status: 'in_review',
    timestamp: min(74),
    category: 'Exfiltration',
    technique: { id: 'T1041', name: 'C2 Exfiltration' },
    sensor: 'NetFlow',
    echoes: 2,
    risk: 71,
    source: { ip: '10.0.4.91', host: 'fin-app-01', countryCode: 'US' },
    adversary: { ip: '45.83.66.12', country: 'Netherlands', countryCode: 'NL' },
    tags: [],
  },
  {
    id: 'a-9d101e',
    name: 'Signed binary proxy execution',
    severity: 'low',
    status: 'auto_review',
    timestamp: min(91),
    category: 'Defense Evasion',
    technique: { id: 'T1218', name: 'Signed Binary Proxy' },
    sensor: 'Sysmon',
    echoes: 27,
    risk: 41,
    source: { ip: '10.4.18.22', host: 'WIN-7F3K2L9Q8X', countryCode: 'US' },
    adversary: { ip: '10.4.18.22', host: 'WIN-7F3K2L9Q8X', countryCode: 'US' },
    tags: [],
  },
  {
    id: 'a-08f3a4',
    name: 'GPO modification by non-administrator',
    severity: 'high',
    status: 'completed',
    timestamp: min(132),
    category: 'Persistence',
    technique: { id: 'T1484', name: 'Domain Policy Modification' },
    sensor: 'AD Audit',
    echoes: 1,
    risk: 78,
    source: { ip: '10.4.18.91', user: 'tferguson', countryCode: 'US' },
    adversary: { ip: '10.4.18.5', host: 'dc-01', countryCode: 'US' },
    tags: [],
  },
]

/* ─── Style maps ───────────────────────────────────────────────────────── */

const SEVERITY_STYLE: Record<Severity, { bar: string; pill: string; label: string }> = {
  critical: {
    bar: 'bg-red-500',
    pill: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300',
    label: 'Critical',
  },
  high: {
    bar: 'bg-orange-500',
    pill: 'bg-orange-500/15 text-orange-600 ring-orange-500/30 dark:text-orange-300',
    label: 'High',
  },
  medium: {
    bar: 'bg-amber-500',
    pill: 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300',
    label: 'Medium',
  },
  low: {
    bar: 'bg-sky-500',
    pill: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300',
    label: 'Low',
  },
}

const STATUS_META: Record<Status, { label: string; pill: string }> = {
  open: { label: 'Open', pill: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300' },
  in_review: { label: 'In review', pill: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300' },
  ignored: { label: 'Ignored', pill: 'bg-muted text-muted-foreground ring-border' },
  completed: { label: 'Completed', pill: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300' },
  auto_review: { label: 'Auto-reviewed', pill: 'bg-violet-500/15 text-violet-600 ring-violet-500/30 dark:text-violet-300' },
  false_positive: { label: 'False positive', pill: 'bg-zinc-500/15 text-zinc-600 ring-zinc-500/30 dark:text-zinc-300' },
}

/* ─── Saved views ──────────────────────────────────────────────────────── */

interface SavedView {
  id: string
  label: string
  count: number
  predicate: (a: Alert) => boolean
}

const SAVED_VIEWS: SavedView[] = [
  { id: 'all', label: 'All alerts', count: ALERTS.length, predicate: () => true },
  { id: 'untriaged', label: 'Untriaged', count: ALERTS.filter((a) => a.status === 'open').length, predicate: (a) => a.status === 'open' },
  { id: 'mine', label: 'My queue', count: 3, predicate: (a) => a.severity === 'critical' },
  { id: 'critical', label: 'Critical only', count: ALERTS.filter((a) => a.severity === 'critical').length, predicate: (a) => a.severity === 'critical' },
  { id: 'campaign', label: 'Scattered Spider campaign', count: ALERTS.filter((a) => a.tags.includes('Scattered Spider')).length, predicate: (a) => a.tags.includes('Scattered Spider') },
]

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function AlertsPage() {
  const [view, setView] = useState<string>('all')
  const [search, setSearch] = useState('')
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [openAlert, setOpenAlert] = useState<Alert | null>(null)

  const filtered = useMemo(() => {
    const v = SAVED_VIEWS.find((s) => s.id === view) ?? SAVED_VIEWS[0]
    return ALERTS.filter(v.predicate).filter((a) =>
      search ? (a.name + a.adversary.ip + a.source.ip + a.tags.join(' ')).toLowerCase().includes(search.toLowerCase()) : true
    )
  }, [view, search])

  const allChecked = filtered.length > 0 && filtered.every((a) => selected.has(a.id))

  const toggle = (id: string) =>
    setSelected((prev) => {
      const next = new Set(prev)
      next.has(id) ? next.delete(id) : next.add(id)
      return next
    })

  const togglePage = () =>
    setSelected((prev) => {
      if (allChecked) {
        const next = new Set(prev)
        filtered.forEach((a) => next.delete(a.id))
        return next
      }
      const next = new Set(prev)
      filtered.forEach((a) => next.add(a.id))
      return next
    })

  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      <Header total={filtered.length} />

      {/* Hero — chart + AI insights side by side */}
      <div className="mt-5 grid grid-cols-12 gap-4">
        <div className="col-span-12 lg:col-span-8">
          <AlertVolumeCard />
        </div>
        <div className="col-span-12 lg:col-span-4">
          <AlertInsightsCard />
        </div>
      </div>

      <div className="mt-5">
        <SavedViewTabs current={view} onChange={setView} />
      </div>

      <Toolbar search={search} onSearch={setSearch} />

      {selected.size > 0 && (
        <BulkActionBar count={selected.size} onClear={() => setSelected(new Set())} />
      )}

      <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
        <TableHeader allChecked={allChecked} onTogglePage={togglePage} />
        <div>
          {filtered.map((a) => (
            <AlertRow
              key={a.id}
              alert={a}
              checked={selected.has(a.id)}
              onToggle={() => toggle(a.id)}
              onOpen={() => setOpenAlert(a)}
            />
          ))}
          {filtered.length === 0 && (
            <div className="px-6 py-16 text-center text-sm text-muted-foreground">
              No alerts match this view.
            </div>
          )}
        </div>
      </div>

      <PaginationBar total={filtered.length} />

      {openAlert && <AlertDrawer alert={openAlert} onClose={() => setOpenAlert(null)} />}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ total }: { total: number }) {
  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold tracking-tight">Alerts</h1>
          <span className="rounded-md bg-muted px-2 py-0.5 font-mono text-[11px] tabular-nums text-muted-foreground">
            {total.toLocaleString()} matching
          </span>
        </div>
        <p className="mt-1 text-xs text-muted-foreground">
          Triage what matters. Auto-resolved alerts are kept for audit but hidden by default.
        </p>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm">
          <TagIcon size={14} className="mr-2" />
          Manage tags
        </Button>
        <Button variant="default" size="sm">
          <Download size={14} className="mr-2" />
          Export report
        </Button>
      </div>
    </header>
  )
}

/* ─── Volume chart ─────────────────────────────────────────────────────── */

function AlertVolumeCard() {
  const buckets = 48
  const data = useMemo(
    () =>
      Array.from({ length: buckets }, (_, i) =>
        Math.round(Math.abs(Math.sin(i / 5)) * 22 + Math.abs(Math.sin(i / 11)) * 8 + (i === 38 ? 18 : 0) + 6)
      ),
    []
  )
  const total = ALERTS.length
  const critical = ALERTS.filter((a) => a.severity === 'critical').length
  const open = ALERTS.filter((a) => a.status === 'open').length

  const w = 1000
  const h = 110
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
            Alerts · last 24 hours
          </div>
          <div className="mt-1 flex items-baseline gap-3">
            <span className="text-3xl font-semibold tabular-nums">{total}</span>
            <span className="text-sm text-muted-foreground">
              <span className="text-red-500">{critical}</span> critical
              <span className="mx-2 text-border">·</span>
              <span className="text-foreground">{open}</span> open
            </span>
          </div>
        </div>
      </div>

      <svg viewBox={`0 0 ${w} ${h}`} className="mt-4 h-24 w-full" preserveAspectRatio="none">
        <defs>
          <linearGradient id="alertVol" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="rgb(244 63 94)" stopOpacity="0.28" />
            <stop offset="100%" stopColor="rgb(244 63 94)" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d={areaPath} fill="url(#alertVol)" />
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

/* ─── AI insights card ─────────────────────────────────────────────────── */

function AlertInsightsCard() {
  return (
    <div className="flex h-full flex-col gap-3 rounded-xl border border-border bg-card p-5">
      <div className="flex items-center gap-2">
        <Sparkles size={14} className="text-fuchsia-500" />
        <h3 className="text-sm font-semibold">SOC AI insights</h3>
      </div>

      <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">2 alerts</span> appear linked to the
        <span className="font-medium"> Scattered Spider</span> campaign.{' '}
        <button className="text-primary hover:underline">Group as incident</button>
      </div>

      <div className="rounded-lg border border-amber-500/30 bg-amber-500/5 p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">312 brute-force attempts</span> from one IP within
        an hour. <button className="text-primary hover:underline">Block at perimeter</button>
      </div>

      <div className="rounded-lg border border-sky-500/30 bg-sky-500/5 p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">27 alerts</span> auto-classified as benign by
        prior rules. <button className="text-primary hover:underline">Review</button>
      </div>
    </div>
  )
}

/* ─── Saved view tabs ──────────────────────────────────────────────────── */

function SavedViewTabs({
  current,
  onChange,
}: {
  current: string
  onChange: (id: string) => void
}) {
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
}: {
  search: string
  onSearch: (s: string) => void
}) {
  return (
    <div className="mt-3 flex flex-wrap items-center gap-2">
      <div className="relative min-w-[280px] flex-1">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search alert name, IP, host, tag, technique…"
          value={search}
          onChange={(e) => onSearch(e.target.value)}
          className="h-9 pl-9"
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
      </Button>
      <div className="ml-auto flex items-center gap-1">
        <button
          title="Configure columns"
          className="flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <Columns3 size={14} />
        </button>
        <button
          title="Refresh"
          className="flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <RefreshCw size={14} />
        </button>
      </div>
    </div>
  )
}

/* ─── Bulk action bar ──────────────────────────────────────────────────── */

function BulkActionBar({ count, onClear }: { count: number; onClear: () => void }) {
  return (
    <div className="sticky top-2 z-10 mt-3 flex flex-wrap items-center gap-2 rounded-xl border border-primary/30 bg-primary/5 px-4 py-2.5 text-sm shadow-sm backdrop-blur">
      <span className="font-medium">
        {count} alert{count === 1 ? '' : 's'} selected
      </span>
      <span className="text-muted-foreground">·</span>
      <Button variant="ghost" size="sm">
        <Flame size={14} className="mr-1.5 text-amber-500" />
        Mark as incident
      </Button>
      <Button variant="ghost" size="sm">
        <BadgeCheck size={14} className="mr-1.5 text-emerald-500" />
        Complete
      </Button>
      <Button variant="ghost" size="sm">
        <X size={14} className="mr-1.5 text-muted-foreground" />
        Ignore
      </Button>
      <button onClick={onClear} className="ml-auto text-xs text-muted-foreground hover:text-foreground">
        Clear selection
      </button>
    </div>
  )
}

/* ─── Table ────────────────────────────────────────────────────────────── */

const COLS = '36px 90px 1fr 280px 64px 64px 110px 90px 36px'

function TableHeader({
  allChecked,
  onTogglePage,
}: {
  allChecked: boolean
  onTogglePage: () => void
}) {
  return (
    <div
      className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: COLS }}
    >
      <Checkbox checked={allChecked} onChange={onTogglePage} />
      <div>Severity</div>
      <div>Alert</div>
      <div>Source → Adversary</div>
      <div className="text-right">Risk</div>
      <div className="text-right">Echoes</div>
      <div>Status</div>
      <div className="text-right">Time</div>
      <div />
    </div>
  )
}

function AlertRow({
  alert,
  checked,
  onToggle,
  onOpen,
}: {
  alert: Alert
  checked: boolean
  onToggle: () => void
  onOpen: () => void
}) {
  const sev = SEVERITY_STYLE[alert.severity]
  return (
    <div
      onClick={onOpen}
      className="group grid cursor-pointer items-center gap-3 border-b border-border px-4 py-3 text-xs hover:bg-muted/40 last:border-b-0"
      style={{ gridTemplateColumns: COLS }}
    >
      <Checkbox
        checked={checked}
        onChange={(e) => {
          e?.stopPropagation()
          onToggle()
        }}
      />

      <div className="flex items-center gap-2">
        <span className={cn('h-3 w-1 rounded-full', sev.bar)} />
        <span
          className={cn(
            'rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset',
            sev.pill
          )}
        >
          {sev.label}
        </span>
      </div>

      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="truncate font-medium">{alert.name}</span>
          {alert.isIncident && (
            <span className="inline-flex items-center gap-1 rounded-md bg-amber-500/15 px-1.5 py-0.5 text-[10px] font-medium text-amber-600 ring-1 ring-inset ring-amber-500/30 dark:text-amber-300">
              <Flame size={10} /> Incident
            </span>
          )}
        </div>
        <div className="mt-0.5 flex items-center gap-2 text-[11px] text-muted-foreground">
          <span>{alert.technique.name}</span>
          <span>·</span>
          <span>{alert.sensor}</span>
        </div>
      </div>

      <FlowCell source={alert.source} adversary={alert.adversary} />

      <div className="text-right">
        <RiskBadge value={alert.risk} />
      </div>

      <div className="text-right font-mono tabular-nums text-muted-foreground">
        {alert.echoes > 0 ? alert.echoes.toLocaleString() : '—'}
      </div>

      <div>
        <span
          className={cn(
            'inline-flex items-center rounded-md px-2 py-0.5 text-[10px] font-medium ring-1 ring-inset',
            STATUS_META[alert.status].pill
          )}
        >
          {STATUS_META[alert.status].label}
        </span>
      </div>

      <div className="text-right">
        <div className="font-mono text-xs tabular-nums">{shortTime(alert.timestamp)}</div>
        <div className="font-mono text-[10px] text-muted-foreground">{relativeTime(alert.timestamp)} ago</div>
      </div>

      <div className="flex justify-end opacity-0 transition-opacity group-hover:opacity-100">
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

function FlowCell({ source, adversary }: { source: Endpoint; adversary: Endpoint }) {
  return (
    <div className="grid grid-cols-[1fr_auto_1fr] items-center gap-2 text-[11px]">
      <EndpointMini ep={source} />
      <ArrowRight size={12} className="text-muted-foreground" />
      <EndpointMini ep={adversary} accent />
    </div>
  )
}

function EndpointMini({ ep, accent }: { ep: Endpoint; accent?: boolean }) {
  return (
    <div className="min-w-0">
      <div className="flex items-center gap-1.5">
        {ep.countryCode && (
          <span aria-hidden className="text-sm leading-none">{flagEmoji(ep.countryCode)}</span>
        )}
        <span className={cn('truncate font-mono text-[11px]', accent && 'text-foreground')}>
          {ep.host ?? ep.ip}
        </span>
      </div>
      <div className="truncate text-[10px] text-muted-foreground">
        {ep.user ? `${ep.user} · ` : ''}
        {ep.host ? ep.ip : ep.country ?? '—'}
      </div>
    </div>
  )
}

function RiskBadge({ value }: { value: number }) {
  const tone =
    value >= 90
      ? 'bg-red-500/15 text-red-500 ring-red-500/30 dark:text-red-300'
      : value >= 70
        ? 'bg-orange-500/15 text-orange-600 ring-orange-500/30 dark:text-orange-300'
        : value >= 50
          ? 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300'
          : 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300'
  return (
    <span
      className={cn(
        'inline-flex w-9 items-center justify-center rounded-md px-1.5 py-0.5 font-mono text-[11px] font-medium ring-1 ring-inset tabular-nums',
        tone
      )}
    >
      {value}
    </span>
  )
}

/* ─── Pagination ───────────────────────────────────────────────────────── */

function PaginationBar({ total }: { total: number }) {
  return (
    <div className="mt-3 flex items-center justify-between text-xs text-muted-foreground">
      <div>
        Showing <span className="font-medium text-foreground">1–{Math.min(50, total)}</span> of{' '}
        <span className="font-medium text-foreground">{total.toLocaleString()}</span>
      </div>
      <div className="flex items-center gap-2">
        <span>50 / page</span>
        <span>·</span>
        <span>Page 1 of 1</span>
        <Button variant="outline" size="sm" disabled>
          <ChevronLeft size={14} />
        </Button>
        <Button variant="outline" size="sm" disabled>
          <ChevronRight size={14} />
        </Button>
      </div>
    </div>
  )
}

/* ─── Detail drawer ────────────────────────────────────────────────────── */

type Tab = 'summary' | 'parties' | 'logs' | 'history'

function AlertDrawer({ alert, onClose }: { alert: Alert; onClose: () => void }) {
  const [tab, setTab] = useState<Tab>('summary')
  const sev = SEVERITY_STYLE[alert.severity]

  return (
    <div
      className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="flex w-full max-w-[920px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                <span className="font-mono">#{alert.id}</span>
                <span>·</span>
                <span>{absTimestamp(alert.timestamp)}</span>
                <span>·</span>
                <span>{alert.sensor}</span>
              </div>
              <div className="mt-1 flex items-center gap-2">
                <span className={cn('h-5 w-1.5 rounded-full', sev.bar)} />
                <h2 className="truncate text-lg font-semibold">{alert.name}</h2>
              </div>
              <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                <span
                  className={cn(
                    'rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset',
                    sev.pill
                  )}
                >
                  {sev.label}
                </span>
                <span
                  className={cn(
                    'rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset',
                    STATUS_META[alert.status].pill
                  )}
                >
                  {STATUS_META[alert.status].label}
                </span>
                <span className="inline-flex items-center gap-1 rounded-md border border-border bg-background/40 px-1.5 py-0.5 font-mono">
                  <Crosshair size={10} className="text-rose-500" />
                  {alert.technique.id} {alert.technique.name}
                </span>
                <span className="inline-flex items-center gap-1 rounded-md border border-border bg-background/40 px-1.5 py-0.5">
                  <Database size={10} />
                  {alert.echoes.toLocaleString()} echoes
                </span>
                <RiskBadge value={alert.risk} />
              </div>
            </div>
            <button
              onClick={onClose}
              className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <X size={16} />
            </button>
          </div>

          {/* Quick actions */}
          <div className="mt-4 flex flex-wrap items-center gap-2">
            <Button size="sm">
              <Flame size={14} className="mr-1.5" />
              Mark as incident
            </Button>
            <Button size="sm" variant="outline">
              <BadgeCheck size={14} className="mr-1.5 text-emerald-500" />
              Complete
            </Button>
            <Button size="sm" variant="outline">
              <X size={14} className="mr-1.5" />
              Ignore
            </Button>
            <Button size="sm" variant="ghost" className="ml-auto">
              <MoreHorizontal size={14} />
            </Button>
          </div>
        </header>

        {/* Tabs */}
        <nav className="flex items-center gap-1 border-b border-border px-6">
          <DrawerTab id="summary" current={tab} onChange={setTab} icon={Sparkles}>
            Summary
          </DrawerTab>
          <DrawerTab id="parties" current={tab} onChange={setTab} icon={Network}>
            Source / Adversary
          </DrawerTab>
          <DrawerTab id="logs" current={tab} onChange={setTab} icon={Database}>
            Logs related
            <span className="ml-1.5 rounded bg-muted px-1.5 py-0.5 font-mono text-[10px]">
              {alert.echoes}
            </span>
          </DrawerTab>
          <DrawerTab id="history" current={tab} onChange={setTab} icon={History}>
            History
          </DrawerTab>
        </nav>

        <div className="flex-1 overflow-y-auto bg-muted/20">
          {tab === 'summary' && <SummaryTab alert={alert} />}
          {tab === 'parties' && <PartiesTab alert={alert} />}
          {tab === 'logs' && <LogsTab />}
          {tab === 'history' && <HistoryTab />}
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

/* ─── Drawer tabs ──────────────────────────────────────────────────────── */

function SummaryTab({ alert }: { alert: Alert }) {
  return (
    <div className="grid grid-cols-3 gap-4 p-6">
      <div className="col-span-2 space-y-4">
        <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-4">
          <div className="mb-2 flex items-center gap-2 text-xs font-medium">
            <Sparkles size={14} className="text-fuchsia-500" />
            SOC AI summary
          </div>
          <p className="text-sm leading-relaxed">
            Service account <span className="font-mono">{alert.source.user ?? alert.source.host}</span> on{' '}
            <span className="font-mono">{alert.source.host ?? alert.source.ip}</span> requested unusually
            many Kerberos service tickets in a short window — a strong signal of <strong>Kerberoasting</strong>.
            This activity correlates with <span className="font-medium">8 other alerts</span> attributed
            to the <span className="font-medium">Scattered Spider</span> campaign in the last 4 hours.
          </p>
          <div className="mt-3 flex flex-wrap gap-2 text-[11px]">
            <button className="rounded-md border border-border bg-background/60 px-2 py-1 hover:bg-background">
              Build detection rule
            </button>
            <button className="rounded-md border border-border bg-background/60 px-2 py-1 hover:bg-background">
              Hunt similar
            </button>
            <button className="rounded-md border border-border bg-background/60 px-2 py-1 hover:bg-background">
              Open in Agent
            </button>
          </div>
        </div>

        <Section title="Description">
          <p className="text-sm text-foreground/90">
            Detected an abnormal pattern of Kerberos TGS-REQ messages from a single host, requesting
            service tickets for several SPNs registered under privileged accounts. The frequency and SPN
            mix suggest credential extraction for offline cracking.
          </p>
        </Section>

        <Section title="Proposed solution">
          <ol className="list-decimal pl-5 text-sm text-foreground/90 space-y-1">
            <li>Disable the affected service account immediately if business-critical workflows allow.</li>
            <li>Rotate Kerberos krbtgt twice within 10 hours to invalidate forged tickets.</li>
            <li>Force AES-only encryption for service accounts and increase ticket complexity.</li>
            <li>Hunt for ticket dumping tools (Rubeus, Mimikatz) on related hosts.</li>
          </ol>
        </Section>
      </div>

      <aside className="col-span-1 space-y-4">
        <KeyValueCard
          title="Detail"
          rows={[
            ['Rule', alert.name],
            ['ID', <span className="font-mono">{alert.id}</span>],
            ['Category', alert.category],
            ['Technique', `${alert.technique.id} ${alert.technique.name}`],
            ['Sensor', alert.sensor],
            ['Generated', absTimestamp(alert.timestamp)],
          ]}
        />

        <div className="rounded-lg border border-border bg-card p-4">
          <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">Impact</div>
          <ImpactBars confidentiality={0.92} integrity={0.74} availability={0.46} />
        </div>
      </aside>
    </div>
  )
}

function PartiesTab({ alert }: { alert: Alert }) {
  return (
    <div className="grid grid-cols-2 gap-4 p-6">
      <PartyCard title="Source" icon={Server} ep={alert.source} accent="text-sky-500" />
      <PartyCard title="Adversary" icon={UserX} ep={alert.adversary} accent="text-red-500" />
      <div className="col-span-2 rounded-lg border border-border bg-card p-4">
        <div className="mb-3 flex items-center justify-between">
          <h4 className="text-sm font-semibold">Geo</h4>
          <span className="text-[11px] text-muted-foreground">Approximate location based on IP</span>
        </div>
        <FauxMap source={alert.source} adversary={alert.adversary} />
      </div>
    </div>
  )
}

function PartyCard({
  title,
  icon: Icon,
  ep,
  accent,
}: {
  title: string
  icon: LucideIcon
  ep: Endpoint
  accent: string
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-3 flex items-center gap-2">
        <Icon size={14} className={accent} />
        <h4 className="text-sm font-semibold">{title}</h4>
        {ep.countryCode && (
          <span aria-hidden className="ml-auto text-lg leading-none">{flagEmoji(ep.countryCode)}</span>
        )}
      </div>
      <dl className="grid grid-cols-[100px_1fr] gap-y-2 text-xs">
        <Row k="IP">
          <span className="font-mono">{ep.ip}</span>
        </Row>
        {ep.host && (
          <Row k="Host">
            <span className="font-mono">{ep.host}</span>
          </Row>
        )}
        {ep.user && (
          <Row k="User">
            <span className="font-mono">{ep.user}</span>
          </Row>
        )}
        {ep.country && (
          <Row k="Country">
            <span>{ep.country}</span>
          </Row>
        )}
        {accent === 'text-red-500' && ep.country && ep.country !== undefined && (
          <Row k="Reputation">
            <span className="inline-flex items-center gap-1.5 rounded bg-red-500/15 px-1.5 py-0.5 text-[10px] text-red-600 ring-1 ring-red-500/30 dark:text-red-300">
              <Target size={10} />
              Known C2 (AlienVault)
            </span>
          </Row>
        )}
      </dl>
    </div>
  )
}

function Row({ k, children }: { k: string; children: React.ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd>{children}</dd>
    </>
  )
}

function FauxMap({ source, adversary }: { source: Endpoint; adversary: Endpoint }) {
  return (
    <div className="relative h-44 overflow-hidden rounded-md border border-border bg-gradient-to-b from-sky-500/5 to-violet-500/10">
      <svg viewBox="0 0 800 200" className="h-full w-full">
        <defs>
          <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
            <path d="M 40 0 L 0 0 0 40" className="fill-none stroke-border/30" strokeWidth="1" />
          </pattern>
        </defs>
        <rect width="800" height="200" fill="url(#grid)" />
        <path
          d="M 540 70 Q 400 -40 220 110"
          fill="none"
          className="stroke-red-500/70"
          strokeWidth="1.5"
          strokeDasharray="3 3"
        />
        <circle cx="220" cy="110" r="6" className="fill-sky-500" />
        <circle cx="220" cy="110" r="14" className="fill-sky-500/20" />
        <text x="220" y="135" textAnchor="middle" fontSize="10" className="fill-foreground">
          {source.countryCode ?? source.host ?? source.ip}
        </text>
        <circle cx="540" cy="70" r="6" className="fill-red-500" />
        <circle cx="540" cy="70" r="14" className="fill-red-500/20 animate-pulse" />
        <text x="540" y="55" textAnchor="middle" fontSize="10" className="fill-foreground">
          {adversary.countryCode ?? adversary.host ?? adversary.ip}
        </text>
      </svg>
    </div>
  )
}

function LogsTab() {
  const logs = [
    { ts: min(4), src: 'win-evt', msg: 'TGS-REQ ticket request for SPN MSSQLSvc/sql-prod-01 from svc_backup' },
    { ts: min(4), src: 'win-evt', msg: 'TGS-REQ ticket request for SPN HOST/dc-01 from svc_backup' },
    { ts: min(5), src: 'sysmon', msg: 'Process create: powershell.exe -enc <encoded>' },
    { ts: min(6), src: 'sysmon', msg: 'Network connection: 10.4.18.22 → 198.51.100.14:443' },
    { ts: min(7), src: 'win-evt', msg: 'Successful logon for svc_backup from 10.4.18.91 (type 3)' },
  ]
  return (
    <div className="p-6">
      <div className="overflow-hidden rounded-lg border border-border bg-card">
        <div className="grid grid-cols-[160px_120px_1fr] gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
          <div>Timestamp</div>
          <div>Source</div>
          <div>Message</div>
        </div>
        {logs.map((l, i) => (
          <div
            key={i}
            className="grid grid-cols-[160px_120px_1fr] gap-3 border-b border-border px-4 py-2 text-xs last:border-b-0 hover:bg-muted/40"
          >
            <div className="font-mono text-[11px] text-muted-foreground">{absTimestamp(l.ts)}</div>
            <div className="text-muted-foreground">{l.src}</div>
            <div className="truncate font-mono text-[11px]">{l.msg}</div>
          </div>
        ))}
      </div>
    </div>
  )
}

function HistoryTab() {
  const events = [
    { ts: min(2), who: 'jsmith@utmstack.com', what: 'changed status to In review' },
    { ts: min(3), who: 'system', what: 'auto-tagged: Scattered Spider' },
    { ts: min(3), who: 'system', what: 'linked to incident #1142' },
    { ts: min(4), who: 'system', what: 'alert generated · risk 96' },
  ]
  return (
    <div className="p-6">
      <ol className="relative space-y-4 border-l border-border pl-6">
        {events.map((e, i) => (
          <li key={i} className="relative">
            <span className="absolute -left-[27px] top-1 flex h-3 w-3 items-center justify-center rounded-full border-2 border-card bg-primary" />
            <div className="text-xs">
              <span className="font-medium">{e.who}</span> {e.what}
            </div>
            <div className="text-[10px] text-muted-foreground">{absTimestamp(e.ts)}</div>
          </li>
        ))}
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
      <dl className="grid grid-cols-[80px_1fr] gap-y-2 text-xs">
        {rows.map(([k, v]) => (
          <Row key={k} k={k}>
            {v}
          </Row>
        ))}
      </dl>
    </div>
  )
}

function ImpactBars({
  confidentiality,
  integrity,
  availability,
}: {
  confidentiality: number
  integrity: number
  availability: number
}) {
  const items: [string, number, string][] = [
    ['Confidentiality', confidentiality, 'from-red-500 to-rose-500'],
    ['Integrity', integrity, 'from-orange-500 to-amber-500'],
    ['Availability', availability, 'from-sky-500 to-violet-500'],
  ]
  return (
    <ul className="space-y-2">
      {items.map(([label, v, grad]) => (
        <li key={label}>
          <div className="mb-1 flex items-center justify-between text-[11px]">
            <span>{label}</span>
            <span className="font-mono tabular-nums text-muted-foreground">
              {Math.round(v * 100)}
            </span>
          </div>
          <div className="h-1.5 overflow-hidden rounded-full bg-muted">
            <div
              className={cn('h-full rounded-full bg-gradient-to-r', grad)}
              style={{ width: `${v * 100}%` }}
            />
          </div>
        </li>
      ))}
    </ul>
  )
}

function Checkbox({
  checked,
  onChange,
}: {
  checked: boolean
  onChange: (e?: React.MouseEvent) => void
}) {
  return (
    <button
      onClick={(e) => {
        e.stopPropagation()
        onChange(e)
      }}
      className={cn(
        'flex h-4 w-4 items-center justify-center rounded border transition-colors',
        checked
          ? 'border-primary bg-primary text-primary-foreground'
          : 'border-border hover:border-foreground/50'
      )}
    >
      {checked && (
        <svg viewBox="0 0 16 16" className="h-3 w-3" fill="none">
          <path
            d="M3.5 8.5l3 3 6-6"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      )}
    </button>
  )
}

/* ─── Helpers ──────────────────────────────────────────────────────────── */

function relativeTime(iso: string) {
  const diff = NOW - new Date(iso).getTime()
  const m = Math.round(diff / 60_000)
  if (m < 1) return 'just now'
  if (m < 60) return `${m}m`
  const h = Math.round(m / 60)
  if (h < 24) return `${h}h`
  return `${Math.round(h / 24)}d`
}

function shortTime(iso: string) {
  return new Date(iso).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
}

function absTimestamp(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleString(undefined, {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
  })
}

function flagEmoji(cc: string) {
  if (!cc || cc.length !== 2) return ''
  const A = 0x1f1e6
  return String.fromCodePoint(
    A - 65 + cc.toUpperCase().charCodeAt(0),
    A - 65 + cc.toUpperCase().charCodeAt(1),
  )
}

// Suppress the Network unused-import warning if any
void Network
