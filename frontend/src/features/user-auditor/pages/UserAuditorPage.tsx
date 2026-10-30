import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Loader2 } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { ResizableDataTable } from '@/shared/components/ui/resizable-data-table'
import { CustomFilterBar } from '@/shared/components/filters/CustomFilterBar'
import type {
  CustomFilter,
  FilterFieldDef,
  FilterOpDef,
} from '@/shared/components/filters/custom-filter.types'
import { useAuth } from '@/features/auth/services/auth.context'
import { adAuditHttpService } from '../services/ad-audit-http.service'
import type {
  ADUserStats,
  ADUserStatus,
  ADUserWithIdentity,
} from '../types/ad-user.types'
import { type ViewId } from '../helpers/user-auditor-display'
import { buildUserColumns } from '../helpers/user-auditor-columns'
import { UserAuditorHeader } from '../components/user-auditor-header'
import { OverviewCard } from '../components/overview-card'
import { SavedViewTabs } from '../components/saved-view-tabs'
import { Toolbar } from '../components/toolbar'
import { UserCard } from '../components/user-card'
import { UserDrawer } from '../components/user-drawer'

const SIZE = 50

const STATUS_VALUES: ADUserStatus[] = ['active', 'disabled', 'deleted', 'stale', 'service']

function LoadingRows() {
  return (
    <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">
      <Loader2 className="h-4 w-4 animate-spin" />
    </div>
  )
}

export function UserAuditorPage() {
  const { t } = useTranslation()
  const { hasPermission } = useAuth()
  const canManage = hasPermission('adaudit.write')
  const [view, setView] = useState<ViewId>('all')
  const [searchInput, setSearchInput] = useState('')
  const [search, setSearch] = useState('')
  const [tenant, setTenant] = useState('') // '' = all
  const [identity, setIdentity] = useState('') // '' = all, 'none' = unassigned, else alias uuid
  const [layout, setLayout] = useState<'list' | 'cards'>('list')
  const [page, setPage] = useState(0)
  const [customFilters, setCustomFilters] = useState<CustomFilter[]>([])

  const [users, setUsers] = useState<ADUserWithIdentity[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [stats, setStats] = useState<ADUserStats | null>(null)
  const [identities, setIdentities] = useState<{ id: string; name: string }[]>([])
  const [openUser, setOpenUser] = useState<ADUserWithIdentity | null>(null)

  const filterFields: FilterFieldDef[] = [
    { field: 'status', label: t('userAuditor.filterFields.status') },
  ]
  const filterOps: FilterOpDef[] = [{ id: 'IS', label: t('userAuditor.ops.is'), needsValue: true }]
  const filterLabels = {
    add: 'userAuditor.filters.add',
    clearAll: 'userAuditor.filters.clearAll',
    filterValues: 'userAuditor.filters.filterValues',
    loadingValues: 'userAuditor.filters.loadingValues',
    noValues: 'userAuditor.filters.noValues',
    pickValue: 'userAuditor.filters.pickValue',
    empty: 'userAuditor.filters.empty',
    cancel: 'userAuditor.filters.cancel',
    addBtn: 'userAuditor.filters.addBtn',
  }
  const fetchFilterValues = async (field: string) => {
    if (field === 'status') return STATUS_VALUES.map((v) => ({ value: v, count: 0 }))
    return []
  }

  // Debounce the search box.
  useEffect(() => {
    const id = setTimeout(() => setSearch(searchInput.trim()), 300)
    return () => clearTimeout(id)
  }, [searchInput])

  // Any filter change resets to the first page.
  useEffect(() => {
    setPage(0)
  }, [view, search, tenant, identity, customFilters])

  const loadStats = useCallback(async () => {
    try {
      setStats(await adAuditHttpService.stats(tenant || undefined))
    } catch {
      // Overview is best-effort; the list below is the source of truth.
    }
  }, [tenant])

  // Identity aliases are best-effort too: the list works without them.
  const loadIdentities = useCallback(async () => {
    try {
      setIdentities(await adAuditHttpService.identities() ?? [])
    } catch {
      setIdentities([])
    }
  }, [])

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      const statusFilter = customFilters.find((f) => f.field === 'status' && f.operator === 'IS')?.value
      const res = await adAuditHttpService.list({
        search: search || undefined,
        tenantId: tenant || undefined,
        source: view === 'all' ? undefined : view,
        status: statusFilter ? (statusFilter as ADUserStatus) : undefined,
        sort: 'recent',
        identity: identity || undefined,
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
  }, [search, tenant, identity, view, page, customFilters])

  useEffect(() => {
    void load()
  }, [load])

  useEffect(() => {
    void loadStats()
  }, [loadStats])

  useEffect(() => {
    void loadIdentities()
  }, [loadIdentities])

  const refresh = () => {
    if (page === 0) void load()
    else setPage(0)
    void loadStats()
    void loadIdentities()
  }

  const onUserChanged = (u: ADUserWithIdentity) => {
    setOpenUser(u)
    loadIdentities()
    setUsers((prev) => prev.map((x) => (x.id === u.id ? u : x)))
  }

  const counts: Record<ViewId, number> = {
    all: stats?.total ?? 0,
    windows: stats?.by_source?.windows ?? 0,
    linux: stats?.by_source?.linux ?? 0,
    o365: stats?.by_source?.o365 ?? 0,
    azure: stats?.by_source?.azure ?? 0,
    gcp: stats?.by_source?.gcp ?? 0,
    aws: stats?.by_source?.aws ?? 0,
  }

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <UserAuditorHeader total={stats?.total ?? total} t={t} />

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
        identity={identity}
        identities={identities}
        onIdentity={setIdentity}
        layout={layout}
        onLayoutChange={setLayout}
        onRefresh={refresh}
        loading={loading}
        t={t}
      />

      <div className="mt-3">
        <CustomFilterBar
          filters={customFilters}
          onAdd={(f) => setCustomFilters((c) => [...c, f])}
          onUpdate={(i, f) => setCustomFilters((c) => c.map((x, idx) => (idx === i ? f : x)))}
          onRemove={(i) => setCustomFilters((c) => c.filter((_, idx) => idx !== i))}
          onClear={() => setCustomFilters([])}
          fields={filterFields}
          operators={filterOps}
          fetchValues={fetchFilterValues}
          labels={filterLabels}
        />
      </div>

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
        <div className="mt-3 overflow-x-auto overflow-y-hidden rounded-xl border border-border bg-card">
          <ResizableDataTable
            columns={buildUserColumns(t)}
            data={users}
            flexColumnId="account"
            storageKey="user-auditor-table-sizing"
            getRowId={(u) => u.id}
            onRowClick={setOpenUser}
            rowClassName={() => 'text-xs last:border-b-0'}
            loading={loading && users.length === 0}
            loadingContent={<LoadingRows />}
            emptyContent={
              <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('userAuditor.empty')}</div>
            }
          />
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

      {openUser && (
        <UserDrawer
          user={openUser}
          canManage={canManage}
          onClose={() => setOpenUser(null)}
          onUserChanged={onUserChanged}
          t={t}
        />
      )}
    </div>
  )
}
