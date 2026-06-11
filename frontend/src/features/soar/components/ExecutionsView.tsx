import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, CheckCircle2, Clock, Loader2, RefreshCw, Search, XCircle } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { Pagination } from '@/shared/components/ui/pagination'
import { useDateFormat } from '@/shared/lib/datetime'
import { soarExecutionsService } from '../services/soar-executions.service'
import type { Execution, ExecutionStatus } from '../types/soar.types'

const STATUSES: (ExecutionStatus | 'all')[] = ['all', 'EXECUTED', 'PENDING', 'FAILED']
const COLS = '90px minmax(160px,1.2fr) minmax(180px,1.6fr) 120px 150px 60px'

const STATUS_META: Record<ExecutionStatus, { icon: typeof CheckCircle2; cls: string }> = {
  EXECUTED: { icon: CheckCircle2, cls: 'text-emerald-500' },
  PENDING: { icon: Clock, cls: 'text-amber-500' },
  FAILED: { icon: XCircle, cls: 'text-red-500' },
}

export function ExecutionsView() {
  const { t } = useTranslation()
  const df = useDateFormat()
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [status, setStatus] = useState<ExecutionStatus | 'all'>('all')
  const [items, setItems] = useState<Execution[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(20)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  useEffect(() => {
    const h = setTimeout(() => {
      setDebounced(search.trim())
      setPage(0)
    }, 300)
    return () => clearTimeout(h)
  }, [search])

  const query = useMemo(
    () => ({
      alertId: debounced || undefined,
      executionStatus: status === 'all' ? undefined : status,
      page,
      size: pageSize,
    }),
    [debounced, status, page, pageSize],
  )

  const load = useCallback(() => {
    setLoading(true)
    setError(false)
    soarExecutionsService
      .list(query)
      .then((r) => {
        setItems(r.data ?? [])
        setTotal(r.total ?? 0)
      })
      .catch(() => setError(true))
      .finally(() => setLoading(false))
  }, [query])
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
            items.map((e) => <ExecutionRow key={e.id} e={e} df={df} t={t} />)
          )}
        </div>
        {total > 0 && (
          <div className="shrink-0 border-t border-border px-3 py-2">
            <Pagination page={page} pageSize={pageSize} total={total} onPageChange={setPage} onPageSizeChange={(s) => { setPageSize(s); setPage(0) }} />
          </div>
        )}
      </div>
    </div>
  )
}

function ExecutionRow({ e, df, t }: { e: Execution; df: ReturnType<typeof useDateFormat>; t: ReturnType<typeof useTranslation>['t'] }) {
  const meta = STATUS_META[e.executionStatus]
  const Icon = meta?.icon ?? Clock
  const flowName = (e.rulePath.split('/').pop() ?? e.rulePath).replace(/\.ya?ml$/i, '')
  return (
    <div className="grid items-center gap-3 border-b border-border px-4 py-2.5 text-sm last:border-0" style={{ gridTemplateColumns: COLS }}>
      <div className={cn('inline-flex items-center gap-1.5 text-[11px] font-medium', meta?.cls)}>
        <Icon size={13} /> {t(`soar.executionStatus.${e.executionStatus}`)}
      </div>
      <div className="min-w-0">
        <div className="truncate text-[13px]" title={e.rulePath}>{flowName}</div>
        {e.alertId && <div className="truncate font-mono text-[10px] text-muted-foreground" title={e.alertId}>{e.alertId}</div>}
      </div>
      <div className="min-w-0">
        <div className="truncate font-mono text-[11px]" title={e.command}>{e.command || '—'}</div>
        {e.nonExecutionCause && <div className="text-[10px] text-red-500">{t(`soar.nonExecutionCause.${e.nonExecutionCause}`)}</div>}
      </div>
      <div className="truncate font-mono text-[11px] text-muted-foreground" title={e.agent}>{e.agent || '—'}</div>
      <div className="text-[11px] text-muted-foreground">{df.formatDateTime(e.executionDate)}</div>
      <div className="text-center text-[11px] text-muted-foreground">{e.executionRetries || 0}</div>
    </div>
  )
}

function Center({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">{children}</div>
}
