import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  AlertTriangle,
  Building2,
  Loader2,
  Plus,
  Search,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { useAuth } from '@/features/auth'
import {
  canReadTenant,
  tenantsHttpService,
} from '../services/tenants-http.service'
import type { Tenant } from '../types/tenant.types'
import { TenantCard } from '../components/TenantCard'
import { CreateTenantDialog } from '../components/CreateTenantDialog'
import { EditTenantDialog } from '../components/EditTenantDialog'
import { TerminateDialog } from '../components/TerminateDialog'
import { PermanentDeleteDialog } from '../components/PermanentDeleteDialog'

/* ─────────────────────────────────────────────────────────────────────────
 * Page
 * ───────────────────────────────────────────────────────────────────────── */

export function TenantsPage() {
  const { t } = useTranslation()
  const { tenantId: ownTenantId } = useAuth()
  const [tenants, setTenants] = useState<Tenant[] | null>(null)
  const [error, setError] = useState(false)
  const [query, setQuery] = useState('')
  const [creating, setCreating] = useState(false)
  const [editing, setEditing] = useState<Tenant | null>(null)
  const [terminating, setTerminating] = useState<Tenant | null>(null)
  const [deletingPermanently, setDeletingPermanently] = useState<Tenant | null>(null)

  const load = useCallback(async () => {
    setError(false)
    try {
      const list = await tenantsHttpService.list({ size: 200 })
      // The default tenant is the platform plane itself, not a customer: it is
      // where this page is being read from, and it cannot be suspended or
      // terminated from here either.
      setTenants(list.filter((x) => x.id !== ownTenantId))
    } catch {
      setError(true)
      setTenants([])
    }
  }, [ownTenantId])

  useEffect(() => {
    void load()
  }, [load])

  const all = tenants ?? []
  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return all
    return all.filter(
      (x) => x.name.toLowerCase().includes(q) || x.domain.toLowerCase().includes(q)
    )
  }, [all, query])

  const active = all.filter((x) => x.status === 'ACTIVE').length
  const open = all.filter(canReadTenant).length

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <header className="flex items-center gap-2 text-xs text-muted-foreground">
        <Building2 size={14} strokeWidth={1.75} />
        <span className="font-medium text-foreground">{t('tenants.title')}</span>
      </header>
      <p className="mt-1 text-xs text-muted-foreground">{t('tenants.subtitle')}</p>

      {all.length > 0 && (
        <div className="mt-4 flex flex-wrap items-center gap-2">
          <StatChip label={t('tenants.chips.total')} value={all.length} />
          <StatChip label={t('tenants.chips.active')} value={active} />
          <StatChip label={t('tenants.chips.readable')} value={open} />
          <div className="relative ml-auto min-w-[240px] flex-1 sm:max-w-xs">
            <Search
              size={14}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder={t('tenants.searchPlaceholder')}
              className="h-9 pl-9"
            />
          </div>
        </div>
      )}

      {tenants === null && (
        <div className="flex items-center justify-center gap-2 py-24 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
          {t('tenants.loading')}
        </div>
      )}

      {tenants !== null && error && (
        <div className="mt-6 flex flex-col items-center gap-3 rounded-xl border border-border bg-card px-6 py-12 text-sm">
          <span className="inline-flex items-center gap-2 text-muted-foreground">
            <AlertTriangle size={16} className="text-amber-500" />
            {t('tenants.loadFailed')}
          </span>
          <Button variant="outline" size="sm" onClick={() => void load()}>
            {t('tenants.retry')}
          </Button>
        </div>
      )}

      {tenants !== null && !error && filtered.length === 0 && query && (
        <div className="mt-6 rounded-xl border border-border bg-card px-6 py-16 text-center text-sm text-muted-foreground">
          {t('tenants.noSearchMatch')}
        </div>
      )}

      {tenants !== null && !error && (
        <div className="mt-5 grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
          {filtered.map((tenant) => (
            <TenantCard
              key={tenant.id}
              tenant={tenant}
              readable={canReadTenant(tenant)}
              onEdit={() => setEditing(tenant)}
              onTerminate={() => setTerminating(tenant)}
              onDelete={() => setDeletingPermanently(tenant)}
            />
          ))}
          {/* Sits in the grid rather than in the header: adding a customer is
              one of the things this page is for, not a corner action. */}
          {!query && <AddTenantCard onClick={() => setCreating(true)} />}
        </div>
      )}

      {creating && (
        <CreateTenantDialog
          onClose={() => setCreating(false)}
          onCreated={() => {
            setCreating(false)
            void load()
          }}
        />
      )}
      {editing && (
        <EditTenantDialog
          tenant={editing}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null)
            void load()
          }}
        />
      )}
      {terminating && (
        <TerminateDialog
          tenant={terminating}
          onClose={() => setTerminating(null)}
          onDone={() => {
            setTerminating(null)
            void load()
          }}
        />
      )}
      {deletingPermanently && (
        <PermanentDeleteDialog
          tenant={deletingPermanently}
          onClose={() => setDeletingPermanently(null)}
          onDone={() => {
            setDeletingPermanently(null)
            void load()
          }}
        />
      )}
    </div>
  )
}

function AddTenantCard({ onClick }: { onClick: () => void }) {
  const { t } = useTranslation()
  return (
    <button
      onClick={onClick}
      className="group flex min-h-[236px] w-full flex-col items-center justify-center gap-4 rounded-xl border-2 border-dashed border-border bg-gradient-to-b from-primary/[0.04] to-transparent p-5 text-center transition-all hover:border-primary/50 hover:from-primary/10 hover:shadow-md"
    >
      <div className="relative flex h-16 w-16 items-center justify-center rounded-2xl bg-primary/10 ring-1 ring-primary/20 transition-transform group-hover:scale-105">
        <Building2 size={28} className="text-primary" strokeWidth={1.75} />
        <span className="absolute -bottom-1.5 -right-1.5 flex h-6 w-6 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-md ring-2 ring-card">
          <Plus size={14} />
        </span>
      </div>
      <div className="px-2">
        <h6 className="text-sm font-semibold text-foreground">{t('tenants.create.title')}</h6>
        <p className="mt-1.5 text-[11px] leading-relaxed text-muted-foreground">
          {t('tenants.create.subtitle')}
        </p>
      </div>
      <span className="inline-flex items-center gap-1.5 rounded-md bg-primary px-4 py-1.5 text-xs font-semibold text-primary-foreground shadow-sm transition-colors group-hover:bg-primary/90">
        <Plus size={13} /> {t('tenants.new')}
      </span>
    </button>
  )
}

function StatChip({ label, value }: { label: string; value: number }) {
  return (
    <div className={cn('flex items-center gap-2 rounded-lg border border-border bg-card px-3 py-1.5')}>
      <span className="text-sm font-semibold tabular-nums">{value}</span>
      <span className="text-[11px] text-muted-foreground">{label}</span>
    </div>
  )
}
