import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import {
  AlertTriangle,
  ChevronDown,
  Clock,
  Copy,
  Download,
  ExternalLink,
  Fingerprint,
  LayoutGrid,
  ListFilter,
  ListIcon,
  Loader2,
  RefreshCw,
  Search,
  Server,
  UserCog,
  UserPlus,
  UserX,
  X,
  type LucideIcon,
} from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { adAuditHttpService } from '../services/ad-audit-http.service'
import type { ADUser, ADUserStats, ADUserStatus } from '../types/ad-user.types'

const SIZE = 50
const STALE_MS = 30 * 86_400_000

type ViewId = 'all' | ADUserStatus
const VIEW_IDS: ViewId[] = ['all', 'active', 'disabled', 'deleted', 'service', 'stale']

/* ─── Derived (per-row, display-only) ──────────────────────────────────── */

type Status = 'active' | 'disabled' | 'deleted'

function statusOf(u: ADUser): Status {
  if (u.accountDeletedAt) return 'deleted'
  if (!u.active) return 'disabled'
  return 'active'
}
function isStale(u: ADUser): boolean {
  return statusOf(u) === 'active' && !!u.lastLogon && Date.now() - new Date(u.lastLogon).getTime() > STALE_MS
}
function isService(u: ADUser): boolean {
  return /^svc/i.test(u.samAccountName)
}

const STATUS_DOT: Record<Status, string> = {
  active: 'bg-emerald-500',
  disabled: 'bg-amber-500',
  deleted: 'bg-red-500',
}
const STATUS_TEXT: Record<Status, string> = {
  active: 'text-emerald-600 dark:text-emerald-400',
  disabled: 'text-amber-600 dark:text-amber-400',
  deleted: 'text-red-600 dark:text-red-400',
}

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function UserAuditorPage() {
  const { t } = useTranslation()
  const [view, setView] = useState<ViewId>('all')
  const [searchInput, setSearchInput] = useState('')
  const [search, setSearch] = useState('')
  const [tenant, setTenant] = useState('') // '' = all
  const [layout, setLayout] = useState<'list' | 'cards'>('list')
  const [page, setPage] = useState(0)

  const [users, setUsers] = useState<ADUser[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [stats, setStats] = useState<ADUserStats | null>(null)
  const [openUser, setOpenUser] = useState<ADUser | null>(null)

  // Debounce the search box.
  useEffect(() => {
    const id = setTimeout(() => setSearch(searchInput.trim()), 300)
    return () => clearTimeout(id)
  }, [searchInput])

  // Any filter change resets to the first page.
  useEffect(() => {
    setPage(0)
  }, [view, search, tenant])

  const loadStats = useCallback(async () => {
    try {
      setStats(await adAuditHttpService.stats(tenant || undefined))
    } catch {
      // Overview is best-effort; the list below is the source of truth.
    }
  }, [tenant])

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      const res = await adAuditHttpService.list({
        search: search || undefined,
        tenantId: tenant || undefined,
        status: view === 'all' ? undefined : view,
        sort: 'recent',
        page,
        size: SIZE,
      })
      setUsers((prev) => (page === 0 ? (res.data ?? []) : [...prev, ...(res.data ?? [])]))
      setTotal(res.total)
    } catch {
      setError(true)
      setUsers([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [search, tenant, view, page])

  useEffect(() => {
    void load()
  }, [load])
  useEffect(() => {
    void loadStats()
  }, [loadStats])

  const refresh = () => {
    void load()
    void loadStats()
  }

  const counts: Record<ViewId, number> = {
    all: stats?.total ?? 0,
    active: stats?.active ?? 0,
    disabled: stats?.disabled ?? 0,
    deleted: stats?.deleted ?? 0,
    service: stats?.service ?? 0,
    stale: stats?.stale ?? 0,
  }

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <Header total={stats?.total ?? total} t={t} />

      <div className="mt-5">
        <OverviewCard stats={stats} t={t} />
      </div>

      <div className="mt-5">
        <SavedViewTabs current={view} counts={counts} onChange={setView} t={t} />
      </div>

      <Toolbar
        search={searchInput}
        onSearch={setSearchInput}
        tenant={tenant}
        tenants={stats?.tenants ?? []}
        onTenant={setTenant}
        layout={layout}
        onLayoutChange={setLayout}
        onRefresh={refresh}
        loading={loading}
        t={t}
      />

      {error ? (
        <div className="mt-3 flex flex-col items-center gap-3 rounded-xl border border-border bg-card px-6 py-12 text-sm">
          <span className="inline-flex items-center gap-2 text-muted-foreground">
            <AlertTriangle size={16} className="text-amber-500" />
            {t('userAuditor.loadError')}
          </span>
          <Button variant="outline" size="sm" onClick={() => void load()}>
            {t('userAuditor.retry')}
          </Button>
        </div>
      ) : layout === 'list' ? (
        <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
          <ListHeader t={t} />
          {loading && users.length === 0 ? (
            <LoadingRows />
          ) : (
            users.map((u) => <UserListRow key={u.id} user={u} onOpen={() => setOpenUser(u)} t={t} />)
          )}
          {!loading && users.length === 0 && (
            <div className="px-6 py-16 text-center text-sm text-muted-foreground">
              {t('userAuditor.empty')}
            </div>
          )}
        </div>
      ) : (
        <div className="mt-3 grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
          {loading && users.length === 0 ? (
            <div className="col-span-full flex items-center justify-center gap-2 rounded-xl border border-border bg-card px-6 py-16 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" />
            </div>
          ) : (
            users.map((u) => <UserCard key={u.id} user={u} onOpen={() => setOpenUser(u)} t={t} />)
          )}
          {!loading && users.length === 0 && (
            <div className="col-span-full rounded-xl border border-border bg-card px-6 py-16 text-center text-sm text-muted-foreground">
              {t('userAuditor.empty')}
            </div>
          )}
        </div>
      )}

      {!error && users.length > 0 && (
        <InfiniteScrollSentinel
          onReach={() => setPage((p) => p + 1)}
          hasMore={users.length < total}
          loading={loading}
          endLabel={t('common.allLoaded', { count: total })}
        />
      )}

      {openUser && <UserDrawer user={openUser} onClose={() => setOpenUser(null)} t={t} />}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ total, t }: { total: number; t: TFunction }) {
  return (
    <header className="flex flex-wrap items-center justify-between gap-3">
      <div className="text-xs text-muted-foreground">
        <span className="font-medium text-foreground">{t('userAuditor.accountsChip', { count: total })}</span>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="default" size="sm">
          <Download size={14} className="mr-2" />
          {t('userAuditor.export')}
        </Button>
      </div>
    </header>
  )
}

/* ─── Overview ─────────────────────────────────────────────────────────── */

function OverviewCard({ stats, t }: { stats: ADUserStats | null; t: TFunction }) {
  const byDomain = stats?.by_domain ?? []
  const maxDomain = Math.max(...byDomain.map((d) => d.count), 1)
  return (
    <div className="grid grid-cols-1 gap-px overflow-hidden rounded-xl border border-border bg-border md:grid-cols-[1.4fr_1fr]">
      <div className="bg-card p-6">
        <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
          {t('userAuditor.overview.title')}
        </div>
        <div className="mt-1 flex items-baseline gap-3">
          <span className="text-3xl font-semibold tabular-nums">{stats?.total ?? 0}</span>
          <span className="text-sm text-muted-foreground">
            <span className="text-emerald-600 dark:text-emerald-400">{stats?.active ?? 0}</span>{' '}
            {t('userAuditor.overview.active')}
            <span className="mx-2 text-border">·</span>
            <span className="text-amber-600 dark:text-amber-400">{stats?.disabled ?? 0}</span>{' '}
            {t('userAuditor.overview.disabled')}
            <span className="mx-2 text-border">·</span>
            <span className="text-red-600 dark:text-red-400">{stats?.deleted ?? 0}</span>{' '}
            {t('userAuditor.overview.deleted')}
          </span>
        </div>
        <div className="mt-4 flex items-center gap-2 text-xs text-muted-foreground">
          <Clock size={13} />
          {t('userAuditor.overview.seen24h', { count: stats?.seen_24h ?? 0 })}
        </div>
      </div>

      <div className="bg-card p-6">
        <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
          {t('userAuditor.overview.byDomain')}
        </div>
        <ul className="mt-3 space-y-2">
          {byDomain.map((d) => (
            <li key={d.domain} className="flex items-center gap-3 text-xs">
              <span className="w-28 shrink-0 truncate font-mono text-muted-foreground">{d.domain || '—'}</span>
              <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-muted">
                <div className="h-full rounded-full bg-sky-500/70" style={{ width: `${(d.count / maxDomain) * 100}%` }} />
              </div>
              <span className="w-6 text-right font-mono tabular-nums text-muted-foreground">{d.count}</span>
            </li>
          ))}
          {byDomain.length === 0 && <li className="text-xs text-muted-foreground">—</li>}
        </ul>
      </div>
    </div>
  )
}

/* ─── Saved view tabs ──────────────────────────────────────────────────── */

function SavedViewTabs({
  current,
  counts,
  onChange,
  t,
}: {
  current: ViewId
  counts: Record<ViewId, number>
  onChange: (id: ViewId) => void
  t: TFunction
}) {
  return (
    <div className="flex flex-wrap items-center gap-1 border-b border-border">
      {VIEW_IDS.map((id) => {
        const active = current === id
        return (
          <button
            key={id}
            onClick={() => onChange(id)}
            className={cn(
              'group relative flex items-center gap-2 px-3 py-2 text-xs transition-colors',
              active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground',
            )}
          >
            <span>{t(`userAuditor.views.${id}`)}</span>
            <span
              className={cn(
                'rounded-md px-1.5 py-0.5 font-mono text-[10px] tabular-nums',
                active ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground',
              )}
            >
              {counts[id] ?? 0}
            </span>
            {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
          </button>
        )
      })}
    </div>
  )
}

/* ─── Toolbar ──────────────────────────────────────────────────────────── */

function Toolbar({
  search,
  onSearch,
  tenant,
  tenants,
  onTenant,
  layout,
  onLayoutChange,
  onRefresh,
  loading,
  t,
}: {
  search: string
  onSearch: (s: string) => void
  tenant: string
  tenants: string[]
  onTenant: (t: string) => void
  layout: 'list' | 'cards'
  onLayoutChange: (l: 'list' | 'cards') => void
  onRefresh: () => void
  loading: boolean
  t: TFunction
}) {
  return (
    <div className="mt-3 flex flex-wrap items-center gap-2">
      <div className="relative min-w-[280px] flex-1">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder={t('userAuditor.searchPlaceholder')}
          value={search}
          onChange={(e) => onSearch(e.target.value)}
          className="h-9 pl-9"
        />
      </div>
      {tenants.length > 0 && (
        <div className="relative">
          <Server size={13} className="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <select
            value={tenant}
            onChange={(e) => onTenant(e.target.value)}
            className="h-9 appearance-none rounded-md border border-border bg-background pl-8 pr-7 text-xs text-foreground"
          >
            <option value="">{t('userAuditor.allTenants')}</option>
            {tenants.map((tn) => (
              <option key={tn} value={tn}>
                {tn}
              </option>
            ))}
          </select>
          <ChevronDown size={12} className="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 opacity-60" />
        </div>
      )}
      <Button variant="outline" size="sm">
        <ListFilter size={14} className="mr-2" />
        {t('userAuditor.filters')}
      </Button>
      <div className="ml-auto flex items-center gap-1 rounded-md border border-border p-0.5">
        <button
          onClick={() => onLayoutChange('list')}
          title={t('userAuditor.layout.list')}
          className={cn(
            'flex h-7 w-7 items-center justify-center rounded transition-colors',
            layout === 'list' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground',
          )}
        >
          <ListIcon size={13} />
        </button>
        <button
          onClick={() => onLayoutChange('cards')}
          title={t('userAuditor.layout.cards')}
          className={cn(
            'flex h-7 w-7 items-center justify-center rounded transition-colors',
            layout === 'cards' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground',
          )}
        >
          <LayoutGrid size={13} />
        </button>
      </div>
      <button
        title={t('userAuditor.refresh')}
        onClick={onRefresh}
        className="flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
      >
        <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
      </button>
    </div>
  )
}

/* ─── List ─────────────────────────────────────────────────────────────── */

const LIST_COLS = '32px 1fr 1.3fr 110px 110px 110px 90px 36px'

function ListHeader({ t }: { t: TFunction }) {
  return (
    <div
      className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: LIST_COLS }}
    >
      <div />
      <div>{t('userAuditor.list.account')}</div>
      <div>{t('userAuditor.list.sid')}</div>
      <div>{t('userAuditor.list.status')}</div>
      <div>{t('userAuditor.list.lastLogon')}</div>
      <div>{t('userAuditor.list.lastSeen')}</div>
      <div>{t('userAuditor.list.tenant')}</div>
      <div />
    </div>
  )
}

function LoadingRows() {
  return (
    <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">
      <Loader2 className="h-4 w-4 animate-spin" />
    </div>
  )
}

function UserListRow({ user, onOpen, t }: { user: ADUser; onOpen: () => void; t: TFunction }) {
  const status = statusOf(user)
  return (
    <div
      onClick={onOpen}
      className="group grid cursor-pointer items-center gap-3 border-b border-border px-4 py-2.5 text-xs hover:bg-muted/40 last:border-b-0"
      style={{ gridTemplateColumns: LIST_COLS }}
    >
      <AccountAvatar user={user} />
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="truncate text-sm font-medium">{user.samAccountName}</span>
          {isService(user) && (
            <span className="rounded bg-violet-500/15 px-1.5 py-0.5 text-[9px] font-medium uppercase text-violet-600 ring-1 ring-violet-500/30 dark:text-violet-300">
              {t('userAuditor.badge.service')}
            </span>
          )}
          {isStale(user) && (
            <span className="rounded bg-amber-500/15 px-1.5 py-0.5 text-[9px] font-medium uppercase text-amber-600 ring-1 ring-amber-500/30 dark:text-amber-300">
              {t('userAuditor.badge.stale')}
            </span>
          )}
        </div>
        <div className="truncate font-mono text-[11px] text-muted-foreground">{user.domain}</div>
      </div>
      <div className="min-w-0 truncate font-mono text-[11px] text-muted-foreground" title={user.sid}>
        {user.sid}
      </div>
      <div>
        <span className={cn('inline-flex items-center gap-1.5 text-[11px]', STATUS_TEXT[status])}>
          <span className={cn('h-1.5 w-1.5 rounded-full', STATUS_DOT[status])} />
          {t(`userAuditor.status.${status}`)}
        </span>
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">{relativeTime(user.lastLogon, t)}</div>
      <div className="font-mono text-[11px] text-muted-foreground">{relativeTime(user.lastSeen, t)}</div>
      <div className="truncate text-[11px] text-muted-foreground">{user.tenantId}</div>
      <div className="flex justify-end opacity-0 group-hover:opacity-100">
        <ExternalLink size={13} className="text-muted-foreground" />
      </div>
    </div>
  )
}

/* ─── Card ─────────────────────────────────────────────────────────────── */

function UserCard({ user, onOpen, t }: { user: ADUser; onOpen: () => void; t: TFunction }) {
  const status = statusOf(user)
  return (
    <button
      onClick={onOpen}
      className="group flex flex-col gap-3 rounded-xl border border-border bg-card p-4 text-left transition-all hover:shadow-md"
    >
      <div className="flex items-start gap-3">
        <AccountAvatar user={user} large />
        <div className="min-w-0 flex-1">
          <div className="truncate text-sm font-semibold">{user.samAccountName}</div>
          <div className="truncate font-mono text-[11px] text-muted-foreground">{user.domain}</div>
          <div className="mt-1">
            <span className={cn('inline-flex items-center gap-1.5 text-[11px]', STATUS_TEXT[status])}>
              <span className={cn('h-1.5 w-1.5 rounded-full', STATUS_DOT[status])} />
              {t(`userAuditor.status.${status}`)}
            </span>
          </div>
        </div>
      </div>
      <div className="truncate border-t border-border pt-2 font-mono text-[10px] text-muted-foreground" title={user.sid}>
        {user.sid}
      </div>
      <div className="flex items-center justify-between text-[11px] text-muted-foreground">
        <span>
          {t('userAuditor.card.lastLogon')}{' '}
          <span className="font-mono text-foreground">{relativeTime(user.lastLogon, t)}</span>
        </span>
        <span className="font-mono">{user.tenantId}</span>
      </div>
    </button>
  )
}

/* ─── Avatar ───────────────────────────────────────────────────────────── */

function AccountAvatar({ user, large }: { user: ADUser; large?: boolean }) {
  const status = statusOf(user)
  const Icon = status === 'deleted' ? UserX : UserCog
  return (
    <div className="relative">
      <div
        className={cn(
          'flex shrink-0 items-center justify-center rounded-lg bg-muted text-foreground/70 ring-2 ring-border',
          large ? 'h-11 w-11' : 'h-7 w-7',
        )}
      >
        <Icon size={large ? 18 : 13} className={isService(user) ? 'text-violet-500' : 'text-sky-500'} />
      </div>
      <span className={cn('absolute -right-0.5 -top-0.5 h-2.5 w-2.5 rounded-full ring-2 ring-card', STATUS_DOT[status])} />
    </div>
  )
}

/* ─── Drawer ───────────────────────────────────────────────────────────── */

function UserDrawer({ user, onClose, t }: { user: ADUser; onClose: () => void; t: TFunction }) {
  const status = statusOf(user)
  const copySid = () => {
    void navigator.clipboard?.writeText(user.sid)
    toast.success(t('userAuditor.drawer.copied'))
  }
  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[680px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="flex min-w-0 flex-1 items-start gap-3">
              <AccountAvatar user={user} large />
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                  <Server size={11} />
                  <span className="font-mono">{user.domain}</span>
                  <span>·</span>
                  <span>{t('userAuditor.drawer.tenantInline', { id: user.tenantId })}</span>
                </div>
                <h2 className="mt-0.5 truncate text-xl font-semibold">{user.samAccountName}</h2>
                <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                  <span className={cn('inline-flex items-center gap-1.5', STATUS_TEXT[status])}>
                    <span className={cn('h-1.5 w-1.5 rounded-full', STATUS_DOT[status])} />
                    {t(`userAuditor.status.${status}`)}
                  </span>
                  {isService(user) && (
                    <span className="rounded-md bg-violet-500/15 px-1.5 py-0.5 font-medium text-violet-600 ring-1 ring-violet-500/30 dark:text-violet-300">
                      {t('userAuditor.drawer.serviceAccount')}
                    </span>
                  )}
                  {isStale(user) && (
                    <span className="rounded-md bg-amber-500/15 px-1.5 py-0.5 font-medium text-amber-600 ring-1 ring-amber-500/30 dark:text-amber-300">
                      {t('userAuditor.drawer.stale')}
                    </span>
                  )}
                </div>
              </div>
            </div>
            <button
              onClick={onClose}
              className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <X size={16} />
            </button>
          </div>

          <div className="mt-4 flex flex-wrap items-center gap-2">
            <Button size="sm" variant="outline" asChild>
              <Link
                to={`/log-explorer?${new URLSearchParams({
                  dataType: 'wineventlog',
                  tenantId: user.tenantId,
                  eventCode: '4720,4726,4624',
                  eventDataTargetSid: user.sid,
                  '@timestamp': '30d',
                }).toString()}`}
              >
                <ExternalLink size={14} className="mr-1.5" />
                {t('userAuditor.drawer.viewInLogs')}
              </Link>
            </Button>
            <Button size="sm" variant="outline" onClick={copySid}>
              <Copy size={14} className="mr-1.5" />
              {t('userAuditor.drawer.copySid')}
            </Button>
          </div>
        </header>

        <div className="flex-1 space-y-4 overflow-y-auto bg-muted/20 p-6">
          <Section title={t('userAuditor.drawer.identity')}>
            <dl className="grid grid-cols-[150px_1fr] gap-y-2 text-xs">
              <Row k={t('userAuditor.drawer.accountName')}>
                <span className="font-mono">{user.samAccountName}</span>
              </Row>
              <Row k={t('userAuditor.drawer.domain')}>
                <span className="font-mono">{user.domain}</span>
              </Row>
              <Row k={t('userAuditor.drawer.sid')}>
                <span className="break-all font-mono">{user.sid}</span>
              </Row>
              <Row k={t('userAuditor.drawer.tenant')}>
                <span className="font-mono">{user.tenantId}</span>
              </Row>
              <Row k={t('userAuditor.drawer.adStatus')}>
                {user.active ? t('userAuditor.drawer.enabled') : t('userAuditor.drawer.disabled')}
              </Row>
            </dl>
          </Section>

          <Section title={t('userAuditor.drawer.lifecycle')}>
            <Timeline user={user} t={t} />
          </Section>
        </div>
      </div>
    </div>
  )
}

/* ─── Lifecycle timeline ───────────────────────────────────────────────── */

function Timeline({ user, t }: { user: ADUser; t: TFunction }) {
  const events: { icon: LucideIcon; label: string; ts?: string; tone: string }[] = [
    { icon: UserPlus, label: t('userAuditor.drawer.created'), ts: user.accountCreatedAt, tone: 'text-sky-500' },
    { icon: Fingerprint, label: t('userAuditor.drawer.lastLogon'), ts: user.lastLogon, tone: 'text-emerald-500' },
    { icon: Clock, label: t('userAuditor.drawer.lastSeen'), ts: user.lastSeen, tone: 'text-muted-foreground' },
    { icon: UserX, label: t('userAuditor.drawer.deleted'), ts: user.accountDeletedAt, tone: 'text-red-500' },
  ].filter((e) => e.ts)

  if (events.length === 0) {
    return <div className="text-xs text-muted-foreground">{t('userAuditor.drawer.noLifecycle')}</div>
  }

  return (
    <ol className="relative space-y-4 pl-6">
      <span className="absolute bottom-1 left-[9px] top-1 w-px bg-border" />
      {events.map((e) => (
        <li key={e.label} className="relative">
          <span className="absolute -left-6 flex h-[18px] w-[18px] items-center justify-center rounded-full bg-card ring-1 ring-border">
            <e.icon size={11} className={e.tone} />
          </span>
          <div className="flex items-baseline justify-between gap-3">
            <span className="text-xs font-medium">{e.label}</span>
            <span className="font-mono text-[11px] text-muted-foreground">{relativeTime(e.ts, t)}</span>
          </div>
          <div className="font-mono text-[11px] text-muted-foreground">{absTimestamp(e.ts!)}</div>
        </li>
      ))}
    </ol>
  )
}

/* ─── Small parts ──────────────────────────────────────────────────────── */

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <h4 className="mb-3 text-sm font-semibold">{title}</h4>
      {children}
    </div>
  )
}

function Row({ k, children }: { k: string; children: React.ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="break-words">{children}</dd>
    </>
  )
}

/* ─── Helpers ──────────────────────────────────────────────────────────── */

function relativeTime(iso: string | undefined, t: TFunction): string {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  const m = Math.round(diff / 60_000)
  if (m < 1) return t('userAuditor.relative.justNow')
  if (m < 60) return t('userAuditor.relative.minutesAgo', { count: m })
  const h = Math.round(m / 60)
  if (h < 24) return t('userAuditor.relative.hoursAgo', { count: h })
  return t('userAuditor.relative.daysAgo', { count: Math.round(h / 24) })
}

function absTimestamp(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
