import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  AlertTriangle,
  BarChart3,
  Braces,
  Calendar,
  Check,
  ChevronDown,
  ChevronRight,
  Clock,
  Columns3,
  Code2,
  Copy,
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
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { Pagination } from '@/shared/components/ui/pagination'
import {
  logExplorerHttpService as svc,
  LogExplorerHttpError,
} from '../services/log-explorer-http.service'
import type {
  ChartView,
  FilterType,
  IndexField,
  IndexPattern,
  LogDocument,
  TopValues,
} from '../types/log-explorer.types'

/* ─── Constants ────────────────────────────────────────────────────────── */

const TS = '@timestamp'

// `interval` must be a valid OpenSearch calendar_interval token (lowercase, singular):
// minute / hour / day / week / month. The backend passes it through verbatim.
const TIME_PRESETS = [
  { id: '15m', label: 'Last 15 minutes', from: 'now-15m', interval: 'minute' },
  { id: '1h', label: 'Last 1 hour', from: 'now-1h', interval: 'minute' },
  { id: '6h', label: 'Last 6 hours', from: 'now-6h', interval: 'hour' },
  { id: '24h', label: 'Last 24 hours', from: 'now-24h', interval: 'hour' },
  { id: '7d', label: 'Last 7 days', from: 'now-7d', interval: 'day' },
  { id: '30d', label: 'Last 30 days', from: 'now-30d', interval: 'day' },
] as const

// Field-name candidates for the compact result columns (read from the flattened doc).
const MSG_FIELDS = ['log.message', 'logx.message', 'message', 'event.original', 'rule.name', 'logx.raw']
// "Source" = where the log came from (the host/origin, which varies row to row).
const SRC_FIELDS = ['dataSource', 'host.name', 'agent.name', 'log.computer', 'source', 'origin']
const LEVEL_FIELDS = ['log.level', 'severity', 'level', 'event.severity', 'logx.severity']

// Metadata/noise fields excluded from the inline document preview (same on every row).
const NOISE_KEYS = new Set(['@timestamp', '@version', 'dataType', 'deviceTime', 'id', 'isAnomaly', 'timestamp'])
const NOISE_PREFIXES = [
  'tenant',
  'globalaccount',
  'log.activityid',
  'log.correlation',
  'log.version',
  'log.opcode',
  'log.task',
  'log.keywords',
  'log.processid',
  'log.threadid',
  'log.recordid',
  'log.providerguid',
  'log.level',
]

function isNoise(k: string): boolean {
  if (NOISE_KEYS.has(k)) return true
  const lower = k.toLowerCase()
  return NOISE_PREFIXES.some((p) => lower.startsWith(p))
}

// Rank fields so the event-specific content (log.data.*) leads the preview.
function fieldRank(k: string): number {
  if (k.startsWith('log.data.')) return 0
  if (k.startsWith('event.')) return 1
  if (k.startsWith('log.') || k.startsWith('logx.') || k.startsWith('alert.')) return 2
  return 3
}

const LEVEL_TONE: Record<string, { dot: string; tone: string }> = {
  critical: { dot: 'bg-red-500', tone: 'text-red-500' },
  high: { dot: 'bg-red-500', tone: 'text-red-500' },
  error: { dot: 'bg-orange-500', tone: 'text-orange-500' },
  warn: { dot: 'bg-amber-500', tone: 'text-amber-500' },
  warning: { dot: 'bg-amber-500', tone: 'text-amber-500' },
  medium: { dot: 'bg-amber-500', tone: 'text-amber-500' },
  info: { dot: 'bg-sky-500', tone: 'text-sky-500' },
  low: { dot: 'bg-sky-500', tone: 'text-sky-500' },
  debug: { dot: 'bg-muted-foreground', tone: 'text-muted-foreground' },
}

/* ─── Helpers ──────────────────────────────────────────────────────────── */

function flattenDoc(obj: unknown, prefix = '', out: Record<string, unknown> = {}): Record<string, unknown> {
  if (obj && typeof obj === 'object' && !Array.isArray(obj)) {
    for (const [k, v] of Object.entries(obj as Record<string, unknown>)) {
      const key = prefix ? `${prefix}.${k}` : k
      if (v && typeof v === 'object' && !Array.isArray(v)) flattenDoc(v, key, out)
      else out[key] = Array.isArray(v) ? v.join(', ') : v
    }
  }
  return out
}

function pick(flat: Record<string, unknown>, fields: string[]): string | undefined {
  for (const f of fields) {
    const v = flat[f]
    if (v != null && v !== '') return String(v)
  }
  return undefined
}

// Ordered, signal-carrying key/value pairs for the inline document preview shown
// when there's no single "message" field (the common case for raw event logs).
function docPreview(flat: Record<string, unknown>): [string, string][] {
  return Object.entries(flat)
    .filter(([k, v]) => v != null && v !== '' && k !== 'dataSource' && !isNoise(k))
    .sort((a, b) => fieldRank(a[0]) - fieldRank(b[0]))
    .slice(0, 8)
    .map(([k, v]) => [k.split('.').pop() ?? k, String(v)])
}

function shortTime(iso: string) {
  const d = new Date(iso)
  return Number.isNaN(d.getTime())
    ? iso
    : d.toLocaleString(undefined, {
        month: 'short',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      })
}

function absTimestamp(iso: string) {
  const d = new Date(iso)
  return Number.isNaN(d.getTime())
    ? iso
    : d.toLocaleString(undefined, {
        month: 'short',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      })
}

// Free-text "search all" maps to a wildcard field by data nature (logx.* / alert.*).
function textPrefix(pattern: string): string {
  return /alert/i.test(pattern) ? 'alert.*' : 'logx.*'
}

/* ─── Page ─────────────────────────────────────────────────────────────── */

const PAGE_SIZE_DEFAULT = 25

export function LogExplorerPage() {
  const { t } = useTranslation()
  const [patterns, setPatterns] = useState<IndexPattern[]>([])
  const [pattern, setPattern] = useState<IndexPattern | null>(null)

  const [presetId, setPresetId] = useState<(typeof TIME_PRESETS)[number]['id']>('24h')
  const preset = TIME_PRESETS.find((p) => p.id === presetId) ?? TIME_PRESETS[3]

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

  /* Load active index patterns once; default to v11-log-* (or first). */
  useEffect(() => {
    let cancelled = false
    svc
      .patterns()
      .then((ps) => {
        if (cancelled) return
        setPatterns(ps)
        setPattern(ps.find((p) => p.pattern === 'v11-log-*') ?? ps[0] ?? null)
      })
      .catch(() => {
        if (!cancelled) setError('explorer:patterns')
      })
    return () => {
      cancelled = true
    }
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
    const out: FilterType[] = [{ field: TS, operator: 'IS_BETWEEN', value: [preset.from, 'now'] }]
    if (appliedQuery.trim() && pattern) {
      out.push({ field: textPrefix(pattern.pattern), operator: 'IS_IN_FIELDS', value: appliedQuery.trim() })
    }
    out.push(...filters)
    return out
  }, [preset.from, appliedQuery, pattern, filters])

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
  }, [pattern, presetId, appliedQuery, appliedSql, sqlMode, filters, page, pageSize, nonce])

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
          presetId={presetId}
          onPreset={(id) => {
            setPresetId(id)
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
        {filters.length > 0 ? (
          <FilterChips filters={filters} onRemove={removeFilter} onClear={() => setFilters([])} />
        ) : (
          <span />
        )}
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
  presetId,
  onPreset,
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
  presetId: string
  onPreset: (id: (typeof TIME_PRESETS)[number]['id']) => void
  onRun: () => void
  onRefresh: () => void
  loading: boolean
  onExport: () => void
}) {
  const { t } = useTranslation()
  const [patternOpen, setPatternOpen] = useState(false)
  const [timeOpen, setTimeOpen] = useState(false)
  const preset = TIME_PRESETS.find((p) => p.id === presetId) ?? TIME_PRESETS[3]

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
      <Dropdown
        open={timeOpen}
        onOpenChange={setTimeOpen}
        align="right"
        trigger={
          <>
            <Clock size={13} className="text-muted-foreground" />
            {t(`logExplorer.time.${preset.id}`)}
            <ChevronDown size={11} className="opacity-60" />
          </>
        }
      >
        {TIME_PRESETS.map((p) => (
          <button
            key={p.id}
            onClick={() => {
              onPreset(p.id)
              setTimeOpen(false)
            }}
            className={cn(
              'flex w-full items-center px-3 py-2 text-left text-sm transition-colors',
              p.id === presetId ? 'bg-muted/60' : 'hover:bg-muted/60'
            )}
          >
            {t(`logExplorer.time.${p.id}`)}
          </button>
        ))}
      </Dropdown>

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
              <span className="max-w-[220px] truncate font-mono font-medium">{String(f.value)}</span>
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

/* ─── Results table ────────────────────────────────────────────────────── */

function gridTemplate(columns: string[]): string {
  if (columns.length === 0) return '24px 4px 195px 150px minmax(0, 1fr)'
  return `24px 4px 195px ${columns.map(() => 'minmax(140px, 1fr)').join(' ')}`
}

function colValue(flat: Record<string, unknown>, c: string): string {
  const v = flat[c]
  if (v == null || v === '') return '—'
  if (c === TS) return shortTime(String(v))
  return String(v)
}

function ResultsHeader({ columns, onRemoveColumn }: { columns: string[]; onRemoveColumn: (c: string) => void }) {
  const { t } = useTranslation()
  return (
    <div
      className="grid items-center gap-4 border-b border-border/70 bg-muted/20 px-5 py-2.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: gridTemplate(columns) }}
    >
      <div />
      <div />
      <div>{t('logExplorer.results.time')}</div>
      {columns.length === 0 ? (
        <>
          <div>{t('logExplorer.results.source')}</div>
          <div>{t('logExplorer.results.message')}</div>
        </>
      ) : (
        columns.map((c) => (
          <div key={c} className="group flex min-w-0 items-center gap-1">
            <span className="truncate" title={c}>
              {c.split('.').pop()}
            </span>
            <button
              onClick={() => onRemoveColumn(c)}
              title={t('logExplorer.results.removeColumn', { field: c })}
              className="shrink-0 opacity-0 transition-opacity hover:text-foreground group-hover:opacity-100"
            >
              <X size={11} />
            </button>
          </div>
        ))
      )}
    </div>
  )
}

function ResultRow({
  doc,
  columns,
  expanded,
  onToggle,
  onAdd,
}: {
  doc: LogDocument
  columns: string[]
  expanded: boolean
  onToggle: () => void
  onAdd: (f: FilterType) => void
}) {
  const flat = useMemo(() => flattenDoc(doc), [doc])
  const ts = (flat[TS] as string) ?? ''
  const source = pick(flat, SRC_FIELDS) ?? '—'
  const level = (pick(flat, LEVEL_FIELDS) ?? '').toLowerCase()
  const tone = LEVEL_TONE[level] ?? { dot: 'bg-muted-foreground/50', tone: 'text-muted-foreground' }
  const message = pick(flat, MSG_FIELDS)
  const preview = useMemo(
    () => (columns.length > 0 || message ? null : docPreview(flat)),
    [columns.length, message, flat]
  )

  return (
    <>
      <div
        onClick={onToggle}
        className={cn(
          'grid cursor-pointer items-center gap-4 border-b border-border/40 px-5 py-3 text-[13px] leading-relaxed transition-colors last:border-b-0',
          expanded ? 'bg-muted/30' : 'hover:bg-muted/20'
        )}
        style={{ gridTemplateColumns: gridTemplate(columns) }}
      >
        <ChevronRight size={14} className={cn('text-muted-foreground/60 transition-transform', expanded && 'rotate-90 text-foreground')} />
        <span className={cn('h-4 w-1 rounded-full', tone.dot)} />
        <div className="font-mono tabular-nums text-muted-foreground">{ts ? shortTime(ts) : '—'}</div>
        {columns.length > 0 ? (
          columns.map((c) => (
            <div key={c} className="truncate font-mono text-foreground/85" title={colValue(flat, c)}>
              {colValue(flat, c)}
            </div>
          ))
        ) : (
          <>
            <div className="truncate font-mono text-foreground/70">{source}</div>
            {message ? (
              <div className="truncate text-foreground">{message}</div>
            ) : (
              <div className="flex items-center overflow-hidden whitespace-nowrap">
                {preview!.map(([k, v], idx) => (
                  <span key={k} className="flex shrink-0 items-center">
                    {idx > 0 && <span className="px-2.5 text-border">·</span>}
                    <span className="text-muted-foreground">{k}</span>
                    <span className="ml-1.5 font-mono text-foreground">{v}</span>
                  </span>
                ))}
              </div>
            )}
          </>
        )}
      </div>
      {expanded && <ExpandedPanel flat={flat} doc={doc} onAdd={onAdd} />}
    </>
  )
}

type DetailTab = 'fields' | 'json'

function ExpandedPanel({
  flat,
  doc,
  onAdd,
}: {
  flat: Record<string, unknown>
  doc: LogDocument
  onAdd: (f: FilterType) => void
}) {
  const { t } = useTranslation()
  const [tab, setTab] = useState<DetailTab>('fields')
  const ts = (flat[TS] as string) ?? ''
  const entries = Object.entries(flat).sort(([a], [b]) => a.localeCompare(b))

  return (
    <div className="border-b border-l-2 border-border/50 border-l-sky-500/50 bg-muted/15 last:border-b-0">
      <div className="flex items-center justify-between gap-4 border-b border-border/40 px-5 py-2.5">
        <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
          {ts && <span className="font-mono">{absTimestamp(ts)}</span>}
          <span className="font-mono">{t('logExplorer.detail.fieldsCount', { count: entries.length })}</span>
        </div>
        <button
          onClick={() => {
            void navigator.clipboard.writeText(JSON.stringify(doc, null, 2))
            toast.success(t('logExplorer.detail.copied'))
          }}
          className="flex h-7 items-center gap-1.5 rounded-md px-2 text-[11px] text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <Copy size={12} /> {t('logExplorer.detail.copyJson')}
        </button>
      </div>

      <div className="flex items-center gap-4 border-b border-border/60 px-5">
        <DetailTabBtn id="fields" current={tab} onChange={setTab}>
          {t('logExplorer.detail.parsedFields')}
        </DetailTabBtn>
        <DetailTabBtn id="json" current={tab} onChange={setTab}>
          {t('logExplorer.detail.json')}
        </DetailTabBtn>
      </div>

      <div className="p-5">
        {tab === 'fields' ? (
          <div className="overflow-hidden rounded-md border border-border bg-card">
            {entries.map(([k, v], i) => (
              <div
                key={k}
                className={cn(
                  'group grid grid-cols-[260px_1fr_60px] items-center gap-4 px-4 py-2 text-[13px] leading-relaxed hover:bg-muted/30',
                  i < entries.length - 1 && 'border-b border-border/60'
                )}
              >
                <div className="truncate font-mono text-xs text-muted-foreground">{k}</div>
                <div className="break-all font-mono text-xs">{String(v)}</div>
                <div className="flex justify-end gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
                  <button
                    title={t('logExplorer.detail.filterFor')}
                    onClick={() => onAdd({ field: k, operator: 'IS', value: String(v) })}
                    className="flex h-5 w-5 items-center justify-center rounded text-emerald-500 hover:bg-emerald-500/15"
                  >
                    <Plus size={10} />
                  </button>
                  <button
                    title={t('logExplorer.detail.filterOut')}
                    onClick={() => onAdd({ field: k, operator: 'IS_NOT', value: String(v) })}
                    className="flex h-5 w-5 items-center justify-center rounded text-red-500 hover:bg-red-500/15"
                  >
                    <Minus size={10} />
                  </button>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <pre className="overflow-x-auto rounded-md border border-border bg-card p-3 font-mono text-[11px] leading-relaxed">
            {JSON.stringify(doc, null, 2)}
          </pre>
        )}
      </div>
    </div>
  )
}

function DetailTabBtn({
  id,
  current,
  onChange,
  children,
}: {
  id: DetailTab
  current: DetailTab
  onChange: (t: DetailTab) => void
  children: React.ReactNode
}) {
  const active = id === current
  return (
    <button
      onClick={() => onChange(id)}
      className={cn(
        'relative -mb-px border-b-2 py-2 text-xs transition-colors',
        active ? 'border-foreground text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground'
      )}
    >
      {children}
    </button>
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
