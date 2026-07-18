import { useCallback, useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, ChevronDown, ChevronUp, FileCode, FlaskConical, Loader2, Lock, Plus, RefreshCw, Search } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { filtersHttpService } from '@/features/data-processing/services/data-processing-http.service'
import type { Filter } from '@/features/data-processing/types/data-processing.types'
import { TestPlaygroundModal } from '@/features/playground/components/TestPlaygroundModal'
import { FilterFormDrawer } from '../components/FilterFormDrawer'
import { displayName } from '../lib/filter-model'

type Tab = 'all' | 'active' | 'inactive' | 'system' | 'user'
const TABS: Tab[] = ['all', 'active', 'inactive', 'system', 'user']

const COLS = 'minmax(180px,1fr) minmax(160px,1.3fr) 100px 80px 60px 64px'

export function ParsingFiltersPage() {
  const { t } = useTranslation()
  const [tab, setTab] = useState<Tab>('all')
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  // Seed from ?dataType on first render so the initial fetch is already filtered
  // (avoids a race where the unfiltered request resolves last and wins).
  const [dataType, setDataType] = useState(() => new URLSearchParams(window.location.search).get('dataType') ?? '')
  const [dataTypeOptions, setDataTypeOptions] = useState<string[]>([])
  const [items, setItems] = useState<Filter[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0) // 0-based
  const [pageSize] = useState(50)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [editing, setEditing] = useState<{ filter: Filter; creating: boolean } | null>(null)
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
    filtersHttpService
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
    filtersHttpService
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

  const toggleActive = async (f: Filter) => {
    const next = !f.active
    setItems((list) => list.map((x) => (x.relPath === f.relPath ? { ...x, active: next } : x)))
    try {
      await filtersHttpService.activate(f.relPath, next)
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
      const r = await filtersHttpService.list({ page: 1, size: 1000 })
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
    const current = items[index]
    const other = items[otherIndex]
    setReordering(true)
    setItems((list) => {
      const next = [...list]
      next[index] = { ...other, order: current.order }
      next[otherIndex] = { ...current, order: other.order }
      return next
    })
    try {
      await Promise.all([
        filtersHttpService.setOrder(current.relPath, other.order),
        filtersHttpService.setOrder(other.relPath, current.order),
      ])
    } catch {
      setItems((list) => {
        const next = [...list]
        next[index] = current
        next[otherIndex] = other
        return next
      })
      toast.error(t('parsingFilters.toast.orderError'))
    } finally {
      setReordering(false)
    }
  }

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
        <div
          className="grid items-center gap-3 border-b border-border bg-muted/30 px-4 py-2.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground"
          style={{ gridTemplateColumns: COLS }}
        >
          <div>{t('parsingFilters.cols.filter')}</div>
          <div>{t('parsingFilters.cols.dataTypes')}</div>
          <div>{t('parsingFilters.cols.type')}</div>
          <div className="text-center">{t('parsingFilters.cols.active')}</div>
          <div />
          <div />
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto">
          {loading && items.length === 0 ? (
            <Center>
              <Loader2 className="h-4 w-4 animate-spin" /> {t('parsingFilters.loading')}
            </Center>
          ) : error ? (
            <Center>
              <AlertTriangle size={16} className="text-amber-500" /> {t('parsingFilters.loadError')}
              <Button variant="outline" size="sm" className="ml-2" onClick={load}>
                {t('parsingFilters.retry')}
              </Button>
            </Center>
          ) : items.length === 0 ? (
            <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('parsingFilters.empty')}</div>
          ) : (
            <>
              {items.map((f, i) => (
                <Row
                  key={f.relPath}
                  f={f}
                  onOpen={() => setEditing({ filter: f, creating: false })}
                  onToggle={() => toggleActive(f)}
                  onMoveUp={() => moveOrder(i, -1)}
                  onMoveDown={() => moveOrder(i, 1)}
                  canMoveUp={i > 0}
                  canMoveDown={i < items.length - 1}
                  reordering={reordering}
                />
              ))}
              <InfiniteScrollSentinel
                onReach={() => setPage((p) => p + 1)}
                hasMore={items.length < total}
                loading={loading}
                endLabel={t('common.allLoaded', { count: total })}
              />
            </>
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
        <TestPlaygroundModal mode="filter" dataTypeOptions={dataTypeOptions} onClose={() => setShowTestModal(false)} />
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

function Row({
  f,
  onOpen,
  onToggle,
  onMoveUp,
  onMoveDown,
  canMoveUp,
  canMoveDown,
  reordering,
}: {
  f: Filter
  onOpen: () => void
  onToggle: () => void
  onMoveUp: () => void
  onMoveDown: () => void
  canMoveUp: boolean
  canMoveDown: boolean
  reordering: boolean
}) {
  const { t } = useTranslation()
  return (
    <div
      className="grid cursor-pointer items-center gap-3 border-b border-border px-4 py-2.5 text-sm transition-colors last:border-0 hover:bg-muted/40"
      style={{ gridTemplateColumns: COLS }}
      onClick={onOpen}
    >
      <div className="flex min-w-0 items-center gap-2" title={f.relPath}>
        <span
          className="inline-flex h-5 min-w-5 shrink-0 items-center justify-center rounded bg-muted px-1 font-mono text-[10px] text-muted-foreground"
          title={t('parsingFilters.cols.order')}
        >
          {f.order}
        </span>
        <FileCode size={14} className="shrink-0 text-muted-foreground" />
        <span className="truncate text-[13px]">{displayName(f.relPath)}</span>
      </div>
      <DataTypeCells dataTypes={f.dataTypes} />
      <div>
        <span
          className={cn(
            'inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] font-medium',
            f.system ? 'bg-violet-500/15 text-violet-500' : 'bg-sky-500/15 text-sky-500',
          )}
        >
          {f.system && <Lock size={9} />}
          {t(f.system ? 'parsingFilters.system' : 'parsingFilters.user')}
        </span>
      </div>
      <div className="flex justify-center" onClick={(e) => e.stopPropagation()}>
        <Toggle checked={f.active} onChange={onToggle} />
      </div>
      <div className="text-right text-[11px] text-muted-foreground">{t('parsingFilters.view')}</div>
      <div className="flex items-center justify-center gap-0.5" onClick={(e) => e.stopPropagation()}>
        <button
          type="button"
          onClick={onMoveUp}
          disabled={!canMoveUp || reordering}
          title={t('parsingFilters.cols.moveUp')}
          className="flex h-5 w-5 items-center justify-center rounded text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:pointer-events-none disabled:opacity-30"
        >
          <ChevronUp size={13} />
        </button>
        <button
          type="button"
          onClick={onMoveDown}
          disabled={!canMoveDown || reordering}
          title={t('parsingFilters.cols.moveDown')}
          className="flex h-5 w-5 items-center justify-center rounded text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:pointer-events-none disabled:opacity-30"
        >
          <ChevronDown size={13} />
        </button>
      </div>
    </div>
  )
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
