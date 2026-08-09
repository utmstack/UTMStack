import { useCallback, useEffect, useState } from 'react'
import {
  Activity,
  AlertTriangle,
  Apple,
  Cloud,
  Database,
  Layers,
  ListIcon,
  LayoutGrid,
  Loader2,
  Network,
  Plus,
  RefreshCw,
  ScrollText,
  Search,
  Server,
  ShieldCheck,
  Tag,
  Terminal,
  Trash2,
  Webhook,
  X,
  type LucideIcon,
} from 'lucide-react'
import { Link, useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { TimeRangePicker, presetRange, type TimeRange } from '@/shared/components/ui/time-range-picker'
import {
  datasourcesHttpService as svc,
  DatasourcesHttpError,
} from '../services/datasources-http.service'
import { AgentConsole } from '../components/AgentConsole'
import { deriveStatus, statusLabel, STATUS_META, IDLE_MS } from '../lib/status'
import type {
  Datasource,
  IngestionBucket,
  ListResponse,
  SourceKind,
  TimelinePoint,
  DatasourceCount,
} from '../types/datasource.types'

/* ─── Kind meta ───────────────────────────────────────────────────────── */

const KIND_META: Record<SourceKind, { label: string; icon: LucideIcon; tone: string }> = {
  agent: { label: 'Agent', icon: Server, tone: 'text-sky-500' },
  puller: { label: 'Integration', icon: Cloud, tone: 'text-violet-500' },
  direct: { label: 'Direct', icon: Webhook, tone: 'text-emerald-500' },
}

function kindMeta(kind?: string) {
  return KIND_META[(kind as SourceKind) ?? 'direct'] ?? { label: kind ?? '—', icon: Database, tone: 'text-muted-foreground' }
}

/* Map a datasource's dataType to a recognizable icon + label so the analyst knows
 * the technology at a glance. Matched in order; first hit wins. */
const DT_RULES: { match: RegExp; label: string; icon: LucideIcon; tone: string }[] = [
  { match: /wineventlog|windows/, label: 'Windows', icon: Server, tone: 'text-sky-500' },
  { match: /linux/, label: 'Linux', icon: Terminal, tone: 'text-amber-500' },
  { match: /macos|darwin|\bmac\b/, label: 'macOS', icon: Apple, tone: 'text-violet-400' },
  { match: /o365|office.?365|\b365\b/, label: 'Microsoft 365', icon: Cloud, tone: 'text-fuchsia-500' },
  { match: /azure/, label: 'Azure', icon: Cloud, tone: 'text-sky-500' },
  { match: /aws/, label: 'AWS', icon: Cloud, tone: 'text-orange-500' },
  { match: /gcp|google/, label: 'Google Cloud', icon: Cloud, tone: 'text-amber-500' },
  { match: /sophos/, label: 'Sophos', icon: Network, tone: 'text-sky-600' },
  { match: /fortigate|fortiweb|forti/, label: 'Fortinet', icon: Network, tone: 'text-red-500' },
  { match: /firepower|cisco|asa/, label: 'Cisco', icon: Network, tone: 'text-sky-500' },
  { match: /palo.?alto|panw/, label: 'Palo Alto', icon: Network, tone: 'text-orange-500' },
  { match: /meraki/, label: 'Meraki', icon: Network, tone: 'text-emerald-500' },
  { match: /mikrotik/, label: 'MikroTik', icon: Network, tone: 'text-red-400' },
  { match: /sonic.?wall/, label: 'SonicWall', icon: Network, tone: 'text-orange-400' },
  { match: /pfsense/, label: 'pfSense', icon: Network, tone: 'text-rose-500' },
  { match: /firewall|\bfw\b/, label: 'Firewall', icon: Network, tone: 'text-rose-500' },
  { match: /eset/, label: 'ESET', icon: ShieldCheck, tone: 'text-sky-500' },
  { match: /kaspersky/, label: 'Kaspersky', icon: ShieldCheck, tone: 'text-emerald-500' },
  { match: /sentinel.?one|sentinel/, label: 'SentinelOne', icon: ShieldCheck, tone: 'text-violet-500' },
  { match: /bitdefender/, label: 'Bitdefender', icon: ShieldCheck, tone: 'text-red-500' },
  { match: /crowdstrike/, label: 'CrowdStrike', icon: ShieldCheck, tone: 'text-red-500' },
  { match: /antivirus|deceptive/, label: 'Antivirus', icon: ShieldCheck, tone: 'text-emerald-500' },
  { match: /vmware|esxi/, label: 'VMware', icon: Server, tone: 'text-violet-500' },
  { match: /netflow/, label: 'NetFlow', icon: Activity, tone: 'text-cyan-500' },
  { match: /syslog/, label: 'Syslog', icon: Network, tone: 'text-muted-foreground' },
  { match: /generic/, label: 'Generic', icon: Database, tone: 'text-muted-foreground' },
]

function dataTypeMeta(dataType?: string): { label: string; icon: LucideIcon; tone: string } {
  const dt = (dataType ?? '').trim().toLowerCase()
  if (!dt) return { label: '', icon: Database, tone: 'text-muted-foreground' }
  for (const r of DT_RULES) if (r.match.test(dt)) return { label: r.label, icon: r.icon, tone: r.tone }
  return { label: dataType!, icon: Database, tone: 'text-muted-foreground' }
}

/** The leading icon: prefer the dataType (more recognizable); fall back to kind. */
function sourceIconMeta(s: Datasource): { icon: LucideIcon; tone: string } {
  if (s.dataType) return dataTypeMeta(s.dataType)
  return kindMeta(s.sourceKind)
}

const TABS = [
  { id: 'all', label: 'All' },
  { id: 'agent', label: 'Agents' },
  { id: 'puller', label: 'Integrations' },
  { id: 'direct', label: 'Direct' },
  { id: 'offline', label: 'Offline' },
] as const
type TabId = (typeof TABS)[number]['id']

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function DataSourcesPage() {
  const { t } = useTranslation()
  const [tab, setTab] = useState<TabId>('all')
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [labelFilter, setLabelFilter] = useState<string | null>(null)
  const [layout, setLayout] = useState<'list' | 'cards'>('list')

  const [page, setPage] = useState(0) // 0-based; sent as page+1, accumulated on scroll
  const [pageSize] = useState(50)
  const [total, setTotal] = useState(0)
  const [sources, setSources] = useState<Datasource[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const [usage, setUsage] = useState<DatasourceCount | null>(null)
  const [stats, setStats] = useState<Record<string, IngestionBucket>>({})
  const [timeline, setTimeline] = useState<TimelinePoint[]>([])
  const [range, setRange] = useState<TimeRange>(() => presetRange('24h'))
  const [counts, setCounts] = useState<Record<TabId, number> | null>(null)
  const [openId, setOpenId] = useState<string | null>(null)

  useEffect(() => {
    const h = setTimeout(() => {
      setDebounced(search.trim())
      setPage(0)
    }, 300)
    return () => clearTimeout(h)
  }, [search])

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      const res = await svc.list({
        page: page + 1,
        size: pageSize,
        search: debounced || undefined,
        kind: tab === 'all' || tab === 'offline' ? undefined : (tab as SourceKind),
        label: labelFilter ?? undefined,
        staleBefore: tab === 'offline' ? new Date(Date.now() - IDLE_MS).toISOString() : undefined,
      })
      setSources((prev) => (page === 0 ? (res.items ?? []) : [...prev, ...(res.items ?? [])]))
      setTotal(res.total_items ?? 0)
    } catch {
      setError(true)
      setSources([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, debounced, tab, labelFilter])

  const filterByLabel = (l: string) => {
    setLabelFilter(l)
    setPage(0)
  }

  useEffect(() => {
    void load()
  }, [load])

  // Side data (usage, ingestion stats) — load once + on manual refresh.
  const loadAux = useCallback(async () => {
    setLoading(true)
    const [u, totals, tl] = await Promise.allSettled([
      svc.count(),
      svc.ingestionTotals(),
      svc.ingestionTimeline(range),
    ])
    if (u.status === 'fulfilled') setUsage(u.value)
    if (totals.status === 'fulfilled') {
      const map: Record<string, IngestionBucket> = {}
      for (const b of totals.value.buckets ?? []) map[b.key] = b
      setStats(map)
    }
    if (tl.status === 'fulfilled') setTimeline(tl.value.points ?? [])

    // Per-tab counts for the health summary + tab badges. Cheap size:1 queries
    // that only read total_items; reflect global health (no search/group filter).
    const staleIso = new Date(Date.now() - IDLE_MS).toISOString()
    const [cAll, cAgent, cPuller, cDirect, cOffline,cNotConnected] = await Promise.allSettled([
      svc.list({ page: 1, size: 1 }),
      svc.list({ page: 1, size: 1, kind: 'agent' }),
      svc.list({ page: 1, size: 1, kind: 'puller'}),
      svc.list({ page: 1, size: 1, kind: 'direct'}),
      svc.list({ page: 1, size: 1, staleBefore: staleIso}),
      svc.list({ page: 1, size: 1, pingNull: true}),
    ])


    const tot = (r: PromiseSettledResult<ListResponse<Datasource>>) =>
      r.status === 'fulfilled' ? r.value.total_items ?? 0 : 0
    setCounts({
      all: tot(cAll),
      agent: tot(cAgent),
      puller: tot(cPuller),
      direct: tot(cDirect),
      offline: tot(cOffline)+tot(cNotConnected),
    })
    setLoading(false)
  }, [range, setLoading])

  useEffect(() => {
    void loadAux()
  }, [loadAux])

  const refresh = () => {
    void loadAux()
    void load()
  }

  const events24h = (name: string) => stats[name]?.count ?? 0

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <Header total={total} />

      <div className="mt-5 grid grid-cols-12 gap-4">
        <div className="col-span-12 lg:col-span-8">
          <IngestionCard timeline={timeline} range={range} onRange={setRange} />
        </div>
        <div className="col-span-12 lg:col-span-4">
          <UsageCard count={usage} />
        </div>
      </div>

      {counts && <HealthSummary counts={counts} />}

      <div className="mt-5">
        <Tabs current={tab} counts={counts} onChange={(t) => { setTab(t); setPage(0) }} />
      </div>

      <Toolbar
        search={search}
        onSearch={setSearch}
        layout={layout}
        onLayout={setLayout}
        onRefresh={refresh}
        loading={loading}
      />

      {labelFilter && (
        <div className="mt-3 flex items-center gap-2 text-xs">
          <span className="text-muted-foreground">{t('datasources.labels.filteredBy')}</span>
          <span className="inline-flex items-center gap-1.5 rounded-full border border-primary/25 bg-primary/5 px-2 py-1">
            <Tag size={11} />
            <span className="font-medium">{labelFilter}</span>
            <button
              onClick={() => { setLabelFilter(null); setPage(0) }}
              className="flex h-4 w-4 items-center justify-center rounded-full text-muted-foreground hover:bg-foreground/10 hover:text-foreground"
            >
              <X size={11} />
            </button>
          </span>
        </div>
      )}

      <div className="mt-3">
        {loading && sources.length === 0 ? (
          <CenterCard>
            <Loader2 className="h-4 w-4 animate-spin" /> {t('datasources.loading')}
          </CenterCard>
        ) : error ? (
          <CenterCard>
            <AlertTriangle size={16} className="text-amber-500" /> {t('datasources.loadFailed')}
            <Button variant="outline" size="sm" className="ml-2" onClick={refresh}>
              {t('datasources.retry')}
            </Button>
          </CenterCard>
        ) : sources.length === 0 ? (
          <CenterCard>{t('datasources.none')}</CenterCard>
        ) : layout === 'list' ? (
          <div className="overflow-hidden rounded-xl border border-border bg-card">
            <ListHeader />
            {sources.map((s) => (
              <SourceListRow key={s.id} source={s} events24h={events24h(s.name)} onOpen={() => setOpenId(s.id)} onLabelClick={filterByLabel} />
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
            {sources.map((s) => (
              <SourceCard key={s.id} source={s} events24h={events24h(s.name)} onOpen={() => setOpenId(s.id)} onLabelClick={filterByLabel} />
            ))}
          </div>
        )}
      </div>

      {!error && sources.length > 0 && (
        <InfiniteScrollSentinel
          onReach={() => setPage((p) => p + 1)}
          hasMore={sources.length < total}
          loading={loading}
          endLabel={t('common.allLoaded', { count: total })}
        />
      )}

      {openId != null && (
        <SourceDrawer
          id={openId}
          events24h={(n) => events24h(n)}
          onClose={() => setOpenId(null)}
          onChanged={() => { void load(); void loadAux() }}
        />
      )}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ total }: { total: number }) {
  const { t } = useTranslation()
  return (
    <header className="flex flex-wrap items-center justify-between gap-3">
      <div className="text-xs text-muted-foreground">
        <span className="font-medium text-foreground">{t('datasources.total', { count: total })}</span>
      </div>
      <Button asChild size="sm">
        <Link to="/integrations">
          <Plus size={14} className="mr-2" />
          {t('datasources.connectSource')}
        </Link>
      </Button>
    </header>
  )
}

/* ─── Ingestion overview (real timeline) ───────────────────────────────── */

function IngestionCard({
  timeline,
  range,
  onRange,
}: {
  timeline: TimelinePoint[]
  range: TimeRange
  onRange: (r: TimeRange) => void
}) {
  const { t } = useTranslation()
  const total = timeline.reduce((a, p) => a + p.count, 0)
  const values = timeline.map((p) => p.count)
  const max = Math.max(1, ...values)
  const w = 1000
  const h = 100
  const xs = values.map((_, i) => (i * w) / Math.max(1, values.length - 1))
  const ys = values.map((v) => h - (v / max) * h)
  const linePath = values.reduce((acc, _, i) => {
    if (i === 0) return `M ${xs[i]} ${ys[i]}`
    const px = xs[i - 1]
    return `${acc} C ${px + (xs[i] - px) / 2} ${ys[i - 1]}, ${xs[i] - (xs[i] - px) / 2} ${ys[i]}, ${xs[i]} ${ys[i]}`
  }, '')
  const areaPath = values.length ? `${linePath} L ${xs[xs.length - 1]} ${h} L ${xs[0]} ${h} Z` : ''

  return (
    <div className="rounded-xl border border-border bg-card p-6">
      <div className="flex items-start justify-between gap-3">
        <div>
          <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
            {t('datasources.ingestion.title')}
          </div>
          <div className="mt-1 flex items-baseline gap-2">
            <span className="text-3xl font-semibold tabular-nums">{formatCount(total)}</span>
            <span className="text-sm text-muted-foreground">{t('datasources.ingestion.received')}</span>
          </div>
        </div>
        <TimeRangePicker value={range} onChange={onRange} align="right" />
      </div>
      {values.length === 0 ? (
        <div className="mt-4 flex h-24 items-center justify-center text-xs text-muted-foreground">
          {t('datasources.ingestion.none')}
        </div>
      ) : (
        <svg viewBox={`0 0 ${w} ${h}`} className="mt-4 h-24 w-full" preserveAspectRatio="none">
          <defs>
            <linearGradient id="dsIngestGrad" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="rgb(34 211 238)" stopOpacity="0.28" />
              <stop offset="100%" stopColor="rgb(34 211 238)" stopOpacity="0" />
            </linearGradient>
          </defs>
          <path d={areaPath} fill="url(#dsIngestGrad)" />
          <path d={linePath} fill="none" stroke="rgb(34 211 238)" strokeOpacity="0.85" strokeWidth="1.75" strokeLinejoin="round" strokeLinecap="round" />
        </svg>
      )}
    </div>
  )
}

function UsageCard({ count }: { count: DatasourceCount | null }) {
  const { t } = useTranslation()
  return (
    <div className="flex h-full flex-col rounded-xl border border-border bg-card p-5">
      <div className="flex items-center gap-2">
        <Layers size={14} className="text-muted-foreground" />
        <h3 className="text-sm font-semibold">{t('datasources.usage.title')}</h3>
      </div>
      {count ? (
        <div className="mt-3 flex items-baseline gap-2">
          <span className="text-2xl font-semibold tabular-nums">{count.count.toLocaleString()}</span>
          <span className="text-sm text-muted-foreground">{t('datasources.usage.configured')}</span>
        </div>
      ) : (
        <div className="mt-3 text-xs text-muted-foreground">—</div>
      )}
    </div>
  )
}

/* ─── Tabs ─────────────────────────────────────────────────────────────── */

function Tabs({
  current,
  counts,
  onChange,
}: {
  current: TabId
  counts: Record<TabId, number> | null
  onChange: (t: TabId) => void
}) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-wrap items-center gap-1 border-b border-border">
      {TABS.map((v) => {
        const active = current === v.id
        const n = counts?.[v.id]
        return (
          <button
            key={v.id}
            onClick={() => onChange(v.id)}
            className={cn(
              'relative flex items-center gap-1.5 px-3 py-2 text-xs transition-colors',
              active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
            )}
          >
            {t(`datasources.tabs.${v.id}`)}
            {n != null && (
              <span
                className={cn(
                  'rounded px-1.5 py-0.5 text-[10px] font-semibold tabular-nums',
                  v.id === 'offline' && n > 0
                    ? 'bg-red-500/15 text-red-600 dark:text-red-400'
                    : 'bg-muted text-muted-foreground'
                )}
              >
                {n.toLocaleString()}
              </span>
            )}
            {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
          </button>
        )
      })}
    </div>
  )
}

// Compact at-a-glance fleet health derived from the per-tab counts. "Online" is
// everything not stale-past-24h; "Offline" is the stale set. Gives the analyst the
// online/offline split the page otherwise hid behind the row-by-row status dots.
function HealthSummary({ counts }: { counts: Record<TabId, number> }) {
  const { t } = useTranslation()
  const offline = counts.offline
  const online = Math.max(0, counts.all - offline)
  const items = [
    { key: 'online', n: online, dot: 'bg-emerald-500', tone: 'text-emerald-600 dark:text-emerald-400' },
    { key: 'offline', n: offline, dot: 'bg-red-500', tone: 'text-red-600 dark:text-red-400' },
    { key: 'total', n: counts.all, dot: 'bg-muted-foreground/40', tone: 'text-foreground' },
  ]
  return (
    <div className="mt-4 flex flex-wrap items-center gap-x-5 gap-y-2 rounded-lg border border-border bg-card px-4 py-2.5">

      {items.map((it) => (
        <div key={it.key} className="flex items-center gap-2 text-sm">
          <span className={cn('h-2 w-2 rounded-full', it.dot)} />
          <span className="font-semibold tabular-nums">{it.n.toLocaleString()}</span>
          <span className="text-xs text-muted-foreground">{t(`datasources.health.${it.key}`)}</span>
        </div>
      ))}
    </div>
  )
}

/* ─── Toolbar ──────────────────────────────────────────────────────────── */

function Toolbar({
  search,
  onSearch,
  layout,
  onLayout,
  onRefresh,
  loading,
}: {
  search: string
  onSearch: (s: string) => void
  layout: 'list' | 'cards'
  onLayout: (l: 'list' | 'cards') => void
  onRefresh: () => void
  loading: boolean
}) {
  const { t } = useTranslation()
  return (
    <div className="mt-3 flex flex-wrap items-center gap-2">
      <div className="relative min-w-[260px] flex-1">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input placeholder={t('datasources.toolbar.searchPlaceholder')} value={search} onChange={(e) => onSearch(e.target.value)} className="h-9 pl-9" />
      </div>
      <div className="ml-auto flex items-center gap-1 rounded-md border border-border p-0.5">
        <button onClick={() => onLayout('list')} title={t('datasources.toolbar.list')} className={cn('flex h-7 w-7 items-center justify-center rounded transition-colors', layout === 'list' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground')}>
          <ListIcon size={13} />
        </button>
        <button onClick={() => onLayout('cards')} title={t('datasources.toolbar.cards')} className={cn('flex h-7 w-7 items-center justify-center rounded transition-colors', layout === 'cards' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground')}>
          <LayoutGrid size={13} />
        </button>
      </div>
      <Button variant="outline" size="sm" onClick={onRefresh} disabled={loading} title={t('datasources.toolbar.refresh')}>
        <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
      </Button>
    </div>
  )
}

/* ─── List ─────────────────────────────────────────────────────────────── */

const LIST_COLS = '36px 1fr 150px 110px 120px 120px'

function ListHeader() {
  const { t } = useTranslation()
  return (
    <div
      className="grid items-center gap-3 border-b border-border bg-muted/30 px-4 py-2.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: LIST_COLS }}
    >
      <div />
      <div>{t('datasources.cols.source')}</div>
      <div>{t('datasources.cols.type')}</div>
      <div>{t('datasources.cols.status')}</div>
      <div className="text-center">{t('datasources.cols.events24h')}</div>
      <div className="text-center">{t('datasources.cols.lastSeen')}</div>
    </div>
  )
}

function SourceListRow({
  source: s,
  events24h,
  onOpen,
  onLabelClick,
}: {
  source: Datasource
  events24h: number
  onOpen: () => void
  onLabelClick: (label: string) => void
}) {
  const { t } = useTranslation()
  const status = deriveStatus(s.lastPingAt)
  const st = STATUS_META[status]
  const typeLabel = dataTypeMeta(s.dataType).label || kindLabel(t, s.sourceKind)
  return (
    <div
      onClick={onOpen}
      className="grid cursor-pointer items-center gap-3 border-b border-border/50 px-4 py-3 text-[13px] hover:bg-muted/20 last:border-b-0"
      style={{ gridTemplateColumns: LIST_COLS }}
    >
      <SourceIcon source={s} />
      <div className="min-w-0">
        <div className="truncate font-medium">{s.name}</div>
        <div className="mt-0.5 flex min-w-0 items-center gap-1.5 text-[11px] text-muted-foreground">
          <span className="truncate">
            {s.ip && <span className="font-mono">{s.ip}</span>}
            {s.ip && s.dataType && ' · '}
            {s.dataType && <span className="font-mono">{s.dataType}</span>}
          </span>
          {labelList(s.labels).slice(0, 3).map((l) => (
            <LabelChip key={l} label={l} onClick={onLabelClick} />
          ))}
        </div>
      </div>
      <div className="truncate text-[12px] text-muted-foreground" title={s.dataType || typeLabel}>
        {typeLabel}
      </div>
      <div>
        <span className="inline-flex items-center gap-1.5 text-[12px] text-muted-foreground">
          <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
          {statusLabel(t, status)}
        </span>
      </div>
      <div className="text-center font-mono tabular-nums text-muted-foreground">
        {events24h > 0 ? events24h.toLocaleString() : '—'}
      </div>
      <div className="text-center font-mono text-[12px] text-muted-foreground">{relativeTime(t, s.lastPingAt)}</div>
    </div>
  )
}

function SourceIcon({ source }: { source: Datasource }) {
  const m = sourceIconMeta(source)
  return (
    <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-muted">
      <m.icon size={13} className={m.tone} />
    </div>
  )
}

/* ─── Cards ────────────────────────────────────────────────────────────── */

function SourceCard({
  source: s,
  events24h,
  onOpen,
  onLabelClick,
}: {
  source: Datasource
  events24h: number
  onOpen: () => void
  onLabelClick: (label: string) => void
}) {
  const { t } = useTranslation()
  const status = deriveStatus(s.lastPingAt)
  const st = STATUS_META[status]
  const typeLabel = dataTypeMeta(s.dataType).label || kindLabel(t, s.sourceKind)
  const labels = labelList(s.labels)
  return (
    <div
      onClick={onOpen}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => (e.key === 'Enter' || e.key === ' ') && onOpen()}
      className="group flex cursor-pointer flex-col gap-3 rounded-xl border border-border bg-card p-4 text-left transition-all hover:shadow-md"
    >
      <div className="flex items-start gap-3">
        <SourceIcon source={s} />
        <div className="min-w-0 flex-1">
          <div className="truncate text-sm font-semibold">{s.name}</div>
          <div className="truncate text-[11px] text-muted-foreground">
            {typeLabel}
            <span className="text-muted-foreground/60"> · {kindLabel(t, s.sourceKind)}</span>
          </div>
        </div>
        <span className="inline-flex items-center gap-1.5 text-[11px] text-muted-foreground">
          <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
          {statusLabel(t, status)}
        </span>
      </div>
      <div className="grid grid-cols-2 gap-2 border-t border-border pt-3 text-[11px]">
        <CardStat label={t('datasources.cols.events24h')} value={events24h > 0 ? formatCount(events24h) : '—'} />
        <CardStat label={t('datasources.cols.lastSeen')} value={relativeTime(t, s.lastPingAt)} />
      </div>
      {labels.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {labels.slice(0, 4).map((l) => (
            <LabelChip key={l} label={l} onClick={onLabelClick} />
          ))}
        </div>
      )}
      {s.ip && <div className="font-mono text-[11px] text-muted-foreground">{s.ip}</div>}
    </div>
  )
}

function LabelChip({ label, onClick }: { label: string; onClick: (label: string) => void }) {
  return (
    <button
      onClick={(e) => {
        e.stopPropagation()
        onClick(label)
      }}
      title={`Filter by ${label}`}
      className="inline-flex shrink-0 items-center gap-1 rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground transition-colors hover:bg-primary/15 hover:text-primary"
    >
      <Tag size={9} />
      {label}
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

function SourceDrawer({
  id,
  events24h,
  onClose,
  onChanged,
}: {
  id: string
  events24h: (name: string) => number
  onClose: () => void
  onChanged: () => void
}) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [src, setSrc] = useState<Datasource | null>(null)
  const [loading, setLoading] = useState(true)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [consoleOpen, setConsoleOpen] = useState(false)
  const [trend, setTrend] = useState<number[] | null>(null)

  // Pivot to Log Explorer scoped to this source's events (Discover-style drill-in).
  const viewLogs = () => {
    if (!src) return
    navigate('/log-explorer', {
      state: {
        socaiFilters: [{ field: 'dataSource', operator: 'IS', value: src.name }],
        socaiTime: '7d',
      },
    })
  }

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    svc
      .get(id)
      .then((d) => {
        if (cancelled) return
        setSrc(d)
      })
      .catch(() => !cancelled && toast.error(t('datasources.drawer.loadFailed')))
      .finally(() => !cancelled && setLoading(false))
    return () => { cancelled = true }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  // Per-source ingestion trend (sparkline) — fetched once the source name is known.
  useEffect(() => {
    if (!src?.name) return
    let cancelled = false
    setTrend(null)
    svc
      .ingestionTimelineFor(src.name)
      .then((tl) => {
        if (!cancelled) setTrend((tl.points ?? []).map((p) => p.count))
      })
      .catch(() => {
        if (!cancelled) setTrend([])
      })
    return () => {
      cancelled = true
    }
  }, [src?.name])

  // Persist a labels array (joined to the comma-separated form the backend stores).
  const saveLabels = async (next: string[]): Promise<boolean> => {
    if (!src) return false
    const joined = next.join(', ')
    try {
      await svc.updateLabels(src.id, joined)
      setSrc({ ...src, labels: joined })
      onChanged()
      return true
    } catch (e) {
      toast.error(e instanceof DatasourcesHttpError ? e.message : t('datasources.toast.labelsFailed'))
      return false
    }
  }

  const saveSensitivity = async (cia: {
    assetConfidentiality: number
    assetIntegrity: number
    assetAvailability: number
  }): Promise<boolean> => {
    if (!src) return false
    try {
      await svc.updateSensitivity(src.id, cia)
      setSrc({ ...src, ...cia })
      toast.success(t('datasources.sensitivity.saved'))
      onChanged()
      return true
    } catch (e) {
      toast.error(e instanceof DatasourcesHttpError ? e.message : t('datasources.sensitivity.failed'))
      return false
    }
  }

  const doDelete = async () => {
    if (!src) return
    try {
      await svc.remove(src.id)
      toast.success(t('datasources.toast.deleted'))
      onChanged()
      onClose()
    } catch {
      toast.error(t('datasources.toast.deleteFailed'))
    }
  }

  const statusKey = src ? deriveStatus(src.lastPingAt) : null
  const st = statusKey ? STATUS_META[statusKey] : null
  const sicon = src ? sourceIconMeta(src) : null
  const typeLabel = src ? dataTypeMeta(src.dataType).label : ''
  const metaEntries = src?.metadata ? Object.entries(src.metadata) : []

  return (
    <>
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div className="flex w-full max-w-[680px] flex-col overflow-hidden border-l border-border bg-card shadow-xl" onClick={(e) => e.stopPropagation()}>
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="flex min-w-0 flex-1 items-start gap-3">
              {sicon && (
                <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-muted">
                  <sicon.icon size={20} className={sicon.tone} />
                </div>
              )}
              <div className="min-w-0 flex-1">
                {src && (
                  <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                    {typeLabel && <span className="font-medium text-foreground/80">{typeLabel}</span>}
                    {typeLabel && <span>·</span>}
                    <span>{kindLabel(t, src.sourceKind)}</span>
                    {src.dataType && (<><span>·</span><span className="font-mono">{src.dataType}</span></>)}
                  </div>
                )}
                <h2 className="mt-0.5 truncate text-xl font-semibold">{src?.name ?? '…'}</h2>
                {src?.ip && <div className="truncate font-mono text-xs text-muted-foreground">{src.ip}</div>}
                {st && statusKey && (
                  <div className="mt-2 flex items-center gap-2 text-[11px]">
                    <span className="inline-flex items-center gap-1.5 text-muted-foreground">
                      <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
                      {statusLabel(t, statusKey)}
                    </span>
                  </div>
                )}
              </div>
            </div>
            <div className="flex shrink-0 items-center gap-2">
              {src && (
                <Button variant="outline" size="sm" onClick={viewLogs} disabled={!src.dataType && !src.name}>
                  <ScrollText size={13} className="mr-1.5" />
                  {t('datasources.drawer.viewLogs')}
                </Button>
              )}
              <button onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
                <X size={16} />
              </button>
            </div>
          </div>
        </header>

        <div className="flex-1 overflow-y-auto bg-muted/10 p-6">
          {loading || !src ? (
            <div className="flex items-center justify-center gap-2 py-16 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" /> {t('datasources.drawer.loading')}
            </div>
          ) : (
            <div className="space-y-4">
              {statusKey === 'offline' && src.lastPingAt && (
                <div className="flex items-start gap-2 rounded-lg border border-amber-500/40 bg-amber-500/10 px-3 py-2.5 text-xs text-amber-700 dark:text-amber-300">
                  <AlertTriangle size={14} className="mt-0.5 shrink-0" />
                  <span>{t('datasources.drawer.logGap', { since: relativeTime(t, src.lastPingAt) })}</span>
                </div>
              )}

              <Section title={t('datasources.drawer.statsIdentity')}>
                <dl className="grid grid-cols-[140px_1fr] gap-y-2 text-xs">
                  <Row k={t('datasources.drawer.rows.events24h')}>{events24h(src.name) > 0 ? events24h(src.name).toLocaleString() : '—'}</Row>
                  <Row k={t('datasources.drawer.rows.lastSeen')}>{relativeTime(t, src.lastPingAt)}</Row>
                  <Row k={t('datasources.drawer.rows.discovered')}>{relativeTime(t, src.discoveredAt)}</Row>
                  <Row k={t('datasources.drawer.rows.kind')}>{kindLabel(t, src.sourceKind)}</Row>
                  {src.dataType && <Row k={t('datasources.drawer.rows.dataType')}><span className="font-mono">{src.dataType}</span></Row>}
                  {src.ip && <Row k={t('datasources.drawer.rows.ip')}><span className="font-mono">{src.ip}</span></Row>}
                </dl>
                <div className="mt-3 border-t border-border/60 pt-3">
                  <div className="mb-1.5 text-[11px] uppercase tracking-wider text-muted-foreground">
                    {t('datasources.drawer.trend')}
                  </div>
                  <Sparkline values={trend} />
                </div>
              </Section>

              {metaEntries.length > 0 && (
                <Section title={t('datasources.drawer.metadata')}>
                  <dl className="grid grid-cols-[140px_1fr] gap-y-2 text-xs">
                    {metaEntries.map(([k, v]) => (
                      <Row key={k} k={k}><span className="break-all font-mono">{String(v)}</span></Row>
                    ))}
                  </dl>
                </Section>
              )}

              <Section title={t('datasources.drawer.labelsSection')}>
                <LabelsEditor value={src.labels} onSave={saveLabels} />
                <p className="mt-2 text-[11px] text-muted-foreground">{t('datasources.labels.hint')}</p>
              </Section>

              <Section title={t('datasources.sensitivity.section')}>
                <SensitivityEditor
                  value={{
                    assetConfidentiality: src.assetConfidentiality ?? 0,
                    assetIntegrity: src.assetIntegrity ?? 0,
                    assetAvailability: src.assetAvailability ?? 0,
                  }}
                  onSave={saveSensitivity}
                />
                <p className="mt-2 text-[11px] text-muted-foreground">{t('datasources.sensitivity.hint')}</p>
              </Section>

              {src.sourceKind === 'agent' && src.metadata?.agentId != null && (
                <Section title={t('datasources.console.section')}>
                  <p className="text-xs text-muted-foreground">{t('datasources.console.sectionHint')}</p>
                  <Button size="sm" variant="outline" className="mt-3" onClick={() => setConsoleOpen(true)}>
                    <Terminal size={13} className="mr-1.5" /> {t('datasources.console.open')}
                  </Button>
                </Section>
              )}

              <Section title={t('datasources.drawer.dangerZone')}>
                <p className="text-xs text-muted-foreground">{t('datasources.drawer.dangerText')}</p>
                {confirmDelete ? (
                  <div className="mt-3 flex items-center gap-2">
                    <Button size="sm" variant="destructive" onClick={() => void doDelete()}>
                      <Trash2 size={13} className="mr-1.5" /> {t('datasources.drawer.confirmDelete')}
                    </Button>
                    <Button size="sm" variant="outline" onClick={() => setConfirmDelete(false)}>{t('datasources.drawer.cancel')}</Button>
                  </div>
                ) : (
                  <button onClick={() => setConfirmDelete(true)} className="mt-3 rounded-md border border-red-500/30 bg-red-500/5 px-3 py-1.5 text-xs font-medium text-red-600 hover:bg-red-500/10 dark:text-red-300">
                    {t('datasources.drawer.delete')}
                  </button>
                )}
              </Section>
            </div>
          )}
        </div>
      </div>
    </div>
    {consoleOpen && src && src.metadata?.agentId != null && (
      <AgentConsole
        agentId={String(src.metadata.agentId)}
        hostname={src.name}
        osPlatform={typeof src.metadata.osPlatform === 'string' ? src.metadata.osPlatform : undefined}
        offline={deriveStatus(src.lastPingAt) === 'offline'}
        onClose={() => setConsoleOpen(false)}
      />
    )}
    </>
  )
}

/* ─── Small parts ──────────────────────────────────────────────────────── */

// Compact per-source ingestion sparkline. null = loading, [] = no data. Gaps in
// the series (zero buckets) make a silent source visually obvious.
function Sparkline({ values }: { values: number[] | null }) {
  const { t } = useTranslation()
  if (values == null) {
    return (
      <div className="flex h-10 items-center gap-1.5 text-[11px] text-muted-foreground">
        <Loader2 className="h-3 w-3 animate-spin" /> {t('datasources.drawer.loading')}
      </div>
    )
  }
  if (values.length === 0 || values.every((v) => v <= 0)) {
    return <div className="flex h-10 items-center text-[11px] text-muted-foreground">{t('datasources.drawer.noTrend')}</div>
  }
  const max = Math.max(1, ...values)
  return (
    <div className="flex h-10 items-end justify-center gap-px">
      {values.map((v, i) => (
        <div key={i} className="flex h-full max-w-[10px] flex-1 items-end">
          <div
            className="w-full rounded-t-[1px] bg-primary/45"
            style={{ height: `${v <= 0 ? 0 : Math.max(3, (v / max) * 100)}%` }}
            title={v.toLocaleString()}
          />
        </div>
      ))}
    </div>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <h4 className="mb-3 text-sm font-semibold">{title}</h4>
      {children}
    </div>
  )
}

type CIA = { assetConfidentiality: number; assetIntegrity: number; assetAvailability: number }

const CIA_AXES: { key: keyof CIA; label: string }[] = [
  { key: 'assetConfidentiality', label: 'datasources.sensitivity.confidentiality' },
  { key: 'assetIntegrity', label: 'datasources.sensitivity.integrity' },
  { key: 'assetAvailability', label: 'datasources.sensitivity.availability' },
]

// Severity tone per level (0 None → 3 High), mirroring the alert severity palette.
const CIA_LEVEL_TONE: Record<number, string> = {
  0: 'border-border bg-muted text-muted-foreground',
  1: 'border-sky-500/40 bg-sky-500/15 text-sky-600 dark:text-sky-400',
  2: 'border-amber-500/40 bg-amber-500/15 text-amber-600 dark:text-amber-400',
  3: 'border-red-500/40 bg-red-500/15 text-red-600 dark:text-red-400',
}

function SensitivityEditor({ value, onSave }: { value: CIA; onSave: (cia: CIA) => Promise<boolean> }) {
  const { t } = useTranslation()
  const [draft, setDraft] = useState<CIA>(value)
  const [saving, setSaving] = useState(false)

  // Re-sync when the persisted value changes (drawer reload / after save).
  useEffect(() => {
    setDraft(value)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value.assetConfidentiality, value.assetIntegrity, value.assetAvailability])

  const dirty =
    draft.assetConfidentiality !== value.assetConfidentiality ||
    draft.assetIntegrity !== value.assetIntegrity ||
    draft.assetAvailability !== value.assetAvailability

  const save = async () => {
    setSaving(true)
    await onSave(draft)
    setSaving(false)
  }

  return (
    <div className="space-y-3">
      {CIA_AXES.map((axis) => (
        <div key={axis.key} className="flex items-center justify-between gap-3">
          <span className="text-xs text-foreground/80">{t(axis.label)}</span>
          <div className="inline-flex gap-1">
            {[0, 1, 2, 3].map((lvl) => (
              <button
                key={lvl}
                type="button"
                onClick={() => setDraft((d) => ({ ...d, [axis.key]: lvl }))}
                title={t(`datasources.sensitivity.levels.${lvl}`)}
                className={cn(
                  'h-7 w-7 rounded-md border text-xs font-semibold transition-colors',
                  draft[axis.key] === lvl
                    ? CIA_LEVEL_TONE[lvl]
                    : 'border-border bg-background/40 text-muted-foreground hover:bg-muted',
                )}
              >
                {lvl}
              </button>
            ))}
          </div>
        </div>
      ))}
      {dirty && (
        <div className="flex justify-end pt-1">
          <Button size="sm" className="h-7" disabled={saving} onClick={() => void save()}>
            {saving && <Loader2 size={12} className="mr-1.5 animate-spin" />}
            {t('datasources.sensitivity.save')}
          </Button>
        </div>
      )}
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

function LabelsEditor({ value, onSave }: { value?: string; onSave: (labels: string[]) => Promise<boolean> }) {
  const { t } = useTranslation()
  const [labels, setLabels] = useState<string[]>(() => labelList(value))
  const [input, setInput] = useState('')
  const [busy, setBusy] = useState(false)

  // Re-sync when the source reloads (e.g. after a successful save upstream).
  useEffect(() => setLabels(labelList(value)), [value])

  const commit = async (next: string[]) => {
    setBusy(true)
    const ok = await onSave(next)
    setBusy(false)
    if (ok) setLabels(next)
  }

  const add = () => {
    const v = input.trim().replace(/,/g, '')
    setInput('')
    if (!v || labels.includes(v)) return
    void commit([...labels, v])
  }
  const remove = (l: string) => void commit(labels.filter((x) => x !== l))

  return (
    <div className="flex flex-wrap items-center gap-1.5 rounded-md border border-input bg-background/40 px-2 py-2">
      {labels.map((l) => (
        <span key={l} className="inline-flex items-center gap-1 rounded-md bg-muted px-2 py-0.5 text-[11px]">
          <Tag size={10} /> {l}
          <button
            onClick={() => remove(l)}
            disabled={busy}
            className="ml-0.5 text-muted-foreground transition-colors hover:text-foreground disabled:opacity-50"
          >
            <X size={11} />
          </button>
        </span>
      ))}
      <input
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ',') {
            e.preventDefault()
            add()
          } else if (e.key === 'Backspace' && !input && labels.length) {
            remove(labels[labels.length - 1])
          }
        }}
        onBlur={add}
        placeholder={labels.length ? t('datasources.labels.add') : t('datasources.labels.addFirst')}
        className="min-w-[110px] flex-1 bg-transparent text-xs outline-none placeholder:text-muted-foreground"
      />
    </div>
  )
}

function CenterCard({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-center gap-2 rounded-xl border border-border bg-card px-6 py-16 text-sm text-muted-foreground">
      {children}
    </div>
  )
}

/* ─── Helpers ──────────────────────────────────────────────────────────── */

function labelList(labels?: string): string[] {
  return (labels ?? '').split(',').map((s) => s.trim()).filter(Boolean)
}

function relativeTime(t: TFunction, iso?: string) {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  if (Number.isNaN(diff)) return '—'
  const m = Math.round(diff / 60_000)
  if (m < 1) return t('datasources.time.justNow')
  if (m < 60) return t('datasources.time.minutesAgo', { count: m })
  const h = Math.round(m / 60)
  if (h < 24) return t('datasources.time.hoursAgo', { count: h })
  return t('datasources.time.daysAgo', { count: Math.round(h / 24) })
}

/** Translated kind label (META keeps icon/colour only; status label lives in lib/status). */
function kindLabel(t: TFunction, kind?: string) {
  const k = (kind as SourceKind) ?? 'direct'
  return KIND_META[k] ? t(`datasources.kind.${k}`) : kind ?? '—'
}

function formatCount(n: number) {
  if (n >= 1_000_000_000) return `${(n / 1_000_000_000).toFixed(1)}B`
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`
  return n.toString()
}
