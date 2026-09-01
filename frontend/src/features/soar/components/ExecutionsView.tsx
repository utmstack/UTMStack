import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, CheckCircle2, Clock, Loader2, RefreshCw, Search, XCircle } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { presetRange, resolveRange, TimeRangePicker, type TimeRange } from '@/shared/components/ui/time-range-picker'
import { useDateFormat } from '@/shared/lib/datetime'
import { datasourcesHttpService } from '@/features/datasources/services/datasources-http.service'
import { soarExecutionsService } from '../services/soar-executions.service'
import type { Execution, ExecutionOrigin, ExecutionStatus, ExecutionListQuery } from '../types/soar.types'

const STATUSES: (ExecutionStatus | 'all')[] = ['all', 'EXECUTED', 'PENDING', 'WAITING', 'EXECUTING', 'FAILED', 'DEAD']
const ORIGINS: (ExecutionOrigin | 'all')[] = ['all', 'FLOW', 'MANUAL']
const COLS = '90px minmax(160px,1.2fr) minmax(180px,1.6fr) 120px 150px 60px'

const STATUS_META: Record<ExecutionStatus, { icon: typeof CheckCircle2; cls: string }> = {
  EXECUTED: { icon: CheckCircle2, cls: 'text-emerald-500' },
  PENDING: { icon: Clock, cls: 'text-amber-500' },
  WAITING: { icon: Clock, cls: 'text-muted-foreground' },
  EXECUTING: { icon: Loader2, cls: 'text-sky-500 [&_svg]:animate-spin' },
  FAILED: { icon: XCircle, cls: 'text-red-500' },
  DEAD: { icon: AlertTriangle, cls: 'text-muted-foreground' },
}

export function ExecutionsView() {
  const { t } = useTranslation()
  const df = useDateFormat()
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [status, setStatus] = useState<ExecutionStatus | 'all'>('all')
  const [origin, setOrigin] = useState<ExecutionOrigin | 'all'>('FLOW')
  const [agent, setAgent] = useState<string>('')
  const [agents, setAgents] = useState<string[]>([])
  const [range, setRange] = useState<TimeRange>(presetRange('7d'))
  const [items, setItems] = useState<Execution[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0)
  const [pageSize] = useState(50)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  useEffect(() => {
    const h = setTimeout(() => {
      setDebounced(search.trim())
      setPage(0)
    }, 300)
    return () => clearTimeout(h)
  }, [search])

  // Same source as FlowEditor / InteractiveConsole — Execution.agent stores the datasource name.
  useEffect(() => {
    datasourcesHttpService
      .list({ page: 1, size: 1000, kind: 'agent', sort: 'asset_name.asc' })
      .then((r) => setAgents((r.items ?? []).map((d) => d.name).filter(Boolean)))
      .catch(() => {})
  }, [])

  const query = useMemo<ExecutionListQuery>(() => {
    const { from, to } = resolveRange(range)
    return {
      alertId: debounced || undefined,
      status: status === 'all' ? undefined : status,
      origin: origin === 'all' ? undefined : origin,
      agent: agent || undefined,
      startedAtFrom: from ?? undefined,
      startedAtTo: to,
      page,
      size: pageSize,
    }
  }, [debounced, status, origin, agent, range, page, pageSize])

  const load = useCallback(() => {
    setLoading(true)
    setError(false)
    soarExecutionsService
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

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="mb-3 flex shrink-0 flex-wrap items-center gap-2">
        <div className="relative">
          <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input value={search} onChange={(e) => setSearch(e.target.value)} placeholder={t('soar.executions.search')} className="w-[260px] pl-8" />
        </div>
        <div className="inline-flex rounded-md border border-border p-0.5">
          {STATUSES.map((s) => (
            <button
              key={s}
              onClick={() => {
                setStatus(s)
                setPage(0)
              }}
              className={cn('rounded px-2.5 py-1 text-xs transition-colors', status === s ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground')}
            >
              {s === 'all' ? t('soar.executions.all') : t(`soar.executionStatus.${s}`)}
            </button>
          ))}
        </div>
        <div className="inline-flex rounded-md border border-border p-0.5">
          {ORIGINS.map((o) => (
            <button
              key={o}
              onClick={() => {
                setOrigin(o)
                setPage(0)
              }}
              className={cn('rounded px-2.5 py-1 text-xs transition-colors', origin === o ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground')}
            >
              {o === 'all' ? t('soar.executions.all') : t(`soar.executionOrigin.${o}`)}
            </button>
          ))}
        </div>
        <select
          value={agent}
          onChange={(e) => {
            setAgent(e.target.value)
            setPage(0)
          }}
          className="h-9 cursor-pointer rounded-md border border-input bg-popover px-2 text-xs text-foreground"
          title={t('soar.executions.filters.source')}
        >
          <option value="">{t('soar.executions.filters.allSources')}</option>
          {agents.map((a) => (
            <option key={a} value={a}>{a}</option>
          ))}
        </select>
        <TimeRangePicker
          value={range}
          onChange={(r) => {
            setRange(r)
            setPage(0)
          }}
          align="right"
        />
        <Button variant="outline" size="sm" onClick={load} disabled={loading} title={t('soar.refresh')}>
          <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
        </Button>
      </div>

      <div className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card">
        <div className="grid items-center gap-3 border-b border-border bg-muted/30 px-4 py-2.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground" style={{ gridTemplateColumns: COLS }}>
          <div>{t('soar.executions.cols.status')}</div>
          <div>{t('soar.executions.cols.flow')}</div>
          <div>{t('soar.executions.cols.command')}</div>
          <div>{t('soar.executions.cols.agent')}</div>
          <div>{t('soar.executions.cols.date')}</div>
          <div className="text-center">{t('soar.executions.cols.retries')}</div>
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto">
          {loading && items.length === 0 ? (
            <Center><Loader2 className="h-4 w-4 animate-spin" /> {t('soar.executions.loading')}</Center>
          ) : error ? (
            <Center>
              <AlertTriangle size={16} className="text-amber-500" /> {t('soar.executions.loadError')}
              <Button variant="outline" size="sm" className="ml-2" onClick={load}>{t('soar.executions.retry')}</Button>
            </Center>
          ) : items.length === 0 ? (
            <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('soar.executions.empty')}</div>
          ) : (
            <>
              {items.map((e) => <ExecutionRow key={e.id} e={e} df={df} t={t} />)}
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
    </div>
  )
}

function ExecutionRow({ e, df, t }: { e: Execution; df: ReturnType<typeof useDateFormat>; t: ReturnType<typeof useTranslation>['t'] }) {
  const meta = STATUS_META[e.status]
  const Icon = meta?.icon ?? Clock
  // A manual run has no flow: what identifies it is who typed it.
  const source =
    e.origin === 'MANUAL'
      ? e.triggeredBy || t('soar.executions.manual')
      : ((e.rulePath ?? '').split('/').pop() ?? '').replace(/\.ya?ml$/i, '') || '—'
  return (
    <div className="grid items-center gap-3 border-b border-border px-4 py-2.5 text-sm last:border-0" style={{ gridTemplateColumns: COLS }}>
      <div className={cn('inline-flex items-center gap-1.5 text-[11px] font-medium', meta?.cls)}>
        <Icon size={13} /> {t(`soar.executionStatus.${e.status}`)}
      </div>
      <div className="min-w-0">
        <div className="truncate text-[13px]" title={e.rulePath ?? e.triggeredBy}>{source}</div>
        {e.alertId && <div className="truncate font-mono text-[10px] text-muted-foreground" title={e.alertId}>{e.alertId}</div>}
      </div>
      <div className="min-w-0">
        <div className="truncate font-mono text-[11px]" title={e.command}>{e.command || '—'}</div>
        {e.nonExecutionCause && <div className="text-[10px] text-red-500">{t(`soar.nonExecutionCause.${e.nonExecutionCause}`)}</div>}
      </div>
      <div className="truncate font-mono text-[11px] text-muted-foreground" title={e.agent}>{e.agent || '—'}</div>
      <div className="text-[11px] text-muted-foreground">{df.formatDateTime(e.startedAt)}</div>
      <div className="text-center text-[11px] text-muted-foreground">{e.retries || 0}</div>
    </div>
  )
}

function Center({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">{children}</div>
}
