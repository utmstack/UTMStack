import { useCallback, useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import { type ColumnDef } from '@tanstack/react-table'
import { AlertTriangle, ChevronDown, ChevronUp, FileCode, FlaskConical, Loader2, Lock, Plus, RefreshCw, Search } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { ResizableDataTable } from '@/shared/components/ui/resizable-data-table'
import { pipelinesHttpService } from '@/features/data-processing/services/data-processing-http.service'
import type { Pipeline } from '@/features/data-processing/types/data-processing.types'
import { TestPlaygroundModal } from '@/features/playground/components/TestPlaygroundModal'
import { PlatformBroadcastButton, broadcast, BULK_PATHS, type BulkSelector } from '@/features/platform-broadcast'
import { FilterFormDrawer } from '../components/FilterFormDrawer'
import { displayName } from '../lib/filter-model'
import { pipelineIdentity, swapInSequence } from '../lib/pipeline-order'

type Tab = 'all' | 'active' | 'inactive' | 'system' | 'user'
const TABS: Tab[] = ['all', 'active', 'inactive', 'system', 'user']

const TH = 'whitespace-nowrap px-3 py-2.5 text-left align-middle font-medium'
const TD = 'whitespace-nowrap px-3 py-2.5 align-middle'
const stopRowClick = (e: React.MouseEvent) => e.stopPropagation()

export function ParsingFiltersPage() {
  const { t } = useTranslation()
  const [tab, setTab] = useState<Tab>('all')
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  // Seed from ?dataType on first render so the initial fetch is already filtered
  // (avoids a race where the unfiltered request resolves last and wins).
  const [dataType, setDataType] = useState(() => new URLSearchParams(window.location.search).get('dataType') ?? '')
  const [dataTypeOptions, setDataTypeOptions] = useState<string[]>([])
  const [items, setItems] = useState<Pipeline[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0) // 0-based
  const [pageSize] = useState(50)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [editing, setEditing] = useState<{ filter: Pipeline; creating: boolean } | null>(null)
  const [preparingNew, setPreparingNew] = useState(false)
  const [showTestModal, setShowTestModal] = useState(false)

  // Deep-link: ?dataType=<value> pre-filters the list to that data type
  // (e.g. opened from an integration's "Filters" button).
  const [searchParams] = useSearchParams()
  useEffect(() => {
    const dt = searchParams.get('dataType')
    if (dt) {
      setDataType(dt)
      setPage(0)
    }
  }, [searchParams])

  useEffect(() => {
    const h = setTimeout(() => {
      setDebounced(search.trim())
      setPage(0)
    }, 300)
    return () => clearTimeout(h)
  }, [search])

  // Distinct dataTypes for the filter dropdown (loaded once).
  useEffect(() => {
    pipelinesHttpService
      .dataTypes()
      .then((d) => setDataTypeOptions(d ?? []))
      .catch(() => setDataTypeOptions([]))
  }, [])

  const query = useMemo(() => {
    const q: {
      relPathContains?: string
      isActive?: boolean
      system?: boolean
      dataType?: string
      page: number
      size: number
    } = {
      relPathContains: debounced || undefined,
      dataType: dataType || undefined,
      page: page + 1, // backend is 1-based
      size: pageSize,
    }
    if (tab === 'active') q.isActive = true
    else if (tab === 'inactive') q.isActive = false
    else if (tab === 'system') q.system = true
    else if (tab === 'user') q.system = false
    return q
  }, [debounced, tab, dataType, page, pageSize])

  const load = useCallback(() => {
    setLoading(true)
    setError(false)
    pipelinesHttpService
      .list(query)
      .then((r) => {
        setItems((prev) => (page === 0 ? (r.data ?? []) : [...prev, ...(r.data ?? [])]))
        setTotal(r.total ?? 0)
      })
      .catch(() => setError(true))
      .finally(() => setLoading(false))
  }, [query])
  useEffect(() => {
    load()
  }, [load])

  const toggleActive = async (f: Pipeline) => {
    const next = !f.active
    setItems((list) => list.map((x) => (x.relPath === f.relPath ? { ...x, active: next } : x)))
    try {
      await pipelinesHttpService.activate(f.relPath, next)
    } catch {
      setItems((list) => list.map((x) => (x.relPath === f.relPath ? { ...x, active: f.active } : x)))
      toast.error(t('parsingFilters.toast.activateError'))
    }
  }

  // New pipelines are appended to the end of the global order — one past
  // whatever the highest order currently in use is.
  const startCreate = async () => {
    setPreparingNew(true)
    let order = 100
    try {
      const r = await pipelinesHttpService.list({ page: 1, size: 1000 })
      const orders = (r.data ?? []).map((f) => f.order ?? 0)
      if (orders.length) order = Math.max(...orders) + 1
    } catch {
      // fall back to the default order band
    } finally {
      setPreparingNew(false)
    }
    setEditing({ filter: { relPath: '', content: '', system: false, active: true, dataTypes: [], order }, creating: true })
  }

  const [reordering, setReordering] = useState(false)

  const moveOrder = async (index: number, direction: -1 | 1) => {
    const otherIndex = index + direction
    if (reordering || otherIndex < 0 || otherIndex >= items.length) return

    const previous = items
    const [a, b] = [items[index], items[otherIndex]]
    const next = [...items]
    next[index] = { ...b, position: a.position }
    next[otherIndex] = { ...a, position: b.position }

    setReordering(true)
    setItems(next)
    try {
      // What is on screen may be a filtered or partial view, but the order is
      // saved as the tenant's whole sequence: swap the two in the full one.
      const all = await pipelinesHttpService.list({ page: 1, size: 1000 })
      const sequence = swapInSequence(
        (all.data ?? []).map((p) => pipelineIdentity(p.relPath)),
        pipelineIdentity(a.relPath),
        pipelineIdentity(b.relPath),
      )
      if (!sequence) throw new Error('pipeline is no longer in the sequence')
      await pipelinesHttpService.setOrder(sequence)
    } catch {
      setItems(previous)
      toast.error(t('parsingFilters.toast.orderError'))
    } finally {
      setReordering(false)
    }
  }

  const onBroadcastDelete = async (f: Pipeline, selector: BulkSelector) => {
    return broadcast(BULK_PATHS.pipelines.delete, selector, { relPath: f.relPath })
  }

  const onBroadcastActivate = async (f: Pipeline, active: boolean, selector: BulkSelector) => {
    return broadcast(BULK_PATHS.pipelines.activate, selector, { relPath: f.relPath, active })
  }

  const columns = useMemo(
    () => buildColumns(t, { items, reordering, moveOrder, toggleActive, onBroadcastDelete, onBroadcastActivate }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [t, items, reordering],
  )

  return (
    <div className="flex h-full min-h-0 w-full flex-col px-6 pb-6 pt-3">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <FileCode size={14} strokeWidth={1.75} />
          <span><span className="font-medium text-foreground">{total}</span> {t('parsingFilters.title').toLowerCase()}</span>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={() => setShowTestModal(true)}>
            <FlaskConical size={14} className="mr-1.5" />
            {t('parsingFilters.testPipeline')}
          </Button>
          <Button size="sm" onClick={() => void startCreate()} disabled={preparingNew}>
            {preparingNew ? <Loader2 size={14} className="mr-1.5 animate-spin" /> : <Plus size={14} className="mr-1.5" />}
            {t('parsingFilters.new')}
          </Button>
        </div>
      </header>

      <div className="mt-4 flex shrink-0 flex-wrap items-center gap-2">
        <div className="relative">
          <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('parsingFilters.search')}
            className="w-[260px] pl-8"
          />
        </div>
        <div className="inline-flex rounded-md border border-border p-0.5">
          {TABS.map((tb) => (
            <button
              key={tb}
              onClick={() => {
                setTab(tb)
                setPage(0)
              }}
              className={cn(
                'rounded px-2.5 py-1 text-xs transition-colors',
                tab === tb ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground',
              )}
            >
              {t(`parsingFilters.tabs.${tb}`)}
            </button>
          ))}
        </div>
        <select
          value={dataType}
          onChange={(e) => {
            setDataType(e.target.value)
            setPage(0)
          }}
          className="h-9 rounded-md border border-input bg-background px-2.5 text-xs text-foreground focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        >
          <option value="">{t('parsingFilters.allDataTypes')}</option>
          {dataTypeOptions.map((dt) => (
            <option key={dt} value={dt}>
              {dt}
            </option>
          ))}
        </select>
        <Button variant="outline" size="sm" onClick={load} disabled={loading} title={t('parsingFilters.refresh')}>
          <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
        </Button>
      </div>

      <div className="mt-3 flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card">
        <div className="min-h-0 flex-1 overflow-auto">
          <ResizableDataTable
            columns={columns}
            data={items}
            flexColumnId="filter"
            storageKey="parsing-filters-table-sizing"
            getRowId={(f) => f.relPath}
            onRowClick={(f) => setEditing({ filter: f, creating: false })}
            loading={loading && items.length === 0}
            loadingContent={
              <Center>
                <Loader2 className="h-4 w-4 animate-spin" /> {t('parsingFilters.loading')}
              </Center>
            }
            error={error}
            errorContent={
              <Center>
                <AlertTriangle size={16} className="text-amber-500" /> {t('parsingFilters.loadError')}
                <Button variant="outline" size="sm" className="ml-2" onClick={load}>
                  {t('parsingFilters.retry')}
                </Button>
              </Center>
            }
            emptyContent={
              <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('parsingFilters.empty')}</div>
            }
          />
          {items.length > 0 && (
            <InfiniteScrollSentinel
              onReach={() => setPage((p) => p + 1)}
              hasMore={items.length < total}
              loading={loading}
              endLabel={t('common.allLoaded', { count: total })}
            />
          )}
        </div>
      </div>

      {editing && (
        <FilterFormDrawer
          filter={editing.filter}
          creating={editing.creating}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null)
            load()
          }}
        />
      )}

      {showTestModal && (
        <TestPlaygroundModal mode="pipeline" titleKey="playground.titleFilter" dataTypeOptions={dataTypeOptions} onClose={() => setShowTestModal(false)} />
      )}
    </div>
  )
}

function DataTypeCells({ dataTypes }: { dataTypes?: string[] }) {
  const dts = dataTypes ?? []
  if (dts.length === 0) return <div className="text-[11px] text-muted-foreground">—</div>
  const shown = dts.slice(0, 3)
  const extra = dts.length - shown.length
  return (
    <div className="flex min-w-0 flex-wrap items-center gap-1" title={dts.join(', ')}>
      {shown.map((dt) => (
        <span key={dt} className="truncate rounded bg-muted px-1.5 py-0.5 font-mono text-[10px] text-foreground/80">
          {dt}
        </span>
      ))}
      {extra > 0 && <span className="text-[10px] text-muted-foreground">+{extra}</span>}
    </div>
  )
}

interface PipelineColumnDeps {
  items: Pipeline[]
  reordering: boolean
  moveOrder: (index: number, direction: -1 | 1) => void
  toggleActive: (f: Pipeline) => void
  onBroadcastDelete: (f: Pipeline, selector: BulkSelector) => Promise<Awaited<ReturnType<typeof broadcast>>>
  onBroadcastActivate: (f: Pipeline, active: boolean, selector: BulkSelector) => Promise<Awaited<ReturnType<typeof broadcast>>>
}

// Column widths/mins/resize-ability live here (as TanStack's size/minSize/
// enableResizing), replacing the old PARSING_FILTERS_TABLE_COLS/_MINS arrays
// — one definition per column instead of three parallel index-aligned arrays.
function buildColumns(
  t: TFunction,
  { items, reordering, moveOrder, toggleActive, onBroadcastDelete, onBroadcastActivate }: PipelineColumnDeps,
): ColumnDef<Pipeline>[] {
  return [
    {
      id: 'filter',
      header: t('parsingFilters.cols.filter'),
      // The flex column (see ResizableDataTable's flexColumnId): `size` is the
      // width it keeps while the rest of the row still has room, `minSize` a
      // floor low enough to truncate a long pipeline name.
      size: 320,
      minSize: 48,
      meta: { headerClassName: TH, cellClassName: TD, cellProps: (f) => ({ title: f.relPath }) },
      cell: ({ row }) => {
        const f = row.original
        return (
          <div className="flex min-w-0 items-center gap-2">
            <span
              className="inline-flex h-5 min-w-5 shrink-0 items-center justify-center rounded bg-muted px-1 font-mono text-[10px] text-muted-foreground"
              title={t('parsingFilters.cols.order')}
            >
              {f.position ?? f.order}
            </span>
            <FileCode size={14} className="shrink-0 text-muted-foreground" />
            <span className="truncate text-[13px]">{displayName(f.relPath)}</span>
          </div>
        )
      },
    },
    {
      id: 'dataTypes',
      header: t('parsingFilters.cols.dataTypes'),
      size: 220,
      minSize: 32,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => <DataTypeCells dataTypes={row.original.dataTypes} />,
    },
    {
      id: 'type',
      header: t('parsingFilters.cols.type'),
      size: 120,
      minSize: 32,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => {
        const f = row.original
        return (
          <span
            className={cn(
              'inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] font-medium',
              f.system ? 'bg-violet-500/15 text-violet-500' : 'bg-sky-500/15 text-sky-500',
            )}
          >
            {f.system && <Lock size={9} />}
            {t(f.system ? 'parsingFilters.system' : 'parsingFilters.user')}
          </span>
        )
      },
    },
    {
      id: 'active',
      header: t('parsingFilters.cols.active'),
      size: 90,
      minSize: 42,
      enableResizing: false,
      meta: { headerClassName: `${TH} text-center`, cellClassName: TD, cellProps: () => ({ onClick: stopRowClick }) },
      cell: ({ row }) => {
        const f = row.original
        return (
          <div className="flex items-center justify-center gap-2">
            <Toggle checked={f.active} onChange={() => toggleActive(f)} />
            <PlatformBroadcastButton
              label={t('platformBroadcast.button')}
              title={t('platformBroadcast.action.activate', { resource: t('platformBroadcast.resource.pipeline') })}
              disabled={false}
              onBroadcast={(selector) => onBroadcastActivate(f, f.active, selector)}
              variant="ghost"
              size="sm"
            />
          </div>
        )
      },
    },
    {
      id: 'actions',
      header: t('parsingFilters.cols.actions'),
      size: 100,
      minSize: 70,
      enableResizing: false,
      meta: { headerClassName: `${TH} text-right`, cellClassName: TD, cellProps: () => ({ onClick: stopRowClick }) },
      cell: ({ row }) => {
        const f = row.original
        return (
          <div className="flex items-center justify-end gap-2">
            <span className="text-right text-[11px] text-muted-foreground">{t('parsingFilters.view')}</span>
            {!f.system && (
              <PlatformBroadcastButton
                label={t('platformBroadcast.button')}
                title={t('platformBroadcast.action.delete', { resource: t('platformBroadcast.resource.pipeline') })}
                disabled={false}
                onBroadcast={(selector) => onBroadcastDelete(f, selector)}
                variant="ghost"
                size="sm"
              />
            )}
          </div>
        )
      },
    },
    {
      id: 'order',
      header: t('parsingFilters.cols.order'),
      size: 80,
      minSize: 72,
      enableResizing: false,
      meta: { headerClassName: `${TH} text-center`, cellClassName: `${TD} text-center`, cellProps: () => ({ onClick: stopRowClick }) },
      cell: ({ row }) => {
        const i = row.index
        return (
          <div className="inline-flex items-center gap-0.5 rounded-md border border-border p-0.5">
            <button
              type="button"
              onClick={() => moveOrder(i, -1)}
              disabled={i <= 0 || reordering}
              title={t('parsingFilters.cols.moveUp')}
              aria-label={t('parsingFilters.cols.moveUp')}
              className="flex h-5 w-5 items-center justify-center rounded text-foreground/70 transition-colors hover:bg-muted hover:text-foreground disabled:pointer-events-none disabled:opacity-30"
            >
              <ChevronUp size={14} strokeWidth={2.25} />
            </button>
            <button
              type="button"
              onClick={() => moveOrder(i, 1)}
              disabled={i >= items.length - 1 || reordering}
              title={t('parsingFilters.cols.moveDown')}
              aria-label={t('parsingFilters.cols.moveDown')}
              className="flex h-5 w-5 items-center justify-center rounded text-foreground/70 transition-colors hover:bg-muted hover:text-foreground disabled:pointer-events-none disabled:opacity-30"
            >
              <ChevronDown size={14} strokeWidth={2.25} />
            </button>
          </div>
        )
      },
    },
  ]
}

function Toggle({ checked, onChange }: { checked: boolean; onChange: () => void }) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      onClick={onChange}
      className={cn(
        'relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors',
        checked ? 'bg-primary' : 'bg-muted-foreground/30',
      )}
    >
      <span className={cn('inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform', checked ? 'translate-x-4' : 'translate-x-0.5')} />
    </button>
  )
}

function Center({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">{children}</div>
}
