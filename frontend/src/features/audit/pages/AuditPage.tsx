import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  AlertTriangle,
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  Copy,
  Filter,
  Hash,
  RefreshCw,
  ScrollText,
  ShieldCheck,
  User,
  X,
} from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { auditHttpService } from '../services/audit-http.service'
import type { AuditListQuery, AuditLog } from '../types/audit.types'

const PAGE_SIZE = 50

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function AuditPage() {
  const [filters, setFilters] = useState<AuditListQuery>({ page: 1, page_size: PAGE_SIZE })
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

  const load = useCallback(async (q: AuditListQuery) => {
    setLoading(true)
    try {
      const resp = await auditHttpService.list(q)
      setData(resp.data)
      setPageInfo({
        page: resp.page_info.page,
        total_pages: resp.page_info.total_pages,
        total_items: resp.page_info.total_items,
        has_next: resp.page_info.has_next,
        has_prev: resp.page_info.has_prev,
      })
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not load audit log')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load(filters)
  }, [load, filters])

  const setFilter = <K extends keyof AuditListQuery>(key: K, value: AuditListQuery[K]) => {
    setFilters((f) => ({ ...f, [key]: value, page: 1 }))
  }

  const clearFilters = () => {
    setFilters({ page: 1, page_size: PAGE_SIZE })
  }

  const hasActiveFilters = useMemo(
    () =>
      Boolean(
        filters.action ||
          filters.status ||
          filters.resource_type ||
          filters.resource_id ||
          filters.user_id ||
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
      if (d.user_login) actors.add(d.user_login)
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

  const toggleAuth = () =>
    setFilter('action', isAuthOnly ? '' : 'auth.')
  const toggleFailures = () =>
    setFilter('status', isFailuresOnly ? '' : 'failure')
  const toggleLast24h = () =>
    setFilters((f) => ({
      ...f,
      from: isLast24h ? undefined : last24h,
      page: 1,
    }))

  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      <Header
        loading={loading}
        total={pageInfo.total_items}
        onRefresh={() => load(filters)}
      />

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
        <TableCard
          data={data}
          loading={loading}
          onSelect={setSelected}
        />
      </div>

      {pageInfo.total_pages > 1 && (
        <div className="mt-3 flex items-center justify-end gap-2 text-xs text-muted-foreground">
          <span>
            Page <span className="font-mono">{pageInfo.page}</span> of{' '}
            <span className="font-mono">{pageInfo.total_pages}</span>
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={!pageInfo.has_prev || loading}
            onClick={() => setFilters((f) => ({ ...f, page: (f.page ?? 1) - 1 }))}
          >
            <ChevronLeft size={14} />
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={!pageInfo.has_next || loading}
            onClick={() => setFilters((f) => ({ ...f, page: (f.page ?? 1) + 1 }))}
          >
            <ChevronRight size={14} />
          </Button>
        </div>
      )}

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
  return (
    <header className="flex items-end justify-between gap-3">
      <div>
        <h1 className="flex items-center gap-2 text-xl font-semibold">
          <ScrollText size={18} strokeWidth={1.75} />
          Audit log
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Tamper-evident record of every security-relevant action.
          {total > 0 && (
            <>
              {' '}
              <span className="font-mono">{total.toLocaleString()}</span> total entries.
            </>
          )}
        </p>
      </div>
      <Button variant="outline" size="sm" onClick={onRefresh} disabled={loading}>
        <RefreshCw size={14} className={cn('mr-2', loading && 'animate-spin')} />
        Refresh
      </Button>
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
  return (
    <section className="rounded-xl border border-border bg-card">
      <div className="grid grid-cols-1 divide-y divide-border sm:grid-cols-4 sm:divide-x sm:divide-y-0">
        <StripStat
          label="Total entries"
          value={<span className="font-mono">{total.toLocaleString()}</span>}
          sub="Across all time"
        />
        <StripStat
          label="Success rate"
          value={
            <span className="inline-flex items-center gap-1.5 text-emerald-500">
              <CheckCircle2 size={16} strokeWidth={2} />
              {successRate}%
            </span>
          }
          sub={`${failures} failure${failures === 1 ? '' : 's'} on this page`}
        />
        <StripStat
          label="Top action"
          value={<span className="font-mono text-base">{topAction}</span>}
          sub={topCount > 0 ? `${topCount} on this page` : '—'}
        />
        <StripStat
          label="Unique actors"
          value={<span className="font-mono">{actors}</span>}
          sub="Including system"
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
  return (
    <section className="rounded-xl border border-border bg-card p-5">
      <header className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Filter size={14} className="text-muted-foreground" />
          <h2 className="text-sm font-semibold">Filters</h2>
        </div>
        {hasActive && (
          <Button variant="ghost" size="sm" onClick={onClear} className="h-7">
            <X size={13} className="mr-1" /> Clear
          </Button>
        )}
      </header>

      {/* Quick filter chips */}
      <div className="mb-4 flex flex-wrap gap-2">
        <Chip active={isLast24h} onClick={onToggleLast24h}>
          Last 24h
        </Chip>
        <Chip active={isFailuresOnly} onClick={onToggleFailures}>
          Failures only
        </Chip>
        <Chip active={isAuthOnly} onClick={onToggleAuth}>
          Auth events
        </Chip>
      </div>

      {/* Filter inputs */}
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-6">
        <Field label="Action prefix" className="sm:col-span-2">
          <Input
            placeholder="e.g. auth., user.update"
            value={filters.action ?? ''}
            onChange={(e) => onChange('action', e.target.value)}
            className="h-9 font-mono text-xs"
          />
        </Field>
        <Field label="Status">
          <select
            value={filters.status ?? ''}
            onChange={(e) => onChange('status', e.target.value as 'success' | 'failure' | '')}
            className="h-9 w-full rounded-md border border-border bg-background/40 px-2 text-sm focus:bg-card focus:outline-none focus:ring-1 focus:ring-ring"
          >
            <option value="">Any</option>
            <option value="success">Success</option>
            <option value="failure">Failure</option>
          </select>
        </Field>
        <Field label="Resource type">
          <Input
            placeholder="user, alert…"
            value={filters.resource_type ?? ''}
            onChange={(e) => onChange('resource_type', e.target.value)}
            className="h-9 text-xs"
          />
        </Field>
        <Field label="Resource id">
          <Input
            placeholder="#"
            value={filters.resource_id ?? ''}
            onChange={(e) => onChange('resource_id', e.target.value)}
            className="h-9 font-mono text-xs"
          />
        </Field>
        <Field label="User id">
          <Input
            type="number"
            placeholder="#"
            value={filters.user_id ?? ''}
            onChange={(e) =>
              onChange('user_id', e.target.value ? Number(e.target.value) : undefined)
            }
            className="h-9 font-mono text-xs"
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
  return (
    <section className="overflow-hidden rounded-xl border border-border bg-card">
      <div className="grid grid-cols-[180px_140px_1fr_180px_90px_140px_60px] gap-3 border-b border-border bg-muted/30 px-4 py-2.5 text-[11px] uppercase tracking-wider text-muted-foreground">
        <div>Timestamp</div>
        <div>Actor</div>
        <div>Action</div>
        <div>Resource</div>
        <div>Status</div>
        <div>IP</div>
        <div className="text-right" />
      </div>
      {loading && data.length === 0 ? (
        <div className="px-4 py-16 text-center text-sm text-muted-foreground">Loading…</div>
      ) : data.length === 0 ? (
        <div className="px-4 py-16 text-center">
          <ShieldCheck size={28} strokeWidth={1.5} className="mx-auto mb-3 text-muted-foreground/60" />
          <div className="text-sm font-medium">No audit entries match these filters.</div>
          <div className="mt-0.5 text-xs text-muted-foreground">
            Try clearing filters or widening the date range.
          </div>
        </div>
      ) : (
        data.map((log) => (
          <button
            key={log.id}
            onClick={() => onSelect(log)}
            className="grid w-full grid-cols-[180px_140px_1fr_180px_90px_140px_60px] gap-3 border-b border-border px-4 py-2.5 text-left text-xs last:border-b-0 transition-colors hover:bg-muted/40"
          >
            <div className="font-mono text-[11px] text-muted-foreground">
              {formatTimestamp(log.timestamp)}
            </div>
            <div className="truncate">
              {log.user_login ? (
                <span className="font-medium">{log.user_login}</span>
              ) : (
                <span className="italic text-muted-foreground">system</span>
              )}
            </div>
            <div className="truncate font-mono text-[11px]">{log.action}</div>
            <div className="truncate text-[11px]">
              {log.resource_type ? (
                <>
                  <span className="text-muted-foreground">{log.resource_type}</span>
                  {log.resource_id && (
                    <span className="font-mono"> #{log.resource_id}</span>
                  )}
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
            <div className="text-right text-[11px] text-primary">View</div>
          </button>
        ))
      )}
    </section>
  )
}

function StatusPill({ status }: { status: 'success' | 'failure' }) {
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
      {status}
    </span>
  )
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleString(undefined, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

/* ─── Detail drawer ────────────────────────────────────────────────────── */

type DrawerTab = 'overview' | 'chain' | 'raw'

function DetailDrawer({ log, onClose }: { log: AuditLog; onClose: () => void }) {
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
                entry #{log.id}
              </span>
            </div>
            <div className="mt-2 break-all font-mono text-base font-semibold">{log.action}</div>
            <div className="mt-1 text-[11px] text-muted-foreground">
              {formatTimestamp(log.timestamp)}
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
              Overview
            </TabButton>
            <TabButton active={tab === 'chain'} onClick={() => setTab('chain')}>
              <Hash size={12} className="mr-1.5" />
              Hash chain
            </TabButton>
            <TabButton active={tab === 'raw'} onClick={() => setTab('raw')}>
              <ScrollText size={12} className="mr-1.5" />
              Raw JSON
            </TabButton>
          </nav>
        </div>

        <div className="flex-1 overflow-y-auto px-6 py-5 text-sm">
          {tab === 'overview' && <OverviewTab log={log} />}
          {tab === 'chain' && <ChainTab log={log} />}
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
  return (
    <div className="space-y-1">
      {log.error_message && (
        <div className="mb-4 flex items-start gap-2 rounded-md border border-red-500/30 bg-red-500/10 px-3 py-2 text-xs text-red-500 dark:text-red-300">
          <AlertTriangle size={14} className="mt-0.5 shrink-0" />
          <div>
            <div className="font-medium">Error</div>
            <div className="mt-0.5 break-words font-mono text-[11px]">{log.error_message}</div>
          </div>
        </div>
      )}

      <DetailRow label="Actor">
        {log.user_login ? (
          <span>
            <span className="font-medium">{log.user_login}</span>
            {log.user_id && (
              <span className="ml-2 font-mono text-[11px] text-muted-foreground">
                #{log.user_id}
              </span>
            )}
          </span>
        ) : (
          <span className="italic text-muted-foreground">system</span>
        )}
      </DetailRow>

      {log.resource_type && (
        <DetailRow label="Resource">
          <span className="text-muted-foreground">{log.resource_type}</span>
          {log.resource_id && (
            <span className="ml-2 font-mono">#{log.resource_id}</span>
          )}
        </DetailRow>
      )}

      <DetailRow label="IP">
        <span className="font-mono text-[12px]">{log.ip || '—'}</span>
      </DetailRow>

      <DetailRow label="User agent">
        <span className="break-all text-[11px] text-muted-foreground">
          {log.user_agent || '—'}
        </span>
      </DetailRow>

      {log.session_id && (
        <DetailRow label="Session">
          <span className="font-mono text-[12px]">#{log.session_id}</span>
        </DetailRow>
      )}

      {log.metadata && Object.keys(log.metadata).length > 0 && (
        <DetailRow label="Metadata">
          <pre className="mt-1 max-h-[280px] overflow-auto rounded-md border border-border bg-muted/30 p-2.5 text-[11px] text-foreground/90">
            {JSON.stringify(log.metadata, null, 2)}
          </pre>
        </DetailRow>
      )}
    </div>
  )
}

function ChainTab({ log }: { log: AuditLog }) {
  return (
    <div className="space-y-4">
      <div className="rounded-md border border-emerald-500/30 bg-emerald-500/10 px-3 py-2.5 text-xs text-emerald-600 dark:text-emerald-300">
        <div className="flex items-center gap-1.5 font-medium">
          <ShieldCheck size={14} />
          Tamper-evident
        </div>
        <p className="mt-1 text-[11px] text-emerald-700/80 dark:text-emerald-300/80">
          Each entry's hash includes the previous entry's hash. Modifying any past entry would
          break the chain on every entry after it.
        </p>
      </div>

      <div>
        <div className="mb-1 flex items-center justify-between">
          <span className="text-[11px] uppercase tracking-wider text-muted-foreground">
            This entry's hash
          </span>
          <CopyButton text={log.hash} />
        </div>
        <code className="block break-all rounded-md border border-border bg-muted/30 p-2.5 font-mono text-[11px]">
          {log.hash}
        </code>
      </div>

      <div>
        <div className="mb-1 flex items-center justify-between">
          <span className="text-[11px] uppercase tracking-wider text-muted-foreground">
            Previous entry's hash
          </span>
          {log.prev_hash && <CopyButton text={log.prev_hash} />}
        </div>
        <code className="block break-all rounded-md border border-border bg-muted/30 p-2.5 font-mono text-[11px] text-muted-foreground">
          {log.prev_hash || <span className="italic">— (genesis entry)</span>}
        </code>
      </div>
    </div>
  )
}

function RawTab({ log }: { log: AuditLog }) {
  const json = JSON.stringify(log, null, 2)
  return (
    <div>
      <div className="mb-2 flex items-center justify-end">
        <CopyButton text={json} label="Copy JSON" />
      </div>
      <pre className="overflow-auto rounded-md border border-border bg-muted/30 p-3 text-[11px] leading-relaxed">
        {json}
      </pre>
    </div>
  )
}

function CopyButton({ text, label }: { text: string; label?: string }) {
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
          Copied
        </>
      ) : (
        <>
          <Copy size={12} />
          {label ?? 'Copy'}
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
