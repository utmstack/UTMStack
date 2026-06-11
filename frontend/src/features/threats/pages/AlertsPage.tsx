import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  AlertTriangle,
  ArrowRight,
  BarChart3,
  Check,
  ChevronDown,
  Download,
  FileSearch,
  Flame,
  ListFilter,
  Loader2,
  Lock,
  Pencil,
  Plus,
  RefreshCw,
  Trash2,
  Rows3,
  Search,
  ShieldAlert,
  Sparkles,
  Tag as TagIcon,
  X,
} from 'lucide-react'
import { toast } from 'sonner'
import { useLocation, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { Pagination } from '@/shared/components/ui/pagination'
import { TimeRangePicker, presetRange, type TimeRange } from '@/shared/components/ui/time-range-picker'
import { LogResults } from '@/features/log-explorer/components/log-results'
import type { LogDocument } from '@/features/log-explorer/types/log-explorer.types'
import { alertsHttpService as svc, AlertsHttpError } from '../services/alerts-http.service'
import {
  incidentsHttpService,
  type AlertLinkItem,
  type Incident,
} from '../services/incidents-http.service'
import {
  SEVERITY_BY_INT,
  SEVERITY_INT,
  STATUS_BY_INT,
  STATUS_INT,
  type Alert,
  type AlertEventItem,
  type AlertTag,
  type FilterType,
  type SeverityKey,
  type Side,
  type StatusKey,
} from '../types/alert.types'

/* ─── Meta / styles ────────────────────────────────────────────────────── */

const SEV_META: Record<SeverityKey, { label: string; bar: string; pill: string }> = {
  high: { label: 'High', bar: 'bg-red-500', pill: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300' },
  medium: { label: 'Medium', bar: 'bg-amber-500', pill: 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300' },
  low: { label: 'Low', bar: 'bg-sky-500', pill: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300' },
}

const ST_META: Record<StatusKey, { label: string; pill: string }> = {
  open: { label: 'Open', pill: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300' },
  in_review: { label: 'In review', pill: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300' },
  completed: { label: 'Completed', pill: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300' },
  auto: { label: 'Automatic review', pill: 'bg-violet-500/15 text-violet-600 ring-violet-500/30 dark:text-violet-300' },
}

function sevKey(a: Alert): SeverityKey {
  return SEVERITY_BY_INT[a.severity ?? 1] ?? 'low'
}
function statusKey(a: Alert): StatusKey {
  return STATUS_BY_INT[a.status ?? 2] ?? 'open'
}

/* A tag chip painted with the tag's own colour (from the tag catalog). The alert
 * only stores tag names, so we look up the colour by name. */
function TagChip({ name, catalog, size = 'sm' }: { name: string; catalog: AlertTag[]; size?: 'sm' | 'xs' }) {
  const color = catalog.find((t) => t.tagName === name)?.tagColor
  const style = color ? { backgroundColor: `${color}26`, color, borderColor: `${color}59` } : undefined
  return (
    <span
      className={cn(
        'inline-flex max-w-[140px] shrink-0 items-center gap-1 whitespace-nowrap rounded-md border font-medium leading-none',
        size === 'xs' ? 'px-1.5 py-0.5 text-[10px]' : 'px-2 py-0.5 text-[11px]',
        !color && 'border-border bg-muted text-muted-foreground'
      )}
      style={style}
      title={name}
    >
      <TagIcon size={size === 'xs' ? 9 : 10} className="shrink-0" />
      <span className="truncate">{name}</span>
    </span>
  )
}

const STATUS_TABS = ['all', 'open', 'in_review', 'completed', 'auto'] as const
type StatusTab = (typeof STATUS_TABS)[number]

const TS = '@timestamp'
const PAGE_SIZE_DEFAULT = 20

// Filterable alert fields (ported from the legacy ALERT_FILTERS_FIELDS).
const FILTER_FIELDS: { label: string; field: string }[] = [
  { label: 'Datasource group', field: 'assetGroupName' },
  { label: 'Category', field: 'category' },
  { label: 'Sensor', field: 'dataSource' },
  { label: 'Tags', field: 'tags' },
  { label: 'Incident name', field: 'incidentDetail.incidentName' },
  { label: 'Availability', field: 'impact.availability' },
  { label: 'Confidentiality', field: 'impact.confidentiality' },
  { label: 'Integrity', field: 'impact.integrity' },
  { label: 'Protocol', field: 'protocol' },
  { label: 'Adversary IP', field: 'adversary.ip' },
  { label: 'Adversary domain', field: 'adversary.domain' },
  { label: 'Adversary URL', field: 'adversary.url' },
  { label: 'Adversary ASN', field: 'adversary.geolocation.asn' },
  { label: 'Adversary ASO', field: 'adversary.geolocation.aso' },
  { label: 'Target IP', field: 'target.ip' },
  { label: 'Target domain', field: 'target.domain' },
  { label: 'Target URL', field: 'target.url' },
  { label: 'Target ASN', field: 'target.geolocation.asn' },
  { label: 'Target ASO', field: 'target.geolocation.aso' },
]

// Field paths contain dots; flatten them to a flat i18n key (dots are the
// i18next nesting separator, so they can't appear inside a single key).
const fieldKey = (field: string) => field.replace(/\./g, '_')

const FILTER_OPS: { id: string; label: string; needsValue: boolean }[] = [
  { id: 'CONTAIN', label: 'contains', needsValue: true },
  { id: 'IS', label: 'is', needsValue: true },
  { id: 'IS_NOT', label: 'is not', needsValue: true },
  { id: 'EXIST', label: 'exists', needsValue: false },
  { id: 'DOES_NOT_EXIST', label: 'does not exist', needsValue: false },
]

interface CustomFilter {
  field: string
  label: string
  operator: string
  value: string
}

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function AlertsPage() {
  const { t } = useTranslation()
  const [view, setView] = useState<'alerts' | 'overview'>('alerts')
  const [statusTab, setStatusTab] = useState<StatusTab>('all')
  const [severity, setSeverity] = useState<SeverityKey | 'all'>('all')
  const [range, setRange] = useState<TimeRange>(() => presetRange('7d'))
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [tagFilter, setTagFilter] = useState<string[]>([])
  const [customFilters, setCustomFilters] = useState<CustomFilter[]>([])

  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(PAGE_SIZE_DEFAULT)
  const [total, setTotal] = useState(0)
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const [sevCounts, setSevCounts] = useState<Record<string, number>>({})
  const [statusCounts, setStatusCounts] = useState<Record<string, number>>({})
  const [timeline, setTimeline] = useState<number[]>([])
  const [openCount, setOpenCount] = useState<number | null>(null)
  const [tagCatalog, setTagCatalog] = useState<AlertTag[]>([])

  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [openAlert, setOpenAlert] = useState<Alert | null>(null)
  const [incidentTargets, setIncidentTargets] = useState<Alert[] | null>(null)

  const selectedAlerts = useMemo(() => alerts.filter((a) => selected.has(a.id)), [alerts, selected])

  // SOC-AI chat navigation: seed the filters + time window the agent emitted.
  const location = useLocation()
  const navigate = useNavigate()
  const seededRef = useRef(false)
  useEffect(() => {
    const state = location.state as { socaiFilters?: FilterType[]; socaiTime?: string } | null
    if (!state?.socaiFilters?.length || seededRef.current) return
    seededRef.current = true
    setCustomFilters(
      state.socaiFilters.map((f) => ({
        field: f.field,
        label: f.field,
        operator: f.operator,
        value: Array.isArray(f.value) ? f.value.join(', ') : String(f.value ?? ''),
      })),
    )
    if (state.socaiTime) setRange(presetRange(state.socaiTime))
    setPage(0)
    // Drop the router state so a refresh doesn't re-seed over the user's edits.
    navigate(location.pathname, { replace: true })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    const h = setTimeout(() => {
      setDebounced(search.trim())
      setPage(0)
    }, 300)
    return () => clearTimeout(h)
  }, [search])

  // Scope filters (time + search). `parentId DOES_NOT_EXIST` hides deduplicated
  // child "echoes" — they roll up under their parent (matches the legacy default).
  const scopeFilters = useMemo<FilterType[]>(() => {
    const f: FilterType[] = [{ field: 'parentId', operator: 'DOES_NOT_EXIST' }]
    if (range.from) f.push({ field: TS, operator: 'IS_BETWEEN', value: [range.from, range.to] })
    if (debounced) f.push({ field: 'name', operator: 'CONTAIN', value: debounced })
    // Tag selector: match alerts carrying ANY of the chosen tags (OR semantics).
    if (tagFilter.length) f.push({ field: 'tags', operator: 'IS_ONE_OF', value: tagFilter })
    for (const cf of customFilters) {
      const needsValue = FILTER_OPS.find((o) => o.id === cf.operator)?.needsValue ?? true
      f.push({ field: cf.field, operator: cf.operator, value: needsValue ? cf.value : undefined })
    }
    return f
  }, [range.from, range.to, debounced, tagFilter, customFilters])

  const addFilter = (cf: CustomFilter) => { setCustomFilters((c) => [...c, cf]); setPage(0) }
  const removeFilter = (i: number) => { setCustomFilters((c) => c.filter((_, idx) => idx !== i)); setPage(0) }

  // List filters add the tab (status) + severity refinement.
  const listFilters = useMemo<FilterType[]>(() => {
    const f = [...scopeFilters]
    if (statusTab !== 'all') f.push({ field: 'status', operator: 'IS', value: STATUS_INT[statusTab as StatusKey] })
    if (severity !== 'all') f.push({ field: 'severity', operator: 'IS', value: SEVERITY_INT[severity] })
    return f
  }, [scopeFilters, statusTab, severity])

  const loadList = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      const { data, total } = await svc.list({ page: page + 1, size: pageSize, filters: listFilters })
      setAlerts(data ?? [])
      setTotal(total)
    } catch {
      setError(true)
      setAlerts([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, listFilters])

  useEffect(() => {
    void loadList()
  }, [loadList])

  const loadStats = useCallback(async () => {
    const [sev, st, tl, open, tags] = await Promise.allSettled([
      svc.counts('severity', scopeFilters),
      svc.counts('status', scopeFilters),
      svc.timeline(scopeFilters, range.interval),
      svc.countOpen(),
      svc.tags(),
    ])
    if (sev.status === 'fulfilled') setSevCounts(Object.fromEntries(sev.value.top.map((b) => [b.value, b.count])))
    if (st.status === 'fulfilled') setStatusCounts(Object.fromEntries(st.value.top.map((b) => [b.value, b.count])))
    if (tl.status === 'fulfilled') setTimeline(tl.value.values ?? [])
    if (open.status === 'fulfilled') setOpenCount(open.value)
    if (tags.status === 'fulfilled') setTagCatalog(tags.value ?? [])
  }, [scopeFilters, range.interval])

  useEffect(() => {
    void loadStats()
  }, [loadStats])

  const refresh = () => {
    void loadList()
    void loadStats()
  }

  const toggleSel = (id: string) =>
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  const allChecked = alerts.length > 0 && alerts.every((a) => selected.has(a.id))
  const togglePage = () =>
    setSelected((prev) => {
      const next = new Set(prev)
      if (allChecked) alerts.forEach((a) => next.delete(a.id))
      else alerts.forEach((a) => next.add(a.id))
      return next
    })

  // Apply an action to a set of ids, then refresh.
  // Re-fetch the open alert after a mutation so its status/tags/history reflect
  // the change. Small delay lets OpenSearch make the updated doc searchable.
  const refetchOpen = (id: string) => {
    setTimeout(() => {
      void svc
        .getById(id)
        .then((fresh) => fresh && setOpenAlert((prev) => (prev && prev.id === id ? fresh : prev)))
        .catch(() => {})
    }, 1200)
  }

  const applyStatus = async (ids: string[], status: number, observation = '', fp = false) => {
    try {
      await svc.updateStatus(ids, status, observation, fp)
      toast.success(t('alerts.toast.statusUpdated'))
      setSelected(new Set())
      refresh()
      if (openAlert && ids.includes(openAlert.id)) {
        setOpenAlert({ ...openAlert, status })
        refetchOpen(openAlert.id)
      }
    } catch (e) {
      toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.toast.statusFailed'))
    }
  }
  const applyTags = async (ids: string[], tags: string[]) => {
    try {
      await svc.updateTags(ids, tags)
      toast.success(t('alerts.toast.tagsUpdated'))
      setSelected(new Set())
      refresh()
      if (openAlert && ids.includes(openAlert.id)) {
        setOpenAlert({ ...openAlert, tags })
        refetchOpen(openAlert.id)
      }
    } catch (e) {
      toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.toast.tagsFailed'))
    }
  }
  const refreshCatalog = async () => setTagCatalog((await svc.tags()) ?? [])
  // Register a new tag in the catalog, then apply it to the given alerts.
  const createAndApplyTag = async (ids: string[], tagName: string, tagColor: string, existing: string[]) => {
    try {
      await svc.createTag(tagName, tagColor)
      await refreshCatalog()
      await applyTags(ids, [...existing, tagName])
    } catch (e) {
      toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.toast.tagCreateFailed'))
    }
  }
  // Register a brand-new tag in the catalog (no alert to apply it to — used by
  // the toolbar tag filter's "create" action).
  const createCatalogTag = async (tagName: string, tagColor: string) => {
    try {
      await svc.createTag(tagName, tagColor)
      await refreshCatalog()
    } catch (e) {
      toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.toast.tagCreateFailed'))
    }
  }
  // Rename / recolor a catalog tag (backend blocks system-owned ones with 403).
  const updateCatalogTag = async (id: number, tagName: string, tagColor: string) => {
    try {
      await svc.updateTag(id, tagName, tagColor)
      await refreshCatalog()
      toast.success(t('alerts.toast.tagUpdated'))
    } catch (e) {
      toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.toast.tagUpdateFailed'))
    }
  }
  // Delete a catalog tag and drop it from the active tag filter.
  const deleteCatalogTag = async (id: number, tagName: string) => {
    try {
      await svc.deleteTag(id)
      await refreshCatalog()
      setTagFilter((tf) => tf.filter((tn) => tn !== tagName))
      toast.success(t('alerts.toast.tagDeleted'))
    } catch (e) {
      toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.toast.tagDeleteFailed'))
    }
  }
  // CSV export of the currently-filtered alert list.
  const exportCsv = async () => {
    try {
      await svc.exportCsv(listFilters)
    } catch (e) {
      toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.toast.exportFailed'))
    }
  }

  return (
    <div className="mx-auto flex h-full min-h-0 w-full max-w-[1100px] flex-col px-6 py-6">
      <Header total={total} openCount={openCount} view={view} onView={setView} />

      <div className="mt-4 shrink-0">
        <Toolbar
          search={search}
          onSearch={setSearch}
          severity={severity}
          onSeverity={(s) => { setSeverity(s); setPage(0) }}
          range={range}
          onRange={(r) => { setRange(r); setPage(0) }}
          tagCatalog={tagCatalog}
          tagFilter={tagFilter}
          onTagFilter={(t) => { setTagFilter(t); setPage(0) }}
          onCreateTag={(name, color) => void createCatalogTag(name, color)}
          onUpdateTag={(id, name, color) => void updateCatalogTag(id, name, color)}
          onDeleteTag={(id, name) => void deleteCatalogTag(id, name)}
          onRefresh={refresh}
          onExport={exportCsv}
          loading={loading}
          showSeverity={view === 'alerts'}
        />
      </div>

      <div className="mt-2 shrink-0">
        <FilterBar filters={customFilters} onAdd={addFilter} onRemove={removeFilter} onClear={() => { setCustomFilters([]); setPage(0) }} />
      </div>

      {view === 'overview' ? (
        <div className="mt-4 grid min-h-0 flex-1 grid-cols-12 gap-4">
          <div className="col-span-12 lg:col-span-7">
            <VolumeCard values={timeline} total={total} />
          </div>
          <div className="col-span-12 lg:col-span-5">
            <BreakdownCard sevCounts={sevCounts} statusCounts={statusCounts} />
          </div>
        </div>
      ) : (
        <>
      <div className="mt-4 shrink-0">
        <StatusTabs current={statusTab} counts={statusCounts} onChange={(s) => { setStatusTab(s); setPage(0) }} />
      </div>

      {selected.size > 0 && (
        <BulkBar
          count={selected.size}
          tagCatalog={tagCatalog}
          onClear={() => setSelected(new Set())}
          onStatus={(s) => void applyStatus([...selected], s)}
          onTag={(tag) => void applyTags([...selected], [tag])}
          onIncident={() => setIncidentTargets(selectedAlerts)}
        />
      )}

      <div className="mt-3 flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card">
        <TableHeader allChecked={allChecked} onTogglePage={togglePage} />
        <div className="min-h-0 flex-1 overflow-y-auto">
          {loading && alerts.length === 0 ? (
            <Center><Loader2 className="h-4 w-4 animate-spin" /> {t('alerts.list.loading')}</Center>
          ) : error ? (
            <Center>
              <AlertTriangle size={16} className="text-amber-500" /> {t('alerts.list.loadError')}
              <Button variant="outline" size="sm" className="ml-2" onClick={refresh}>{t('alerts.list.retry')}</Button>
            </Center>
          ) : alerts.length === 0 ? (
            <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('alerts.list.empty')}</div>
          ) : (
            alerts.map((a) => (
              <AlertRow
                key={a.id}
                alert={a}
                tagCatalog={tagCatalog}
                checked={selected.has(a.id)}
                onToggle={() => toggleSel(a.id)}
                onOpen={() => setOpenAlert(a)}
              />
            ))
          )}
        </div>
      </div>

      {!loading && !error && total > 0 && (
        <div className="shrink-0">
          <Pagination
            page={page}
            pageSize={pageSize}
            total={total}
            loading={loading}
            onPageChange={(p) => setPage(p)}
            onPageSizeChange={(s) => { setPageSize(s); setPage(0) }}
          />
        </div>
      )}
        </>
      )}

      {openAlert && (
        <AlertDrawer
          alert={openAlert}
          tagCatalog={tagCatalog}
          onClose={() => setOpenAlert(null)}
          onStatus={(s, obs, fp) => void applyStatus([openAlert.id], s, obs, fp)}
          onTags={(tags) => void applyTags([openAlert.id], tags)}
          onCreateTag={(name, color) => void createAndApplyTag([openAlert.id], name, color, openAlert.tags ?? [])}
          onUpdateTag={(id, name, color) => void updateCatalogTag(id, name, color)}
          onDeleteTag={(id, name) => void deleteCatalogTag(id, name)}
          onIncident={() => setIncidentTargets([openAlert])}
          onNotes={async (notes) => {
            try {
              await svc.updateNotes(openAlert.id, notes)
              toast.success(t('alerts.toast.notesSaved'))
              setOpenAlert({ ...openAlert, notes })
              refetchOpen(openAlert.id)
            } catch {
              toast.error(t('alerts.toast.notesFailed'))
            }
          }}
        />
      )}

      {incidentTargets && (
        <IncidentModal
          alerts={incidentTargets}
          onClose={() => setIncidentTargets(null)}
          onDone={() => {
            setIncidentTargets(null)
            setSelected(new Set())
            refresh()
          }}
        />
      )}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ total, openCount, view, onView }: { total: number; openCount: number | null; view: 'alerts' | 'overview'; onView: (v: 'alerts' | 'overview') => void }) {
  const { t } = useTranslation()
  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold tracking-tight">{t('alerts.title')}</h1>
          <span className="rounded-md bg-muted px-2 py-0.5 font-mono text-[11px] tabular-nums text-muted-foreground">
            {t('alerts.matching', { count: total.toLocaleString() })}
          </span>
        </div>
        <p className="mt-1 text-xs text-muted-foreground">
          {t('alerts.subtitle')}
          {openCount != null && (
            <> <span className="font-medium text-foreground">{openCount.toLocaleString()}</span> {t('alerts.openNow')}</>
          )}
        </p>
      </div>
      <ViewToggle view={view} onView={onView} />
    </header>
  )
}

function ViewToggle({ view, onView }: { view: 'alerts' | 'overview'; onView: (v: 'alerts' | 'overview') => void }) {
  const { t } = useTranslation()
  const opts = [
    { id: 'alerts' as const, icon: Rows3, label: t('alerts.view.alerts') },
    { id: 'overview' as const, icon: BarChart3, label: t('alerts.view.overview') },
  ]
  return (
    <div className="inline-flex shrink-0 items-center rounded-md border border-border bg-card p-0.5 text-xs">
      {opts.map(({ id, icon: Icon, label }) => (
        <button
          key={id}
          onClick={() => onView(id)}
          className={cn('flex items-center gap-1.5 rounded px-2.5 py-1 transition-colors', view === id ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground')}
        >
          <Icon size={13} /> {label}
        </button>
      ))}
    </div>
  )
}

/* ─── Volume card (timeline) ───────────────────────────────────────────── */

function VolumeCard({ values, total }: { values: number[]; total: number }) {
  const { t } = useTranslation()
  const max = Math.max(1, ...values)
  const w = 1000
  const h = 90
  const n = values.length || 1
  const slot = w / n
  const bw = Math.max(1, slot - 2)
  return (
    <div className="flex h-full flex-col rounded-xl border border-border bg-card p-6">
      <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{t('alerts.overview.overTime')}</div>
      <div className="mt-1 flex items-baseline gap-2">
        <span className="text-3xl font-semibold tabular-nums">{total.toLocaleString()}</span>
        <span className="text-sm text-muted-foreground">{t('alerts.overview.matchingAlerts')}</span>
      </div>
      {values.length === 0 ? (
        <div className="mt-4 flex min-h-[100px] flex-1 items-center justify-center text-xs text-muted-foreground">{t('alerts.overview.noData')}</div>
      ) : (
        <div className="mt-4 min-h-[100px] flex-1">
          <svg viewBox={`0 0 ${w} ${h}`} className="h-full w-full" preserveAspectRatio="none">
            {values.map((v, i) => {
              if (v <= 0) return null
              const bh = Math.max(2, (v / max) * h)
              return <rect key={i} x={i * slot + (slot - bw) / 2} y={h - bh} width={bw} height={bh} rx={1.5} className="fill-primary/55" />
            })}
          </svg>
        </div>
      )}
    </div>
  )
}

/* ─── Breakdown card (severity + status counts) ────────────────────────── */

function BreakdownCard({ sevCounts, statusCounts }: { sevCounts: Record<string, number>; statusCounts: Record<string, number> }) {
  const { t } = useTranslation()
  const sevRows = (['high', 'medium', 'low'] as SeverityKey[]).map((k) => ({ k, label: t(`alerts.severity.${k}`), bar: SEV_META[k].bar, count: sevCounts[String(SEVERITY_INT[k])] ?? 0 }))
  const stRows = (['open', 'in_review', 'completed', 'auto'] as StatusKey[]).map((k) => ({ k, label: t(`alerts.status.${k}`), count: statusCounts[String(STATUS_INT[k])] ?? 0 }))
  const sevMax = Math.max(1, ...sevRows.map((r) => r.count))
  return (
    <div className="flex h-full flex-col gap-4 rounded-xl border border-border bg-card p-5">
      <div>
        <div className="mb-2 flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
          <ShieldAlert size={13} /> {t('alerts.overview.bySeverity')}
        </div>
        <div className="space-y-2">
          {sevRows.map((r) => (
            <div key={r.k} className="flex items-center gap-2 text-xs">
              <span className="w-14 text-muted-foreground">{r.label}</span>
              <div className="h-2 flex-1 overflow-hidden rounded-full bg-muted">
                <div className={cn('h-full rounded-full', r.bar)} style={{ width: `${Math.max(2, (r.count / sevMax) * 100)}%` }} />
              </div>
              <span className="w-12 text-right font-mono tabular-nums">{r.count.toLocaleString()}</span>
            </div>
          ))}
        </div>
      </div>
      <div className="border-t border-border pt-3">
        <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">{t('alerts.overview.byStatus')}</div>
        <div className="grid grid-cols-2 gap-2">
          {stRows.map((r) => (
            <div key={r.k} className="rounded-md border border-border bg-background/40 px-2.5 py-1.5">
              <div className="truncate text-[10px] text-muted-foreground">{r.label}</div>
              <div className="font-semibold tabular-nums">{r.count.toLocaleString()}</div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

/* ─── Status tabs ──────────────────────────────────────────────────────── */

function StatusTabs({ current, counts, onChange }: { current: StatusTab; counts: Record<string, number>; onChange: (s: StatusTab) => void }) {
  const { t } = useTranslation()
  const label = (id: StatusTab) => (id === 'all' ? t('alerts.statusTabs.all') : t(`alerts.status.${id}`))
  const count = (id: StatusTab) => (id === 'all' ? undefined : counts[String(STATUS_INT[id as StatusKey])])
  return (
    <div className="flex flex-wrap items-center gap-1 border-b border-border">
      {STATUS_TABS.map((id) => {
        const active = current === id
        const c = count(id)
        return (
          <button
            key={id}
            onClick={() => onChange(id)}
            className={cn('relative flex items-center gap-2 px-3 py-2 text-xs transition-colors', active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground')}
          >
            {label(id)}
            {c != null && (
              <span className={cn('rounded-md px-1.5 py-0.5 font-mono text-[10px] tabular-nums', active ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground')}>{c}</span>
            )}
            {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
          </button>
        )
      })}
    </div>
  )
}

/* ─── Filter bar (field/operator/value builder) ────────────────────────── */

function FilterBar({
  filters,
  onAdd,
  onRemove,
  onClear,
}: {
  filters: CustomFilter[]
  onAdd: (f: CustomFilter) => void
  onRemove: (i: number) => void
  onClear: () => void
}) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-wrap items-center gap-1.5">
      {filters.map((f, i) => {
        const op = FILTER_OPS.find((o) => o.id === f.operator)
        return (
          <span key={i} className="inline-flex items-center gap-1.5 rounded-full border border-primary/25 bg-primary/5 py-1 pl-3 pr-1.5 text-xs">
            <span className="text-muted-foreground">{t(`alerts.fields.${fieldKey(f.field)}`)}</span>
            <span className="text-[11px] text-muted-foreground/70">{op ? t(`alerts.ops.${op.id}`) : f.operator}</span>
            {op?.needsValue && <span className="max-w-[200px] truncate font-mono font-medium">{f.value}</span>}
            <button onClick={() => onRemove(i)} className="flex h-5 w-5 items-center justify-center rounded-full text-muted-foreground hover:bg-foreground/10 hover:text-foreground">
              <X size={12} />
            </button>
          </span>
        )
      })}
      <AddFilterButton onAdd={onAdd} />
      {filters.length > 1 && (
        <button onClick={onClear} className="px-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground hover:underline">{t('alerts.filters.clearAll')}</button>
      )}
    </div>
  )
}

function AddFilterButton({ onAdd }: { onAdd: (f: CustomFilter) => void }) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [field, setField] = useState(FILTER_FIELDS[0].field)
  const [operator, setOperator] = useState('IS')
  const [values, setValues] = useState<{ value: string; count: number }[]>([])
  const [loadingValues, setLoadingValues] = useState(false)
  const [vq, setVq] = useState('')
  const ref = useRef<HTMLDivElement>(null)

  const op = FILTER_OPS.find((o) => o.id === operator)
  const needsValue = op?.needsValue ?? true

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  // Fetch the field's real existing values when the field changes.
  useEffect(() => {
    if (!open || !needsValue) return
    setLoadingValues(true)
    setValues([])
    svc
      .fieldValues(field)
      .then(setValues)
      .catch(() => setValues([]))
      .finally(() => setLoadingValues(false))
  }, [open, field, needsValue])

  const add = (value: string) => {
    const fdef = FILTER_FIELDS.find((f) => f.field === field)!
    onAdd({ field, label: fdef.label, operator, value })
    setOpen(false)
    setVq('')
  }

  const filtered = values.filter((v) => (vq ? String(v.value).toLowerCase().includes(vq.toLowerCase()) : true))

  return (
    <div className="relative" ref={ref}>
      <button onClick={() => setOpen((v) => !v)} className="inline-flex items-center gap-1.5 rounded-full border border-dashed border-border px-3 py-1 text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">
        <ListFilter size={12} /> {t('alerts.filters.add')}
      </button>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 w-80 rounded-md border border-border bg-popover p-3 shadow-lg">
          <div className="space-y-2">
            <select value={field} onChange={(e) => setField(e.target.value)} className={cn(SELECT_CLS, 'w-full')}>
              {FILTER_FIELDS.map((f) => (
                <option key={f.field} value={f.field}>{t(`alerts.fields.${fieldKey(f.field)}`)}</option>
              ))}
            </select>
            <select value={operator} onChange={(e) => setOperator(e.target.value)} className={cn(SELECT_CLS, 'w-full')}>
              {FILTER_OPS.map((o) => (
                <option key={o.id} value={o.id}>{t(`alerts.ops.${o.id}`)}</option>
              ))}
            </select>

            {needsValue ? (
              <>
                <div className="relative">
                  <Search size={13} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
                  <Input value={vq} onChange={(e) => setVq(e.target.value)} placeholder={t('alerts.filters.filterValues')} className="h-8 pl-8 text-xs" autoFocus />
                </div>
                <div className="max-h-48 overflow-y-auto rounded-md border border-border">
                  {loadingValues ? (
                    <div className="flex items-center gap-1.5 px-3 py-3 text-xs text-muted-foreground"><Loader2 className="h-3.5 w-3.5 animate-spin" /> {t('alerts.filters.loadingValues')}</div>
                  ) : filtered.length === 0 ? (
                    <div className="px-3 py-3 text-xs text-muted-foreground">{t('alerts.filters.noValues')}</div>
                  ) : (
                    filtered.map((v) => (
                      <button
                        key={v.value}
                        onClick={() => add(v.value)}
                        className="flex w-full items-center justify-between gap-2 px-3 py-1.5 text-left text-xs hover:bg-muted"
                      >
                        <span className="truncate font-mono">{v.value || t('alerts.filters.empty')}</span>
                        <span className="shrink-0 font-mono text-[10px] text-muted-foreground">{v.count.toLocaleString()}</span>
                      </button>
                    ))
                  )}
                </div>
                <p className="text-[10px] text-muted-foreground">{t('alerts.filters.pickValue')}</p>
              </>
            ) : (
              <div className="flex justify-end gap-2 pt-1">
                <Button variant="outline" size="sm" onClick={() => setOpen(false)}>{t('alerts.filters.cancel')}</Button>
                <Button size="sm" onClick={() => add('')}>{t('alerts.filters.addBtn')}</Button>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

/* ─── Toolbar ──────────────────────────────────────────────────────────── */

const SELECT_CLS = 'h-9 cursor-pointer rounded-md border border-input bg-background/40 px-2 text-sm transition-colors focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring'

function Toolbar({
  search,
  onSearch,
  severity,
  onSeverity,
  range,
  onRange,
  tagCatalog,
  tagFilter,
  onTagFilter,
  onCreateTag,
  onUpdateTag,
  onDeleteTag,
  onRefresh,
  onExport,
  loading,
  showSeverity = true,
}: {
  search: string
  onSearch: (s: string) => void
  severity: SeverityKey | 'all'
  onSeverity: (s: SeverityKey | 'all') => void
  range: TimeRange
  onRange: (r: TimeRange) => void
  tagCatalog: AlertTag[]
  tagFilter: string[]
  onTagFilter: (tags: string[]) => void
  onCreateTag: (tagName: string, tagColor: string) => void
  onUpdateTag: (id: number, tagName: string, tagColor: string) => void
  onDeleteTag: (id: number, tagName: string) => void
  onRefresh: () => void
  onExport: () => void
  loading: boolean
  showSeverity?: boolean
}) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-wrap items-center gap-2">
      <div className="relative min-w-[240px] flex-1">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input placeholder={t('alerts.toolbar.searchPlaceholder')} value={search} onChange={(e) => onSearch(e.target.value)} className="h-9 pl-9" />
      </div>
      {showSeverity && (
        <select value={severity} onChange={(e) => onSeverity(e.target.value as SeverityKey | 'all')} className={SELECT_CLS}>
          <option value="all">{t('alerts.toolbar.allSeverities')}</option>
          <option value="high">{t('alerts.severity.high')}</option>
          <option value="medium">{t('alerts.severity.medium')}</option>
          <option value="low">{t('alerts.severity.low')}</option>
        </select>
      )}
      <TagFilter catalog={tagCatalog} selected={tagFilter} onSelected={onTagFilter} onCreateTag={onCreateTag} onUpdateTag={onUpdateTag} onDeleteTag={onDeleteTag} />
      <TimeRangePicker value={range} onChange={onRange} allowAllTime align="right" />
      <Button variant="outline" size="sm" onClick={onExport} title={t('alerts.toolbar.export')}>
        <Download size={14} />
      </Button>
      <Button variant="outline" size="sm" onClick={onRefresh} disabled={loading} title={t('alerts.toolbar.refresh')}>
        <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
      </Button>
    </div>
  )
}

/* ─── Tag filter (toolbar multiselect) ─────────────────────────────────── */

function TagFilter({
  catalog,
  selected,
  onSelected,
  onCreateTag,
  onUpdateTag,
  onDeleteTag,
}: {
  catalog: AlertTag[]
  selected: string[]
  onSelected: (tags: string[]) => void
  onCreateTag: (tagName: string, tagColor: string) => void
  onUpdateTag: (id: number, tagName: string, tagColor: string) => void
  onDeleteTag: (id: number, tagName: string) => void
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editName, setEditName] = useState('')
  const [editColor, setEditColor] = useState(TAG_COLORS[5])
  const [name, setName] = useState('')
  const [color, setColor] = useState(TAG_COLORS[5])
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const trimmed = name.trim()
  const exists = catalog.some((tg) => tg.tagName.toLowerCase() === trimmed.toLowerCase())
  const canCreate = trimmed.length > 0 && !exists
  const create = () => {
    if (!canCreate) return
    onCreateTag(trimmed, color)
    setName('')
  }

  const toggle = (name: string) =>
    onSelected(selected.includes(name) ? selected.filter((t) => t !== name) : [...selected, name])
  const startEdit = (tg: AlertTag) => {
    setEditingId(tg.id)
    setEditName(tg.tagName)
    setEditColor(tg.tagColor || TAG_COLORS[5])
  }
  const saveEdit = () => {
    const n = editName.trim()
    if (!n || editingId == null) return
    onUpdateTag(editingId, n, editColor)
    setEditingId(null)
  }

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        className={cn(
          'inline-flex h-9 items-center gap-1.5 rounded-md border px-3 text-sm transition-colors',
          selected.length ? 'border-primary bg-primary/5 text-foreground' : 'border-input bg-background hover:bg-muted'
        )}
      >
        <TagIcon size={14} className="text-muted-foreground" />
        {selected.length ? t('alerts.tagFilter.tag', { count: selected.length }) : t('alerts.tagFilter.tags')}
        <ChevronDown size={12} className="opacity-60" />
      </button>
      {open && (
        <div className="absolute right-0 top-full z-30 mt-1 w-64 rounded-md border border-border bg-popover py-1 shadow-lg">
          {selected.length > 0 && (
            <button
              onClick={() => onSelected([])}
              className="mb-1 flex w-full items-center gap-2 border-b border-border px-3 py-1.5 text-left text-xs text-muted-foreground hover:bg-muted"
            >
              <X size={12} /> {t('alerts.tagFilter.clearSelection')}
            </button>
          )}
          <div className="max-h-56 overflow-y-auto">
            {catalog.length === 0 && <div className="px-3 py-1.5 text-xs text-muted-foreground">{t('alerts.tagFilter.noTags')}</div>}
            {catalog.map((tg) => {
              const on = selected.includes(tg.tagName)
              if (editingId === tg.id) {
                return (
                  <div key={tg.id} className="px-3 py-1.5">
                    <div className="flex items-center gap-1.5">
                      <input
                        value={editName}
                        onChange={(e) => setEditName(e.target.value)}
                        onKeyDown={(e) => { if (e.key === 'Enter') saveEdit(); if (e.key === 'Escape') setEditingId(null) }}
                        autoFocus
                        className="h-7 min-w-0 flex-1 rounded-md border border-input bg-background px-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                      />
                      <button onClick={saveEdit} disabled={!editName.trim()} title={t('alerts.tagEditor.save')} className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-primary text-primary-foreground disabled:opacity-40">
                        <Check size={13} />
                      </button>
                      <button onClick={() => setEditingId(null)} title={t('alerts.tagEditor.cancel')} className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md border border-input hover:bg-muted">
                        <X size={13} />
                      </button>
                    </div>
                    <div className="mt-1.5 flex flex-wrap gap-1">
                      {TAG_COLORS.map((c) => (
                        <button
                          key={c}
                          onClick={() => setEditColor(c)}
                          className={cn('h-4 w-4 rounded-full ring-offset-1 ring-offset-popover transition', editColor === c && 'ring-2 ring-ring')}
                          style={{ backgroundColor: c }}
                          aria-label={t('alerts.tagEditor.color', { color: c })}
                        />
                      ))}
                    </div>
                  </div>
                )
              }
              return (
                <div key={tg.id} className="group flex w-full items-center gap-2 px-3 py-1.5 text-sm hover:bg-muted">
                  <button onClick={() => toggle(tg.tagName)} className="flex min-w-0 flex-1 items-center gap-2 text-left">
                    <span className="h-2.5 w-2.5 shrink-0 rounded-full" style={{ backgroundColor: tg.tagColor || '#64748b' }} />
                    <span className="truncate">{tg.tagName}</span>
                    {on && <Check size={14} className="ml-auto shrink-0 text-primary" />}
                  </button>
                  {tg.systemOwner ? (
                    <Lock size={11} className="shrink-0 text-muted-foreground/50" aria-label={t('alerts.tagEditor.systemLocked')} />
                  ) : (
                    <span className="flex shrink-0 items-center gap-0.5 opacity-0 transition group-hover:opacity-100">
                      <button onClick={() => startEdit(tg)} title={t('alerts.tagEditor.edit')} className="rounded p-1 text-muted-foreground hover:bg-background hover:text-foreground">
                        <Pencil size={12} />
                      </button>
                      <button onClick={() => onDeleteTag(tg.id, tg.tagName)} title={t('alerts.tagEditor.delete')} className="rounded p-1 text-muted-foreground hover:bg-background hover:text-red-500">
                        <Trash2 size={12} />
                      </button>
                    </span>
                  )}
                </div>
              )
            })}
          </div>
          <div className="mt-1 border-t border-border px-3 pb-2 pt-2">
            <div className="mb-1.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">{t('alerts.tagEditor.createTag')}</div>
            <div className="flex items-center gap-1.5">
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && create()}
                placeholder={t('alerts.tagEditor.newTagName')}
                className="h-7 min-w-0 flex-1 rounded-md border border-input bg-background px-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
              <button
                onClick={create}
                disabled={!canCreate}
                className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-primary text-primary-foreground disabled:opacity-40"
                title={t('alerts.tagEditor.createTag')}
              >
                <Plus size={14} />
              </button>
            </div>
            <div className="mt-2 flex flex-wrap gap-1.5">
              {TAG_COLORS.map((c) => (
                <button
                  key={c}
                  onClick={() => setColor(c)}
                  className={cn('h-5 w-5 rounded-full ring-offset-1 ring-offset-popover transition', color === c && 'ring-2 ring-ring')}
                  style={{ backgroundColor: c }}
                  aria-label={t('alerts.tagEditor.color', { color: c })}
                />
              ))}
            </div>
            {exists && trimmed.length > 0 && <div className="mt-1.5 text-[10px] text-amber-500">{t('alerts.tagEditor.alreadyExists')}</div>}
          </div>
        </div>
      )}
    </div>
  )
}

/* ─── Bulk bar ─────────────────────────────────────────────────────────── */

function BulkBar({
  count,
  tagCatalog,
  onClear,
  onStatus,
  onTag,
  onIncident,
}: {
  count: number
  tagCatalog: AlertTag[]
  onClear: () => void
  onStatus: (status: number) => void
  onTag: (tag: string) => void
  onIncident: () => void
}) {
  const { t } = useTranslation()
  return (
    <div className="mt-3 flex flex-wrap items-center gap-2 rounded-lg border border-primary/30 bg-primary/5 px-3 py-2 text-xs">
      <span className="font-medium">{t('alerts.bulk.selected', { count })}</span>
      <span className="text-border">·</span>
      <Menu trigger={<>{t('alerts.bulk.setStatus')} <ChevronDown size={12} /></>}>
        {(['open', 'in_review', 'completed'] as StatusKey[]).map((k) => (
          <button key={k} onClick={() => onStatus(STATUS_INT[k])} className="block w-full px-3 py-1.5 text-left text-sm hover:bg-muted">
            {t(`alerts.status.${k}`)}
          </button>
        ))}
      </Menu>
      <Menu trigger={<><TagIcon size={12} /> {t('alerts.bulk.addTag')} <ChevronDown size={12} /></>}>
        {tagCatalog.length === 0 && <div className="px-3 py-1.5 text-xs text-muted-foreground">{t('alerts.bulk.noTags')}</div>}
        {tagCatalog.map((tg) => (
          <button key={tg.id} onClick={() => onTag(tg.tagName)} className="block w-full px-3 py-1.5 text-left text-sm hover:bg-muted">
            {tg.tagName}
          </button>
        ))}
      </Menu>
      <button onClick={onIncident} className="inline-flex items-center gap-1.5 rounded-md border border-border bg-card px-2 py-1 hover:bg-muted">
        <Flame size={12} className="text-red-500" /> {t('alerts.bulk.addToIncident')}
      </button>
      <button onClick={onClear} className="ml-auto text-muted-foreground hover:text-foreground">{t('alerts.bulk.clear')}</button>
    </div>
  )
}

function Menu({ trigger, children }: { trigger: React.ReactNode; children: React.ReactNode }) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])
  return (
    <div className="relative" ref={ref}>
      <button onClick={() => setOpen((v) => !v)} className="inline-flex items-center gap-1 rounded-md border border-border bg-card px-2 py-1 hover:bg-muted">
        {trigger}
      </button>
      {open && (
        <div onClick={() => setOpen(false)} className="absolute left-0 top-full z-30 mt-1 max-h-64 w-48 overflow-y-auto rounded-md border border-border bg-popover py-1 shadow-lg">
          {children}
        </div>
      )}
    </div>
  )
}

// Palette offered when creating a new tag.
const TAG_COLORS = ['#ef4444', '#f97316', '#eab308', '#22c55e', '#06b6d4', '#3b82f6', '#8b5cf6', '#ec4899', '#64748b']

/**
 * Tag picker for the drawer: toggle existing catalog tags, or create a brand-new
 * tag (name + color) that gets registered and applied. Unlike `Menu`, inner
 * clicks don't dismiss the popover — only an outside click does.
 */
function TagEditor({
  tags,
  catalog,
  onTags,
  onCreateTag,
  onUpdateTag,
  onDeleteTag,
}: {
  tags: string[]
  catalog: AlertTag[]
  onTags: (tags: string[]) => void
  onCreateTag: (tagName: string, tagColor: string) => void
  onUpdateTag: (id: number, tagName: string, tagColor: string) => void
  onDeleteTag: (id: number, tagName: string) => void
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [name, setName] = useState('')
  const [color, setColor] = useState(TAG_COLORS[5])
  // Inline edit of an existing catalog tag.
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editName, setEditName] = useState('')
  const [editColor, setEditColor] = useState(TAG_COLORS[5])
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const trimmed = name.trim()
  const exists = catalog.some((t) => t.tagName.toLowerCase() === trimmed.toLowerCase())
  const canCreate = trimmed.length > 0 && !exists
  const create = () => {
    if (!canCreate) return
    onCreateTag(trimmed, color)
    setName('')
  }
  const startEdit = (tg: AlertTag) => {
    setEditingId(tg.id)
    setEditName(tg.tagName)
    setEditColor(tg.tagColor || TAG_COLORS[5])
  }
  const saveEdit = () => {
    const n = editName.trim()
    if (!n || editingId == null) return
    onUpdateTag(editingId, n, editColor)
    setEditingId(null)
  }

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        className="inline-flex items-center gap-1 rounded-md border border-border bg-card px-2 py-1 text-sm hover:bg-muted"
      >
        <TagIcon size={12} /> {t('alerts.tagEditor.tag')} <ChevronDown size={12} />
      </button>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 w-64 rounded-md border border-border bg-popover py-1 shadow-lg">
          <div className="max-h-48 overflow-y-auto">
            {catalog.length === 0 && <div className="px-3 py-1.5 text-xs text-muted-foreground">{t('alerts.tagEditor.noTagsYet')}</div>}
            {catalog.map((tg) => {
              const has = tags.includes(tg.tagName)
              if (editingId === tg.id) {
                return (
                  <div key={tg.id} className="px-3 py-1.5">
                    <div className="flex items-center gap-1.5">
                      <input
                        value={editName}
                        onChange={(e) => setEditName(e.target.value)}
                        onKeyDown={(e) => { if (e.key === 'Enter') saveEdit(); if (e.key === 'Escape') setEditingId(null) }}
                        autoFocus
                        className="h-7 min-w-0 flex-1 rounded-md border border-input bg-background px-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                      />
                      <button onClick={saveEdit} disabled={!editName.trim()} title={t('alerts.tagEditor.save')} className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-primary text-primary-foreground disabled:opacity-40">
                        <Check size={13} />
                      </button>
                      <button onClick={() => setEditingId(null)} title={t('alerts.tagEditor.cancel')} className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md border border-input hover:bg-muted">
                        <X size={13} />
                      </button>
                    </div>
                    <div className="mt-1.5 flex flex-wrap gap-1">
                      {TAG_COLORS.map((c) => (
                        <button
                          key={c}
                          onClick={() => setEditColor(c)}
                          className={cn('h-4 w-4 rounded-full ring-offset-1 ring-offset-popover transition', editColor === c && 'ring-2 ring-ring')}
                          style={{ backgroundColor: c }}
                          aria-label={t('alerts.tagEditor.color', { color: c })}
                        />
                      ))}
                    </div>
                  </div>
                )
              }
              return (
                <div key={tg.id} className="group flex w-full items-center gap-2 px-3 py-1.5 text-sm hover:bg-muted">
                  <button
                    onClick={() => onTags(has ? tags.filter((x) => x !== tg.tagName) : [...tags, tg.tagName])}
                    className="flex min-w-0 flex-1 items-center gap-2 text-left"
                  >
                    <span className="h-2.5 w-2.5 shrink-0 rounded-full" style={{ backgroundColor: tg.tagColor || '#64748b' }} />
                    <span className="truncate">{tg.tagName}</span>
                    {has && <Check size={14} className="ml-auto shrink-0 text-primary" />}
                  </button>
                  {tg.systemOwner ? (
                    <Lock size={11} className="shrink-0 text-muted-foreground/50" aria-label={t('alerts.tagEditor.systemLocked')} />
                  ) : (
                    <span className="flex shrink-0 items-center gap-0.5 opacity-0 transition group-hover:opacity-100">
                      <button onClick={() => startEdit(tg)} title={t('alerts.tagEditor.edit')} className="rounded p-1 text-muted-foreground hover:bg-background hover:text-foreground">
                        <Pencil size={12} />
                      </button>
                      <button onClick={() => onDeleteTag(tg.id, tg.tagName)} title={t('alerts.tagEditor.delete')} className="rounded p-1 text-muted-foreground hover:bg-background hover:text-red-500">
                        <Trash2 size={12} />
                      </button>
                    </span>
                  )}
                </div>
              )
            })}
          </div>
          <div className="mt-1 border-t border-border px-3 pb-2 pt-2">
            <div className="mb-1.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">{t('alerts.tagEditor.createTag')}</div>
            <div className="flex items-center gap-1.5">
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && create()}
                placeholder={t('alerts.tagEditor.newTagName')}
                className="h-7 min-w-0 flex-1 rounded-md border border-input bg-background px-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
              <button
                onClick={create}
                disabled={!canCreate}
                className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-primary text-primary-foreground disabled:opacity-40"
                title={t('alerts.tagEditor.createApply')}
              >
                <Plus size={14} />
              </button>
            </div>
            <div className="mt-2 flex flex-wrap gap-1.5">
              {TAG_COLORS.map((c) => (
                <button
                  key={c}
                  onClick={() => setColor(c)}
                  className={cn('h-5 w-5 rounded-full ring-offset-1 ring-offset-popover transition', color === c && 'ring-2 ring-ring')}
                  style={{ backgroundColor: c }}
                  aria-label={t('alerts.tagEditor.color', { color: c })}
                />
              ))}
            </div>
            {exists && trimmed.length > 0 && <div className="mt-1.5 text-[10px] text-amber-500">{t('alerts.tagEditor.alreadyExists')}</div>}
          </div>
        </div>
      )}
    </div>
  )
}

/* ─── Table ────────────────────────────────────────────────────────────── */

const COLS = '36px 4px 1fr 130px 150px 200px 70px 90px'

function TableHeader({ allChecked, onTogglePage }: { allChecked: boolean; onTogglePage: () => void }) {
  const { t } = useTranslation()
  return (
    <div className="grid items-center gap-3 border-b border-border bg-muted/30 px-4 py-2.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground" style={{ gridTemplateColumns: COLS }}>
      <button onClick={onTogglePage} className="flex h-4 w-4 items-center justify-center rounded border border-input">
        {allChecked && <span className="h-2 w-2 rounded-sm bg-primary" />}
      </button>
      <div />
      <div>{t('alerts.table.alert')}</div>
      <div>{t('alerts.table.status')}</div>
      <div>{t('alerts.table.technique')}</div>
      <div>{t('alerts.table.sourceAdversary')}</div>
      <div className="text-center">{t('alerts.table.risk')}</div>
      <div className="text-center">{t('alerts.table.time')}</div>
    </div>
  )
}

function AlertRow({ alert: a, tagCatalog, checked, onToggle, onOpen }: { alert: Alert; tagCatalog: AlertTag[]; checked: boolean; onToggle: () => void; onOpen: () => void }) {
  const { t } = useTranslation()
  const sm = SEV_META[sevKey(a)]
  const stm = ST_META[statusKey(a)]
  const stmLabel = t(`alerts.status.${statusKey(a)}`)
  return (
    <div className="grid cursor-pointer items-center gap-3 border-b border-border/50 px-4 py-2.5 text-[13px] hover:bg-muted/20 last:border-b-0" style={{ gridTemplateColumns: COLS }} onClick={onOpen}>
      <button
        onClick={(e) => { e.stopPropagation(); onToggle() }}
        className={cn('flex h-4 w-4 items-center justify-center rounded border', checked ? 'border-primary bg-primary' : 'border-input')}
      >
        {checked && <span className="h-2 w-2 rounded-sm bg-primary-foreground" />}
      </button>
      <span className={cn('h-7 w-1 rounded-full', sm.bar)} />
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="truncate font-medium">{a.name || '—'}</span>
          {(isAiNote(a.notes) || isAiNote(a.statusObservation)) && <Sparkles size={11} className="shrink-0 text-fuchsia-500" aria-label={t('alerts.badge.aiAssessed')} />}
          {a.isIncident && <span className="shrink-0 rounded bg-red-500/15 px-1 py-0.5 text-[9px] font-semibold uppercase text-red-500">{t('alerts.badge.incident')}</span>}
        </div>
        <div className="mt-0.5 flex items-center gap-1.5 overflow-hidden text-[11px] text-muted-foreground">
          <span className="truncate">
            {a.category}
            {a.dataSource && ` · ${a.dataSource}`}
          </span>
          {(a.tags ?? []).slice(0, 2).map((t) => (
            <TagChip key={t} name={t} catalog={tagCatalog} size="xs" />
          ))}
          {(a.tags ?? []).length > 2 && (
            <span
              className="shrink-0 whitespace-nowrap rounded-md border border-border bg-muted px-1.5 py-0.5 text-[10px] font-medium leading-none text-muted-foreground"
              title={(a.tags ?? []).slice(2).join(', ')}
            >
              +{(a.tags ?? []).length - 2}
            </span>
          )}
        </div>
      </div>
      <div>
        <span className={cn('inline-flex items-center rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset', stm.pill)}>{stmLabel}</span>
      </div>
      <div className="truncate font-mono text-[11px] text-muted-foreground" title={a.technique}>{a.technique || '—'}</div>
      <FlowCell source={a.target} adversary={a.adversary} />
      <div className="text-center font-mono tabular-nums">{riskOf(a)}</div>
      <div className="text-center font-mono text-[11px] text-muted-foreground">{relativeTime(a[TS])}</div>
    </div>
  )
}

function FlowCell({ source, adversary }: { source?: Side; adversary?: Side }) {
  return (
    <div className="flex min-w-0 items-center gap-1.5 text-[11px]">
      <EndpointMini ep={source} />
      <ArrowRight size={11} className="shrink-0 text-muted-foreground/60" />
      <EndpointMini ep={adversary} accent />
    </div>
  )
}

function EndpointMini({ ep, accent }: { ep?: Side; accent?: boolean }) {
  if (!ep || (!ep.host && !ep.ip && !ep.user)) return <span className="text-muted-foreground/50">—</span>
  const cc = ep.geolocation?.countryCode
  const flag = flagEmoji(cc)
  return (
    <span className={cn('inline-flex min-w-0 items-center gap-1', accent ? 'text-foreground/90' : 'text-muted-foreground')}>
      {flag && <span title={ep.geolocation?.country || cc}>{flag}</span>}
      <span className="min-w-0 truncate font-mono">{ep.host || ep.user || ep.ip}</span>
    </span>
  )
}

function riskOf(a: Alert): string {
  const r = a.impactScore ?? a.impact?.score
  return r != null ? String(Math.round(r)) : '—'
}

/**
 * The events the engine attached inline to the alert (a sample, capped at ~11),
 * rendered with the exact same table + Fields/JSON detail as the Log Explorer.
 * The "View all related logs" button reproduces the engine's correlation search
 * (uncapped) in the Log Explorer.
 */
function RelatedEventsSection({ events, onViewAll, loadingAll }: { events: AlertEventItem[]; onViewAll: () => void; loadingAll: boolean }) {
  const { t } = useTranslation()
  // Each stored event IS a log document; normalise its time field to @timestamp
  // (it's stored as `timestamp`) so the shared renderer shows it.
  const docs = useMemo<LogDocument[]>(
    () =>
      events.map((e) => ({
        ...(e as Record<string, unknown>),
        '@timestamp': e['@timestamp'] || e.timestamp || e.deviceTime || '',
      })),
    [events]
  )

  return (
    <div className="space-y-3">
      <div className="flex items-start justify-between gap-3">
        <p className="text-[11px] leading-relaxed text-muted-foreground">
          {events.length > 0 ? t('alerts.related.sample') : t('alerts.related.empty')}
        </p>
        <Button size="sm" variant="outline" onClick={onViewAll} disabled={loadingAll} className="shrink-0">
          {loadingAll ? <Loader2 size={13} className="mr-1.5 animate-spin" /> : <FileSearch size={13} className="mr-1.5" />}
          {t('alerts.related.viewAll')}
        </Button>
      </div>
      {events.length > 0 && <LogResults docs={docs} />}
    </div>
  )
}

/* ─── Drawer ───────────────────────────────────────────────────────────── */

type Tab = 'summary' | 'parties' | 'events' | 'history'

function AlertDrawer({
  alert: a,
  tagCatalog,
  onClose,
  onStatus,
  onTags,
  onCreateTag,
  onUpdateTag,
  onDeleteTag,
  onIncident,
  onNotes,
}: {
  alert: Alert
  tagCatalog: AlertTag[]
  onClose: () => void
  onStatus: (status: number, observation: string, fp: boolean) => void
  onTags: (tags: string[]) => void
  onCreateTag: (tagName: string, tagColor: string) => void
  onUpdateTag: (id: number, tagName: string, tagColor: string) => void
  onDeleteTag: (id: number, tagName: string) => void
  onIncident: () => void
  onNotes: (notes: string) => void
}) {
  const { t } = useTranslation()
  const [tab, setTab] = useState<Tab>('summary')
  const navigate = useNavigate()
  const [loadingRelated, setLoadingRelated] = useState(false)
  // Reproduce the engine's correlation search server-side (no 10-hit cap) and
  // open every related log in the Log Explorer (filtered by _id).
  const viewRelatedLogs = async () => {
    setLoadingRelated(true)
    try {
      const r = await svc.relatedLogs(a.id)
      if (!r.ids?.length) {
        toast.info(t('alerts.toast.noRelatedLogs'))
        return
      }
      navigate('/log-explorer', {
        state: {
          relatedLogs: {
            ids: r.ids,
            indexPattern: r.indexPattern,
            timeFrom: r.timeFrom,
            timeTo: r.timeTo,
            alertName: a.name,
            truncated: r.truncated,
          },
        },
      })
    } catch (e) {
      toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.toast.relatedLogsFailed'))
    } finally {
      setLoadingRelated(false)
    }
  }
  // `notes` stores only the analyst's own free text; any AI assessment block is
  // kept separate so the SOC AI note is preserved when the analyst saves.
  const [notes, setNotes] = useState(userNotePart(a.notes))
  // The SOC AI agent may write its assessment into notes OR statusObservation.
  const aiNote = parseAiNote(a.notes) ?? parseAiNote(a.statusObservation)
  const sm = SEV_META[sevKey(a)]
  const stm = ST_META[statusKey(a)]
  const smLabel = t(`alerts.severity.${sevKey(a)}`)
  const stmLabel = t(`alerts.status.${statusKey(a)}`)
  const tags = a.tags ?? []

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div className="flex w-full max-w-[760px] flex-col overflow-hidden border-l border-border bg-card shadow-xl" onClick={(e) => e.stopPropagation()}>
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                <span className={cn('inline-flex items-center rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset', sm.pill)}>{smLabel}</span>
                <span className={cn('inline-flex items-center rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset', stm.pill)}>{stmLabel}</span>
                {a.category && <span>· {a.category}</span>}
                {a.isIncident && <span className="rounded bg-red-500/15 px-1 py-0.5 text-[9px] font-semibold uppercase text-red-500">{t('alerts.badge.incident')}</span>}
              </div>
              <h2 className="mt-1 text-xl font-semibold">{a.name || '—'}</h2>
              <div className="mt-1 flex items-center gap-2 font-mono text-[11px] text-muted-foreground">
                {a.technique && <span>{a.technique}</span>}
                {a.dataSource && <span>· {a.dataSource}</span>}
                <span>· {absTime(a[TS])}</span>
              </div>
            </div>
            <button onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
              <X size={16} />
            </button>
          </div>

          {/* Actions */}
          <div className="mt-4 flex flex-wrap items-center gap-2">
            <Menu trigger={<>{t('alerts.drawer.setStatus')} <ChevronDown size={12} /></>}>
              {(['open', 'in_review', 'completed'] as StatusKey[]).map((k) => (
                <button key={k} onClick={() => onStatus(STATUS_INT[k], '', false)} className="block w-full px-3 py-1.5 text-left text-sm hover:bg-muted">{t(`alerts.status.${k}`)}</button>
              ))}
              <button onClick={() => onStatus(STATUS_INT.completed, 'Marked as false positive', true)} className="block w-full border-t border-border px-3 py-1.5 text-left text-sm text-muted-foreground hover:bg-muted">
                {t('alerts.drawer.completeFalsePositive')}
              </button>
            </Menu>
            <TagEditor tags={tags} catalog={tagCatalog} onTags={onTags} onCreateTag={onCreateTag} onUpdateTag={onUpdateTag} onDeleteTag={onDeleteTag} />
            <Button size="sm" variant="outline" onClick={onIncident}>
              <Flame size={13} className="mr-1.5 text-red-500" /> {t('alerts.drawer.addToIncident')}
            </Button>
          </div>

          {tags.length > 0 && (
            <div className="mt-2 flex flex-wrap gap-1.5">
              {tags.map((tg) => (
                <TagChip key={tg} name={tg} catalog={tagCatalog} />
              ))}
            </div>
          )}
        </header>

        <nav className="flex items-center gap-1 border-b border-border px-6">
          {(['summary', 'parties', 'events', 'history'] as Tab[]).map((id) => (
            <button key={id} onClick={() => setTab(id)} className={cn('relative px-3 py-2.5 text-xs capitalize transition-colors', tab === id ? 'text-foreground' : 'text-muted-foreground hover:text-foreground')}>
              {id === 'events' ? `${t('alerts.drawer.tabs.events')}${a.events?.length ? ` (${a.events.length})` : ''}` : t(`alerts.drawer.tabs.${id}`)}
              {tab === id && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
            </button>
          ))}
        </nav>

        <div className="flex-1 overflow-y-auto bg-muted/10 p-6">
          {tab === 'summary' && (
            <div className="space-y-4">
              {a.isIncident && a.incidentDetail && (
                <div className="rounded-lg border border-red-500/30 bg-red-500/5 p-4">
                  <div className="flex items-center gap-2 text-xs font-medium text-red-600 dark:text-red-300">
                    <Flame size={13} /> {t('alerts.drawer.partOfIncident')}
                  </div>
                  <dl className="mt-2 grid grid-cols-[120px_1fr] gap-y-1.5 text-xs">
                    {a.incidentDetail.incidentName && <Row k={t('alerts.drawer.incidentDetail.incident')}>{a.incidentDetail.incidentName}</Row>}
                    {a.incidentDetail.incidentId != null && <Row k={t('alerts.drawer.incidentDetail.id')}><span className="font-mono">#{String(a.incidentDetail.incidentId)}</span></Row>}
                    {a.incidentDetail.createdBy && <Row k={t('alerts.drawer.incidentDetail.createdBy')}>{a.incidentDetail.createdBy}</Row>}
                    {a.incidentDetail.creationDate && <Row k={t('alerts.drawer.incidentDetail.created')}>{absTime(a.incidentDetail.creationDate)}</Row>}
                  </dl>
                </div>
              )}
              {a.description && <Section title={t('alerts.drawer.section.description')}><p className="text-xs leading-relaxed text-muted-foreground">{a.description}</p></Section>}
              {a.solution && <Section title={t('alerts.drawer.section.solution')}><p className="text-xs leading-relaxed text-muted-foreground">{a.solution}</p></Section>}
              {(a.reference ?? []).length > 0 && (
                <Section title={t('alerts.drawer.section.references')}>
                  <ul className="space-y-1 text-xs">
                    {(a.reference ?? []).map((r) => (
                      <li key={r}>
                        <a href={r} target="_blank" rel="noopener noreferrer" className="break-all text-primary hover:underline">{r}</a>
                      </li>
                    ))}
                  </ul>
                </Section>
              )}
              <Section title={t('alerts.drawer.section.details')}>
                <dl className="grid grid-cols-[140px_1fr] gap-y-2 text-xs">
                  <Row k={t('alerts.drawer.details.risk')}>{riskOf(a)}</Row>
                  <Row k={t('alerts.drawer.details.category')}>{a.category || '—'}</Row>
                  <Row k={t('alerts.drawer.details.technique')}><span className="font-mono">{a.technique || '—'}</span></Row>
                  <Row k={t('alerts.drawer.details.sensor')}>{a.dataSource || '—'}</Row>
                  <Row k={t('alerts.drawer.details.dataType')}><span className="font-mono">{a.dataType || '—'}</span></Row>
                  <Row k={t('alerts.drawer.details.alertId')}><span className="break-all font-mono">{a.id}</span></Row>
                  {a.statusObservation && !isAiNote(a.statusObservation) && <Row k={t('alerts.drawer.details.statusNote')}>{a.statusObservation}</Row>}
                </dl>
              </Section>
              {/* The AI SOC assessment is rendered read-only above; the analyst's
                  own notes are always editable below it and saved alongside the
                  AI block (which is preserved on save). */}
              {aiNote && <AiAssessment note={aiNote} />}
              <Section title={aiNote ? t('alerts.drawer.section.yourNotes') : t('alerts.drawer.section.notes')}>
                <textarea
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  rows={3}
                  placeholder={t('alerts.drawer.notesPlaceholder')}
                  className="w-full rounded-md border border-input bg-background/40 p-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                />
                <div className="mt-2 flex justify-end">
                  <Button
                    size="sm"
                    disabled={notes.trim() === userNotePart(a.notes)}
                    onClick={() => onNotes(combineUserNote(notes, a.notes))}
                  >
                    {t('alerts.drawer.saveNotes')}
                  </Button>
                </div>
              </Section>
            </div>
          )}
          {tab === 'parties' && (
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <PartyCard title={t('alerts.party.source')} ep={a.target} />
              <PartyCard title={t('alerts.party.adversary')} ep={a.adversary} accent />
            </div>
          )}
          {tab === 'events' && <RelatedEventsSection events={a.events ?? []} onViewAll={viewRelatedLogs} loadingAll={loadingRelated} />}
          {tab === 'history' && <HistoryTab alert={a} />}
        </div>
      </div>
    </div>
  )
}

function PartyCard({ title, ep, accent }: { title: string; ep?: Side; accent?: boolean }) {
  const { t } = useTranslation()
  return (
    <div className={cn('rounded-lg border bg-card p-4', accent ? 'border-red-500/30' : 'border-border')}>
      <h4 className="mb-3 text-sm font-semibold">{title}</h4>
      {!ep ? (
        <p className="text-xs text-muted-foreground">{t('alerts.party.noData')}</p>
      ) : (
        <dl className="grid grid-cols-[80px_1fr] gap-y-2 text-xs">
          {ep.ip && <Row k={t('alerts.party.ip')}><span className="font-mono">{ep.ip}</span></Row>}
          {ep.host && <Row k={t('alerts.party.host')}><span className="font-mono">{ep.host}</span></Row>}
          {ep.user && <Row k={t('alerts.party.user')}><span className="font-mono">{ep.user}</span></Row>}
          {ep.domain && <Row k={t('alerts.party.domain')}><span className="font-mono">{ep.domain}</span></Row>}
          {ep.geolocation?.country && (
            <Row k={t('alerts.party.country')}>
              {flagEmoji(ep.geolocation.countryCode)} {ep.geolocation.country}
              {ep.geolocation.city ? ` · ${ep.geolocation.city}` : ''}
            </Row>
          )}
        </dl>
      )}
    </div>
  )
}

function HistoryTab({ alert: a }: { alert: Alert }) {
  const { t } = useTranslation()
  const history = a.history ?? []
  if (history.length === 0) return <p className="text-xs text-muted-foreground">{t('alerts.history.empty')}</p>
  return (
    <div className="space-y-2">
      {history
        .slice()
        .reverse()
        .map((h, i) => {
          const detail = historyDetail(h, t)
          return (
            <div key={i} className="rounded-md border border-border bg-card px-3 py-2 text-xs">
              <div className="flex items-center justify-between gap-2">
                <span className="font-medium">{actionLabel(h.action, t)}</span>
                <span className="font-mono text-[10px] text-muted-foreground">{relativeTime(h.timestamp)}</span>
              </div>
              {(detail || h.user) && (
                <div className="mt-0.5 text-muted-foreground">
                  {detail}
                  {h.user && <span>{detail ? ' · ' : ''}{t('alerts.history.by', { user: h.user })}</span>}
                </div>
              )}
            </div>
          )
        })}
    </div>
  )
}

const ACTION_KEYS = ['UPDATE_STATUS', 'UPDATE_TAGS', 'UPDATE_NOTES', 'UPDATE_SOLUTION', 'MARK_AS_INCIDENT']
function actionLabel(a: string | undefined, t: TFunction) {
  if (a && ACTION_KEYS.includes(a)) return t(`alerts.history.actions.${a}`)
  return (a ?? 'change').replace(/_/g, ' ').toLowerCase()
}

// A clean one-line detail — never dumps the raw AI assessment or the newValue JSON.
function historyDetail(h: { message?: string; newValue?: string }, t: TFunction): string {
  const msg = (h.message ?? '').trim()
  if (msg && !isAiNote(msg) && !msg.startsWith('{')) return msg
  try {
    const v = JSON.parse(h.newValue || '{}') as Record<string, unknown>
    const parts: string[] = []
    if (typeof v.status === 'number') {
      const stKey = STATUS_BY_INT[v.status]
      parts.push(t('alerts.history.detail.statusTo', { status: stKey ? t(`alerts.status.${stKey}`) : String(v.status) }))
    }
    if (v.tags != null) parts.push(t('alerts.history.detail.tags', { tags: Array.isArray(v.tags) ? (v.tags as string[]).join(', ') : String(v.tags) }))
    if (typeof v.statusObservation === 'string' && v.statusObservation) parts.push(isAiNote(v.statusObservation) ? t('alerts.history.detail.aiAdded') : t('alerts.history.detail.observationAdded'))
    if (v.notes != null && !isAiNote(String(v.notes))) parts.push(t('alerts.history.detail.notesUpdated'))
    return parts.join(' · ')
  } catch {
    return ''
  }
}

/* ─── Incident modal ───────────────────────────────────────────────────── */

function IncidentModal({ alerts, onClose, onDone }: { alerts: Alert[]; onClose: () => void; onDone: () => void }) {
  const { t } = useTranslation()
  const [mode, setMode] = useState<'new' | 'existing'>('new')
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [busy, setBusy] = useState(false)
  const [incidents, setIncidents] = useState<Incident[]>([])
  const [incidentId, setIncidentId] = useState<number | ''>('')
  const [loadingInc, setLoadingInc] = useState(false)

  const alertList = useMemo<AlertLinkItem[]>(
    () => alerts.map((a) => ({ alertId: a.id, alertName: a.name || a.id, alertSeverity: a.severity ?? 1, alertStatus: a.status })),
    [alerts]
  )

  useEffect(() => {
    if (mode !== 'existing' || incidents.length) return
    setLoadingInc(true)
    incidentsHttpService
      .list()
      .then((r) => setIncidents(r.data ?? []))
      .catch(() => setIncidents([]))
      .finally(() => setLoadingInc(false))
  }, [mode, incidents.length])

  const canSubmit = mode === 'new' ? !!name.trim() : !!incidentId
  const submit = async () => {
    if (!canSubmit || busy) return
    setBusy(true)
    try {
      const inc =
        mode === 'new'
          ? await incidentsHttpService.create({ incidentName: name.trim(), incidentDescription: description.trim() || undefined, alertList })
          : await incidentsHttpService.addAlerts(Number(incidentId), alertList)
      await svc.convertToIncident(alerts.map((a) => a.id), inc.incidentName, inc.id)
      toast.success(mode === 'new' ? t('alerts.toast.incidentCreated') : t('alerts.toast.alertsAdded'))
      onDone()
    } catch (e) {
      toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.toast.incidentFailed'))
      setBusy(false)
    }
  }

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm" onClick={onClose}>
      <div className="flex w-full max-w-md flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl" onClick={(e) => e.stopPropagation()}>
        <header className="flex items-center justify-between gap-4 border-b border-border px-5 py-4">
          <h2 className="flex items-center gap-2 text-base font-semibold">
            <Flame size={17} className="text-red-500" />
            {t('alerts.incident.titleAdd', { count: alerts.length })}
          </h2>
          <button onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
            <X size={16} />
          </button>
        </header>

        <div className="flex items-center gap-1 border-b border-border px-5 pt-2">
          {(['new', 'existing'] as const).map((m) => (
            <button
              key={m}
              onClick={() => setMode(m)}
              className={cn('relative px-3 py-2 text-xs transition-colors', mode === m ? 'text-foreground' : 'text-muted-foreground hover:text-foreground')}
            >
              {m === 'new' ? t('alerts.incident.new') : t('alerts.incident.existing')}
              {mode === m && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
            </button>
          ))}
        </div>

        <div className="space-y-3 px-5 py-4">
          {mode === 'new' ? (
            <>
              <div>
                <label className="mb-1 block text-xs font-medium text-foreground/80">{t('alerts.incident.nameLabel')}</label>
                <Input value={name} onChange={(e) => setName(e.target.value)} placeholder={t('alerts.incident.namePlaceholder')} autoFocus />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-foreground/80">{t('alerts.incident.descriptionLabel')}</label>
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  rows={3}
                  placeholder={t('alerts.incident.descriptionPlaceholder')}
                  className="w-full rounded-md border border-input bg-background/40 p-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                />
              </div>
            </>
          ) : (
            <div>
              <label className="mb-1 block text-xs font-medium text-foreground/80">{t('alerts.incident.pickLabel')}</label>
              {loadingInc ? (
                <div className="flex items-center gap-2 py-2 text-xs text-muted-foreground"><Loader2 className="h-3.5 w-3.5 animate-spin" /> {t('alerts.incident.loading')}</div>
              ) : (
                <select value={incidentId} onChange={(e) => setIncidentId(e.target.value ? Number(e.target.value) : '')} className={cn(SELECT_CLS, 'w-full')}>
                  <option value="">{t('alerts.incident.selectPlaceholder')}</option>
                  {incidents.map((i) => (
                    <option key={i.id} value={i.id}>#{i.id} · {i.incidentName}</option>
                  ))}
                </select>
              )}
            </div>
          )}
          <p className="text-[11px] text-muted-foreground">{t('alerts.incident.willLink', { count: alerts.length })}</p>
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-border px-5 py-3">
          <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>{t('alerts.incident.cancel')}</Button>
          <Button size="sm" disabled={!canSubmit || busy} onClick={() => void submit()}>
            {busy ? '…' : mode === 'new' ? t('alerts.incident.create') : t('alerts.incident.add')}
          </Button>
        </footer>
      </div>
    </div>
  )
}

/* ─── AI SOC Agent note ────────────────────────────────────────────────── */

// Notes written by the AI SOC agent start with this marker and pack labelled
// sections separated by " | ". We parse them into a structured assessment.
const AI_NOTE_MARKER = '[AI SOC Agent]'

// A stored notes value sometimes arrives JSON-encoded (wrapped in double quotes)
// from the pipeline that persisted the AI assessment. Unwrap it so the marker
// sits at the start and the analyst's textarea doesn't show a stray leading `"`.
function unquoteNotes(s?: string): string {
  if (!s) return ''
  const v = s
  if (v.length >= 2 && v[0] === '"' && v[v.length - 1] === '"') {
    try {
      const parsed = JSON.parse(v)
      if (typeof parsed === 'string') return parsed
    } catch {
      return v.slice(1, -1)
    }
  }
  return v
}

// Robust detection: the marker may appear with leading whitespace/quotes or a
// slightly different case, so we look for it anywhere, case-insensitively.
function aiMarkerIndex(s?: string): number {
  const v = unquoteNotes(s)
  return v ? v.toUpperCase().indexOf(AI_NOTE_MARKER.toUpperCase()) : -1
}
function isAiNote(s?: string): boolean {
  return aiMarkerIndex(s) >= 0
}

// A stored `notes` value may contain the analyst's own text followed by the AI
// SOC block. We keep them separable so the analyst can keep adding notes while
// the AI assessment (rendered read-only) is preserved untouched.
function aiBlock(notes?: string): string {
  const v = unquoteNotes(notes)
  const idx = aiMarkerIndex(notes)
  return idx >= 0 ? v.slice(idx) : ''
}
function userNotePart(notes?: string): string {
  const v = unquoteNotes(notes)
  if (!v) return ''
  const idx = aiMarkerIndex(notes)
  return (idx >= 0 ? v.slice(0, idx) : v).trim()
}
function combineUserNote(userText: string, original?: string): string {
  const ai = aiBlock(original)
  const u = userText.trim()
  if (!ai) return u
  return u ? `${u}\n\n${ai}` : ai
}

interface AiNote {
  headline: string
  score?: number
  risk?: string
  sections: { label: string; value: string }[]
}

function parseAiNote(notes?: string): AiNote | null {
  const v = unquoteNotes(notes)
  const idx = aiMarkerIndex(notes)
  if (idx < 0 || !v) return null
  const body = v.slice(idx + AI_NOTE_MARKER.length)
  const parts = body
    .split('|')
    .map((s) => s.trim())
    .filter(Boolean)
  const headline = parts[0] ?? ''
  const sections = parts.slice(1).map((p) => {
    const i = p.indexOf(':')
    return i > 0 ? { label: p.slice(0, i).trim(), value: p.slice(i + 1).trim() } : { label: '', value: p }
  })
  const scoreM = headline.match(/Score:\s*([\d.]+)/i)
  const riskM = headline.match(/(low|medium|high)\s*risk/i)
  return {
    headline,
    score: scoreM ? Number(scoreM[1]) : undefined,
    risk: riskM ? riskM[1].toLowerCase() : undefined,
    sections,
  }
}

function AiAssessment({ note }: { note: AiNote }) {
  const { t } = useTranslation()
  const riskTone = note.risk === 'high' ? 'text-red-500' : note.risk === 'medium' ? 'text-amber-500' : 'text-emerald-500'
  return (
    <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-4">
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2 text-xs font-medium">
          <Sparkles size={14} className="text-fuchsia-500" /> {t('alerts.ai.title')}
        </div>
        <div className="flex items-center gap-2 text-sm">
          {note.score != null && <span className="font-semibold tabular-nums">{t('alerts.ai.score', { score: note.score })}</span>}
          {note.risk && <span className={cn('text-[10px] font-semibold uppercase tracking-wider', riskTone)}>{t('alerts.ai.risk', { risk: t(`alerts.severity.${note.risk}`) })}</span>}
        </div>
      </div>
      <dl className="mt-3 space-y-2.5 text-xs">
        {note.sections.map((s, i) => (
          <div key={i}>
            {s.label && <dt className="font-medium text-foreground/80">{s.label}</dt>}
            <dd className="leading-relaxed text-muted-foreground">{s.value}</dd>
          </div>
        ))}
      </dl>
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

function Row({ k, children }: { k: string; children: React.ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="break-words">{children}</dd>
    </>
  )
}

function Center({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">{children}</div>
}

function relativeTime(iso?: string) {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  if (Number.isNaN(diff)) return '—'
  const m = Math.round(diff / 60_000)
  if (m < 1) return 'now'
  if (m < 60) return `${m}m`
  const h = Math.round(m / 60)
  if (h < 24) return `${h}h`
  return `${Math.round(h / 24)}d`
}

function absTime(iso?: string) {
  if (!iso) return '—'
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? iso : d.toLocaleString(undefined, { month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

/** ISO 3166-1 alpha-2 → regional-indicator flag emoji. */
function flagEmoji(cc?: string): string {
  if (!cc || !/^[a-zA-Z]{2}$/.test(cc)) return ''
  const base = 0x1f1e6
  const up = cc.toUpperCase()
  return String.fromCodePoint(base + up.charCodeAt(0) - 65, base + up.charCodeAt(1) - 65)
}
