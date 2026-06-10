import { Fragment, useMemo, useState } from 'react'
import {
  AlertTriangle,
  ArrowRight,
  BadgeCheck,
  Bookmark,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Clock,
  Columns3,
  Crosshair,
  Download,
  Flame,
  GitMerge,
  History,
  LayoutGrid,
  ListFilter,
  ListIcon,
  MessageSquare,
  MoreHorizontal,
  Network,
  Plus,
  RefreshCw,
  Search,
  Send,
  ShieldAlert,
  Sparkles,
  Tag as TagIcon,
  Timer,
  TrendingUp,
  UserCircle2,
  Users,
  Workflow,
  X,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type Severity = 'critical' | 'high' | 'medium' | 'low'
type Status = 'open' | 'in_review' | 'completed' | 'merged'

interface Analyst {
  id: string
  name: string
  initials: string
  tone: string
}

interface Incident {
  id: number
  name: string
  description: string
  solution: string
  severity: Severity
  status: Status
  createdAt: string
  updatedAt: string
  assignee: Analyst | null
  alertsCount: number
  affectedAssets: number
  tags: string[]
  mitre: { id: string; name: string }[]
  sla: { dueAt: string; breached: boolean }
  origin: 'manual' | 'auto-grouped' | 'playbook'
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const NOW = new Date('2026-05-01T16:42:00Z').getTime()
const min = (m: number) => new Date(NOW - m * 60_000).toISOString()
const future = (m: number) => new Date(NOW + m * 60_000).toISOString()

const ANALYSTS: Analyst[] = [
  { id: 'jsmith', name: 'Jordan Smith', initials: 'JS', tone: 'bg-sky-500' },
  { id: 'mthomas', name: 'Mara Thomas', initials: 'MT', tone: 'bg-emerald-500' },
  { id: 'tferguson', name: 'Tomás Ferguson', initials: 'TF', tone: 'bg-amber-500' },
  { id: 'aluna', name: 'Ana Luna', initials: 'AL', tone: 'bg-violet-500' },
]

const INCIDENTS: Incident[] = [
  {
    id: 1142,
    name: 'Scattered Spider campaign — DC compromise',
    description:
      'Service account credential extraction followed by lateral movement. 8 alerts grouped by adversary IP and shared technique chain. Contains TGS-REQ burst from WIN-7F3K2L9Q8X.',
    solution:
      'Disable affected accounts, rotate krbtgt, hunt for ticket dumping tools on related hosts, isolate WIN-7F3K2L9Q8X from the network.',
    severity: 'critical',
    status: 'in_review',
    createdAt: min(46),
    updatedAt: min(2),
    assignee: ANALYSTS[0],
    alertsCount: 8,
    affectedAssets: 4,
    tags: ['Scattered Spider', 'campaign-44', 'lateral-movement'],
    mitre: [
      { id: 'T1558.003', name: 'Kerberoasting' },
      { id: 'T1059.001', name: 'PowerShell' },
      { id: 'T1071.001', name: 'Web Protocols' },
    ],
    sla: { dueAt: future(74), breached: false },
    origin: 'auto-grouped',
  },
  {
    id: 1141,
    name: 'Unusual exfiltration to external Dutch IP',
    description:
      'Sustained outbound traffic from fin-app-01 to 45.83.66.12 over the last 4 hours. Volume 18× the 30-day baseline.',
    solution:
      'Block destination at perimeter, snapshot fin-app-01, review TLS proxy logs for unencrypted artifacts.',
    severity: 'high',
    status: 'open',
    createdAt: min(74),
    updatedAt: min(74),
    assignee: null,
    alertsCount: 3,
    affectedAssets: 1,
    tags: ['exfiltration', 'untriaged'],
    mitre: [{ id: 'T1041', name: 'C2 Exfiltration' }],
    sla: { dueAt: future(20), breached: false },
    origin: 'auto-grouped',
  },
  {
    id: 1140,
    name: 'Phishing wave targeting Finance OU',
    description:
      '14 employees received Microsoft 365 credential harvesting emails. 3 clicked, 1 entered credentials.',
    solution:
      'Reset compromised credentials, enable phish-resistant MFA on Finance OU, push awareness reminder.',
    severity: 'high',
    status: 'in_review',
    createdAt: min(180),
    updatedAt: min(45),
    assignee: ANALYSTS[2],
    alertsCount: 14,
    affectedAssets: 3,
    tags: ['phishing'],
    mitre: [{ id: 'T1566.002', name: 'Spearphishing Link' }],
    sla: { dueAt: future(-15), breached: true },
    origin: 'manual',
  },
  {
    id: 1139,
    name: 'Suspicious GPO modification by non-admin',
    description:
      'tferguson@corp made a GPO modification outside change-window. Action reversed within 2 minutes.',
    solution:
      'Confirmed authorized — change ticket CHG-2231 was misrouted. No action needed beyond documentation.',
    severity: 'medium',
    status: 'completed',
    createdAt: min(420),
    updatedAt: min(120),
    assignee: ANALYSTS[3],
    alertsCount: 1,
    affectedAssets: 1,
    tags: ['policy', 'benign'],
    mitre: [{ id: 'T1484', name: 'Domain Policy Modification' }],
    sla: { dueAt: future(-200), breached: false },
    origin: 'manual',
  },
  {
    id: 1138,
    name: 'Failed brute force from external Brazilian IP',
    description:
      '312 failed logins against mail-srv-02 from 203.0.113.7 culminating in one successful auth for mthomas.',
    solution: 'Reset mthomas password, force logout of all sessions, enable conditional access geo-block on inbound auth.',
    severity: 'high',
    status: 'in_review',
    createdAt: min(28),
    updatedAt: min(8),
    assignee: ANALYSTS[1],
    alertsCount: 5,
    affectedAssets: 2,
    tags: ['brute-force', 'external'],
    mitre: [{ id: 'T1110', name: 'Brute Force' }],
    sla: { dueAt: future(95), breached: false },
    origin: 'auto-grouped',
  },
  {
    id: 1137,
    name: 'Routine FP burst from Rule #992',
    description: 'Alert rule "outbound DNS to dynamic domain" generated 47 false positives in 5 minutes.',
    solution: 'Tuned rule to exclude internal CDN domain. Monitoring for re-occurrence.',
    severity: 'low',
    status: 'merged',
    createdAt: min(900),
    updatedAt: min(700),
    assignee: ANALYSTS[0],
    alertsCount: 47,
    affectedAssets: 0,
    tags: ['rule-tuning', 'merged-into-1136'],
    mitre: [],
    sla: { dueAt: future(-700), breached: false },
    origin: 'auto-grouped',
  },
  {
    id: 1136,
    name: 'New scheduled task on critical asset',
    description: 'Persistence indicator on fin-app-01. Task points to legitimate update script.',
    solution: 'Confirmed via change-management. Whitelist added.',
    severity: 'medium',
    status: 'open',
    createdAt: min(58),
    updatedAt: min(58),
    assignee: null,
    alertsCount: 1,
    affectedAssets: 1,
    tags: [],
    mitre: [{ id: 'T1053.005', name: 'Scheduled Task' }],
    sla: { dueAt: future(180), breached: false },
    origin: 'auto-grouped',
  },
  {
    id: 1135,
    name: 'Mimikatz signature on workstation',
    description:
      'Defender flagged Mimikatz variant on WIN-7F3K2L9Q8X. Likely component of campaign-44.',
    solution: 'Isolate host, full forensic capture, link to incident #1142.',
    severity: 'critical',
    status: 'open',
    createdAt: min(62),
    updatedAt: min(62),
    assignee: null,
    alertsCount: 1,
    affectedAssets: 1,
    tags: ['malware', 'campaign-44'],
    mitre: [{ id: 'T1003', name: 'OS Credential Dumping' }],
    sla: { dueAt: future(58), breached: false },
    origin: 'auto-grouped',
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

const STATUS_META: Record<
  Status,
  { label: string; pill: string; dot: string; column: string }
> = {
  open: {
    label: 'Open',
    pill: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300',
    dot: 'bg-red-500',
    column: 'border-red-500/30',
  },
  in_review: {
    label: 'In review',
    pill: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300',
    dot: 'bg-sky-500',
    column: 'border-sky-500/30',
  },
  completed: {
    label: 'Completed',
    pill: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300',
    dot: 'bg-emerald-500',
    column: 'border-emerald-500/30',
  },
  merged: {
    label: 'Merged',
    pill: 'bg-violet-500/15 text-violet-600 ring-violet-500/30 dark:text-violet-300',
    dot: 'bg-violet-500',
    column: 'border-violet-500/30',
  },
}

const STATUS_ORDER: Status[] = ['open', 'in_review', 'completed', 'merged']

/* ─── Saved views ──────────────────────────────────────────────────────── */

interface SavedView {
  id: string
  label: string
  count: number
  predicate: (i: Incident) => boolean
}

const SAVED_VIEWS: SavedView[] = [
  { id: 'all', label: 'All incidents', count: INCIDENTS.length, predicate: () => true },
  { id: 'mine', label: 'My queue', count: INCIDENTS.filter((i) => i.assignee?.id === 'jsmith').length, predicate: (i) => i.assignee?.id === 'jsmith' },
  { id: 'unassigned', label: 'Unassigned', count: INCIDENTS.filter((i) => !i.assignee).length, predicate: (i) => !i.assignee },
  { id: 'open', label: 'Open', count: INCIDENTS.filter((i) => i.status === 'open').length, predicate: (i) => i.status === 'open' },
  { id: 'critical', label: 'Critical', count: INCIDENTS.filter((i) => i.severity === 'critical').length, predicate: (i) => i.severity === 'critical' },
  { id: 'sla', label: 'SLA at risk', count: INCIDENTS.filter((i) => i.sla.breached || (new Date(i.sla.dueAt).getTime() - NOW) < 60 * 60_000).length, predicate: (i) => i.sla.breached || (new Date(i.sla.dueAt).getTime() - NOW) < 60 * 60_000 },
]

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function IncidentsPage() {
  const [view, setView] = useState<string>('all')
  const [search, setSearch] = useState('')
  const [layout, setLayout] = useState<'table' | 'board'>('table')
  const [openIncident, setOpenIncident] = useState<Incident | null>(null)

  const filtered = useMemo(() => {
    const v = SAVED_VIEWS.find((s) => s.id === view) ?? SAVED_VIEWS[0]
    return INCIDENTS.filter(v.predicate).filter((i) =>
      search ? (i.name + i.id + i.tags.join(' ')).toLowerCase().includes(search.toLowerCase()) : true
    )
  }, [view, search])

  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 py-6">
      <Header total={filtered.length} />

      <div className="mt-5 grid grid-cols-12 gap-4">
        <div className="col-span-12 lg:col-span-7">
          <WorkflowFlowCard />
        </div>
        <div className="col-span-12 lg:col-span-5">
          <ResponseSummaryCard />
        </div>
      </div>

      <div className="mt-5">
        <SavedViewTabs current={view} onChange={setView} />
      </div>

      <Toolbar search={search} onSearch={setSearch} layout={layout} onLayoutChange={setLayout} />

      <ActiveFilterChips />

      {layout === 'table' ? (
        <>
          <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
            <TableHeader />
            <div>
              {filtered.map((i) => (
                <IncidentRow key={i.id} incident={i} onOpen={() => setOpenIncident(i)} />
              ))}
              {filtered.length === 0 && (
                <div className="px-6 py-16 text-center text-sm text-muted-foreground">
                  No incidents match this view.
                </div>
              )}
            </div>
          </div>
          <PaginationBar total={filtered.length} />
        </>
      ) : (
        <BoardView incidents={filtered} onOpen={setOpenIncident} />
      )}

      {openIncident && <IncidentDrawer incident={openIncident} onClose={() => setOpenIncident(null)} />}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ total }: { total: number }) {
  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold tracking-tight">Incidents</h1>
          <span className="rounded-md bg-muted px-2 py-0.5 font-mono text-[11px] tabular-nums text-muted-foreground">
            {total.toLocaleString()} matching
          </span>
        </div>
        <p className="mt-1 text-xs text-muted-foreground">
          Cases under investigation. Each groups one or more correlated alerts and tracks resolution.
        </p>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm">
          <Workflow size={14} className="mr-2" />
          Playbooks
        </Button>
        <Button variant="outline" size="sm">
          <Bookmark size={14} className="mr-2" />
          Save view
        </Button>
        <Button variant="outline" size="sm">
          <Download size={14} className="mr-2" />
          Export
        </Button>
        <Button variant="default" size="sm">
          <Plus size={14} className="mr-2" />
          New incident
        </Button>
      </div>
    </header>
  )
}

/* ─── Workflow flow card ───────────────────────────────────────────────── */

function WorkflowFlowCard() {
  const counts = STATUS_ORDER.map((s) => ({
    s,
    n: INCIDENTS.filter((i) => i.status === s).length,
  }))
  return (
    <div className="rounded-xl border border-border bg-card p-5">
      <div className="flex items-center justify-between">
        <h3 className="flex items-center gap-2 text-sm font-semibold">
          <Workflow size={15} className="text-violet-500" />
          Response workflow
        </h3>
        <span className="text-[10px] uppercase tracking-wider text-muted-foreground">last 7 days</span>
      </div>
      <div className="mt-5 grid grid-cols-[1fr_auto_1fr_auto_1fr_auto_1fr] items-center gap-2">
        {counts.map((c, idx) => (
          <Fragment key={c.s}>
            <div
              className={cn(
                'flex flex-col items-center gap-1.5 rounded-lg border bg-background/40 p-3',
                STATUS_META[c.s].column
              )}
            >
              <div className="flex items-center gap-1.5 text-[11px] uppercase tracking-wider text-muted-foreground">
                <span className={cn('h-2 w-2 rounded-full', STATUS_META[c.s].dot)} />
                {STATUS_META[c.s].label}
              </div>
              <div className="text-2xl font-semibold tabular-nums">{c.n}</div>
              <div className="text-[10px] text-muted-foreground">
                {c.s === 'open' && 'awaiting triage'}
                {c.s === 'in_review' && 'analyst working'}
                {c.s === 'completed' && 'resolved'}
                {c.s === 'merged' && 'rolled up'}
              </div>
            </div>
            {idx < counts.length - 1 && (
              <ArrowRight size={16} className="text-muted-foreground" />
            )}
          </Fragment>
        ))}
      </div>
      <div className="mt-4 grid grid-cols-3 gap-3 border-t border-border pt-4 text-xs">
        <FlowStat label="MTTR" value="2h 14m" trend="-18%" up />
        <FlowStat label="MTTA" value="11m" trend="-23%" up />
        <FlowStat label="Re-opened" value="3" trend="+1" up={false} />
      </div>
    </div>
  )
}

function FlowStat({
  label,
  value,
  trend,
  up,
}: {
  label: string
  value: string
  trend: string
  up: boolean
}) {
  return (
    <div className="flex items-center justify-between rounded-md border border-border bg-background/40 px-3 py-2">
      <div>
        <div className="text-[10px] uppercase tracking-wider text-muted-foreground">{label}</div>
        <div className="text-base font-semibold tabular-nums">{value}</div>
      </div>
      <span
        className={cn(
          'rounded px-1.5 py-0.5 text-[10px] font-medium',
          up ? 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-300' : 'bg-red-500/15 text-red-500 dark:text-red-300'
        )}
      >
        <TrendingUp size={10} className="inline" /> {trend}
      </span>
    </div>
  )
}

/* ─── Response summary ─────────────────────────────────────────────────── */

function ResponseSummaryCard() {
  return (
    <div className="flex h-full flex-col gap-3 rounded-xl border border-border bg-card p-5">
      <div className="flex items-center justify-between">
        <h3 className="flex items-center gap-2 text-sm font-semibold">
          <Sparkles size={15} className="text-fuchsia-500" />
          What needs your attention
        </h3>
      </div>

      <div className="rounded-lg border border-red-500/30 bg-gradient-to-br from-red-500/10 via-rose-500/5 to-transparent p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">1 incident</span> just breached SLA —{' '}
        <span className="font-medium">#1140 Phishing wave</span>. Owner: Tomás Ferguson.{' '}
        <button className="text-primary hover:underline">Escalate</button>
      </div>

      <div className="rounded-lg border border-amber-500/30 bg-amber-500/5 p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">3 incidents</span> remain unassigned. Auto-routing
        suggests <span className="font-medium">Mara Thomas</span> based on technique.{' '}
        <button className="text-primary hover:underline">Assign all</button>
      </div>

      <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">Incidents #1142 and #1135</span> share the same
        adversary IP. Likely same campaign — consider merging.{' '}
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
            {active && (
              <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />
            )}
          </button>
        )
      })}
      <button className="ml-auto flex items-center gap-1 px-2 py-2 text-[11px] text-muted-foreground hover:text-foreground">
        <Bookmark size={12} />
        Save current as view
      </button>
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
  layout: 'table' | 'board'
  onLayoutChange: (l: 'table' | 'board') => void
}) {
  return (
    <div className="mt-3 flex flex-wrap items-center gap-2">
      <div className="relative min-w-[280px] flex-1">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search incident name, ID, tag…"
          value={search}
          onChange={(e) => onSearch(e.target.value)}
          className="h-9 pl-9 pr-24"
        />
      </div>
      <Button variant="outline" size="sm">
        <Clock size={14} className="mr-2" />
        Last 7 days
        <ChevronDown size={12} className="ml-1.5 opacity-60" />
      </Button>
      <Button variant="outline" size="sm">
        <UserCircle2 size={14} className="mr-2" />
        Assignee
        <ChevronDown size={12} className="ml-1.5 opacity-60" />
      </Button>
      <Button variant="outline" size="sm">
        <ListFilter size={14} className="mr-2" />
        Filters
        <span className="ml-2 rounded-md bg-primary/15 px-1.5 py-0.5 font-mono text-[10px] text-primary">2</span>
      </Button>

      <div className="ml-auto flex items-center gap-1 rounded-md border border-border p-0.5">
        <button
          onClick={() => onLayoutChange('table')}
          title="Table view"
          className={cn(
            'flex h-7 w-7 items-center justify-center rounded transition-colors',
            layout === 'table' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'
          )}
        >
          <ListIcon size={13} />
        </button>
        <button
          onClick={() => onLayoutChange('board')}
          title="Board view"
          className={cn(
            'flex h-7 w-7 items-center justify-center rounded transition-colors',
            layout === 'board' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'
          )}
        >
          <LayoutGrid size={13} />
        </button>
      </div>
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
  )
}

function ActiveFilterChips() {
  const chips = [
    { label: 'severity', value: 'critical, high' },
    { label: 'time', value: 'last 7d' },
  ]
  return (
    <div className="mt-3 flex flex-wrap items-center gap-1.5">
      {chips.map((c) => (
        <span
          key={c.label}
          className="inline-flex items-center gap-1.5 rounded-md border border-border bg-card px-2 py-1 text-[11px]"
        >
          <span className="text-muted-foreground">{c.label}:</span>
          <span className="font-medium">{c.value}</span>
          <button className="text-muted-foreground hover:text-foreground">
            <X size={11} />
          </button>
        </span>
      ))}
      <button className="px-2 py-1 text-[11px] text-muted-foreground hover:text-foreground">
        Clear all
      </button>
    </div>
  )
}

/* ─── Table view ───────────────────────────────────────────────────────── */

const COLS = '60px 90px 1fr 110px 130px 90px 100px 100px 36px'

function TableHeader() {
  return (
    <div
      className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: COLS }}
    >
      <div>ID</div>
      <div>Severity</div>
      <div>Incident</div>
      <div>Status</div>
      <div>Assignee</div>
      <div className="text-right">Alerts</div>
      <div>SLA</div>
      <div>Updated</div>
      <div />
    </div>
  )
}

function IncidentRow({ incident, onOpen }: { incident: Incident; onOpen: () => void }) {
  const sev = SEVERITY_STYLE[incident.severity]
  const st = STATUS_META[incident.status]
  return (
    <div
      onClick={onOpen}
      className="group grid cursor-pointer items-center gap-3 border-b border-border px-4 py-3 text-xs hover:bg-muted/40 last:border-b-0"
      style={{ gridTemplateColumns: COLS }}
    >
      <div className="font-mono tabular-nums text-muted-foreground">#{incident.id}</div>

      <div className="flex items-center gap-2">
        <span className={cn('h-3 w-1 rounded-full', sev.bar)} />
        <span className={cn('rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset', sev.pill)}>
          {sev.label}
        </span>
      </div>

      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="truncate font-medium">{incident.name}</span>
          {incident.origin === 'auto-grouped' && (
            <span className="inline-flex items-center gap-1 rounded border border-border bg-background/40 px-1 py-0.5 text-[10px] text-muted-foreground">
              <Sparkles size={9} /> auto
            </span>
          )}
        </div>
        <div className="mt-0.5 flex flex-wrap items-center gap-1.5 text-[11px] text-muted-foreground">
          {incident.mitre.slice(0, 2).map((m) => (
            <span key={m.id} className="inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5 font-mono text-[10px]">
              {m.id}
            </span>
          ))}
          {incident.mitre.length > 2 && (
            <span className="text-[10px]">+{incident.mitre.length - 2} more</span>
          )}
          {incident.tags.length > 0 && (
            <>
              <span>·</span>
              {incident.tags.slice(0, 2).map((t) => (
                <span key={t} className="rounded bg-muted px-1.5 py-0.5 text-[10px]">
                  {t}
                </span>
              ))}
            </>
          )}
        </div>
      </div>

      <div>
        <span className={cn('inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 text-[10px] font-medium ring-1 ring-inset', st.pill)}>
          <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
          {st.label}
        </span>
      </div>

      <div>
        <AssigneeChip analyst={incident.assignee} />
      </div>

      <div className="text-right tabular-nums">
        <span className="inline-flex items-center gap-1">
          <ShieldAlert size={11} className="text-orange-500" />
          {incident.alertsCount}
        </span>
      </div>

      <div>
        <SLAIndicator sla={incident.sla} status={incident.status} />
      </div>

      <div className="font-mono text-[11px] text-muted-foreground">{relativeTime(incident.updatedAt)}</div>

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

function AssigneeChip({ analyst }: { analyst: Analyst | null }) {
  if (!analyst) {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-md border border-dashed border-border px-2 py-0.5 text-[11px] text-muted-foreground">
        <UserCircle2 size={12} />
        Unassigned
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1.5">
      <span className={cn('flex h-5 w-5 items-center justify-center rounded-full text-[9px] font-semibold text-white', analyst.tone)}>
        {analyst.initials}
      </span>
      <span className="truncate text-[11px]">{analyst.name}</span>
    </span>
  )
}

function SLAIndicator({ sla, status }: { sla: Incident['sla']; status: Status }) {
  if (status === 'completed' || status === 'merged') {
    return <span className="text-[11px] text-muted-foreground">—</span>
  }
  const ms = new Date(sla.dueAt).getTime() - NOW
  const minLeft = Math.round(ms / 60_000)
  if (sla.breached || minLeft < 0) {
    return (
      <span className="inline-flex items-center gap-1 rounded bg-red-500/15 px-1.5 py-0.5 text-[10px] font-medium text-red-600 ring-1 ring-red-500/30 dark:text-red-300">
        <Timer size={10} />
        breached {fmtDur(-minLeft)}
      </span>
    )
  }
  if (minLeft < 60) {
    return (
      <span className="inline-flex items-center gap-1 rounded bg-amber-500/15 px-1.5 py-0.5 text-[10px] font-medium text-amber-600 ring-1 ring-amber-500/30 dark:text-amber-300">
        <Timer size={10} />
        {fmtDur(minLeft)} left
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1 text-[11px] text-muted-foreground">
      <Timer size={10} />
      {fmtDur(minLeft)} left
    </span>
  )
}

function fmtDur(min: number) {
  if (min < 60) return `${min}m`
  const h = Math.round(min / 60)
  if (h < 24) return `${h}h`
  return `${Math.round(h / 24)}d`
}

/* ─── Board view ───────────────────────────────────────────────────────── */

function BoardView({
  incidents,
  onOpen,
}: {
  incidents: Incident[]
  onOpen: (i: Incident) => void
}) {
  const grouped = STATUS_ORDER.map((s) => ({
    s,
    items: incidents.filter((i) => i.status === s),
  }))

  return (
    <div className="mt-3 grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-4">
      {grouped.map(({ s, items }) => {
        const meta = STATUS_META[s]
        return (
          <div key={s} className={cn('flex flex-col rounded-xl border bg-muted/20', meta.column)}>
            <div className="flex items-center justify-between border-b border-border/60 px-3 py-2">
              <div className="flex items-center gap-2 text-xs font-semibold">
                <span className={cn('h-2 w-2 rounded-full', meta.dot)} />
                {meta.label}
                <span className="rounded bg-muted px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground">
                  {items.length}
                </span>
              </div>
              <button className="flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-background hover:text-foreground">
                <Plus size={12} />
              </button>
            </div>
            <div className="flex flex-col gap-2 p-2">
              {items.map((i) => (
                <BoardCard key={i.id} incident={i} onOpen={() => onOpen(i)} />
              ))}
              {items.length === 0 && (
                <div className="rounded-lg border border-dashed border-border bg-background/30 p-4 text-center text-[11px] text-muted-foreground">
                  Nothing here
                </div>
              )}
            </div>
          </div>
        )
      })}
    </div>
  )
}

function BoardCard({ incident, onOpen }: { incident: Incident; onOpen: () => void }) {
  const sev = SEVERITY_STYLE[incident.severity]
  return (
    <button
      onClick={onOpen}
      className="group flex flex-col gap-2 rounded-lg border border-border bg-card p-3 text-left transition-shadow hover:shadow-sm"
    >
      <div className="flex items-start justify-between gap-2">
        <span className="font-mono text-[10px] text-muted-foreground">#{incident.id}</span>
        <span className={cn('rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset', sev.pill)}>
          {sev.label}
        </span>
      </div>
      <div className="text-xs font-medium leading-snug">{incident.name}</div>
      <div className="flex flex-wrap gap-1 text-[10px] text-muted-foreground">
        {incident.mitre.slice(0, 2).map((m) => (
          <span key={m.id} className="rounded bg-muted px-1 py-0.5 font-mono">
            {m.id}
          </span>
        ))}
      </div>
      <div className="mt-1 flex items-center justify-between border-t border-border pt-2 text-[11px]">
        <AssigneeChip analyst={incident.assignee} />
        <span className="inline-flex items-center gap-1 text-muted-foreground">
          <ShieldAlert size={10} className="text-orange-500" />
          {incident.alertsCount}
        </span>
      </div>
      <SLAIndicator sla={incident.sla} status={incident.status} />
    </button>
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

/* ─── Drawer ───────────────────────────────────────────────────────────── */

type Tab = 'overview' | 'timeline' | 'alerts' | 'notes' | 'actions' | 'history'

function IncidentDrawer({ incident, onClose }: { incident: Incident; onClose: () => void }) {
  const [tab, setTab] = useState<Tab>('overview')
  const sev = SEVERITY_STYLE[incident.severity]
  const st = STATUS_META[incident.status]

  return (
    <div
      className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="flex w-full max-w-[960px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                <span className="font-mono">#{incident.id}</span>
                <span>·</span>
                <span>opened {absTimestamp(incident.createdAt)}</span>
                <span>·</span>
                <span className="capitalize">{incident.origin.replace('-', ' ')}</span>
              </div>
              <div className="mt-1 flex items-center gap-2">
                <span className={cn('h-5 w-1.5 rounded-full', sev.bar)} />
                <h2 className="truncate text-lg font-semibold">{incident.name}</h2>
              </div>
              <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                <span className={cn('rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset', sev.pill)}>
                  {sev.label}
                </span>
                <span className={cn('inline-flex items-center gap-1.5 rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset', st.pill)}>
                  <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
                  {st.label}
                </span>
                <span className="inline-flex items-center gap-1 rounded-md border border-border bg-background/40 px-1.5 py-0.5">
                  <ShieldAlert size={10} className="text-orange-500" />
                  {incident.alertsCount} alerts
                </span>
                <span className="inline-flex items-center gap-1 rounded-md border border-border bg-background/40 px-1.5 py-0.5">
                  <Network size={10} />
                  {incident.affectedAssets} assets
                </span>
                <SLAIndicator sla={incident.sla} status={incident.status} />
                {incident.tags.map((t) => (
                  <span key={t} className="rounded-md bg-muted px-1.5 py-0.5">
                    {t}
                  </span>
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
              <BadgeCheck size={14} className="mr-1.5" />
              Mark complete
            </Button>
            <Button size="sm" variant="outline">
              {st.label}
              <ChevronDown size={12} className="ml-1.5 opacity-60" />
            </Button>
            <Button size="sm" variant="outline">
              <Users size={14} className="mr-1.5" />
              {incident.assignee?.name ?? 'Assign'}
            </Button>
            <Button size="sm" variant="outline">
              <GitMerge size={14} className="mr-1.5" />
              Merge
            </Button>
            <Button size="sm" variant="outline">
              <Flame size={14} className="mr-1.5 text-amber-500" />
              Escalate
            </Button>
            <Button size="sm" variant="ghost" className="ml-auto">
              <MoreHorizontal size={14} />
            </Button>
          </div>
        </header>

        <nav className="flex items-center gap-1 border-b border-border px-6">
          <DrawerTab id="overview" current={tab} onChange={setTab} icon={Sparkles}>
            Overview
          </DrawerTab>
          <DrawerTab id="timeline" current={tab} onChange={setTab} icon={Clock}>
            Timeline
          </DrawerTab>
          <DrawerTab id="alerts" current={tab} onChange={setTab} icon={AlertTriangle}>
            Alerts
            <span className="ml-1.5 rounded bg-muted px-1.5 py-0.5 font-mono text-[10px]">
              {incident.alertsCount}
            </span>
          </DrawerTab>
          <DrawerTab id="notes" current={tab} onChange={setTab} icon={MessageSquare}>
            Notes
          </DrawerTab>
          <DrawerTab id="actions" current={tab} onChange={setTab} icon={Workflow}>
            Actions taken
          </DrawerTab>
          <DrawerTab id="history" current={tab} onChange={setTab} icon={History}>
            History
          </DrawerTab>
        </nav>

        <div className="flex-1 overflow-y-auto bg-muted/20">
          {tab === 'overview' && <OverviewTab incident={incident} />}
          {tab === 'timeline' && <TimelineTab incident={incident} />}
          {tab === 'alerts' && <AlertsTab />}
          {tab === 'notes' && <NotesTab />}
          {tab === 'actions' && <ActionsTab />}
          {tab === 'history' && <HistoryTab incident={incident} />}
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

function OverviewTab({ incident }: { incident: Incident }) {
  return (
    <div className="grid grid-cols-3 gap-4 p-6">
      <div className="col-span-2 space-y-4">
        <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-4">
          <div className="mb-2 flex items-center gap-2 text-xs font-medium">
            <Sparkles size={14} className="text-fuchsia-500" />
            SOC AI brief
          </div>
          <p className="text-sm leading-relaxed">
            This incident shows a clear chain of <strong>credential extraction</strong> followed by{' '}
            <strong>command-and-control</strong>. Adversary IP{' '}
            <span className="font-mono">198.51.100.14</span> appears in 6 of 8 grouped alerts and is
            tagged as <span className="font-medium">Scattered Spider</span> in 3 threat-intel feeds.
          </p>
          <div className="mt-3 flex flex-wrap gap-2 text-[11px]">
            <button className="rounded-md border border-border bg-background/60 px-2 py-1 hover:bg-background">
              Run containment playbook
            </button>
            <button className="rounded-md border border-border bg-background/60 px-2 py-1 hover:bg-background">
              Hunt similar incidents
            </button>
            <button className="rounded-md border border-border bg-background/60 px-2 py-1 hover:bg-background">
              Draft executive summary
            </button>
          </div>
        </div>

        <Section title="Description">
          <p className="text-sm text-foreground/90">{incident.description}</p>
        </Section>

        <Section title="Proposed solution">
          <p className="text-sm text-foreground/90">{incident.solution}</p>
        </Section>

        <Section title="MITRE ATT&CK chain">
          <div className="flex flex-wrap gap-2">
            {incident.mitre.map((m, idx) => (
              <span key={m.id} className="inline-flex items-center gap-1.5">
                <span className="inline-flex items-center gap-1.5 rounded-md border border-border bg-background/40 px-2 py-1 text-xs">
                  <Crosshair size={11} className="text-rose-500" />
                  <span className="font-mono">{m.id}</span>
                  {m.name}
                </span>
                {idx < incident.mitre.length - 1 && <ArrowRight size={12} className="text-muted-foreground" />}
              </span>
            ))}
          </div>
        </Section>
      </div>

      <aside className="col-span-1 space-y-4">
        <KeyValueCard
          title="Detail"
          rows={[
            ['Incident', `#${incident.id}`],
            ['Opened', absTimestamp(incident.createdAt)],
            ['Last update', relativeTime(incident.updatedAt)],
            ['Origin', incident.origin],
            ['Alerts', `${incident.alertsCount}`],
            ['Affected', `${incident.affectedAssets} assets`],
          ]}
        />

        <div className="rounded-lg border border-border bg-card p-4">
          <div className="mb-2 flex items-center justify-between">
            <span className="text-[11px] uppercase tracking-wider text-muted-foreground">SLA</span>
            <span className="text-[10px] text-muted-foreground">resolve within 4h</span>
          </div>
          <SLAProgress sla={incident.sla} />
        </div>

        <div className="rounded-lg border border-border bg-card p-4">
          <div className="mb-2 flex items-center justify-between">
            <span className="text-[11px] uppercase tracking-wider text-muted-foreground">Owner</span>
            <button className="text-[11px] text-primary hover:underline">Reassign</button>
          </div>
          {incident.assignee ? (
            <div className="flex items-center gap-2">
              <span className={cn('flex h-9 w-9 items-center justify-center rounded-full text-xs font-semibold text-white', incident.assignee.tone)}>
                {incident.assignee.initials}
              </span>
              <div>
                <div className="text-sm font-medium">{incident.assignee.name}</div>
                <div className="text-[11px] text-muted-foreground">on shift · responds in &lt;10m</div>
              </div>
            </div>
          ) : (
            <div className="text-xs text-muted-foreground italic">No owner assigned yet.</div>
          )}
        </div>

        <div className="rounded-lg border border-border bg-card p-4">
          <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">
            Related incidents
          </div>
          <ul className="space-y-1.5 text-xs">
            <li className="flex items-center justify-between hover:underline">
              <span>
                <span className="font-mono text-muted-foreground">#1135</span> Mimikatz signature on…
              </span>
              <span className="rounded bg-red-500/15 px-1.5 py-0.5 text-[10px] text-red-600 ring-1 ring-red-500/30 dark:text-red-300">
                critical
              </span>
            </li>
            <li className="flex items-center justify-between hover:underline">
              <span>
                <span className="font-mono text-muted-foreground">#1141</span> Unusual exfiltration…
              </span>
              <span className="rounded bg-orange-500/15 px-1.5 py-0.5 text-[10px] text-orange-600 ring-1 ring-orange-500/30 dark:text-orange-300">
                high
              </span>
            </li>
          </ul>
        </div>
      </aside>
    </div>
  )
}

function SLAProgress({ sla }: { sla: Incident['sla'] }) {
  const total = 4 * 60
  const elapsed = total - Math.max(0, Math.round((new Date(sla.dueAt).getTime() - NOW) / 60_000))
  const pct = Math.min(100, Math.max(0, (elapsed / total) * 100))
  const breached = sla.breached || elapsed > total
  return (
    <div>
      <div className="mb-1.5 flex items-center justify-between text-[11px]">
        <span className="text-muted-foreground">{breached ? 'Breached' : 'Time used'}</span>
        <span className={cn('font-mono tabular-nums', breached ? 'text-red-500' : 'text-foreground')}>
          {Math.round((elapsed / total) * 100)}%
        </span>
      </div>
      <div className="h-1.5 overflow-hidden rounded-full bg-muted">
        <div
          className={cn(
            'h-full rounded-full bg-gradient-to-r',
            breached
              ? 'from-red-500 to-rose-500'
              : pct > 75
                ? 'from-amber-500 to-orange-500'
                : 'from-emerald-500 to-sky-500'
          )}
          style={{ width: `${pct}%` }}
        />
      </div>
      <div className="mt-1 text-[10px] text-muted-foreground">
        Due {absTimestamp(sla.dueAt)}
      </div>
    </div>
  )
}

/* ─── Drawer: Timeline ─────────────────────────────────────────────────── */

function TimelineTab({ incident }: { incident: Incident }) {
  const events: { ts: string; kind: 'alert' | 'analyst' | 'system' | 'playbook'; text: string; who?: string }[] = [
    { ts: min(46), kind: 'alert', text: 'First alert correlated · Kerberoasting burst detected' },
    { ts: min(44), kind: 'system', text: 'Auto-grouped 8 alerts → incident #1142' },
    { ts: min(40), kind: 'analyst', who: 'Jordan Smith', text: 'Acknowledged · changed status to In review' },
    { ts: min(35), kind: 'playbook', text: 'Playbook "Containment Lite" triggered — isolated WIN-7F3K2L9Q8X' },
    { ts: min(28), kind: 'alert', text: 'New related alert: Mimikatz variant on same host' },
    { ts: min(20), kind: 'analyst', who: 'Jordan Smith', text: 'Added note: "Ticket dump confirmed via Sysmon EID 10"' },
    { ts: min(8), kind: 'system', text: 'Threat intel match: adversary IP 198.51.100.14 → Scattered Spider' },
    { ts: min(2), kind: 'analyst', who: 'Jordan Smith', text: 'Rotated krbtgt password (round 1 of 2)' },
  ]
  const kindStyle: Record<string, string> = {
    alert: 'bg-red-500',
    analyst: 'bg-sky-500',
    system: 'bg-violet-500',
    playbook: 'bg-emerald-500',
  }
  return (
    <div className="p-6">
      <ol className="relative space-y-4 border-l border-border pl-6">
        {events.map((e, i) => (
          <li key={i} className="relative">
            <span className={cn('absolute -left-[27px] top-1 flex h-3 w-3 rounded-full border-2 border-card', kindStyle[e.kind])} />
            <div className="flex items-baseline justify-between gap-3">
              <div className="text-xs">
                {e.who && <span className="font-medium">{e.who} </span>}
                {e.text}
              </div>
              <span className="shrink-0 font-mono text-[10px] text-muted-foreground">
                {relativeTime(e.ts)}
              </span>
            </div>
            <div className="mt-0.5 text-[10px] uppercase tracking-wider text-muted-foreground">
              {e.kind}
            </div>
          </li>
        ))}
        {/* Incident bookmark */}
        <li className="relative pt-4">
          <span className="absolute -left-[31px] top-3 flex h-5 w-5 items-center justify-center rounded-full border-2 border-card bg-primary text-primary-foreground">
            <Send size={9} />
          </span>
          <div className="text-xs italic text-muted-foreground">
            Incident still active · last activity {relativeTime(incident.updatedAt)}
          </div>
        </li>
      </ol>
    </div>
  )
}

/* ─── Drawer: Alerts ───────────────────────────────────────────────────── */

function AlertsTab() {
  const linked = [
    { id: 'a-9f04c2', sev: 'critical' as Severity, name: 'Suspected Kerberoasting against domain controller', technique: 'T1558.003', echoes: 142, ts: min(46) },
    { id: 'a-c014f0', sev: 'critical' as Severity, name: 'Outbound connection to known C2 IP', technique: 'T1071.001', echoes: 8, ts: min(46) },
    { id: 'a-1ab32e', sev: 'high' as Severity, name: 'PowerShell execution with encoded command', technique: 'T1059.001', echoes: 47, ts: min(44) },
    { id: 'a-44a811', sev: 'high' as Severity, name: 'Antivirus signature match — Mimikatz variant', technique: 'T1003', echoes: 1, ts: min(28) },
    { id: 'a-3e88a1', sev: 'medium' as Severity, name: 'Suspicious SMB share enumeration', technique: 'T1135', echoes: 12, ts: min(38) },
  ]
  return (
    <div className="p-6">
      <div className="mb-3 flex items-center justify-between">
        <h4 className="text-sm font-semibold">Linked alerts</h4>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm">
            <Plus size={12} className="mr-1.5" />
            Add alert
          </Button>
          <Button variant="ghost" size="sm">
            <X size={12} className="mr-1.5" />
            Unlink selected
          </Button>
        </div>
      </div>
      <div className="overflow-hidden rounded-lg border border-border bg-card">
        <div className="grid grid-cols-[80px_90px_1fr_120px_70px_90px] gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
          <div>ID</div>
          <div>Severity</div>
          <div>Alert</div>
          <div>Technique</div>
          <div className="text-right">Echoes</div>
          <div>Time</div>
        </div>
        {linked.map((l) => {
          const sev = SEVERITY_STYLE[l.sev]
          return (
            <div
              key={l.id}
              className="grid cursor-pointer grid-cols-[80px_90px_1fr_120px_70px_90px] gap-3 border-b border-border px-4 py-2.5 text-xs hover:bg-muted/40 last:border-b-0"
            >
              <div className="font-mono text-muted-foreground">{l.id}</div>
              <div>
                <span className={cn('rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset', sev.pill)}>
                  {sev.label}
                </span>
              </div>
              <div className="truncate">{l.name}</div>
              <div className="font-mono text-[11px] text-muted-foreground">{l.technique}</div>
              <div className="text-right tabular-nums text-muted-foreground">{l.echoes}</div>
              <div className="font-mono text-[11px] text-muted-foreground">{relativeTime(l.ts)}</div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

/* ─── Drawer: Notes (comment thread) ───────────────────────────────────── */

function NotesTab() {
  const notes = [
    {
      author: ANALYSTS[0],
      ts: min(20),
      body: 'Ticket dump confirmed via Sysmon EID 10 (lsass.exe access from PROCMON). Building Mimikatz signature into the next detection ruleset.',
      reactions: ['👀 2', '🔥 1'],
    },
    {
      author: ANALYSTS[1],
      ts: min(15),
      body: 'I can take the lateral-movement angle — checking GPO and shared drive accesses on neighboring hosts.',
      reactions: ['👍 1'],
    },
    {
      author: ANALYSTS[0],
      ts: min(2),
      body: 'Round 1 of krbtgt rotation done. Round 2 scheduled in 10h per Microsoft guidance.',
      reactions: [],
    },
  ]
  return (
    <div className="flex h-full flex-col">
      <div className="flex-1 space-y-3 overflow-y-auto p-6">
        {notes.map((n, i) => (
          <div key={i} className="flex gap-3">
            <span className={cn('mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-[10px] font-semibold text-white', n.author.tone)}>
              {n.author.initials}
            </span>
            <div className="min-w-0 flex-1">
              <div className="flex items-baseline gap-2 text-[11px]">
                <span className="font-medium">{n.author.name}</span>
                <span className="text-muted-foreground">{relativeTime(n.ts)}</span>
              </div>
              <div className="mt-1 rounded-lg border border-border bg-card px-3 py-2 text-sm leading-relaxed">
                {n.body}
              </div>
              {n.reactions.length > 0 && (
                <div className="mt-1 flex gap-1">
                  {n.reactions.map((r) => (
                    <span key={r} className="rounded-full border border-border bg-card px-2 py-0.5 text-[10px]">
                      {r}
                    </span>
                  ))}
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
      <div className="border-t border-border bg-card p-4">
        <div className="flex items-center gap-2">
          <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-sky-500 text-[10px] font-semibold text-white">
            JS
          </span>
          <div className="relative flex-1">
            <Input placeholder="Add a note · @mention to notify · Shift+Enter for newline" className="h-9 pr-12" />
            <button className="absolute right-1 top-1 flex h-7 w-7 items-center justify-center rounded-md bg-primary text-primary-foreground hover:opacity-90">
              <Send size={13} />
            </button>
          </div>
        </div>
        <div className="mt-2 flex items-center gap-3 text-[10px] text-muted-foreground">
          <span>
            <kbd className="rounded border border-border bg-muted px-1 font-mono">/</kbd> commands
          </span>
          <span>
            <TagIcon size={10} className="inline" /> tag
          </span>
          <span>
            <UserCircle2 size={10} className="inline" /> assign
          </span>
          <span>
            <Workflow size={10} className="inline" /> trigger playbook
          </span>
        </div>
      </div>
    </div>
  )
}

/* ─── Drawer: Actions taken ────────────────────────────────────────────── */

function ActionsTab() {
  const items = [
    { ts: min(35), kind: 'playbook' as const, name: 'Containment Lite', state: 'success' as const, detail: 'Isolated WIN-7F3K2L9Q8X via firewall API' },
    { ts: min(30), kind: 'manual' as const, name: 'Disabled account svc_backup', state: 'success' as const, detail: 'Performed by Jordan Smith via AD console' },
    { ts: min(8), kind: 'auto' as const, name: 'IOC pushed to perimeter blocklist', state: 'success' as const, detail: 'Adversary IP 198.51.100.14 now blocked at edge-fw-bsa-01' },
    { ts: min(4), kind: 'manual' as const, name: 'krbtgt rotation (1/2)', state: 'success' as const, detail: 'Round 2 scheduled in 10h' },
    { ts: min(2), kind: 'manual' as const, name: 'Forensic capture of fin-app-01', state: 'pending' as const, detail: 'Snapshot in progress, ETA 18m' },
  ]
  const kindStyle: Record<string, string> = {
    playbook: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300',
    manual: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300',
    auto: 'bg-violet-500/15 text-violet-600 ring-violet-500/30 dark:text-violet-300',
  }
  return (
    <div className="space-y-3 p-6">
      <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-3 text-xs">
        <div className="flex items-center gap-2 font-medium">
          <Sparkles size={13} className="text-fuchsia-500" />
          Suggested next actions
        </div>
        <ul className="mt-2 space-y-1.5 text-foreground/90">
          <li>· Run "Hunt for Mimikatz indicators" across all assets in the affected OU.</li>
          <li>· Open ticket with help-desk to reset passwords for users in /Finance OU.</li>
          <li>· Schedule retrospective once incident closes.</li>
        </ul>
      </div>
      {items.map((a, i) => (
        <div key={i} className="rounded-lg border border-border bg-card p-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-sm font-medium">
              <span className={cn('rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset', kindStyle[a.kind])}>
                {a.kind}
              </span>
              {a.name}
            </div>
            <div className="flex items-center gap-2">
              <span
                className={cn(
                  'rounded px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset',
                  a.state === 'success'
                    ? 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300'
                    : 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300'
                )}
              >
                {a.state}
              </span>
              <span className="font-mono text-[10px] text-muted-foreground">{relativeTime(a.ts)}</span>
            </div>
          </div>
          <p className="mt-1.5 pl-1 text-xs text-muted-foreground">{a.detail}</p>
        </div>
      ))}
    </div>
  )
}

/* ─── Drawer: History ──────────────────────────────────────────────────── */

function HistoryTab({ incident }: { incident: Incident }) {
  const events = [
    { ts: min(46), who: 'system', what: `incident created from 8 correlated alerts (origin: ${incident.origin})` },
    { ts: min(45), who: 'system', what: `severity set to ${incident.severity}` },
    { ts: min(40), who: 'jsmith', what: 'changed status from open to in_review' },
    { ts: min(40), who: 'system', what: 'auto-tagged: Scattered Spider, campaign-44' },
    { ts: min(28), who: 'system', what: 'linked alert a-44a811 (Mimikatz variant)' },
    { ts: min(20), who: 'jsmith', what: 'added note' },
    { ts: min(2), who: 'jsmith', what: 'recorded action: krbtgt rotation (1/2)' },
  ]
  return (
    <div className="p-6">
      <div className="overflow-hidden rounded-lg border border-border bg-card">
        <div className="grid grid-cols-[140px_120px_1fr] gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
          <div>Timestamp</div>
          <div>Actor</div>
          <div>Action</div>
        </div>
        {events.map((e, i) => (
          <div key={i} className="grid grid-cols-[140px_120px_1fr] gap-3 border-b border-border px-4 py-2 text-xs last:border-b-0">
            <div className="font-mono text-[11px] text-muted-foreground">{absTimestamp(e.ts)}</div>
            <div className="truncate font-mono text-[11px]">{e.who}</div>
            <div>{e.what}</div>
          </div>
        ))}
      </div>
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
  if (Math.abs(m) < 60) return `${m}m ago`
  const h = Math.round(m / 60)
  if (Math.abs(h) < 24) return `${h}h ago`
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
