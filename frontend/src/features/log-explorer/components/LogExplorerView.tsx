import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { AlertTriangle, Loader2, X } from 'lucide-react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { useLocation, useNavigate } from 'react-router-dom'
import { Button } from '@/shared/components/ui/button'
import { presetRange, resolveRange, type TimeRange } from '@/shared/components/ui/time-range-picker'
import { looksLikeSql } from '../domain/sql-sync'
import { ResultsHeader, ResultRow } from './log-results'
import { MSG_FIELDS, SRC_FIELDS, flattenDoc, pick, previewText } from '../domain/flatten'
import { CustomFilterBar } from '@/shared/components/filters/CustomFilterBar'
import type { CustomFilter, FilterOpDef } from '@/shared/components/filters/custom-filter.types'
import { QueryBar } from './QueryBar'
import { FieldSidebar } from './FieldSidebar'
import { RowMessage } from './RowMessage'
import { ViewToggle } from './ViewToggle'
import { SavedSearches, type SavedSearchState } from './SavedSearches'
import { HistogramStrip } from './HistogramStrip'
import { ChartPanel } from './ChartPanel'
import { OP_KEY, TS } from './log-explorer.constants'
import {
  logExplorerHttpService as svc,
  LogExplorerHttpError,
  DEFAULT_DATASET,
  type ExportColumn,
} from '../services/log-explorer-http.service'
import type { DatasetTypes } from './DatasetSelector'
import type {
  FilterOperator,
  FilterType,
  IndexField,
  LogDocument,
  LogExplorerTabConfig,
} from '../types/log-explorer.types'

/* ─── Constants ────────────────────────────────────────────────────────── */

/* ─── Helpers ──────────────────────────────────────────────────────────── */

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
  dataType: string
  timeFrom: string
  timeTo: string
  alertName?: string
  truncated?: boolean
}

const BUILDER_OPS: FilterOpDef[] = [
  { id: 'IS', label: 'is', needsValue: true },
  { id: 'IS_NOT', label: 'is not', needsValue: true },
  { id: 'CONTAIN', label: 'contains', needsValue: true },
  { id: 'EXIST', label: 'exists', needsValue: false },
]

function toCustom(f: FilterType): CustomFilter {
  return { field: f.field, label: f.field, operator: f.operator, value: typeof f.value === 'string' ? f.value : '' }
}

function toFilter(cf: CustomFilter): FilterType {
  const op = BUILDER_OPS.find((o) => o.id === cf.operator)
  return { field: cf.field, operator: cf.operator as FilterOperator, value: op?.needsValue ? cf.value : undefined }
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
  // Two axes: which table, and which kind of record inside it.
  const [sources, setSources] = useState<DatasetTypes[]>([])
  const [dataset, setDataset] = useState<string>(initial.dataset ?? DEFAULT_DATASET)
  const [pattern, setPattern] = useState<string | null>(null)

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
  // Which dataset the loaded fields describe. Until it matches the one on
  // screen, nothing may reason about the shape of what it is searching.
  const [fieldsFor, setFieldsFor] = useState<string | null>(null)
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
      dataset,
      patternStr: pattern ?? null,
      range,
      filters,
      searchInput,
      appliedQuery,
    }),
    [pattern, range, filters, searchInput, appliedQuery]
  )

  const loadSnapshot = useCallback(
    (s: SavedSearchState) => {
      if (s.dataset) setDataset(s.dataset)
      setPattern(s.patternStr ?? null)
      setRange(s.range)
      setFilters(s.filters ?? [])
      setSearchInput(s.searchInput ?? '')
      setAppliedQuery(s.appliedQuery ?? '')
      setExpanded(null)
    },
    []
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
      .datasets()
      .then(async (names) => {
        const list = await Promise.all(
          names.map(async (d) => ({ dataset: d, dataTypes: await svc.dataTypes(d).catch(() => []) })),
        )
        if (cancelled) return
        setSources(list)
        const ps = list.find((l) => l.dataset === (initial.dataset ?? DEFAULT_DATASET))?.dataTypes ?? []
        const state = location.state as {
          relatedLogs?: RelatedLogsSeed
          socaiFilters?: FilterType[]
          socaiTime?: string
        } | null
        const seed = state?.relatedLogs
        if (seed?.ids?.length && !seededRef.current) {
          seededRef.current = true
          setPattern(ps.find((p) => p === seed.dataType) ?? null)
          setFilters([{ field: 'id', operator: 'IS_ONE_OF_TERMS', value: seed.ids }])
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
          setPattern(null)
          setFilters(state.socaiFilters)
          if (state.socaiTime) setRange(presetRange(state.socaiTime))
          navigate(location.pathname, { replace: true })
          setReadyToPersist(true)
          return
        }
        // Default: resolve the tab's saved data type against the live ones.
        const target = initial.patternStr
        // A saved tab keeps its data type; anything else starts on all of them.
        setPattern(target ? (ps.find((p) => p === target) ?? null) : null)
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

  /* Field list for the sidebar follows the selected dataset. */
  useEffect(() => {
    let cancelled = false
    svc
      .fields(dataset)
      .then((f) => {
        if (cancelled) return
        setFields(f ?? [])
        setFieldsFor(dataset)
      })
      .catch(() => {
        if (cancelled) return
        setFields([])
        setFieldsFor(dataset)
      })
    return () => {
      cancelled = true
    }
  }, [dataset])

  /* Persist any change to the tab's config (after the first patterns load). */
  useEffect(() => {
    if (!readyToPersist) return
    onConfigChange({
      dataset,
      patternStr: pattern ?? null,
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
  // The store searches one column, `raw`; a dataset without it cannot be
  // searched by text at all.
  const supportsTextSearch = useMemo(
    () => fieldsFor === dataset && fields.some((f) => f.name === 'raw'),
    [fieldsFor, dataset, fields],
  )

  const buildFilters = useCallback((): FilterType[] => {
    const out: FilterType[] = []
    const abs = resolveRange(range)
    if (abs.from) out.push({ field: TS, operator: 'IS_BETWEEN', value: [abs.from, abs.to] })
    // Free text reads the record as it arrived, which only some datasets keep.
    // Sending it at a dataset that has no such column is an error, so the box
    // is disabled there rather than the text being dropped on the floor.
    if (appliedQuery.trim() && supportsTextSearch) {
      out.push({ field: 'raw', operator: 'IS_IN_FIELDS', value: appliedQuery.trim() })
    }
    out.push(...filters)
    return out
  }, [range.from, range.to, appliedQuery, supportsTextSearch, filters])

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

  // The CSV is the table: the analyst's own columns when they picked any, the
  // default source/important/message layout when they didn't. Exporting a fixed
  // pair of fields instead was how a download of one populated column happened.
  const exportColumns = useMemo<ExportColumn[]>(() => {
    const field = (f: string) => ({
      label: f,
      value: (flat: Record<string, unknown>) => (flat[f] == null ? '' : String(flat[f])),
    })
    const time = { ...field(TS), label: 'Time' }
    if (columns.length > 0) return [time, ...columns.map(field)]
    return [
      time,
      { label: 'Source', value: (flat: Record<string, unknown>) => pick(flat, SRC_FIELDS) ?? '' },
      ...autoColumns.map(field),
      // Same fallback the row uses: most normalized logs carry no message
      // field, and the cell shows a summary of the record instead.
      {
        label: 'Message',
        value: (flat: Record<string, unknown>) => pick(flat, MSG_FIELDS) ?? previewText(flat),
      },
    ]
  }, [columns, autoColumns])

  /* Fetch one page. page 1 replaces the list (fresh query); later pages append
     (infinite scroll). The histogram fetches separately. */
  const fetchPage = useCallback(
    async (pageNum: number) => {
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
            dataset,
            dataType: pattern,
            filters: buildFilters(),
            // The view counts pages from 1; the endpoint from 0.
            page: pageNum - 1,
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
    [dataset, pattern, sqlMode, appliedSql, buildFilters]
  )

  // Fresh load whenever the query inputs change → reset to page 1 (replace).
  useEffect(() => {
    setExpanded(null)
    void fetchPage(1)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dataset, pattern, range.from, range.to, appliedQuery, appliedSql, sqlMode, filters, nonce])

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
      return
    }

    // A statement typed into the search box runs as one, and the toggle moves
    // so the mode is visible rather than inferred: the box behaves differently
    // from here on, and a person should be able to see why.
    if (looksLikeSql(searchInput)) {
      setSqlInput(searchInput)
      setAppliedSql(searchInput)
      setSearchInput('')
      setSqlMode(true)
      return
    }

    if (appliedQuery === searchInput) setNonce((n) => n + 1)
    else setAppliedQuery(searchInput)
  }

  // Stable identities: memoized children (FieldSidebar/FieldItem/HistogramStrip/
  // ResultRow) skip re-render on SQL keystrokes only if their callback props are
  // reference-stable. Functional setState updaters let these be dep-free.
  const addFilter = useCallback((f: FilterType) => {
    setFilters((cur) =>
      cur.some((c) => c.field === f.field && c.operator === f.operator && c.value === f.value) ? cur : [...cur, f]
    )
  }, [])
  const removeFilter = useCallback((i: number) => {
    setFilters((cur) => cur.filter((_, idx) => idx !== i))
  }, [])
  const toggleExpanded = useCallback((i: number) => {
    setExpanded((prev) => (prev === i ? null : i))
  }, [])

  // Adapters: CustomFilterBar works with CustomFilter; FilterType is the internal type.
  // IS_ONE_OF_TERMS (array value) is excluded — rendered separately as terms chips.
  const simpleFilters = useMemo(() => filters.filter((f) => f.operator !== 'IS_ONE_OF_TERMS'), [filters])
  const termsFilters = useMemo(() => filters.filter((f) => f.operator === 'IS_ONE_OF_TERMS'), [filters])

  const customFilters = useMemo(() => simpleFilters.map(toCustom), [simpleFilters])

  const barFields = useMemo(
    () =>
      fields
        .filter((f) => !f.name.endsWith('.keyword'))
        .sort((a, b) => a.name.localeCompare(b.name))
        .map((f) => ({ field: f.name, label: f.name })),
    [fields]
  )

  const barOperators = useMemo(
    () => BUILDER_OPS.map((o) => ({ ...o, label: t(`logExplorer.ops.${OP_KEY[o.id] ?? o.id}`) })),
    [t]
  )

  const fetchValues = useCallback(
    (field: string) => {
      const fieldDef = fields.find((f) => f.name === field)
      const aggField = fieldDef?.type === 'text' && !field.endsWith('.keyword') ? `${field}.keyword` : field
      return svc.topValues(dataset, pattern, aggField, activeFilterList, 100).then((r) => r.top ?? [])
    },
    [fields, pattern, activeFilterList]
  )

  const barLabels = useMemo(
    () => ({
      add: t('logExplorer.builder.add'),
      clearAll: t('logExplorer.filters.clearAll'),
      filterValues: t('logExplorer.builder.filterValues'),
      loadingValues: t('logExplorer.builder.loadingValues'),
      noValues: t('logExplorer.builder.noValues'),
      pickValue: t('logExplorer.builder.pickValue'),
      empty: t('logExplorer.fields.empty'),
      cancel: t('logExplorer.builder.cancel'),
      addBtn: t('logExplorer.builder.confirm'),
    }),
    [t]
  )

  const onBarAdd = useCallback(
    (cf: CustomFilter) => addFilter(toFilter(cf)),
    [addFilter]
  )

  const onBarUpdate = useCallback(
    (i: number, cf: CustomFilter) => {
      // i is the index within simpleFilters; map back to the full filters array
      const target = simpleFilters[i]
      setFilters((cur) => cur.map((f) => (f === target ? toFilter(cf) : f)))
    },
    [simpleFilters]
  )

  const onBarRemove = useCallback(
    (i: number) => {
      const target = simpleFilters[i]
      setFilters((cur) => cur.filter((f) => f !== target))
    },
    [simpleFilters]
  )

  const onBarClear = useCallback(() => setFilters([]), [])

  return (
    <div className="flex h-full min-h-0 flex-col px-6 pb-4 pt-3">
      <div>
        <QueryBar
          sources={sources}
          dataset={dataset}
          dataType={pattern}
          textSearchable={supportsTextSearch}
          onSelect={(ds: string, dt: string | null) => {
            setDataset(ds)
            setPattern(dt)
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
            (sqlMode
              ? svc.exportSqlCsv(appliedSql.trim())
              : svc.exportCsv({
                  dataset,
                  dataType: pattern,
                  filters: buildFilters(),
                  columns: exportColumns,
                })
            ).catch(() => toast.error(t('logExplorer.toast.exportFailed')))
          }
        />
      </div>

      <div className="mt-3 flex items-center justify-between gap-2">
        <div className="flex flex-wrap items-center gap-2">
          {!sqlMode && pattern && (
            <CustomFilterBar
              filters={customFilters}
              onAdd={onBarAdd}
              onUpdate={onBarUpdate}
              onRemove={onBarRemove}
              onClear={onBarClear}
              fields={barFields}
              operators={barOperators}
              fetchValues={fetchValues}
              labels={barLabels}
            />
          )}
          {termsFilters.map((f, idx) => (
            <span
              key={idx}
              className="inline-flex items-center gap-1.5 rounded-full border border-primary/25 bg-primary/5 py-1 pl-3 pr-1.5 text-xs"
            >
              <span className="font-mono text-muted-foreground">{f.field}</span>
              <span className="text-[11px] text-muted-foreground/70">{t('logExplorer.ops.isOneOf')}</span>
              <span className="font-mono font-medium">
                {Array.isArray(f.value) ? t('logExplorer.related.nLogs', { count: f.value.length }) : String(f.value)}
              </span>
              <button
                onClick={() => removeFilter(filters.indexOf(f))}
                className="flex h-5 w-5 items-center justify-center rounded-full text-muted-foreground hover:bg-foreground/10 hover:text-foreground"
              >
                <X size={12} />
              </button>
            </span>
          ))}
          {!sqlMode && <SavedSearches snapshot={currentSnapshot} onLoad={loadSnapshot} />}
        </div>
        <div className="flex items-center gap-3">
          <span className="whitespace-nowrap text-xs text-muted-foreground">
            {loading ? (
              t('logExplorer.searching')
            ) : (
              <>
                <span className="font-medium text-foreground">{total.toLocaleString()}</span> {t('logExplorer.eventsIn')}{' '}
                <span className="font-mono">{pattern ?? '—'}</span>
              </>
            )}
          </span>
          {!sqlMode && <ViewToggle mode={viewMode} onChange={setViewMode} />}
        </div>
      </div>

      <div className="mt-2 flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card">
        {viewMode === 'chart' && !sqlMode ? (
          <ChartPanel dataset={dataset} pattern={pattern} fields={fields} filters={activeFilterList} />
        ) : (
          <>
            {pattern &&  (
              <HistogramStrip dataset={dataset} pattern={pattern} filters={activeFilterList} range={range} />
            )}
            <div className="flex min-h-0 flex-1">
              <FieldSidebar
                fields={fields}
                dataset={dataset} pattern={pattern}
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
                      ? t('logExplorer.results.dataTypesFailed')
                      : error === 'explorer:failed'
                        ? t('logExplorer.results.searchFailed')
                        : error}
                    <Button variant="outline" size="sm" className="ml-2" onClick={() => setNonce((n) => n + 1)}>
                      {t('logExplorer.results.retry')}
                    </Button>
                  </RowMessage>
                ) : rows.length === 0 ? (
                  <div className="px-6 py-16 text-center text-sm text-muted-foreground">
                    <div>{t('logExplorer.results.none')}</div>
                    {/* A data type still narrows the search even when the filter
                        names a single id, and nothing on the empty state said so:
                        "no alert of this type has that id" looked identical to
                        "that id does not exist". */}
                    <div className="mt-1 text-xs">
                      {t('logExplorer.results.scope', { dataset, dataType: pattern ?? t('logExplorer.query.allDataTypes') })}
                      {pattern && (
                        <button
                          onClick={() => setPattern(null)}
                          className="ml-2 underline underline-offset-2 hover:text-foreground"
                        >
                          {t('logExplorer.results.searchAllDataTypes')}
                        </button>
                      )}
                    </div>
                  </div>
                ) : (
                  <>
                    {rows.map((doc, i) => (
                      <ResultRow
                        key={i}
                        index={i}
                        doc={doc}
                        columns={columns}
                        autoColumns={autoColumns}
                        expanded={expanded === i}
                        onToggle={toggleExpanded}
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
