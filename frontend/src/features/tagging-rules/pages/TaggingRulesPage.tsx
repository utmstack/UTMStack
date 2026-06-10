import { useMemo, useState } from 'react'
import {
  ChevronDown,
  Copy,
  ExternalLink,
  GripVertical,
  History,
  ListFilter,
  MoreHorizontal,
  Plus,
  Search,
  Sparkles,
  Tag as TagIcon,
  Upload,
  X,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type Operator = 'eq' | 'in' | 'not_in' | 'gt' | 'lt' | 'matches' | 'between' | 'not_between'

interface Condition {
  field: string
  op: Operator
  value: string
}

interface TaggingRule {
  id: string
  name: string
  description: string
  enabled: boolean
  priority: number
  conditions: Condition[]
  // logic between conditions: 'all' = AND, 'any' = OR
  logic: 'all' | 'any'
  tags: { name: string; color: string }[]
  applied24h: number
  appliedTotal: number
  lastApplied?: string
  createdAt: string
  modifiedAt: string
  modifiedBy: string
  author: 'built-in' | 'custom'
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const NOW = new Date('2026-05-01T16:42:00Z').getTime()
const min = (m: number) => new Date(NOW - m * 60_000).toISOString()
const hr = (h: number) => new Date(NOW - h * 3_600_000).toISOString()
const days = (d: number) => new Date(NOW - d * 86_400_000).toISOString()

const RULES: TaggingRule[] = [
  {
    id: 'tr-scattered-spider',
    name: 'Scattered Spider campaign',
    description: 'Tag any alert whose adversary IP appears in the Scattered Spider IOC list. Used to group activity into a campaign.',
    enabled: true,
    priority: 1,
    logic: 'any',
    conditions: [
      { field: 'adversary.ip', op: 'in', value: 'feed:scattered-spider-c2' },
      { field: 'tags', op: 'matches', value: 'octo-tempest' },
    ],
    tags: [
      { name: 'Scattered Spider', color: 'bg-fuchsia-500' },
      { name: 'campaign-44', color: 'bg-violet-500' },
    ],
    applied24h: 23,
    appliedTotal: 412,
    lastApplied: min(4),
    createdAt: days(120),
    modifiedAt: hr(8),
    modifiedBy: 'jsmith',
    author: 'custom',
  },
  {
    id: 'tr-external-brute',
    name: 'External brute-force attempts',
    description: 'Identify brute-force attempts originating from outside the corporate network so they can be triaged separately.',
    enabled: true,
    priority: 2,
    logic: 'all',
    conditions: [
      { field: 'source.ip', op: 'not_in', value: 'list:internal-ranges' },
      { field: 'technique.id', op: 'eq', value: 'T1110' },
    ],
    tags: [
      { name: 'external', color: 'bg-orange-500' },
      { name: 'brute-force', color: 'bg-red-500' },
    ],
    applied24h: 41,
    appliedTotal: 1284,
    lastApplied: min(23),
    createdAt: days(220),
    modifiedAt: days(7),
    modifiedBy: 'system',
    author: 'built-in',
  },
  {
    id: 'tr-lateral',
    name: 'Internal lateral movement',
    description: 'Mark internal-to-internal activity matching known lateral-movement techniques (SMB enum, RDP, WMI exec).',
    enabled: true,
    priority: 3,
    logic: 'all',
    conditions: [
      { field: 'source.ip', op: 'in', value: 'list:internal-ranges' },
      { field: 'target.ip', op: 'in', value: 'list:internal-ranges' },
      { field: 'technique.id', op: 'in', value: 'T1135, T1021, T1077, T1570' },
    ],
    tags: [
      { name: 'lateral-movement', color: 'bg-rose-500' },
    ],
    applied24h: 12,
    appliedTotal: 184,
    lastApplied: min(38),
    createdAt: days(180),
    modifiedAt: days(30),
    modifiedBy: 'system',
    author: 'built-in',
  },
  {
    id: 'tr-critical-asset',
    name: 'Hit on critical asset',
    description: 'Boost visibility of any alert involving an asset tagged as critical-asset by adding the same tag to the alert.',
    enabled: true,
    priority: 4,
    logic: 'any',
    conditions: [
      { field: 'source.host.tags', op: 'matches', value: 'critical-asset' },
      { field: 'target.host.tags', op: 'matches', value: 'critical-asset' },
    ],
    tags: [
      { name: 'critical-asset', color: 'bg-red-500' },
    ],
    applied24h: 18,
    appliedTotal: 624,
    lastApplied: min(11),
    createdAt: days(360),
    modifiedAt: days(60),
    modifiedBy: 'system',
    author: 'built-in',
  },
  {
    id: 'tr-off-hours',
    name: 'Off-hours activity',
    description: 'Flag alerts that fire outside the 08:00–18:00 business window. Useful for prioritizing review the next morning.',
    enabled: true,
    priority: 5,
    logic: 'any',
    conditions: [
      { field: '@timestamp.hour', op: 'not_between', value: '08, 18' },
      { field: '@timestamp.day_of_week', op: 'in', value: 'sat, sun' },
    ],
    tags: [
      { name: 'off-hours', color: 'bg-amber-500' },
    ],
    applied24h: 84,
    appliedTotal: 4_120,
    lastApplied: min(2),
    createdAt: days(540),
    modifiedAt: days(120),
    modifiedBy: 'system',
    author: 'built-in',
  },
  {
    id: 'tr-privileged',
    name: 'Privileged user activity',
    description: 'Tag any alert involving a Domain Admin or Schema Admin so it surfaces in the privileged-activity dashboard.',
    enabled: true,
    priority: 6,
    logic: 'any',
    conditions: [
      { field: 'source.user.groups', op: 'matches', value: 'Domain Admins|Schema Admins' },
      { field: 'target.user.groups', op: 'matches', value: 'Domain Admins|Schema Admins' },
    ],
    tags: [
      { name: 'privileged', color: 'bg-amber-500' },
    ],
    applied24h: 4,
    appliedTotal: 88,
    lastApplied: hr(2),
    createdAt: days(360),
    modifiedAt: days(180),
    modifiedBy: 'tferguson',
    author: 'custom',
  },
  {
    id: 'tr-finance',
    name: 'Finance department alert',
    description: 'Tag alerts where the user belongs to the Finance OU so the Finance team can filter their own queue.',
    enabled: true,
    priority: 7,
    logic: 'any',
    conditions: [
      { field: 'source.user.department', op: 'eq', value: 'Finance' },
      { field: 'target.user.department', op: 'eq', value: 'Finance' },
    ],
    tags: [
      { name: 'finance-ou', color: 'bg-emerald-500' },
    ],
    applied24h: 7,
    appliedTotal: 142,
    lastApplied: hr(4),
    createdAt: days(180),
    modifiedAt: days(45),
    modifiedBy: 'aluna',
    author: 'custom',
  },
  {
    id: 'tr-tor',
    name: 'Tor exit node source',
    description: 'Mark alerts where the source IP is a known Tor exit node. Often a strong indicator of intent to obscure.',
    enabled: true,
    priority: 8,
    logic: 'all',
    conditions: [
      { field: 'source.ip', op: 'in', value: 'feed:tor-exit-nodes' },
    ],
    tags: [
      { name: 'tor', color: 'bg-violet-500' },
      { name: 'anonymized', color: 'bg-zinc-500' },
    ],
    applied24h: 0,
    appliedTotal: 14,
    lastApplied: days(2),
    createdAt: days(540),
    modifiedAt: days(360),
    modifiedBy: 'system',
    author: 'built-in',
  },
  {
    id: 'tr-noisy',
    name: 'Noisy duplicate (auto-suppress)',
    description: 'When the same rule fires more than 100× from the same source in 24h, tag the alert noisy and auto-suppress it.',
    enabled: true,
    priority: 9,
    logic: 'all',
    conditions: [
      { field: 'duplicate.count_24h', op: 'gt', value: '100' },
      { field: 'duplicate.same_source', op: 'eq', value: 'true' },
    ],
    tags: [
      { name: 'noisy', color: 'bg-zinc-500' },
      { name: 'auto-suppress', color: 'bg-zinc-400' },
    ],
    applied24h: 312,
    appliedTotal: 18_204,
    lastApplied: min(8),
    createdAt: days(120),
    modifiedAt: days(15),
    modifiedBy: 'jsmith',
    author: 'custom',
  },
  {
    id: 'tr-fp',
    name: 'Known false-positive (Rule #992)',
    description: 'Specific tuning: alerts from rule "outbound DNS to dynamic domain" matching the internal CDN are false positives.',
    enabled: false,
    priority: 10,
    logic: 'all',
    conditions: [
      { field: 'rule.id', op: 'eq', value: 'r-992' },
      { field: 'destination.domain', op: 'matches', value: '*.cdn-internal.local' },
    ],
    tags: [
      { name: 'false-positive', color: 'bg-zinc-400' },
    ],
    applied24h: 0,
    appliedTotal: 47,
    lastApplied: days(7),
    createdAt: days(45),
    modifiedAt: days(7),
    modifiedBy: 'jsmith',
    author: 'custom',
  },
]

const OP_LABEL: Record<Operator, string> = {
  eq: '=',
  in: 'in',
  not_in: 'not in',
  gt: '>',
  lt: '<',
  matches: 'matches',
  between: 'between',
  not_between: 'not between',
}

/* ─── Saved views ──────────────────────────────────────────────────────── */

interface SavedView {
  id: string
  label: string
  count: number
  predicate: (r: TaggingRule) => boolean
}

const SAVED_VIEWS: SavedView[] = [
  { id: 'all', label: 'All', count: RULES.length, predicate: () => true },
  { id: 'enabled', label: 'Enabled', count: RULES.filter((r) => r.enabled).length, predicate: (r) => r.enabled },
  { id: 'disabled', label: 'Disabled', count: RULES.filter((r) => !r.enabled).length, predicate: (r) => !r.enabled },
  { id: 'custom', label: 'Custom', count: RULES.filter((r) => r.author === 'custom').length, predicate: (r) => r.author === 'custom' },
  { id: 'high-impact', label: 'High impact', count: RULES.filter((r) => r.applied24h >= 20).length, predicate: (r) => r.applied24h >= 20 },
]

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function TaggingRulesPage() {
  const [view, setView] = useState<string>('all')
  const [search, setSearch] = useState('')
  const [openRule, setOpenRule] = useState<TaggingRule | null>(null)
  const [enabledOverrides, setEnabledOverrides] = useState<Record<string, boolean>>({})

  const isEnabled = (r: TaggingRule) => enabledOverrides[r.id] ?? r.enabled

  const filtered = useMemo(() => {
    const v = SAVED_VIEWS.find((s) => s.id === view) ?? SAVED_VIEWS[0]
    return RULES.filter(v.predicate).filter((r) =>
      search
        ? (r.name + r.description + r.tags.map((t) => t.name).join(' '))
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
          <ApplyOverviewCard />
        </div>
        <div className="col-span-12 lg:col-span-4">
          <TagsInsightsCard />
        </div>
      </div>

      <div className="mt-5">
        <SavedViewTabs current={view} onChange={setView} />
      </div>

      <Toolbar search={search} onSearch={setSearch} />

      <div className="mt-3 text-[11px] text-muted-foreground">
        Rules are evaluated <span className="font-medium text-foreground">top to bottom</span>{' '}
        in priority order. All matching rules apply their tags.
      </div>

      <div className="mt-2 overflow-hidden rounded-xl border border-border bg-card">
        <ListHeader />
        {filtered.map((r) => (
          <RuleRow
            key={r.id}
            rule={r}
            enabled={isEnabled(r)}
            onToggle={() =>
              setEnabledOverrides((prev) => ({ ...prev, [r.id]: !isEnabled(r) }))
            }
            onOpen={() => setOpenRule(r)}
          />
        ))}
        {filtered.length === 0 && (
          <div className="px-6 py-16 text-center text-sm text-muted-foreground">
            No tagging rules match this view.
          </div>
        )}
      </div>

      {openRule && (
        <RuleDrawer
          rule={openRule}
          enabled={isEnabled(openRule)}
          onClose={() => setOpenRule(null)}
        />
      )}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ total }: { total: number }) {
  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold tracking-tight">Tagging rules</h1>
          <span className="rounded-md bg-muted px-2 py-0.5 font-mono text-[11px] tabular-nums text-muted-foreground">
            {total.toLocaleString()} rules
          </span>
        </div>
        <p className="mt-1 text-xs text-muted-foreground">
          Apply tags to alerts automatically based on conditions. Tags drive saved views, routing,
          and dashboards.
        </p>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm">
          <Upload size={14} className="mr-2" />
          Import
        </Button>
        <Button size="sm">
          <Plus size={14} className="mr-2" />
          New tagging rule
        </Button>
      </div>
    </header>
  )
}

/* ─── Apply overview ───────────────────────────────────────────────────── */

function ApplyOverviewCard() {
  const buckets = 60
  const data = useMemo(
    () =>
      Array.from({ length: buckets }, (_, i) =>
        Math.max(0, Math.round(Math.abs(Math.sin(i / 6)) * 28 + Math.abs(Math.sin(i / 11)) * 8 + 6))
      ),
    []
  )
  const total24h = RULES.reduce((s, r) => s + r.applied24h, 0)
  const enabled = RULES.filter((r) => r.enabled).length

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
            Tags applied · last 24 hours
          </div>
          <div className="mt-1 flex items-baseline gap-3">
            <span className="text-3xl font-semibold tabular-nums">
              {total24h.toLocaleString()}
            </span>
            <span className="text-sm text-muted-foreground">
              from <span className="text-foreground">{enabled}</span> active rules
            </span>
          </div>
        </div>
      </div>

      <svg viewBox={`0 0 ${w} ${h}`} className="mt-4 h-24 w-full" preserveAspectRatio="none">
        <defs>
          <linearGradient id="tagsGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="rgb(168 85 247)" stopOpacity="0.28" />
            <stop offset="100%" stopColor="rgb(168 85 247)" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d={areaPath} fill="url(#tagsGrad)" />
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

/* ─── Insights ─────────────────────────────────────────────────────────── */

function TagsInsightsCard() {
  return (
    <div className="flex h-full flex-col gap-3 rounded-xl border border-border bg-card p-5">
      <div className="flex items-center gap-2">
        <Sparkles size={14} className="text-fuchsia-500" />
        <h3 className="text-sm font-semibold">Tag activity</h3>
      </div>

      <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">Top tag in 24h:</span>{' '}
        <span className="font-medium">noisy</span> · 312 alerts (auto-suppressed).{' '}
        <button className="text-primary hover:underline">Review</button>
      </div>

      <div className="rounded-lg border border-amber-500/30 bg-amber-500/5 p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">2 rules</span> haven't applied any tag in 30
        days. Consider reviewing or disabling.{' '}
        <button className="text-primary hover:underline">View</button>
      </div>

      <div className="rounded-lg border border-border bg-background/40 p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">14 unique tags</span> in use across all rules.{' '}
        <button className="text-primary hover:underline">Manage tags</button>
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

function Toolbar({ search, onSearch }: { search: string; onSearch: (s: string) => void }) {
  return (
    <div className="mt-3 flex flex-wrap items-center gap-2">
      <div className="relative min-w-[280px] flex-1">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input
          value={search}
          onChange={(e) => onSearch(e.target.value)}
          placeholder="Search by name, description, tag…"
          className="h-9 pl-9"
        />
      </div>
      <Button variant="outline" size="sm">
        <ListFilter size={14} className="mr-2" />
        Filters
      </Button>
    </div>
  )
}

/* ─── List ─────────────────────────────────────────────────────────────── */

const COLS = '20px 30px 44px 1fr 1fr 90px 110px 110px 36px'

function ListHeader() {
  return (
    <div
      className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: COLS }}
    >
      <div />
      <div className="text-right">#</div>
      <div />
      <div>Rule</div>
      <div>Applies tags</div>
      <div className="text-right">24h</div>
      <div>Last applied</div>
      <div>Modified</div>
      <div />
    </div>
  )
}

function RuleRow({
  rule: r,
  enabled,
  onToggle,
  onOpen,
}: {
  rule: TaggingRule
  enabled: boolean
  onToggle: () => void
  onOpen: () => void
}) {
  return (
    <div
      onClick={onOpen}
      className={cn(
        'group grid cursor-pointer items-center gap-3 border-b border-border px-4 py-3 text-xs transition-colors hover:bg-muted/40 last:border-b-0',
        !enabled && 'opacity-60'
      )}
      style={{ gridTemplateColumns: COLS }}
    >
      <span
        className="cursor-grab text-muted-foreground/50 hover:text-foreground"
        onClick={(e) => e.stopPropagation()}
        title="Drag to reorder"
      >
        <GripVertical size={12} />
      </span>
      <div className="text-right font-mono tabular-nums text-muted-foreground">
        {r.priority}
      </div>
      <Toggle
        enabled={enabled}
        onChange={(e) => {
          e.stopPropagation()
          onToggle()
        }}
      />
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="truncate text-sm font-medium">{r.name}</span>
          {r.author === 'custom' && (
            <span className="rounded bg-emerald-500/15 px-1.5 py-0.5 text-[9px] font-medium uppercase text-emerald-600 ring-1 ring-emerald-500/30 dark:text-emerald-300">
              custom
            </span>
          )}
        </div>
        <div className="mt-0.5 truncate text-[11px] text-muted-foreground">
          {r.conditions.length} condition{r.conditions.length === 1 ? '' : 's'}{' '}
          {r.logic === 'all' ? '· match all' : '· match any'}
        </div>
      </div>
      <div className="flex flex-wrap items-center gap-1">
        {r.tags.map((t) => (
          <TagPill key={t.name} name={t.name} color={t.color} />
        ))}
      </div>
      <div className={cn('text-right tabular-nums', r.applied24h > 0 ? 'text-foreground' : 'text-muted-foreground')}>
        {r.applied24h > 0 ? r.applied24h.toLocaleString() : '—'}
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">
        {r.lastApplied ? relativeTime(r.lastApplied) : '—'}
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">
        {relativeTime(r.modifiedAt)}
        <div className="text-[10px] text-muted-foreground/70">{r.modifiedBy}</div>
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

function TagPill({ name, color }: { name: string; color: string }) {
  return (
    <span className="inline-flex items-center gap-1 rounded-md bg-muted px-1.5 py-0.5 text-[10px]">
      <span className={cn('h-1.5 w-1.5 rounded-full', color)} />
      {name}
    </span>
  )
}

function Toggle({
  enabled,
  onChange,
}: {
  enabled: boolean
  onChange: (e: React.MouseEvent) => void
}) {
  return (
    <button
      onClick={onChange}
      className={cn(
        'relative h-4 w-7 shrink-0 rounded-full transition-colors',
        enabled ? 'bg-primary' : 'bg-muted'
      )}
    >
      <span
        className={cn(
          'absolute top-0.5 h-3 w-3 rounded-full bg-white shadow transition-all',
          enabled ? 'left-3.5' : 'left-0.5'
        )}
      />
    </button>
  )
}

/* ─── Drawer ───────────────────────────────────────────────────────────── */

type Tab = 'overview' | 'logic' | 'history'

function RuleDrawer({
  rule: r,
  enabled,
  onClose,
}: {
  rule: TaggingRule
  enabled: boolean
  onClose: () => void
}) {
  const [tab, setTab] = useState<Tab>('overview')

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
                <span className="font-mono">{r.id}</span>
                <span>·</span>
                <span>Priority {r.priority}</span>
                <span>·</span>
                <span className="capitalize">{r.author}</span>
              </div>
              <h2 className="mt-1 truncate text-lg font-semibold">{r.name}</h2>
              <p className="mt-1 text-xs leading-relaxed text-muted-foreground">{r.description}</p>
              <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                <span className="text-muted-foreground">Tags applied:</span>
                {r.tags.map((t) => (
                  <TagPill key={t.name} name={t.name} color={t.color} />
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
              <ExternalLink size={13} className="mr-1.5" />
              View tagged alerts
            </Button>
            <Button size="sm" variant="outline">
              <Copy size={13} className="mr-1.5" />
              Duplicate
            </Button>
            <span className="ml-auto inline-flex items-center gap-2 text-[11px] text-muted-foreground">
              <Toggle enabled={enabled} onChange={(e) => e.stopPropagation()} />
              {enabled ? 'Enabled' : 'Disabled'}
            </span>
          </div>
        </header>

        <nav className="flex items-center gap-1 border-b border-border px-6">
          <DrawerTab id="overview" current={tab} onChange={setTab} icon={Sparkles}>
            Overview
          </DrawerTab>
          <DrawerTab id="logic" current={tab} onChange={setTab} icon={TagIcon}>
            Conditions & tags
          </DrawerTab>
          <DrawerTab id="history" current={tab} onChange={setTab} icon={History}>
            History
          </DrawerTab>
        </nav>

        <div className="flex-1 overflow-y-auto bg-muted/20">
          {tab === 'overview' && <OverviewTab rule={r} />}
          {tab === 'logic' && <LogicTab rule={r} />}
          {tab === 'history' && <HistoryTab rule={r} />}
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

function OverviewTab({ rule: r }: { rule: TaggingRule }) {
  return (
    <div className="grid grid-cols-3 gap-4 p-6">
      <div className="col-span-2 space-y-4">
        <Section title="Description">
          <p className="text-sm leading-relaxed">{r.description}</p>
        </Section>

        <Section title="Apply trend · 14 days">
          <SparkCalendar />
        </Section>
      </div>

      <aside className="col-span-1 space-y-4">
        <KeyValueCard
          title="Stats"
          rows={[
            ['24h', r.applied24h.toLocaleString()],
            ['Total', r.appliedTotal.toLocaleString()],
            ['Last applied', r.lastApplied ? relativeTime(r.lastApplied) : '—'],
          ]}
        />

        <KeyValueCard
          title="Lifecycle"
          rows={[
            ['Priority', String(r.priority)],
            ['Created', relativeTime(r.createdAt)],
            ['Modified', relativeTime(r.modifiedAt)],
            ['By', <span className="font-mono">{r.modifiedBy}</span>],
            ['Author', <span className="capitalize">{r.author}</span>],
          ]}
        />
      </aside>
    </div>
  )
}

function SparkCalendar() {
  const days14 = Array.from({ length: 14 }, (_, i) =>
    Math.max(0, Math.round(Math.abs(Math.sin((i + 3) / 1.7)) * 18 + (i === 13 ? 8 : 0)))
  )
  const max = Math.max(...days14, 1)
  return (
    <div className="flex items-end gap-1.5">
      {days14.map((v, i) => {
        const intensity = v / max
        return (
          <div key={i} className="flex flex-1 flex-col items-center gap-1">
            <div
              className={cn(
                'w-full rounded-sm',
                v === 0 ? 'bg-muted' : intensity > 0.7 ? 'bg-violet-500/80' : intensity > 0.4 ? 'bg-violet-500/60' : 'bg-violet-500/40'
              )}
              style={{ height: `${4 + intensity * 36}px` }}
              title={`${v} applications`}
            />
            <span className="text-[9px] tabular-nums text-muted-foreground">
              {i === 13 ? 'now' : `-${13 - i}d`}
            </span>
          </div>
        )
      })}
    </div>
  )
}

/* ─── Drawer: Logic ────────────────────────────────────────────────────── */

function LogicTab({ rule: r }: { rule: TaggingRule }) {
  return (
    <div className="space-y-4 p-6">
      <Section title={`Conditions · match ${r.logic}`}>
        <div className="space-y-2">
          {r.conditions.map((c, i) => (
            <div
              key={i}
              className="flex items-center gap-2 rounded-md border border-border bg-background/40 px-3 py-2 text-xs"
            >
              <span className="font-mono text-muted-foreground">{c.field}</span>
              <span className="rounded bg-muted px-1.5 py-0.5 font-mono text-[10px]">
                {OP_LABEL[c.op]}
              </span>
              <span className="font-mono">{c.value}</span>
              {i < r.conditions.length - 1 && (
                <span className="ml-auto text-[10px] uppercase tracking-wider text-muted-foreground">
                  {r.logic === 'all' ? 'AND' : 'OR'}
                </span>
              )}
            </div>
          ))}
        </div>
      </Section>

      <Section title="Tags applied when matched">
        <div className="flex flex-wrap gap-2">
          {r.tags.map((t) => (
            <span
              key={t.name}
              className="inline-flex items-center gap-1.5 rounded-md border border-border bg-card px-2.5 py-1.5 text-xs"
            >
              <span className={cn('h-2 w-2 rounded-full', t.color)} />
              {t.name}
            </span>
          ))}
        </div>
      </Section>

      <Section title="Equivalent expression">
        <pre className="overflow-x-auto rounded-md bg-muted/40 p-3 font-mono text-[11px] leading-relaxed">
          {`when ${r.logic === 'all' ? 'all' : 'any'} of:
${r.conditions.map((c) => `  ${c.field} ${OP_LABEL[c.op]} ${c.value}`).join('\n')}
then apply tags:
${r.tags.map((t) => `  + ${t.name}`).join('\n')}`}
        </pre>
        <div className="mt-2 flex items-center gap-3">
          <button className="inline-flex items-center gap-1 text-[11px] text-primary hover:underline">
            <Copy size={11} /> Copy
          </button>
          <button className="inline-flex items-center gap-1 text-[11px] text-primary hover:underline">
            <ExternalLink size={11} /> Open in editor
          </button>
        </div>
      </Section>
    </div>
  )
}

/* ─── Drawer: History ──────────────────────────────────────────────────── */

function HistoryTab({ rule: r }: { rule: TaggingRule }) {
  if (!r.lastApplied || r.applied24h === 0) {
    return (
      <div className="p-6">
        <div className="rounded-lg border border-dashed border-border bg-card px-6 py-12 text-center text-sm text-muted-foreground">
          This rule hasn't applied any tag in the selected time range.
        </div>
      </div>
    )
  }
  const events = [
    { ts: r.lastApplied, alert: 'Alert a-9f04c2', host: 'WIN-7F3K2L9Q8X' },
    { ts: hr(2), alert: 'Alert a-c014f0', host: 'fin-app-01' },
    { ts: hr(7), alert: 'Alert a-44a811', host: 'WIN-7F3K2L9Q8X' },
    { ts: hr(14), alert: 'Alert a-1ab32e', host: 'fin-app-01' },
  ]
  return (
    <div className="p-6">
      <div className="overflow-hidden rounded-lg border border-border bg-card">
        <div
          className="grid grid-cols-[140px_1fr_180px_36px] gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
        >
          <div>When</div>
          <div>Alert</div>
          <div>Host</div>
          <div />
        </div>
        {events.map((e, i) => (
          <div
            key={i}
            className="grid cursor-pointer grid-cols-[140px_1fr_180px_36px] gap-3 border-b border-border/60 px-4 py-2 text-xs hover:bg-muted/30 last:border-b-0"
          >
            <div className="font-mono text-[11px] text-muted-foreground">{relativeTime(e.ts)}</div>
            <div className="truncate font-mono text-[11px]">{e.alert}</div>
            <div className="truncate font-mono text-[11px]">{e.host}</div>
            <div className="flex justify-end">
              <ExternalLink size={11} className="text-muted-foreground" />
            </div>
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

// Suppress unused-import warning
void ChevronDown
