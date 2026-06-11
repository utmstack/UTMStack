import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  AlertTriangle,
  BarChart3,
  Braces,
  Calendar,
  Check,
  ChevronDown,
  ChevronRight,
  Columns3,
  Code2,
  Database,
  Download,
  Filter,
  Globe,
  Hash,
  Loader2,
  Minus,
  Play,
  Plus,
  RefreshCw,
  Search,
  Table as TableIcon,
  Tag,
  ToggleLeft,
  Type,
  X,
  type LucideIcon,
} from 'lucide-react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { useLocation, useNavigate } from 'react-router-dom'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { Pagination } from '@/shared/components/ui/pagination'
import { TimeRangePicker, presetRange, type TimeRange } from '@/shared/components/ui/time-range-picker'
import { ResultsHeader, ResultRow } from '../components/log-results'
import {
  logExplorerHttpService as svc,
  LogExplorerHttpError,
} from '../services/log-explorer-http.service'
import type {
  ChartView,
  FilterOperator,
  FilterType,
  IndexField,
  IndexPattern,
  LogDocument,
  TopValues,
} from '../types/log-explorer.types'

/* ─── Constants ────────────────────────────────────────────────────────── */

const TS = '@timestamp'


// Field-name candidates for the compact result columns (read from the flattened doc).
const MSG_FIELDS = ['log.message', 'logx.message', 'message', 'event.original', 'rule.name', 'logx.raw']

/* ─── Helpers ──────────────────────────────────────────────────────────── */

// Free-text "search all" maps to a wildcard field by data nature (logx.* / alert.*).
function textPrefix(pattern: string): string {
  return /alert/i.test(pattern) ? 'alert.*' : 'logx.*'
}

/* ─── Page ─────────────────────────────────────────────────────────────── */

const PAGE_SIZE_DEFAULT = 25

/** Router state passed by an alert's "view all related logs" action. */
interface RelatedLogsSeed {
  ids: string[]
  indexPattern: string
  timeFrom: string
  timeTo: string
  alertName?: string
  truncated?: boolean
}

export function LogExplorerPage() {
  const { t } = useTranslation()
  const location = useLocation()
  const navigate = useNavigate()
  const seededRef = useRef(false)
  const [patterns, setPatterns] = useState<IndexPattern[]>([])
  const [pattern, setPattern] = useState<IndexPattern | null>(null)

  const [range, setRange] = useState<TimeRange>(() => presetRange('24h'))

  const [searchInput, setSearchInput] = useState('')
  const [appliedQuery, setAppliedQuery] = useState('')
  const [sqlMode, setSqlMode] = useState(false)
  const [sqlInput, setSqlInput] = useState('')
  const [appliedSql, setAppliedSql] = useState('')

  const [filters, setFilters] = useState<FilterType[]>([])

  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(PAGE_SIZE_DEFAULT)
  const [total, setTotal] = useState(0)

  const [rows, setRows] = useState<LogDocument[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [fields, setFields] = useState<IndexField[]>([])
  const [expanded, setExpanded] = useState<number | null>(null)
  const [nonce, setNonce] = useState(0)

  const [viewMode, setViewMode] = useState<'table' | 'chart'>('table')

  // Table columns (Discover-style). Empty → the compact "Document" summary view.
  const [columns, setColumns] = useState<string[]>([])
  const toggleColumn = useCallback(
    (name: string) =>
      setColumns((cur) => (cur.includes(name) ? cur.filter((c) => c !== name) : [...cur, name])),
    []
  )

  /* Load active index patterns once; default to v11-log-* (or first).
     When arriving from an alert's "view all related logs" deep-link, seed the
     pattern, the _id filter and the time window from router state instead. */
  useEffect(() => {
    let cancelled = false
    svc
      .patterns()
      .then((ps) => {
        if (cancelled) return
        setPatterns(ps)
        const state = location.state as {
          relatedLogs?: RelatedLogsSeed
          socaiFilters?: FilterType[]
          socaiTime?: string
        } | null
        const seed = state?.relatedLogs
        if (seed?.ids?.length && !seededRef.current) {
          seededRef.current = true
          setPattern(ps.find((p) => p.pattern === seed.indexPattern) ?? ps.find((p) => p.pattern === 'v11-log-*') ?? ps[0] ?? null)
          setFilters([{ field: '_id', operator: 'IS_ONE_OF_TERMS', value: seed.ids }])
          setRange({ from: seed.timeFrom, to: seed.timeTo, interval: 'hour' })
          if (seed.truncated) {
            toast.info(t('logExplorer.related.truncated', { count: seed.ids.length }))
          }
          // Drop the router state so a refresh doesn't re-seed over the user's edits.
          navigate(location.pathname, { replace: true })
          return
        }
        // SOC-AI chat navigation: apply the agent's filters + time window.
        if (state?.socaiFilters?.length && !seededRef.current) {
          seededRef.current = true
          setPattern(ps.find((p) => p.pattern === 'v11-log-*') ?? ps[0] ?? null)
          setFilters(state.socaiFilters)
          if (state.socaiTime) setRange(presetRange(state.socaiTime))
          navigate(location.pathname, { replace: true })
          return
        }
        // Integration "View events" deep-link: ?dataType=<value> → filter + 30d + column.
        const dataType = new URLSearchParams(location.search).get('dataType')
        if (dataType && !seededRef.current) {
          seededRef.current = true
          setPattern(ps.find((p) => p.pattern === 'v11-log-*') ?? ps[0] ?? null)
          setFilters([{ field: 'dataType', operator: 'IS', value: dataType }])
          setRange(presetRange('30d'))
          setColumns(['dataType'])
          navigate(location.pathname, { replace: true })
          return
        }
        setPattern(ps.find((p) => p.pattern === 'v11-log-*') ?? ps[0] ?? null)
      })
      .catch(() => {
        if (!cancelled) setError('explorer:patterns')
      })
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  /* Field list for the sidebar follows the selected pattern. */
  useEffect(() => {
    if (!pattern) return
    let cancelled = false
    svc
      .fields(pattern.pattern)
      .then((f) => !cancelled && setFields(f ?? []))
      .catch(() => !cancelled && setFields([]))
    return () => {
      cancelled = true
    }
  }, [pattern])

  /* The filter array sent to the backend: time + free-text + chips. */
  const buildFilters = useCallback((): FilterType[] => {
    const out: FilterType[] = []
    if (range.from) out.push({ field: TS, operator: 'IS_BETWEEN', value: [range.from, range.to] })
    if (appliedQuery.trim() && pattern) {
      out.push({ field: textPrefix(pattern.pattern), operator: 'IS_IN_FIELDS', value: appliedQuery.trim() })
    }
    out.push(...filters)
    return out
  }, [range.from, range.to, appliedQuery, pattern, filters])

  const activeFilterList = useMemo(() => buildFilters(), [buildFilters])

  /* Main fetch — results (+ histogram, except in SQL mode). */
  const run = useCallback(async () => {
    if (!pattern) return
    setLoading(true)
    setError(null)
    try {
      if (sqlMode) {
        if (!appliedSql.trim()) {
          setRows([])
          setTotal(0)
          return
        }
        const { data, total } = await svc.searchSql(appliedSql.trim(), page + 1, pageSize)
        setRows(data ?? [])
        setTotal(total)
      } else {
        const { data, total } = await svc.search({
          indexPattern: pattern.pattern,
          filters: buildFilters(),
          page: page + 1,
          size: pageSize,
        })
        setRows(data ?? [])
        setTotal(total)
      }
    } catch (e) {
      setError(e instanceof LogExplorerHttpError ? e.message : 'explorer:failed')
      setRows([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [pattern, sqlMode, appliedSql, page, pageSize, buildFilters])

  useEffect(() => {
    void run()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pattern, range.from, range.to, appliedQuery, appliedSql, sqlMode, filters, page, pageSize, nonce])

  /* Commit the input box and run (Enter / Run button). */
  const submit = () => {
    setExpanded(null)
    setPage(0)
    if (sqlMode) {
      if (appliedSql === sqlInput) setNonce((n) => n + 1)
      else setAppliedSql(sqlInput)
    } else {
      if (appliedQuery === searchInput) setNonce((n) => n + 1)
      else setAppliedQuery(searchInput)
    }
  }

  const addFilter = (f: FilterType) => {
    setPage(0)
    setFilters((cur) =>
      cur.some((c) => c.field === f.field && c.operator === f.operator && c.value === f.value) ? cur : [...cur, f]
    )
  }
  const removeFilter = (i: number) => {
    setPage(0)
    setFilters((cur) => cur.filter((_, idx) => idx !== i))
  }

  return (
    <div className="flex h-full min-h-0 flex-col px-6 py-6">
      <Header pattern={pattern} total={total} loading={loading} />

      <div className="mt-4">
        <QueryBar
          patterns={patterns}
          pattern={pattern}
          onPattern={(p) => {
            setPattern(p)
            setPage(0)
            setFilters([])
            setColumns([])
          }}
          searchInput={searchInput}
          onSearchInput={setSearchInput}
          sqlMode={sqlMode}
          onSqlMode={(v) => {
            setSqlMode(v)
            setPage(0)
          }}
          sqlInput={sqlInput}
          onSqlInput={setSqlInput}
          range={range}
          onRange={(r) => {
            setRange(r)
            setPage(0)
          }}
          onRun={submit}
          loading={loading}
          onRefresh={() => setNonce((n) => n + 1)}
          onExport={() =>
            pattern &&
            svc
              .exportCsv({
                indexPattern: pattern.pattern,
                filters: buildFilters(),
                columns: [
                  { label: 'Timestamp', field: TS, type: 'date', visible: true },
                  { label: 'Message', field: MSG_FIELDS[0], type: 'text', visible: true },
                ],
              })
              .catch(() => toast.error(t('logExplorer.toast.exportFailed')))
          }
        />
      </div>

      <div className="mt-4 flex items-center justify-between gap-2">
        <div className="flex flex-wrap items-center gap-2">
          {filters.length > 0 && <FilterChips filters={filters} onRemove={removeFilter} onClear={() => setFilters([])} />}
          {!sqlMode && pattern && (
            <AddFilterButton pattern={pattern} fields={fields} filters={activeFilterList} onAdd={addFilter} />
          )}
        </div>
        {!sqlMode && <ViewToggle mode={viewMode} onChange={setViewMode} />}
      </div>

      <div className="mt-2 flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card">
        {viewMode === 'chart' && !sqlMode ? (
          <ChartPanel pattern={pattern} fields={fields} filters={activeFilterList} />
        ) : (
          <div className="flex min-h-0 flex-1">
            {!sqlMode && (
              <FieldSidebar
                fields={fields}
                pattern={pattern}
                filters={activeFilterList}
                columns={columns}
                onAdd={addFilter}
                onToggleColumn={toggleColumn}
              />
            )}
            <div className="flex min-w-0 flex-1 flex-col border-l border-border">
              <ResultsHeader columns={columns} onRemoveColumn={toggleColumn} />
              <div className="min-h-0 flex-1 overflow-y-auto">
                {loading && rows.length === 0 ? (
                  <RowMessage>
                    <Loader2 className="h-4 w-4 animate-spin" /> {t('logExplorer.results.searching')}
                  </RowMessage>
                ) : error ? (
                  <RowMessage>
                    <AlertTriangle size={16} className="text-amber-500" />
                    {error === 'explorer:patterns'
                      ? t('logExplorer.results.patternsFailed')
                      : error === 'explorer:failed'
                        ? t('logExplorer.results.searchFailed')
                        : error}
                    <Button variant="outline" size="sm" className="ml-2" onClick={() => setNonce((n) => n + 1)}>
                      {t('logExplorer.results.retry')}
                    </Button>
                  </RowMessage>
                ) : rows.length === 0 ? (
                  <div className="px-6 py-16 text-center text-sm text-muted-foreground">
                    {t('logExplorer.results.none')}
                  </div>
                ) : (
                  rows.map((doc, i) => (
                    <ResultRow
                      key={i}
                      doc={doc}
                      columns={columns}
                      expanded={expanded === i}
                      onToggle={() => setExpanded(expanded === i ? null : i)}
                      onAdd={addFilter}
                    />
                  ))
                )}
              </div>
            </div>
          </div>
        )}
      </div>

      {viewMode === 'table' && !loading && !error && total > 0 && (
        <div className="shrink-0">
          <Pagination
            page={page}
            pageSize={pageSize}
            total={total}
            loading={loading}
            align="right"
            onPageChange={(p) => {
              setExpanded(null)
              setPage(p)
            }}
            onPageSizeChange={(s) => {
              setPageSize(s)
              setPage(0)
            }}
          />
        </div>
      )}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ pattern, total, loading }: { pattern: IndexPattern | null; total: number; loading: boolean }) {
  const { t } = useTranslation()
  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">{t('logExplorer.title')}</h1>
        <p className="mt-1 text-xs text-muted-foreground">
          {loading ? (
            t('logExplorer.searching')
          ) : (
            <>
              <span className="font-medium text-foreground">{total.toLocaleString()}</span> {t('logExplorer.eventsIn')}{' '}
              <span className="font-mono">{pattern?.pattern ?? '—'}</span>
            </>
          )}
        </p>
      </div>
    </header>
  )
}

/* ─── Query bar ────────────────────────────────────────────────────────── */

function QueryBar({
  patterns,
  pattern,
  onPattern,
  searchInput,
  onSearchInput,
  sqlMode,
  onSqlMode,
  sqlInput,
  onSqlInput,
  range,
  onRange,
  onRun,
  onRefresh,
  loading,
  onExport,
}: {
  patterns: IndexPattern[]
  pattern: IndexPattern | null
  onPattern: (p: IndexPattern) => void
  searchInput: string
  onSearchInput: (q: string) => void
  sqlMode: boolean
  onSqlMode: (b: boolean) => void
  sqlInput: string
  onSqlInput: (q: string) => void
  range: TimeRange
  onRange: (r: TimeRange) => void
  onRun: () => void
  onRefresh: () => void
  loading: boolean
  onExport: () => void
}) {
  const { t } = useTranslation()
  const [patternOpen, setPatternOpen] = useState(false)

  return (
    <div className="flex flex-wrap items-center gap-2 rounded-xl border border-border bg-card p-2">
      {/* Index pattern selector */}
      <Dropdown
        open={patternOpen}
        onOpenChange={setPatternOpen}
        trigger={
          <>
            <Database size={13} className="text-muted-foreground" />
            <span className="font-mono">{pattern?.pattern ?? '—'}</span>
            <ChevronDown size={12} className="text-muted-foreground" />
          </>
        }
      >
        <div className="border-b border-border px-3 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
          {t('logExplorer.query.indexPatterns')}
        </div>
        {patterns.length === 0 && (
          <div className="px-3 py-2 text-xs text-muted-foreground">{t('logExplorer.query.noPatterns')}</div>
        )}
        {patterns.map((p) => (
          <button
            key={p.id}
            onClick={() => {
              onPattern(p)
              setPatternOpen(false)
            }}
            className={cn(
              'flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors',
              p.id === pattern?.id ? 'bg-muted/60' : 'hover:bg-muted/60'
            )}
          >
            <span className="font-mono">{p.pattern}</span>
          </button>
        ))}
      </Dropdown>

      <div className="h-5 w-px bg-border" />

      {/* Search input — free text or SQL */}
      <div className="relative min-w-[300px] flex-1">
        {sqlMode ? (
          <Input
            value={sqlInput}
            onChange={(e) => onSqlInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && onRun()}
            placeholder="SELECT * FROM &quot;v11-log-*&quot; ORDER BY @timestamp DESC"
            className="h-9 border-0 bg-transparent font-mono text-xs shadow-none focus-visible:ring-0"
          />
        ) : (
          <>
            <Search size={13} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={searchInput}
              onChange={(e) => onSearchInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && onRun()}
              placeholder={t('logExplorer.query.searchPlaceholder')}
              className="h-9 border-0 bg-transparent pl-8 font-mono text-xs shadow-none focus-visible:ring-0"
            />
          </>
        )}
      </div>

      <div className="h-5 w-px bg-border" />

      {/* Time range */}
      <TimeRangePicker value={range} onChange={onRange} align="right" />

      <button
        onClick={() => onSqlMode(!sqlMode)}
        className={cn(
          'flex h-9 items-center gap-1.5 rounded-md px-2.5 text-xs transition-colors',
          sqlMode
            ? 'bg-violet-500/15 text-violet-600 dark:text-violet-300'
            : 'text-muted-foreground hover:bg-muted hover:text-foreground'
        )}
        title="Toggle SQL mode"
      >
        <Code2 size={13} />
        SQL
      </button>

      <Button variant="outline" size="sm" onClick={onRefresh} disabled={loading} title={t('logExplorer.query.refresh')}>
        <RefreshCw size={13} className={cn(loading && 'animate-spin')} />
      </Button>

      <Button variant="outline" size="sm" onClick={onExport} title={t('logExplorer.query.exportCsv')}>
        <Download size={13} />
      </Button>

      <Button size="sm" onClick={onRun}>
        <Play size={12} className="mr-1.5" />
        {t('logExplorer.query.run')}
      </Button>
    </div>
  )
}

function Dropdown({
  open,
  onOpenChange,
  trigger,
  align = 'left',
  children,
}: {
  open: boolean
  onOpenChange: (v: boolean) => void
  trigger: React.ReactNode
  align?: 'left' | 'right'
  children: React.ReactNode
}) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onOpenChange(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open, onOpenChange])

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => onOpenChange(!open)}
        className={cn(
          'flex h-9 items-center gap-2 rounded-md px-3 text-sm transition-colors',
          open ? 'bg-muted' : 'hover:bg-muted'
        )}
      >
        {trigger}
      </button>
      {open && (
        <div
          className={cn(
            'absolute top-full z-30 mt-1 max-h-80 w-72 overflow-y-auto rounded-md border border-border bg-popover shadow-lg',
            align === 'right' ? 'right-0' : 'left-0'
          )}
        >
          {children}
        </div>
      )}
    </div>
  )
}

/* ─── Active filter chips ──────────────────────────────────────────────── */

const OP_KEY: Record<string, string> = {
  IS: 'is',
  IS_NOT: 'isNot',
  CONTAIN: 'contains',
  EXIST: 'exists',
  IS_BETWEEN: 'between',
  IS_IN_FIELDS: 'search',
  IS_ONE_OF_TERMS: 'isOneOf',
}

/* Explicit field/operator/value filter builder — pick a field, an operator, and
 * a value from the field's real existing values (fetched via top-x-values). */
const BUILDER_OPS: { id: FilterOperator; label: string; needsValue: boolean }[] = [
  { id: 'IS', label: 'is', needsValue: true },
  { id: 'IS_NOT', label: 'is not', needsValue: true },
  { id: 'CONTAIN', label: 'contains', needsValue: true },
  { id: 'EXIST', label: 'exists', needsValue: false },
]

function AddFilterButton({
  pattern,
  fields,
  filters,
  onAdd,
}: {
  pattern: IndexPattern
  fields: IndexField[]
  filters: FilterType[]
  onAdd: (f: FilterType) => void
}) {
  const { t } = useTranslation()
  const selectable = useMemo(
    () => fields.filter((f) => !f.name.endsWith('.keyword')).sort((a, b) => a.name.localeCompare(b.name)),
    [fields]
  )
  const [open, setOpen] = useState(false)
  const [field, setField] = useState('')
  const [operator, setOperator] = useState<FilterOperator>('IS')
  const [values, setValues] = useState<TopValues['top']>([])
  const [loadingValues, setLoadingValues] = useState(false)
  const [vq, setVq] = useState('')
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (open && !field && selectable.length) setField(selectable[0].name)
  }, [open, field, selectable])

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const op = BUILDER_OPS.find((o) => o.id === operator)
  const needsValue = op?.needsValue ?? true
  const fieldDef = selectable.find((f) => f.name === field)
  const aggField = fieldDef?.type === 'text' && !field.endsWith('.keyword') ? `${field}.keyword` : field

  useEffect(() => {
    if (!open || !needsValue || !field) return
    setLoadingValues(true)
    setValues([])
    svc
      .topValues(pattern.pattern, aggField, filters, 100)
      .then((r) => setValues(r.top ?? []))
      .catch(() => setValues([]))
      .finally(() => setLoadingValues(false))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, field, needsValue])

  const add = (value: string) => {
    onAdd({ field, operator, value: needsValue ? value : undefined })
    setOpen(false)
    setVq('')
  }
  const filtered = values.filter((v) => (vq ? String(v.value).toLowerCase().includes(vq.toLowerCase()) : true))

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        className="inline-flex items-center gap-1.5 rounded-full border border-dashed border-border px-3 py-1 text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
      >
        <Filter size={12} /> {t('logExplorer.builder.add')}
      </button>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 w-80 rounded-md border border-border bg-popover p-3 shadow-lg">
          <div className="space-y-2">
            <select value={field} onChange={(e) => setField(e.target.value)} className={cn(SELECT_CLS, 'w-full font-mono')}>
              {selectable.map((f) => (
                <option key={f.name} value={f.name}>{f.name}</option>
              ))}
            </select>
            <select value={operator} onChange={(e) => setOperator(e.target.value as FilterOperator)} className={cn(SELECT_CLS, 'w-full')}>
              {BUILDER_OPS.map((o) => (
                <option key={o.id} value={o.id}>{t(`logExplorer.ops.${OP_KEY[o.id] ?? o.id}`)}</option>
              ))}
            </select>
            {needsValue ? (
              <>
                <div className="relative">
                  <Search size={13} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
                  <Input value={vq} onChange={(e) => setVq(e.target.value)} placeholder={t('logExplorer.builder.filterValues')} className="h-8 pl-8 text-xs" autoFocus />
                </div>
                <div className="max-h-48 overflow-y-auto rounded-md border border-border">
                  {loadingValues ? (
                    <div className="flex items-center gap-1.5 px-3 py-3 text-xs text-muted-foreground"><Loader2 className="h-3.5 w-3.5 animate-spin" /> {t('logExplorer.builder.loadingValues')}</div>
                  ) : filtered.length === 0 ? (
                    <div className="px-3 py-3 text-xs text-muted-foreground">{t('logExplorer.builder.noValues')}</div>
                  ) : (
                    filtered.map((v) => (
                      <button key={v.value} onClick={() => add(v.value)} className="flex w-full items-center justify-between gap-2 px-3 py-1.5 text-left text-xs hover:bg-muted">
                        <span className="truncate font-mono">{v.value || t('logExplorer.fields.empty')}</span>
                        <span className="shrink-0 font-mono text-[10px] text-muted-foreground">{v.count.toLocaleString()}</span>
                      </button>
                    ))
                  )}
                </div>
              </>
            ) : (
              <div className="flex justify-end gap-2 pt-1">
                <Button variant="outline" size="sm" onClick={() => setOpen(false)}>{t('logExplorer.builder.cancel')}</Button>
                <Button size="sm" onClick={() => add('')}>{t('logExplorer.builder.confirm')}</Button>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

function FilterChips({
  filters,
  onRemove,
  onClear,
}: {
  filters: FilterType[]
  onRemove: (i: number) => void
  onClear: () => void
}) {
  const { t } = useTranslation()
  return (
    <div className="mt-3 flex flex-wrap items-center gap-2">
      <span className="inline-flex items-center gap-1.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/70">
        <Filter size={12} /> {t('logExplorer.filters.label')}
      </span>
      {filters.map((f, i) => {
        const neg = f.operator === 'IS_NOT'
        return (
          <span
            key={i}
            className={cn(
              'inline-flex items-center gap-1.5 rounded-full border py-1 pl-3 pr-1.5 text-xs',
              neg ? 'border-red-500/30 bg-red-500/10' : 'border-primary/25 bg-primary/5'
            )}
          >
            <span className="font-mono text-muted-foreground">{f.field}</span>
            <span className={cn('text-[11px]', neg ? 'text-red-500' : 'text-muted-foreground/70')}>
              {OP_KEY[f.operator] ? t(`logExplorer.ops.${OP_KEY[f.operator]}`) : f.operator}
            </span>
            {f.value != null && f.operator !== 'EXIST' && (
              <span className="max-w-[220px] truncate font-mono font-medium">
                {Array.isArray(f.value) ? t('logExplorer.related.nLogs', { count: f.value.length }) : String(f.value)}
              </span>
            )}
            <button
              onClick={() => onRemove(i)}
              className="flex h-5 w-5 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-foreground/10 hover:text-foreground"
            >
              <X size={12} />
            </button>
          </span>
        )
      })}
      {filters.length > 1 && (
        <button
          onClick={onClear}
          className="px-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground hover:underline"
        >
          {t('logExplorer.filters.clearAll')}
        </button>
      )}
    </div>
  )
}

/* ─── Field sidebar ────────────────────────────────────────────────────── */

function FieldSidebar({
  fields,
  pattern,
  filters,
  columns,
  onAdd,
  onToggleColumn,
}: {
  fields: IndexField[]
  pattern: IndexPattern | null
  filters: FilterType[]
  columns: string[]
  onAdd: (f: FilterType) => void
  onToggleColumn: (name: string) => void
}) {
  const { t } = useTranslation()
  const [q, setQ] = useState('')
  const [openField, setOpenField] = useState<string | null>(null)

  // Hide raw .keyword variants — the base field covers them.
  const visible = useMemo(
    () =>
      fields
        .filter((f) => !f.name.endsWith('.keyword'))
        .filter((f) => (q ? f.name.toLowerCase().includes(q.toLowerCase()) : true))
        .sort((a, b) => a.name.localeCompare(b.name)),
    [fields, q]
  )

  // Selected fields (in column order) on top, the rest below — like the legacy.
  const selected = columns
    .map((name) => visible.find((f) => f.name === name))
    .filter((f): f is IndexField => !!f)
  const available = visible.filter((f) => !columns.includes(f.name))

  const renderItem = (f: IndexField) => (
    <FieldItem
      key={f.name}
      field={f}
      pattern={pattern}
      filters={filters}
      isColumn={columns.includes(f.name)}
      open={openField === f.name}
      onToggle={() => setOpenField(openField === f.name ? null : f.name)}
      onAdd={onAdd}
      onToggleColumn={() => onToggleColumn(f.name)}
    />
  )

  return (
    <aside className="hidden w-64 shrink-0 flex-col bg-muted/10 lg:flex">
      <div className="flex items-center justify-between border-b border-border/70 px-3 pt-2.5 pb-1.5">
        <span className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
          {t('logExplorer.fields.title')}
        </span>
        <span className="font-mono text-[11px] text-muted-foreground/60">{visible.length}</span>
      </div>
      <div className="border-b border-border/70 p-2">
        <div className="relative">
          <Search size={13} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder={t('logExplorer.fields.filterPlaceholder')}
            className="h-8 border-0 bg-transparent pl-7 text-xs shadow-none focus-visible:bg-card focus-visible:ring-1 focus-visible:ring-border"
          />
        </div>
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto p-1.5">
        {visible.length === 0 && (
          <div className="px-2 py-6 text-center text-xs text-muted-foreground">{t('logExplorer.fields.none')}</div>
        )}

        {selected.length > 0 && (
          <>
            <SidebarSectionLabel>{t('logExplorer.fields.selected')}</SidebarSectionLabel>
            {selected.map(renderItem)}
            <SidebarSectionLabel className="mt-3">{t('logExplorer.fields.available')}</SidebarSectionLabel>
          </>
        )}
        {available.map(renderItem)}
      </div>
    </aside>
  )
}

function SidebarSectionLabel({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn('px-2 pb-1 pt-1 text-[10px] font-medium uppercase tracking-wider text-muted-foreground/55', className)}>
      {children}
    </div>
  )
}

const TYPE_META: Record<string, { icon: LucideIcon; color: string; label: string }> = {
  date: { icon: Calendar, color: 'text-violet-500', label: 'Date' },
  keyword: { icon: Tag, color: 'text-sky-500', label: 'Keyword' },
  text: { icon: Type, color: 'text-emerald-500', label: 'Text' },
  ip: { icon: Globe, color: 'text-fuchsia-500', label: 'IP' },
  boolean: { icon: ToggleLeft, color: 'text-rose-500', label: 'Boolean' },
}
const NUMBER_TYPES = new Set(['long', 'integer', 'short', 'byte', 'double', 'float', 'half_float', 'scaled_float'])

function typeMeta(type: string) {
  if (TYPE_META[type]) return TYPE_META[type]
  if (NUMBER_TYPES.has(type)) return { icon: Hash, color: 'text-amber-500', label: 'Number' }
  return { icon: Braces, color: 'text-muted-foreground', label: type || 'object' }
}

function TypeBadge({ type }: { type: string }) {
  const m = typeMeta(type)
  const Icon = m.icon
  return (
    <span
      className="flex h-5 w-5 shrink-0 items-center justify-center rounded bg-muted"
      title={m.label}
    >
      <Icon size={11} className={m.color} />
    </span>
  )
}

function FieldItem({
  field,
  pattern,
  filters,
  isColumn,
  open,
  onToggle,
  onAdd,
  onToggleColumn,
}: {
  field: IndexField
  pattern: IndexPattern | null
  filters: FilterType[]
  isColumn: boolean
  open: boolean
  onToggle: () => void
  onAdd: (f: FilterType) => void
  onToggleColumn: () => void
}) {
  const { t } = useTranslation()
  const [top, setTop] = useState<TopValues | null>(null)
  const [loading, setLoading] = useState(false)

  // Aggregations need the keyword sub-field for text types.
  const aggField =
    field.type === 'text' && !field.name.endsWith('.keyword') ? `${field.name}.keyword` : field.name

  useEffect(() => {
    if (!open || !pattern || top) return
    setLoading(true)
    svc
      .topValues(pattern.pattern, aggField, filters, 5)
      .then(setTop)
      .catch(() => setTop({ total: 0, top: [] }))
      .finally(() => setLoading(false))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  return (
    <div className={cn('group/field rounded-md', open && 'bg-card shadow-sm ring-1 ring-border/70')}>
      <div className={cn('flex items-center gap-2.5 rounded-md px-2 py-2', !open && 'hover:bg-card/70')}>
        <button onClick={onToggle} className="flex min-w-0 flex-1 items-center gap-2.5 text-left">
          <TypeBadge type={field.type} />
          <span className="flex-1 truncate font-mono text-xs" title={field.name}>
            {field.name}
          </span>
        </button>
        <button
          onClick={onToggleColumn}
          title={isColumn ? t('logExplorer.fields.removeColumn') : t('logExplorer.fields.addColumn')}
          className={cn(
            'flex h-6 w-6 shrink-0 items-center justify-center rounded transition-colors',
            isColumn
              ? 'text-primary hover:bg-primary/10'
              : 'text-muted-foreground opacity-0 hover:bg-muted group-hover/field:opacity-100'
          )}
        >
          {isColumn ? <Check size={13} /> : <Columns3 size={13} />}
        </button>
        <button onClick={onToggle} className="shrink-0">
          <ChevronRight size={13} className={cn('text-muted-foreground/60 transition-transform', open && 'rotate-90')} />
        </button>
      </div>
      {open && (
        <div className="px-2 pb-2.5 pt-1">
          <div className="mb-1.5 px-1 text-[10px] uppercase tracking-wider text-muted-foreground/60">
            {t('logExplorer.fields.topValues', { count: Math.min(5, top?.top.length ?? 0) })}
          </div>
          {loading ? (
            <div className="flex items-center gap-1.5 px-1 py-2 text-[11px] text-muted-foreground">
              <Loader2 className="h-3.5 w-3.5 animate-spin" /> {t('logExplorer.fields.loading')}
            </div>
          ) : !top || top.top.length === 0 ? (
            <div className="px-1 py-2 text-[11px] text-muted-foreground">{t('logExplorer.fields.noValues')}</div>
          ) : (
            <div className="space-y-1">
              {top.top.slice(0, 5).map((v) => (
                <div key={v.value} className="group rounded px-1.5 py-1.5 hover:bg-muted/50">
                  <div className="flex items-center justify-between gap-2">
                    <span className="truncate font-mono text-[11px]" title={v.value}>
                      {v.value || t('logExplorer.fields.empty')}
                    </span>
                    <span className="shrink-0 font-mono text-[10px] text-muted-foreground group-hover:hidden">
                      {Math.round(v.percent)}%
                    </span>
                    <div className="hidden shrink-0 items-center gap-1 group-hover:flex">
                      <button
                        title={t('logExplorer.fields.filterFor')}
                        onClick={() => onAdd({ field: field.name, operator: 'IS', value: v.value })}
                        className="flex h-5 w-5 items-center justify-center rounded text-emerald-500 hover:bg-emerald-500/15"
                      >
                        <Plus size={12} />
                      </button>
                      <button
                        title={t('logExplorer.fields.filterOut')}
                        onClick={() => onAdd({ field: field.name, operator: 'IS_NOT', value: v.value })}
                        className="flex h-5 w-5 items-center justify-center rounded text-red-500 hover:bg-red-500/15"
                      >
                        <Minus size={12} />
                      </button>
                    </div>
                  </div>
                  <div className="mt-1.5 h-1 overflow-hidden rounded-full bg-muted">
                    <div className="h-full rounded-full bg-primary/50" style={{ width: `${Math.max(3, v.percent)}%` }} />
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function RowMessage({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">{children}</div>
  )
}

/* ─── Chart mode ───────────────────────────────────────────────────────── */

// Valid OpenSearch calendar_interval tokens (lowercase) for the date histogram.
const CALENDAR_INTERVALS = [
  { id: 'minute', label: 'Minute' },
  { id: 'hour', label: 'Hour' },
  { id: 'day', label: 'Day' },
  { id: 'week', label: 'Week' },
  { id: 'month', label: 'Month' },
  { id: 'quarter', label: 'Quarter' },
  { id: 'year', label: 'Year' },
]

const SELECT_CLS =
  'h-8 cursor-pointer rounded-md border border-input bg-background/40 px-2 text-xs transition-colors focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring'

function ViewToggle({ mode, onChange }: { mode: 'table' | 'chart'; onChange: (m: 'table' | 'chart') => void }) {
  const { t } = useTranslation()
  const opts = [
    { id: 'table' as const, icon: TableIcon, label: t('logExplorer.view.table') },
    { id: 'chart' as const, icon: BarChart3, label: t('logExplorer.view.chart') },
  ]
  return (
    <div className="inline-flex shrink-0 items-center rounded-md border border-border bg-card p-0.5 text-xs">
      {opts.map(({ id, icon: Icon, label }) => (
        <button
          key={id}
          onClick={() => onChange(id)}
          className={cn(
            'flex items-center gap-1.5 rounded px-2.5 py-1 transition-colors',
            mode === id ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'
          )}
        >
          <Icon size={13} /> {label}
        </button>
      ))}
    </div>
  )
}

function chartTimeLabel(c: string) {
  const d = new Date(c)
  return Number.isNaN(d.getTime())
    ? c
    : d.toLocaleString(undefined, { month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function ChartPanel({
  pattern,
  fields,
  filters,
}: {
  pattern: IndexPattern | null
  fields: IndexField[]
  filters: FilterType[]
}) {
  const { t } = useTranslation()
  const selectable = useMemo(
    () => fields.filter((f) => !f.name.endsWith('.keyword')).sort((a, b) => a.name.localeCompare(b.name)),
    [fields]
  )
  const [fieldName, setFieldName] = useState('')
  const [interval, setInterval] = useState('day')
  const [data, setData] = useState<ChartView | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(false)

  // Default to @timestamp (time histogram) when present, else the first field.
  useEffect(() => {
    if (fieldName || selectable.length === 0) return
    setFieldName(selectable.find((f) => f.name === TS)?.name ?? selectable[0].name)
  }, [selectable, fieldName])

  const field = selectable.find((f) => f.name === fieldName) ?? null
  const isDate = field?.type === 'date'
  const aggField = field ? (field.type === 'text' ? `${field.name}.keyword` : field.name) : ''

  useEffect(() => {
    if (!pattern || !field) return
    setLoading(true)
    setError(false)
    svc
      .chartView({
        indexPattern: pattern.pattern,
        field: aggField,
        fieldDataType: field.type,
        filters,
        interval: isDate ? interval : '',
        top: 20,
      })
      .then(setData)
      .catch(() => {
        setData(null)
        setError(true)
      })
      .finally(() => setLoading(false))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pattern, fieldName, interval, filters])

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="flex flex-wrap items-center gap-2 border-b border-border/60 px-4 py-2.5 text-xs">
        <span className="text-muted-foreground">{t('logExplorer.chart.aggregateOn')}</span>
        <select value={fieldName} onChange={(e) => setFieldName(e.target.value)} className={cn(SELECT_CLS, 'min-w-[200px] font-mono')}>
          {selectable.map((f) => (
            <option key={f.name} value={f.name}>
              {f.name}
            </option>
          ))}
        </select>
        {isDate ? (
          <>
            <span className="text-muted-foreground">{t('logExplorer.chart.per')}</span>
            <select value={interval} onChange={(e) => setInterval(e.target.value)} className={SELECT_CLS}>
              {CALENDAR_INTERVALS.map((i) => (
                <option key={i.id} value={i.id}>
                  {t(`logExplorer.intervals.${i.id}`)}
                </option>
              ))}
            </select>
          </>
        ) : (
          <span className="text-muted-foreground">{t('logExplorer.chart.topValues')}</span>
        )}
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto p-5">
        {loading ? (
          <div className="flex h-full items-center justify-center gap-2 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" /> {t('logExplorer.chart.building')}
          </div>
        ) : error ? (
          <div className="flex h-full items-center justify-center gap-2 text-sm text-muted-foreground">
            <AlertTriangle size={16} className="text-amber-500" /> {t('logExplorer.chart.failed')}
          </div>
        ) : !data || data.values.length === 0 ? (
          <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
            {t('logExplorer.chart.noData')}
          </div>
        ) : isDate ? (
          <TimeChart data={data} />
        ) : (
          <TermsChart data={data} />
        )}
      </div>
    </div>
  )
}

function TermsChart({ data }: { data: ChartView }) {
  const { t } = useTranslation()
  const max = Math.max(1, ...data.values)
  return (
    <div className="space-y-2">
      {data.categories.map((cat, i) => {
        const v = data.values[i] ?? 0
        const pct = (v / max) * 100
        return (
          <div key={`${cat}-${i}`} className="flex items-center gap-3 text-[13px]">
            <div className="w-56 shrink-0 truncate font-mono text-foreground" title={cat}>
              {cat || t('logExplorer.fields.empty')}
            </div>
            <div className="relative h-6 flex-1 overflow-hidden rounded bg-muted/50">
              <div className="h-full rounded bg-primary/40" style={{ width: `${Math.max(1.5, pct)}%` }} />
            </div>
            <div className="w-24 shrink-0 text-right font-mono tabular-nums text-muted-foreground">
              {v.toLocaleString()}
            </div>
          </div>
        )
      })}
    </div>
  )
}

function TimeChart({ data }: { data: ChartView }) {
  const values = data.values
  const max = Math.max(1, ...values)
  const w = 1200
  const h = 300
  const n = values.length || 1
  const slot = w / n
  const bw = Math.max(1, slot - 3)
  return (
    <div>
      <svg viewBox={`0 0 ${w} ${h}`} className="h-[300px] w-full" preserveAspectRatio="none">
        {values.map((v, i) => {
          if (v <= 0) return null
          const bh = Math.max(2, (v / max) * h)
          const x = i * slot + (slot - bw) / 2
          return (
            <rect key={i} x={x} y={h - bh} width={bw} height={bh} rx={2} className="fill-primary/45">
              <title>{`${data.categories[i]}: ${v.toLocaleString()}`}</title>
            </rect>
          )
        })}
      </svg>
      {data.categories.length > 1 && (
        <div className="mt-2 flex justify-between text-[10px] text-muted-foreground">
          <span className="font-mono">{chartTimeLabel(data.categories[0])}</span>
          <span className="font-mono">{chartTimeLabel(data.categories[data.categories.length - 1])}</span>
        </div>
      )}
    </div>
  )
}
