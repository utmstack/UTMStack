import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Loader2, Lock, Plus, RefreshCw, Search, Terminal } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { PlatformBroadcastButton, broadcast, BULK_PATHS } from '@/features/platform-broadcast'
import { soarFlowsService } from '../services/soar-flows.service'
import { soarExecutionsService } from '../services/soar-executions.service'
import { FlowEditor } from '../components/FlowEditor'
import { StartFromModal } from '@/shared/components/StartFromModal'
import { copyOfFlow } from '../lib/duplicate'
import type { Flow } from '../types/soar.types'

type ListTab = 'all' | 'active' | 'inactive' | 'system' | 'user'
const LIST_TABS: ListTab[] = ['all', 'active', 'inactive', 'system', 'user']
const COLS = 'minmax(200px,1fr) 110px 78px 84px 130px 72px 52px'

function flowName(relPath: string): string {
  return (relPath.split('/').pop() ?? relPath).replace(/\.ya?ml$/i, '')
}

// Per-flow runtime telemetry, aggregated client-side from the most recent
// executions (flows themselves carry no run history). Keyed by rulePath = relPath.
interface FlowStat {
  last?: string
  runs: number
  failed: number
}

function relativeTime(iso?: string): string {
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

export function FlowsPage() {
  const { t } = useTranslation()
  const [listTab, setListTab] = useState<ListTab>('all')
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [items, setItems] = useState<Flow[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0)
  const [pageSize] = useState(50)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [editing, setEditing] = useState<{ flow?: Flow; creating: boolean } | null>(null)
  const [starting, setStarting] = useState(false)
  // Every flow, not the page being browsed: what you can copy must not depend
  // on where you had scrolled to.
  const [copyable, setCopyable] = useState<Flow[]>([])
  // Bumped whenever a flow's state changes, so the counters refetch. The total
  // cannot serve as that signal: switching one on does not change how many
  // flows exist.
  const [stateVersion, setStateVersion] = useState(0)
  const [stats, setStats] = useState<Record<string, FlowStat>>({})

  const openStartFrom = () => {
    setStarting(true)
    soarFlowsService
      .list({ page: 0, size: 500 })
      .then((r) => setCopyable(r.data ?? []))
      .catch(() => setCopyable([]))
  }

  const startFromFlow = (relPath: string) => {
    const src = copyable.find((f) => f.relPath === relPath)
    setStarting(false)
    if (src) setEditing({ flow: copyOfFlow(src), creating: true })
  }

  useEffect(() => {
    const h = setTimeout(() => {
      setDebounced(search.trim())
      setPage(0)
    }, 300)
    return () => clearTimeout(h)
  }, [search])

  const query = useMemo(() => {
    const q: { name?: string; active?: boolean; systemOwner?: boolean; page: number; size: number } = {
      name: debounced || undefined,
      page,
      size: pageSize,
    }
    if (listTab === 'active') q.active = true
    else if (listTab === 'inactive') q.active = false
    else if (listTab === 'system') q.systemOwner = true
    else if (listTab === 'user') q.systemOwner = false
    return q
  }, [debounced, listTab, page, pageSize])

  const load = useCallback(() => {
    setLoading(true)
    setError(false)
    soarFlowsService
      .list(query)
      .then((r) => {
        setItems((prev) => (page === 0 ? (r.data ?? []) : [...prev, ...(r.data ?? [])]))
        setTotal(r.total ?? 0)
      })
      .catch(() => setError(true))
      .finally(() => setLoading(false))
  }, [query, page])
  useEffect(() => {
    load()
  }, [load])

  // Aggregate recent executions into per-flow stats (last run / runs / failures).
  // Capped at the most recent 500 executions; refreshed alongside the flows list.
  useEffect(() => {
    let cancelled = false
    soarExecutionsService
      .list({ page: 0, size: 500 })
      .then((r) => {
        if (cancelled) return
        const map: Record<string, FlowStat> = {}
        for (const e of r.data ?? []) {
          if (!e.rulePath) continue // a manual run belongs to no flow
          const s = (map[e.rulePath] ??= { runs: 0, failed: 0 })
          s.runs++
          if (e.status === 'FAILED') s.failed++
          if (!s.last || (e.startedAt && e.startedAt > s.last)) s.last = e.startedAt
        }
        setStats(map)
      })
      .catch(() => {
        if (!cancelled) setStats({})
      })
    return () => {
      cancelled = true
    }
  }, [load])

  const toggleActive = async (f: Flow) => {
    const next = !f.active
    setItems((list) => list.map((x) => (x.relPath === f.relPath ? { ...x, active: next } : x)))
    try {
      await soarFlowsService.setEnabled(f.relPath, next)
      setStateVersion((v) => v + 1)
    } catch {
      setItems((list) => list.map((x) => (x.relPath === f.relPath ? { ...x, active: f.active } : x)))
      toast.error(t('soar.toast.activateError'))
    }
  }

  return (
    <div className="flex h-full min-h-0 w-full flex-col px-6 pb-6 pt-3">
      {!editing &&(
      <>
      <FlowKpis refreshKey={total + stateVersion} />
      <div className="flex shrink-0 flex-wrap items-center gap-2">
        <div className="relative">
          <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input value={search} onChange={(e) => setSearch(e.target.value)} placeholder={t('soar.search')} className="w-[260px] pl-8" />
        </div>
        <div className="inline-flex rounded-md border border-border p-0.5">
          {LIST_TABS.map((tb) => (
            <button
              key={tb}
              onClick={() => {
                setListTab(tb)
                setPage(0)
              }}
              className={cn('rounded px-2.5 py-1 text-xs transition-colors', listTab === tb ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground')}
            >
              {t(`soar.tabs.${tb}`)}
            </button>
          ))}
        </div>
        <Button variant="outline" size="sm" onClick={load} disabled={loading} title={t('soar.refresh')}>
          <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
        </Button>
        <div className="ml-auto">
          <Button size="sm" onClick={openStartFrom}>
            <Plus size={14} className="mr-1.5" /> {t('soar.new')}
          </Button>
        </div>
      </div>

      <div className="mt-3 flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card">
        <div className="grid items-center gap-3 border-b border-border bg-muted/30 px-4 py-2.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground" style={{ gridTemplateColumns: COLS }}>
          <div>{t('soar.cols.flow')}</div>
          <div>{t('soar.cols.platform')}</div>
          <div className="text-center">{t('soar.cols.conditions')}</div>
          <div className="text-center">{t('soar.cols.commands')}</div>
          <div>{t('soar.cols.lastRun')}</div>
          <div className="text-center">{t('soar.cols.active')}</div>
          <div />
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto">
          {loading && items.length === 0 ? (
            <Center><Loader2 className="h-4 w-4 animate-spin" /> {t('soar.loading')}</Center>
          ) : error ? (
            <Center>
              <AlertTriangle size={16} className="text-amber-500" /> {t('soar.loadError')}
              <Button variant="outline" size="sm" className="ml-2" onClick={load}>{t('soar.retry')}</Button>
            </Center>
          ) : items.length === 0 ? (
            <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('soar.empty')}</div>
          ) : (
            <>
              {items.map((f) => (
                <FlowRow key={f.relPath} f={f} stat={stats[f.relPath]} onOpen={() => setEditing({ flow: f, creating: false })} onToggle={() => toggleActive(f)} t={t} />
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

      </>
      )}

      {editing && (
        <FlowEditor
          flow={editing.flow}
          creating={editing.creating}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null)
            load()
          }}
        />
      )}

      {starting && (
        <StartFromModal
          title={t('soar.new')}
          options={copyable.map((f) => ({ id: f.relPath, name: f.name }))}
          onScratch={() => {
            setStarting(false)
            setEditing({ creating: true })
          }}
          onCopy={startFromFlow}
          onClose={() => setStarting(false)}
        />
      )}
    </div>
  )
}

// At-a-glance runtime KPIs. Flows carry no execution telemetry, so these come from
// cheap count queries (size:1, read total) against the flows + executions endpoints.
function FlowKpis({ refreshKey }: { refreshKey: number }) {
  const { t } = useTranslation()
  const [k, setK] = useState<{ flows: number; active: number; runs24h: number; failed24h: number } | null>(null)

  useEffect(() => {
    let cancelled = false
    const since = new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString()
    Promise.allSettled([
      soarFlowsService.list({ page: 0, size: 1 }),
      soarFlowsService.list({ page: 0, size: 1, active: true }),
      soarExecutionsService.list({ page: 0, size: 1, startedAtFrom: since }),
      soarExecutionsService.list({ page: 0, size: 1, status: 'FAILED', startedAtFrom: since }),
    ]).then(([f, a, r, x]) => {
      if (cancelled) return
      const tot = (res: PromiseSettledResult<{ total: number }>) => (res.status === 'fulfilled' ? res.value.total : 0)
      setK({ flows: tot(f), active: tot(a), runs24h: tot(r), failed24h: tot(x) })
    })
    return () => {
      cancelled = true
    }
  }, [refreshKey])

  const cards = [
    { key: 'flows', n: k?.flows, tone: 'text-foreground' },
    { key: 'active', n: k?.active, tone: 'text-emerald-600 dark:text-emerald-400' },
    { key: 'runs24h', n: k?.runs24h, tone: 'text-foreground' },
    { key: 'failed24h', n: k?.failed24h, tone: (k?.failed24h ?? 0) > 0 ? 'text-red-600 dark:text-red-400' : 'text-foreground' },
  ]
  return (
    <div className="mb-3 grid shrink-0 grid-cols-2 gap-3 sm:grid-cols-4">
      {cards.map((c) => (
        <div key={c.key} className="rounded-lg border border-border bg-card px-4 py-3">
          <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{t(`soar.kpi.${c.key}`)}</div>
          <div className={cn('mt-1 text-2xl font-semibold tabular-nums', c.tone)}>
            {c.n == null ? '—' : c.n.toLocaleString()}
          </div>
        </div>
      ))}
    </div>
  )
}

function FlowRow({ f, stat, onOpen, onToggle, t }: { f: Flow; stat?: FlowStat; onOpen: () => void; onToggle: () => void; t: ReturnType<typeof useTranslation>['t'] }) {
  return (
    <div className="grid cursor-pointer items-center gap-3 border-b border-border px-4 py-2.5 text-sm transition-colors last:border-0 hover:bg-muted/40" style={{ gridTemplateColumns: COLS }} onClick={onOpen}>
      <div className="flex min-w-0 items-center gap-2" title={f.relPath}>
        <Terminal size={14} className="shrink-0 text-muted-foreground" />
        <div className="min-w-0">
          <div className="truncate text-[13px]">{f.name || flowName(f.relPath)}</div>
          {f.description && <div className="truncate text-[11px] text-muted-foreground">{f.description}</div>}
        </div>
      </div>
      <div>
        {f.agentPlatform ? (
          <span className="inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5 font-mono text-[10px]">{f.agentPlatform}</span>
        ) : (
          <span className="text-[11px] text-muted-foreground">—</span>
        )}
      </div>
      <div className="text-center font-mono text-[11px] text-muted-foreground">{f.conditions?.length ?? 0}</div>
      <div className="text-center font-mono text-[11px] text-muted-foreground">{f.commands?.length ?? 0}</div>
      <div className="min-w-0 text-[11px]" title={stat?.last ? new Date(stat.last).toLocaleString() : undefined}>
        {stat?.last ? (
          <div className="flex items-center gap-1.5">
            <span className="font-mono text-muted-foreground">{relativeTime(stat.last)}</span>
            {stat.failed > 0 && (
              <span className="rounded bg-red-500/15 px-1 py-0.5 text-[9px] font-semibold text-red-600 dark:text-red-400" title={t('soar.cols.failedRuns', { count: stat.failed })}>
                {stat.failed} ✕
              </span>
            )}
          </div>
        ) : (
          <span className="text-muted-foreground/50">{t('soar.neverRun')}</span>
        )}
      </div>
      <div className="flex justify-center" onClick={(e) => e.stopPropagation()}>
        <Toggle checked={f.active} onChange={onToggle} flow={f} />
      </div>
      <div className="flex items-center justify-end gap-1.5">
        {f.systemOwner && <Lock size={11} className="text-muted-foreground/60" />}
        <span className="text-[11px] text-muted-foreground">{t('soar.view')}</span>
      </div>
    </div>
  )
}

function Toggle({ checked, onChange, flow }: { checked: boolean; onChange: () => void; flow: Flow }) {
  const { t } = useTranslation()
  return (
    <div className="flex items-center gap-1.5" onClick={(e) => e.stopPropagation()}>
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        // The row opens the editor on click. Without stopping here, switching a
        // flow on also opens it — two things from one press, and the panel lands
        // on top of the row you were aiming at.
        onClick={(e) => {
          e.stopPropagation()
          onChange()
        }}
        className={cn('relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors', checked ? 'bg-primary' : 'bg-muted-foreground/30')}
      >
        <span className={cn('inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform', checked ? 'translate-x-4' : 'translate-x-0.5')} />
      </button>
      <PlatformBroadcastButton
        label={t('platformBroadcast.button')}
        title={t('platformBroadcast.action.enable', { resource: t('platformBroadcast.resource.soarFlow') })}
        onBroadcast={(selector) => broadcast(BULK_PATHS.soarRules.enable, selector, { relPath: flow.relPath, enabled: !checked })}
        size="sm"
      />
    </div>
  )
}

function Center({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">{children}</div>
}

