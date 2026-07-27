import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { AlertTriangle, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { useLocation, useNavigate } from 'react-router-dom'
import { Button } from '@/shared/components/ui/button'
import { presetRange, type TimeRange } from '@/shared/components/ui/time-range-picker'
import { ResultsHeader, ResultRow, flattenDoc } from './log-results'
import { QueryBar } from './QueryBar'
import { AddFilterButton } from './AddFilterButton'
import { FilterChips } from './FilterChips'
import { FieldSidebar } from './FieldSidebar'
import { RowMessage } from './RowMessage'
import { ViewToggle } from './ViewToggle'
import { SavedSearches, type SavedSearchState } from './SavedSearches'
import { HistogramStrip } from './HistogramStrip'
import { ChartPanel } from './ChartPanel'
import { TS } from './log-explorer.constants'
import {
  logExplorerHttpService as svc,
  LogExplorerHttpError,
} from '../services/log-explorer-http.service'
import type {
  FilterType,
  IndexField,
  IndexPattern,
  LogDocument,
  LogExplorerTabConfig,
} from '../types/log-explorer.types'

/* ─── Constants ────────────────────────────────────────────────────────── */

// Field-name candidates for the compact result columns (read from the flattened doc).
const MSG_FIELDS = ['log.message', 'logx.message', 'message', 'event.original', 'rule.name', 'logx.raw']

/* ─── Helpers ──────────────────────────────────────────────────────────── */

// Free-text "search all" maps to a wildcard field by data nature (logx.* / alert.*).
function textPrefix(pattern: string): string {
  return /alert/i.test(pattern) ? 'alert.*' : 'logx.*'
}

/* ─── Page ─────────────────────────────────────────────────────────────── */

// Larger page = fewer round-trips while scrolling; the list is virtualization-free
// but rows are light, so 50 keeps scroll smooth.
const PAGE_SIZE_DEFAULT = 50

// SIEM-important fields, in priority order. When the analyst hasn't picked columns,
// the table auto-shows the first few of these that actually carry values in the
// current results — so firewall logs surface src/dst IP, auth logs surface the
// user, etc., without manual setup. Mirrors the legacy "summary" field list.
const IMPORTANT_FIELDS = [
  'severity',
  'dataType',
  'action',
  'actionResult',
  'origin.user',
  'origin.ip',
  'origin.host',
  'target.user',
  'target.ip',
  'target.host',
  'target.port',
  'protocol',
  'statusCode',
  'connectionStatus',
  'command',
  'origin.url',
  'target.url',
]
const MAX_AUTO_COLUMNS = 5

/** Router state passed by an alert's "view all related logs" action. */
interface RelatedLogsSeed {
  ids: string[]
  indexPattern: string
  timeFrom: string
  timeTo: string
  alertName?: string
  truncated?: boolean
}

interface LogExplorerViewProps {
  initial: LogExplorerTabConfig
  onConfigChange: (patch: Partial<LogExplorerTabConfig>) => void
}

export function LogExplorerView({ initial, onConfigChange }: LogExplorerViewProps) {
  const { t } = useTranslation()
  const location = useLocation()
  const navigate = useNavigate()
  const seededRef = useRef(false)
  const [patterns, setPatterns] = useState<IndexPattern[]>([])
  const [pattern, setPattern] = useState<IndexPattern | null>(null)

  const [range, setRange] = useState<TimeRange>(initial.range)

  const [searchInput, setSearchInput] = useState(initial.searchInput)
  const [appliedQuery, setAppliedQuery] = useState(initial.appliedQuery)
  const [sqlMode, setSqlMode] = useState(initial.sqlMode)
  const [sqlInput, setSqlInput] = useState(initial.sqlInput)
  const [appliedSql, setAppliedSql] = useState(initial.appliedSql)

  const [filters, setFilters] = useState<FilterType[]>(initial.filters)

  const [total, setTotal] = useState(0)

  const [rows, setRows] = useState<LogDocument[]>([])
  const [loading, setLoading] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [fields, setFields] = useState<IndexField[]>([])
  const [expanded, setExpanded] = useState<number | null>(null)
  const [nonce, setNonce] = useState(0)
  // Infinite scroll: pages are accumulated, not replaced. pageRef tracks the last
  // page fetched (1-based); a ref so the scroll handler reads it without re-binding.
  const pageRef = useRef(0)
  const scrollRef = useRef<HTMLDivElement | null>(null)

  const [viewMode, setViewMode] = useState<'table' | 'chart'>(initial.viewMode)

  // Table columns (Discover-style). Empty → the compact "Document" summary view.
  const [columns, setColumns] = useState<string[]>(initial.columns)
  const toggleColumn = useCallback(
    (name: string) =>
      setColumns((cur) => (cur.includes(name) ? cur.filter((c) => c !== name) : [...cur, name])),
    []
  )

  // Capture the current query as a reusable saved search.
  const currentSnapshot = useCallback(
    (): SavedSearchState => ({
      patternStr: pattern?.pattern ?? null,
      range,
      filters,
      searchInput,
      appliedQuery,
    }),
    [pattern, range, filters, searchInput, appliedQuery]
  )

  const loadSnapshot = useCallback(
    (s: SavedSearchState) => {
      if (s.patternStr) {
        const p = patterns.find((x) => x.pattern === s.patternStr)
        if (p) setPattern(p)
      }
      setRange(s.range)
      setFilters(s.filters ?? [])
      setSearchInput(s.searchInput ?? '')
      setAppliedQuery(s.appliedQuery ?? '')
      setExpanded(null)
    },
    [patterns]
  )

  // "View surrounding events": re-scope the view to a ±15-minute window around the
  // selected event, narrowed to its source, so the analyst can read what happened
  // immediately before and after it without leaving the table.
  const viewSurrounding = useCallback((ts: string, srcField?: string, srcVal?: string) => {
    const center = new Date(ts).getTime()
    if (Number.isNaN(center)) return
    const W = 15 * 60 * 1000
    setExpanded(null)
    setFilters(srcField && srcVal ? [{ field: srcField, operator: 'IS', value: srcVal }] : [])
    setRange({
      from: new Date(center - W).toISOString(),
      to: new Date(center + W).toISOString(),
      interval: 'minute',
    })
  }, [])

  // Persistence is gated until the initial patterns load resolves — otherwise the
  // first render would write `patternStr: null` and wipe the saved tab pattern.
  const [readyToPersist, setReadyToPersist] = useState(false)

  /* Load active index patterns once; resolve the tab's saved patternStr against
     the live list. When arriving from an alert's "view all related logs" or a
     SOC-AI chat deep-link, mutate the current view (the active tab). */
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
          setReadyToPersist(true)
          return
        }
        // SOC-AI chat navigation: apply the agent's filters + time window.
        if (state?.socaiFilters?.length && !seededRef.current) {
          seededRef.current = true
          setPattern(ps.find((p) => p.pattern === 'v11-log-*') ?? ps[0] ?? null)
          setFilters(state.socaiFilters)
          if (state.socaiTime) setRange(presetRange(state.socaiTime))
          navigate(location.pathname, { replace: true })
          setReadyToPersist(true)
          return
        }
        // Default: resolve the tab's saved patternStr to a live IndexPattern.
        const target = initial.patternStr
        setPattern(
          (target ? ps.find((p) => p.pattern === target) : null) ??
            ps.find((p) => p.pattern === 'v11-log-*') ??
            ps[0] ??
            null
        )
        setReadyToPersist(true)
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

  /* Persist any change to the tab's config (after the first patterns load). */
  useEffect(() => {
    if (!readyToPersist) return
    onConfigChange({
      patternStr: pattern?.pattern ?? null,
      range,
      filters,
      columns,
      searchInput,
      appliedQuery,
      sqlMode,
      sqlInput,
      appliedSql,
      viewMode,
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    readyToPersist,
    pattern,
    range,
    filters,
    columns,
    searchInput,
    appliedQuery,
    sqlMode,
    sqlInput,
    appliedSql,
    viewMode,
  ])

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

  // Auto-detected default columns: the important fields present in the current
  // results. Only used when the analyst hasn't picked their own columns.
  const autoColumns = useMemo(() => {
    if (columns.length > 0 || rows.length === 0) return []
    const sample = rows.slice(0, 50).map((d) => flattenDoc(d))
    const present = IMPORTANT_FIELDS.filter((f) =>
      sample.some((flat) => {
        const v = flat[f]
        return v != null && v !== ''
      })
    )
    return present.slice(0, MAX_AUTO_COLUMNS)
  }, [columns.length, rows])

  /* Fetch one page. page 1 replaces the list (fresh query); later pages append
     (infinite scroll). The histogram fetches separately. */
  const fetchPage = useCallback(
    async (pageNum: number) => {
      if (!pattern) return
      const fresh = pageNum <= 1
      if (fresh) {
        setLoading(true)
        setError(null)
      } else {
        setLoadingMore(true)
      }
      try {
        let data: LogDocument[] = []
        let totalCount = 0
        if (sqlMode) {
          if (!appliedSql.trim()) {
            setRows([])
            setTotal(0)
            pageRef.current = 1
            return
          }
          const r = await svc.searchSql(appliedSql.trim(), pageNum, PAGE_SIZE_DEFAULT)
          data = r.data ?? []
          totalCount = r.total
        } else {
          const r = await svc.search({
            indexPattern: pattern.pattern,
            filters: buildFilters(),
            page: pageNum,
            size: PAGE_SIZE_DEFAULT,
          })
          data = r.data ?? []
          totalCount = r.total
        }
        pageRef.current = pageNum
        setTotal(totalCount)
        setRows((prev) => (fresh ? data : [...prev, ...data]))
      } catch (e) {
        if (fresh) {
          setError(e instanceof LogExplorerHttpError ? e.message : 'explorer:failed')
          setRows([])
          setTotal(0)
        }
      } finally {
        setLoading(false)
        setLoadingMore(false)
      }
    },
    [pattern, sqlMode, appliedSql, buildFilters]
  )

  // Fresh load whenever the query inputs change → reset to page 1 (replace).
  useEffect(() => {
    setExpanded(null)
    void fetchPage(1)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pattern, range.from, range.to, appliedQuery, appliedSql, sqlMode, filters, nonce])

  const loadMore = useCallback(() => {
    if (loading || loadingMore) return
    if (rows.length >= total) return
    void fetchPage(pageRef.current + 1)
  }, [loading, loadingMore, rows.length, total, fetchPage])

  // Auto-load the next page as the results list nears its bottom.
  const onScroll = useCallback(
    (e: React.UIEvent<HTMLDivElement>) => {
      const el = e.currentTarget
      if (el.scrollHeight - el.scrollTop - el.clientHeight < 400) loadMore()
    },
    [loadMore]
  )

  /* Commit the input box and run (Enter / Run button). */
  const submit = () => {
    setExpanded(null)
    if (sqlMode) {
      if (appliedSql === sqlInput) setNonce((n) => n + 1)
      else setAppliedSql(sqlInput)
    } else {
      if (appliedQuery === searchInput) setNonce((n) => n + 1)
      else setAppliedQuery(searchInput)
    }
  }

  const addFilter = (f: FilterType) => {
    setFilters((cur) =>
      cur.some((c) => c.field === f.field && c.operator === f.operator && c.value === f.value) ? cur : [...cur, f]
    )
  }
  const removeFilter = (i: number) => {
    setFilters((cur) => cur.filter((_, idx) => idx !== i))
  }

  return (
    <div className="flex h-full min-h-0 flex-col px-6 pb-4 pt-3">
      <div>
        <QueryBar
          patterns={patterns}
          pattern={pattern}
          onPattern={(p) => {
            setPattern(p)
            setFilters([])
            setColumns([])
          }}
          searchInput={searchInput}
          onSearchInput={setSearchInput}
          sqlMode={sqlMode}
          onSqlMode={(v) => setSqlMode(v)}
          sqlInput={sqlInput}
          onSqlInput={setSqlInput}
          fields={fields}
          range={range}
          onRange={(r) => setRange(r)}
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

      <div className="mt-3 flex items-center justify-between gap-2">
        <div className="flex flex-wrap items-center gap-2">
          {!sqlMode && pattern && (
            <AddFilterButton pattern={pattern} fields={fields} filters={activeFilterList} onAdd={addFilter} />
          )}
          {!sqlMode && <SavedSearches snapshot={currentSnapshot} onLoad={loadSnapshot} />}
          {filters.length > 0 && <FilterChips filters={filters} onRemove={removeFilter} onClear={() => setFilters([])} />}
        </div>
        <div className="flex items-center gap-3">
          <span className="whitespace-nowrap text-xs text-muted-foreground">
            {loading ? (
              t('logExplorer.searching')
            ) : (
              <>
                <span className="font-medium text-foreground">{total.toLocaleString()}</span> {t('logExplorer.eventsIn')}{' '}
                <span className="font-mono">{pattern?.pattern ?? '—'}</span>
              </>
            )}
          </span>
          {!sqlMode && <ViewToggle mode={viewMode} onChange={setViewMode} />}
        </div>
      </div>

      <div className="mt-2 flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card">
        {viewMode === 'chart' && !sqlMode ? (
          <ChartPanel pattern={pattern} fields={fields} filters={activeFilterList} />
        ) : (
          <>
            {pattern && (
              <HistogramStrip pattern={pattern} filters={activeFilterList} range={range} />
            )}
            <div className="flex min-h-0 flex-1">
              <FieldSidebar
                fields={fields}
                pattern={pattern}
                filters={activeFilterList}
                columns={columns}
                onAdd={addFilter}
                onToggleColumn={toggleColumn}
              />
            <div className="flex min-w-0 flex-1 flex-col border-l border-border">
              <div ref={scrollRef} onScroll={onScroll} className="min-h-0 flex-1 overflow-auto">
                <ResultsHeader columns={columns} autoColumns={autoColumns} onRemoveColumn={toggleColumn} />
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
                  <>
                    {rows.map((doc, i) => (
                      <ResultRow
                        key={i}
                        doc={doc}
                        columns={columns}
                        autoColumns={autoColumns}
                        expanded={expanded === i}
                        onToggle={() => setExpanded(expanded === i ? null : i)}
                        onAdd={addFilter}
                        onSurrounding={viewSurrounding}
                      />
                    ))}
                    {loadingMore && (
                      <div className="flex items-center justify-center gap-2 py-4 text-xs text-muted-foreground">
                        <Loader2 className="h-3.5 w-3.5 animate-spin" /> {t('logExplorer.results.loadingMore')}
                      </div>
                    )}
                    {!loadingMore && rows.length >= total && (
                      <div className="py-4 text-center text-[11px] text-muted-foreground/70">
                        {t('logExplorer.results.endOfResults', { count: total.toLocaleString() })}
                      </div>
                    )}
                  </>
                )}
              </div>
            </div>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
