import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  AlertTriangle,
  CheckCircle2,
  Copy,
  Filter,
  RefreshCw,
  ScrollText,
  ShieldCheck,
  User,
  X,
} from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { useDateFormat } from '@/shared/lib/datetime'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { auditHttpService } from '../services/audit-http.service'
import { humanizeAction } from '../lib'
import type { AuditListQuery, AuditLog } from '../types/audit.types'

const DEFAULT_PAGE_SIZE = 50

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function AuditPage() {
  const { t } = useTranslation()
  const [filters, setFilters] = useState<AuditListQuery>({ page: 1, page_size: DEFAULT_PAGE_SIZE })
  const [data, setData] = useState<AuditLog[]>([])
  const [pageInfo, setPageInfo] = useState<{
    page: number
    total_pages: number
    total_items: number
    has_next: boolean
    has_prev: boolean
  }>({ page: 1, total_pages: 1, total_items: 0, has_next: false, has_prev: false })
  const [loading, setLoading] = useState(true)
  const [selected, setSelected] = useState<AuditLog | null>(null)

  const load = useCallback(
    async (q: AuditListQuery) => {
      setLoading(true)
      try {
        const resp = await auditHttpService.list(q)
        setData((prev) => ((q.page ?? 1) === 1 ? resp.items : [...prev, ...resp.items]))
        setPageInfo({
          page: resp.page_number,
          total_pages: resp.total_pages,
          total_items: resp.total_items,
          has_next: resp.page_number < resp.total_pages,
          has_prev: resp.page_number > 1,
        })
      } catch (err) {
        toast.error(err instanceof Error ? err.message : t('audit.loadError'))
      } finally {
        setLoading(false)
      }
    },
    [t],
  )

  useEffect(() => {
    load(filters)
  }, [load, filters])

  const setFilter = <K extends keyof AuditListQuery>(key: K, value: AuditListQuery[K]) => {
    setFilters((f) => ({ ...f, [key]: value, page: 1 }))
  }

  const clearFilters = () => {
    setFilters((f) => ({ page: 1, page_size: f.page_size ?? DEFAULT_PAGE_SIZE }))
  }

  const hasActiveFilters = useMemo(
    () =>
      Boolean(
        filters.action ||
          filters.status ||
          filters.resource_type ||
          filters.resource_id ||
          filters.user_email ||
          filters.from ||
          filters.to,
      ),
    [filters],
  )

  /* derived stats from currently loaded page */
  const pageStats = useMemo(() => {
    const failures = data.filter((d) => d.status === 'failure').length
    const successes = data.length - failures
    const successRate = data.length === 0 ? 100 : Math.round((successes / data.length) * 1000) / 10
    const actors = new Set<string>()
    data.forEach((d) => {
      if (d.user_email) actors.add(d.user_email)
      else actors.add('__system__')
    })
    const actionCounts = new Map<string, number>()
    data.forEach((d) => {
      const ns = d.action.split('.')[0] || d.action
      actionCounts.set(ns, (actionCounts.get(ns) ?? 0) + 1)
    })
    let topAction = '—'
    let topCount = 0
    actionCounts.forEach((count, key) => {
      if (count > topCount) {
        topCount = count
        topAction = key
      }
    })
    return { failures, successRate, actors: actors.size, topAction, topCount }
  }, [data])

  /* quick filter helpers */
  const last24h = useMemo(() => new Date(Date.now() - 24 * 3_600_000).toISOString(), [])
  const isAuthOnly = filters.action === 'auth.'
  const isFailuresOnly = filters.status === 'failure'
  const isLast24h = !!filters.from && filters.from === last24h

  const toggleAuth = () => setFilter('action', isAuthOnly ? '' : 'auth.')
  const toggleFailures = () => setFilter('status', isFailuresOnly ? '' : 'failure')
  const toggleLast24h = () =>
    setFilters((f) => ({
      ...f,
      from: isLast24h ? undefined : last24h,
      page: 1,
    }))

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <Header loading={loading} total={pageInfo.total_items} onRefresh={() => load(filters)} />

      <div className="mt-6">
        <StatsStrip
          total={pageInfo.total_items}
          failures={pageStats.failures}
          successRate={pageStats.successRate}
          actors={pageStats.actors}
          topAction={pageStats.topAction}
          topCount={pageStats.topCount}
        />
      </div>

      <div className="mt-6">
        <FilterCard
          filters={filters}
          isAuthOnly={isAuthOnly}
          isFailuresOnly={isFailuresOnly}
          isLast24h={isLast24h}
          onToggleAuth={toggleAuth}
          onToggleFailures={toggleFailures}
          onToggleLast24h={toggleLast24h}
          onChange={setFilter}
          onClear={clearFilters}
          hasActive={hasActiveFilters}
        />
      </div>

      <div className="mt-5">
        <TableCard data={data} loading={loading} onSelect={setSelected} />
        {data.length > 0 && (
          <InfiniteScrollSentinel
            onReach={() => setFilters((f) => ({ ...f, page: (f.page ?? 1) + 1 }))}
            hasMore={data.length < pageInfo.total_items}
            loading={loading}
            endLabel={t('common.allLoaded', { count: pageInfo.total_items })}
          />
        )}
      </div>

      {selected && <DetailDrawer log={selected} onClose={() => setSelected(null)} />}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({
  loading,
  total,
  onRefresh,
}: {
  loading: boolean
  total: number
  onRefresh: () => void
}) {
  const { t } = useTranslation()
  return (
    <header className="flex flex-wrap items-center justify-between gap-3">
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <ScrollText size={14} strokeWidth={1.75} />
        <span>
          <span className="font-medium text-foreground">{total.toLocaleString()}</span> {t('audit.title').toLowerCase()}
        </span>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm" onClick={onRefresh} disabled={loading}>
          <RefreshCw size={14} className={cn('mr-2', loading && 'animate-spin')} />
          {t('common.actions.refresh')}
        </Button>
      </div>
    </header>
  )
}

/* ─── Stats strip ──────────────────────────────────────────────────────── */

function StatsStrip({
  total,
  failures,
  successRate,
  actors,
  topAction,
  topCount,
}: {
  total: number
  failures: number
  successRate: number
  actors: number
  topAction: string
  topCount: number
}) {
  const { t } = useTranslation()
  return (
    <section className="rounded-xl border border-border bg-card">
      <div className="grid grid-cols-1 divide-y divide-border sm:grid-cols-4 sm:divide-x sm:divide-y-0">
        <StripStat
          label={t('audit.stats.totalEntries')}
          value={<span className="font-mono">{total.toLocaleString()}</span>}
          sub={t('audit.stats.acrossAllTime')}
        />
        <StripStat
          label={t('audit.stats.successRate')}
          value={
            <span className="inline-flex items-center gap-1.5 text-emerald-500">
              <CheckCircle2 size={16} strokeWidth={2} />
              {successRate}%
            </span>
          }
          sub={t('audit.stats.failuresOnPage', { count: failures })}
        />
        <StripStat
          label={t('audit.stats.topAction')}
          value={<span className="font-mono text-base">{topAction}</span>}
          sub={topCount > 0 ? t('audit.stats.onPage', { count: topCount }) : '—'}
        />
        <StripStat
          label={t('audit.stats.uniqueActors')}
          value={<span className="font-mono">{actors}</span>}
          sub={t('audit.stats.includingSystem')}
        />
      </div>
    </section>
  )
}

function StripStat({
  label,
  value,
  sub,
}: {
  label: string
  value: React.ReactNode
  sub: string
}) {
  return (
    <div className="px-5 py-4">
      <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className="mt-1 text-xl font-semibold">{value}</div>
      <div className="mt-0.5 text-[11px] text-muted-foreground">{sub}</div>
    </div>
  )
}

/* ─── Filter card ──────────────────────────────────────────────────────── */

function FilterCard({
  filters,
  isAuthOnly,
  isFailuresOnly,
  isLast24h,
  onToggleAuth,
  onToggleFailures,
  onToggleLast24h,
  onChange,
  onClear,
  hasActive,
}: {
  filters: AuditListQuery
  isAuthOnly: boolean
  isFailuresOnly: boolean
  isLast24h: boolean
  onToggleAuth: () => void
  onToggleFailures: () => void
  onToggleLast24h: () => void
  onChange: <K extends keyof AuditListQuery>(key: K, value: AuditListQuery[K]) => void
  onClear: () => void
  hasActive: boolean
}) {
  const { t } = useTranslation()
  return (
    <section className="rounded-xl border border-border bg-card p-5">
      <header className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Filter size={14} className="text-muted-foreground" />
          <h2 className="text-sm font-semibold">{t('audit.filters.title')}</h2>
        </div>
        {hasActive && (
          <Button variant="ghost" size="sm" onClick={onClear} className="h-7">
            <X size={13} className="mr-1" /> {t('audit.filters.clear')}
          </Button>
        )}
      </header>

      {/* Quick filter chips */}
      <div className="mb-4 flex flex-wrap gap-2">
        <Chip active={isLast24h} onClick={onToggleLast24h}>
          {t('audit.filters.last24h')}
        </Chip>
        <Chip active={isFailuresOnly} onClick={onToggleFailures}>
          {t('audit.filters.failuresOnly')}
        </Chip>
        <Chip active={isAuthOnly} onClick={onToggleAuth}>
          {t('audit.filters.authEvents')}
        </Chip>
      </div>

      {/* Filter inputs */}
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-6">
        <Field label={t('audit.filters.actionPrefix')} className="sm:col-span-2">
          <Input
            placeholder={t('audit.filters.actionPlaceholder')}
            value={filters.action ?? ''}
            onChange={(e) => onChange('action', e.target.value)}
            className="h-9 font-mono text-xs"
          />
        </Field>
        <Field label={t('audit.filters.status')}>
          <select
            value={filters.status ?? ''}
            onChange={(e) => onChange('status', e.target.value as 'success' | 'failure' | '')}
            className="h-9 w-full rounded-md border border-border bg-background/40 px-2 text-sm focus:bg-card focus:outline-none focus:ring-1 focus:ring-ring"
          >
            <option value="">{t('audit.filters.any')}</option>
            <option value="success">{t('audit.statusValue.success')}</option>
            <option value="failure">{t('audit.statusValue.failure')}</option>
          </select>
        </Field>
        <Field label={t('audit.filters.resourceType')}>
          <Input
            placeholder={t('audit.filters.resourceTypePlaceholder')}
            value={filters.resource_type ?? ''}
            onChange={(e) => onChange('resource_type', e.target.value)}
            className="h-9 text-xs"
          />
        </Field>
        <Field label={t('audit.filters.resourceId')}>
          <Input
            placeholder="#"
            value={filters.resource_id ?? ''}
            onChange={(e) => onChange('resource_id', e.target.value)}
            className="h-9 font-mono text-xs"
          />
        </Field>
        <Field label={t('audit.filters.user')}>
          <Input
            placeholder={t('audit.filters.userPlaceholder')}
            value={filters.user_email ?? ''}
            onChange={(e) => onChange('user_email', e.target.value)}
            className="h-9 text-xs"
          />
        </Field>
      </div>
    </section>
  )
}

function Chip({
  children,
  active,
  onClick,
}: {
  children: React.ReactNode
  active: boolean
  onClick: () => void
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        'inline-flex items-center rounded-full border px-3 py-1 text-xs font-medium transition-colors',
        active
          ? 'border-primary/40 bg-primary/10 text-primary'
          : 'border-border bg-background/40 hover:bg-muted/40',
      )}
    >
      {children}
    </button>
  )
}

function Field({
  label,
  children,
  className,
}: {
  label: string
  children: React.ReactNode
  className?: string
}) {
  return (
    <div className={className}>
      <label className="mb-1.5 block text-[11px] uppercase tracking-wider text-muted-foreground">
        {label}
      </label>
      {children}
    </div>
  )
}

/* ─── Table card ───────────────────────────────────────────────────────── */

function TableCard({
  data,
  loading,
  onSelect,
}: {
  data: AuditLog[]
  loading: boolean
  onSelect: (log: AuditLog) => void
}) {
  const { t } = useTranslation()
  const { formatDateTime } = useDateFormat()
  return (
    <section className="overflow-hidden rounded-xl border border-border bg-card">
      <div className="grid grid-cols-[180px_140px_1fr_180px_90px_140px_60px] gap-3 border-b border-border bg-muted/30 px-4 py-2.5 text-[11px] uppercase tracking-wider text-muted-foreground">
        <div>{t('audit.table.timestamp')}</div>
        <div>{t('audit.table.actor')}</div>
        <div>{t('audit.table.action')}</div>
        <div>{t('audit.table.resource')}</div>
        <div>{t('audit.table.status')}</div>
        <div>{t('audit.table.ip')}</div>
        <div className="text-right" />
      </div>
      {loading && data.length === 0 ? (
        <div className="px-4 py-16 text-center text-sm text-muted-foreground">
          {t('audit.table.loading')}
        </div>
      ) : data.length === 0 ? (
        <div className="px-4 py-16 text-center">
          <ShieldCheck size={28} strokeWidth={1.5} className="mx-auto mb-3 text-muted-foreground/60" />
          <div className="text-sm font-medium">{t('audit.table.emptyTitle')}</div>
          <div className="mt-0.5 text-xs text-muted-foreground">{t('audit.table.emptyHint')}</div>
        </div>
      ) : (
        data.map((log) => (
          <button
            key={log.id}
            onClick={() => onSelect(log)}
            className="grid w-full grid-cols-[180px_140px_1fr_180px_90px_140px_60px] gap-3 border-b border-border px-4 py-2.5 text-left text-xs last:border-b-0 transition-colors hover:bg-muted/40"
          >
            <div className="font-mono text-[11px] text-muted-foreground">
              {formatDateTime(log.timestamp)}
            </div>
            <div className="truncate">
              {log.user_email ? (
                <span className="font-medium">{log.user_email}</span>
              ) : (
                <span className="italic text-muted-foreground">{t('audit.table.system')}</span>
              )}
            </div>
            <div className="truncate text-[11px]" title={log.action}>
              {humanizeAction(log.action)}
            </div>
            <div className="truncate text-[11px]">
              {log.resource_type ? (
                <>
                  <span className="text-muted-foreground">{log.resource_type}</span>
                  {log.resource_id && <span className="font-mono"> #{log.resource_id}</span>}
                </>
              ) : (
                <span className="text-muted-foreground">—</span>
              )}
            </div>
            <div>
              <StatusPill status={log.status} />
            </div>
            <div className="truncate font-mono text-[11px] text-muted-foreground">
              {log.ip || '—'}
            </div>
            <div className="text-right text-[11px] text-primary">{t('audit.table.view')}</div>
          </button>
        ))
      )}
    </section>
  )
}

function StatusPill({ status }: { status: 'success' | 'failure' }) {
  const { t } = useTranslation()
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-[10px] font-medium ring-1 ring-inset',
        status === 'success'
          ? 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300'
          : 'bg-red-500/15 text-red-500 ring-red-500/30 dark:text-red-300',
      )}
    >
      <span
        className={cn(
          'h-1.5 w-1.5 rounded-full',
          status === 'success' ? 'bg-emerald-500' : 'bg-red-500',
        )}
      />
      {status === 'success' ? t('audit.statusValue.success') : t('audit.statusValue.failure')}
    </span>
  )
}


/* ─── Detail drawer ────────────────────────────────────────────────────── */

type DrawerTab = 'overview' | 'raw'

function DetailDrawer({ log, onClose }: { log: AuditLog; onClose: () => void }) {
  const { t } = useTranslation()
  const { formatDateTime } = useDateFormat()
  const [tab, setTab] = useState<DrawerTab>('overview')

  return (
    <div
      className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/50 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="flex w-full max-w-[860px] flex-col overflow-hidden border-l border-border bg-card shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-3 border-b border-border px-6 py-4">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <StatusPill status={log.status} />
              <span className="font-mono text-[11px] text-muted-foreground">
                {t('audit.drawer.entry', { id: log.id })}
              </span>
            </div>
            <div className="mt-2 text-base font-semibold">{humanizeAction(log.action)}</div>
            <div className="mt-0.5 break-all font-mono text-[11px] text-muted-foreground">
              {log.action}
            </div>
            <div className="mt-1 text-[11px] text-muted-foreground">
              {formatDateTime(log.timestamp)}
            </div>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X size={16} />
          </Button>
        </header>

        <div className="border-b border-border px-6">
          <nav className="-mb-px flex gap-5 text-xs">
            <TabButton active={tab === 'overview'} onClick={() => setTab('overview')}>
              <User size={12} className="mr-1.5" />
              {t('audit.drawer.overview')}
            </TabButton>
            <TabButton active={tab === 'raw'} onClick={() => setTab('raw')}>
              <ScrollText size={12} className="mr-1.5" />
              {t('audit.drawer.raw')}
            </TabButton>
          </nav>
        </div>

        <div className="flex-1 overflow-y-auto px-6 py-5 text-sm">
          {tab === 'overview' && <OverviewTab log={log} />}
          {tab === 'raw' && <RawTab log={log} />}
        </div>
      </div>
    </div>
  )
}

function TabButton({
  active,
  onClick,
  children,
}: {
  active: boolean
  onClick: () => void
  children: React.ReactNode
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        'inline-flex items-center border-b-2 px-1 py-3 font-medium transition-colors',
        active
          ? 'border-primary text-foreground'
          : 'border-transparent text-muted-foreground hover:text-foreground',
      )}
    >
      {children}
    </button>
  )
}

function OverviewTab({ log }: { log: AuditLog }) {
  const { t } = useTranslation()
  return (
    <div className="space-y-1">
      {log.error_message && (
        <div className="mb-4 flex items-start gap-2 rounded-md border border-red-500/30 bg-red-500/10 px-3 py-2 text-xs text-red-500 dark:text-red-300">
          <AlertTriangle size={14} className="mt-0.5 shrink-0" />
          <div>
            <div className="font-medium">{t('audit.drawer.error')}</div>
            <div className="mt-0.5 break-words font-mono text-[11px]">{log.error_message}</div>
          </div>
        </div>
      )}

      <DetailRow label={t('audit.drawer.actor')}>
        {log.user_email ? (
          <span className="font-medium">{log.user_email}</span>
        ) : (
          <span className="italic text-muted-foreground">{t('audit.table.system')}</span>
        )}
      </DetailRow>

      {log.resource_type && (
        <DetailRow label={t('audit.drawer.resource')}>
          <span className="text-muted-foreground">{log.resource_type}</span>
          {log.resource_id && <span className="ml-2 font-mono">#{log.resource_id}</span>}
        </DetailRow>
      )}

      <DetailRow label={t('audit.drawer.ip')}>
        <span className="font-mono text-[12px]">{log.ip || '—'}</span>
      </DetailRow>

      <DetailRow label={t('audit.drawer.userAgent')}>
        <span className="break-all text-[11px] text-muted-foreground">{log.user_agent || '—'}</span>
      </DetailRow>

      {log.session_id && (
        <DetailRow label={t('audit.drawer.session')}>
          <span className="font-mono text-[12px]">#{log.session_id}</span>
        </DetailRow>
      )}

      {log.metadata && Object.keys(log.metadata).length > 0 && (
        <DetailRow label={t('audit.drawer.metadata')}>
          <pre className="mt-1 max-h-[280px] overflow-auto rounded-md border border-border bg-muted/30 p-2.5 text-[11px] text-foreground/90">
            {JSON.stringify(log.metadata, null, 2)}
          </pre>
        </DetailRow>
      )}
    </div>
  )
}

function RawTab({ log }: { log: AuditLog }) {
  const { t } = useTranslation()
  const json = JSON.stringify(log, null, 2)
  return (
    <div>
      <div className="mb-2 flex items-center justify-end">
        <CopyButton text={json} label={t('audit.drawer.copyJson')} />
      </div>
      <pre className="overflow-auto rounded-md border border-border bg-muted/30 p-3 text-[11px] leading-relaxed">
        {json}
      </pre>
    </div>
  )
}

function CopyButton({ text, label }: { text: string; label?: string }) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)
  const onCopy = () => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 1200)
    })
  }
  return (
    <Button variant="ghost" size="sm" onClick={onCopy} className="h-6 gap-1 px-2 text-[11px]">
      {copied ? (
        <>
          <CheckCircle2 size={12} className="text-emerald-500" />
          {t('audit.drawer.copied')}
        </>
      ) : (
        <>
          <Copy size={12} />
          {label ?? t('audit.drawer.copy')}
        </>
      )}
    </Button>
  )
}

function DetailRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="grid grid-cols-[120px_1fr] gap-3 border-b border-border/60 py-2.5 last:border-b-0">
      <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className="min-w-0">{children}</div>
    </div>
  )
}
