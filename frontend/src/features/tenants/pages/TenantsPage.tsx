import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import {
  AlertTriangle,
  Building2,
  Check,
  Globe,
  Loader2,
  Lock,
  LogIn,
  Pencil,
  Plus,
  Power,
  Search,
  ShieldCheck,
  Trash2,
  X,
  type LucideIcon,
} from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { useAuth } from '@/features/auth'
import { setSupportTenant } from '@/shared/lib/current-tenant'
import {
  canReadTenant,
  fetchTenantStats,
  TenantsHttpError,
  tenantsHttpService,
} from '../services/tenants-http.service'
import type {
  CreateTenantRequest,
  SupportAccess,
  Tenant,
  TenantStats,
  TenantStatus,
} from '../types/tenant.types'

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
    <div className="flex items-center gap-2 rounded-lg border border-border bg-card px-3 py-1.5">
      <span className="text-sm font-semibold tabular-nums">{value}</span>
      <span className="text-[11px] text-muted-foreground">{label}</span>
    </div>
  )
}

/* ─────────────────────────────────────────────────────────────────────────
 * Card
 * ───────────────────────────────────────────────────────────────────────── */

/**
 * Enter the tenant, then reload rather than navigate.
 *
 * Everything already fetched — react-query caches, the branding, the
 * notification feed — belongs to the operator's own tenant. A soft navigation
 * would leave that on screen next to the customer's data, which is exactly the
 * confusion a support session must not create.
 */
function enterTenant(tenant: Tenant): void {
  setSupportTenant({
    id: tenant.id,
    name: tenant.name,
    access: tenant.supportAccess === 'FULL' ? 'FULL' : 'READ',
  })
  window.location.assign('/home')
}

// Deterministic hue per tenant, so each card keeps its own accent across loads.
function hueOf(s: string): number {
  let h = 0
  for (let i = 0; i < s.length; i++) h = (h * 31 + s.charCodeAt(i)) % 360
  return h
}

function TenantCard({
  tenant,
  readable,
  onEdit,
  onTerminate,
}: {
  tenant: Tenant
  readable: boolean
  onEdit: () => void
  onTerminate: () => void
}) {
  const { t } = useTranslation()
  const [stats, setStats] = useState<TenantStats | null>(null)
  const [loadingStats, setLoadingStats] = useState(readable)

  useEffect(() => {
    if (!readable) {
      setLoadingStats(false)
      return
    }
    let cancelled = false
    setLoadingStats(true)
    fetchTenantStats(tenant.id)
      .then((s) => {
        if (!cancelled) setStats(s)
      })
      .finally(() => {
        if (!cancelled) setLoadingStats(false)
      })
    return () => {
      cancelled = true
    }
  }, [tenant.id, readable])

  const hue = hueOf(tenant.name || tenant.domain)
  const initial = (tenant.name || tenant.domain).charAt(0).toUpperCase()
  const terminated = tenant.status === 'TERMINATED'

  return (
    <div
      className={cn(
        'group relative flex flex-col overflow-hidden rounded-xl border border-border bg-card shadow-sm transition-all duration-200',
        readable ? 'hover:border-primary/40 hover:shadow-md' : 'opacity-60 saturate-[0.35]',
        terminated && 'opacity-50'
      )}
    >
      <span
        aria-hidden
        className="absolute inset-x-0 top-0 h-1"
        style={{
          background: readable
            ? `linear-gradient(90deg, hsl(${hue} 80% 55%), hsl(${(hue + 40) % 360} 80% 55%))`
            : 'hsl(var(--border))',
        }}
      />

      <div className="flex items-start gap-3 p-5 pb-3">
        <span
          className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl text-base font-semibold text-white shadow-sm ring-1 ring-white/10"
          style={{
            background: readable
              ? `linear-gradient(135deg, hsl(${hue} 72% 52%), hsl(${(hue + 40) % 360} 72% 46%))`
              : 'linear-gradient(135deg, hsl(215 10% 55%), hsl(215 10% 45%))',
          }}
        >
          {initial || <Building2 size={20} />}
        </span>

        <div className="min-w-0 flex-1">
          <h3 className="truncate text-sm font-semibold">{tenant.name}</h3>
          <div className="mt-0.5 flex items-center gap-1 text-[11px] text-muted-foreground">
            <Globe size={10} className="shrink-0" />
            <span className="truncate font-mono">{tenant.domain}</span>
          </div>
        </div>

        <div className="flex shrink-0 items-center gap-1">
          <button
            type="button"
            onClick={onEdit}
            title={t('tenants.card.edit')}
            className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <Pencil size={13} />
          </button>
          {!terminated && (
            <button
              type="button"
              onClick={onTerminate}
              title={t('tenants.card.terminate')}
              className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-red-500/10 hover:text-red-500"
            >
              <Trash2 size={13} />
            </button>
          )}
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-1.5 px-5 pb-3">
        <StatusBadge status={tenant.status} />
        <SupportBadge level={tenant.supportAccess} />
        {readable && !terminated && (
          <button
            type="button"
            onClick={() => enterTenant(tenant)}
            className="ml-auto inline-flex items-center gap-1 rounded-md border border-input px-2 py-0.5 text-[11px] font-medium transition-colors hover:bg-primary/10 hover:text-primary"
          >
            <LogIn size={11} />
            {t('tenants.card.enter')}
          </button>
        )}
      </div>

      <div className="mt-auto border-t border-border bg-muted/20 px-5 py-4">
        {!readable ? (
          // Not a display rule of ours: the tenant chose this, and the backend
          // answers 403 to any read we attempt until they change it.
          <div className="flex items-center gap-2 py-5 text-[11px] text-muted-foreground">
            <Lock size={12} className="shrink-0" />
            {t('tenants.card.noAccess')}
          </div>
        ) : loadingStats ? (
          <div className="flex items-center gap-2 py-5 text-[11px] text-muted-foreground">
            <Loader2 className="h-3 w-3 animate-spin" />
            {t('tenants.card.loadingStats')}
          </div>
        ) : (
          <div className="grid grid-cols-4 gap-1">
            <Ring label={t('tenants.stats.users')} value={stats?.users} tone="sky" />
            <Ring label={t('tenants.stats.datasources')} value={stats?.datasources} tone="violet" />
            <Ring label={t('tenants.stats.openAlerts')} value={stats?.openAlerts} tone="amber" />
            <Ring
              label={t('tenants.stats.ai')}
              value={stats?.ai?.used ?? null}
              tone="emerald"
              // The only one of the four with a real denominator, so the only
              // one whose arc means a proportion rather than "there is a number".
              ratio={
                stats?.ai && stats.ai.limit > 0
                  ? Math.min(1, stats.ai.used / stats.ai.limit)
                  : undefined
              }
              caption={aiHint(stats, t)}
            />
          </div>
        )}
      </div>
    </div>
  )
}

// A limit of 0 or less is the backend saying "no cap of its own": the tenant
// simply spends against whatever the instance licence allows.
function aiHint(stats: TenantStats | null, t: TFunction): string | undefined {
  if (!stats?.ai) return undefined
  return stats.ai.limit > 0 ? `/ ${stats.ai.limit}` : t('tenants.stats.noLimit')
}

const RING_TONE: Record<string, string> = {
  sky: 'stroke-sky-500',
  violet: 'stroke-violet-500',
  amber: 'stroke-amber-500',
  emerald: 'stroke-emerald-500',
}

/**
 * One counter as a dial.
 *
 * `ratio` fills the arc proportionally and is only passed where a real
 * denominator exists. Everywhere else the ring is a frame around the number,
 * not a measurement — an arc drawn from an invented maximum would read as a
 * percentage of nothing.
 */
function Ring({
  label,
  value,
  tone,
  ratio,
  caption,
}: {
  label: string
  value: number | null | undefined
  tone: keyof typeof RING_TONE
  ratio?: number
  caption?: string
}) {
  const missing = value == null
  const circumference = 2 * Math.PI * 22
  const filled = ratio == null ? 1 : ratio

  return (
    <div className="flex min-w-0 flex-col items-center gap-1">
      <div className="relative h-14 w-14">
        <svg viewBox="0 0 52 52" className="h-full w-full -rotate-90">
          <circle cx="26" cy="26" r="22" fill="none" strokeWidth="3" className="stroke-border" />
          {!missing && (
            <circle
              cx="26"
              cy="26"
              r="22"
              fill="none"
              strokeWidth="3"
              strokeLinecap="round"
              className={RING_TONE[tone]}
              strokeDasharray={circumference}
              strokeDashoffset={circumference * (1 - filled)}
            />
          )}
        </svg>
        <span className="absolute inset-0 flex items-center justify-center text-sm font-semibold tabular-nums">
          {/* An unreachable subsystem shows a dash. Zero is a real answer and
              has to look different from "we could not ask". */}
          {value ?? '—'}
        </span>
      </div>
      <span className="w-full truncate text-center text-[10px] leading-tight text-muted-foreground">
        {label}
      </span>
      {caption && (
        <span className="w-full truncate text-center text-[9px] leading-tight text-muted-foreground/70">
          {caption}
        </span>
      )}
    </div>
  )
}

const STATUS_STYLE: Record<TenantStatus, string> = {
  ACTIVE: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300',
  SUSPENDED: 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300',
  TERMINATED: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300',
}

function StatusBadge({ status }: { status: TenantStatus }) {
  const { t } = useTranslation()
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset',
        STATUS_STYLE[status] ?? STATUS_STYLE.TERMINATED
      )}
    >
      <Power size={9} />
      {t(`tenants.status.${status}`, { defaultValue: status })}
    </span>
  )
}

function SupportBadge({ level }: { level: SupportAccess }) {
  const { t } = useTranslation()
  const denied = level === 'NONE'
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset',
        denied
          ? 'bg-muted text-muted-foreground ring-border'
          : 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300'
      )}
      title={t('tenants.support.hint')}
    >
      {denied ? <Lock size={9} /> : <ShieldCheck size={9} />}
      {t(`tenants.support.${level}`, { defaultValue: level })}
    </span>
  )
}

/* ─────────────────────────────────────────────────────────────────────────
 * Dialogs
 * ───────────────────────────────────────────────────────────────────────── */

function Modal({
  title,
  icon: Icon,
  subtitle,
  onClose,
  children,
  footer,
}: {
  title: string
  icon: LucideIcon
  subtitle?: string
  onClose: () => void
  children: React.ReactNode
  footer: React.ReactNode
}) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="flex w-full max-w-lg flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <div>
            <h2 className="flex items-center gap-2 text-lg font-semibold">
              <Icon size={18} />
              {title}
            </h2>
            {subtitle && <p className="mt-1 text-xs text-muted-foreground">{subtitle}</p>}
          </div>
          <button
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>
        <div className="space-y-4 px-6 py-5">{children}</div>
        <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
          {footer}
        </footer>
      </div>
    </div>
  )
}

function Field({
  label,
  hint,
  children,
}: {
  label: string
  hint?: string
  children: React.ReactNode
}) {
  return (
    <div className="space-y-1.5">
      <label className="block text-xs font-medium text-foreground/80">{label}</label>
      {children}
      {hint && <p className="text-[11px] text-muted-foreground">{hint}</p>}
    </div>
  )
}

function CreateTenantDialog({
  onClose,
  onCreated,
}: {
  onClose: () => void
  onCreated: () => void
}) {
  const { t } = useTranslation()
  const [name, setName] = useState('')
  const [domain, setDomain] = useState('')
  const [adminEmail, setAdminEmail] = useState('')
  const [busy, setBusy] = useState(false)

  const valid =
    name.trim().length >= 2 && domain.trim().length >= 3 && /.+@.+\..+/.test(adminEmail.trim())

  const submit = async () => {
    if (!valid || busy) return
    setBusy(true)
    try {
      const body: CreateTenantRequest = {
        name: name.trim(),
        domain: domain.trim().toLowerCase(),
        adminEmail: adminEmail.trim(),
      }
      await tenantsHttpService.create(body)
      toast.success(t('tenants.toast.created'))
      onCreated()
    } catch (err) {
      toast.error(tenantError(err, t))
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal
      title={t('tenants.create.title')}
      subtitle={t('tenants.create.subtitle')}
      icon={Building2}
      onClose={onClose}
      footer={
        <>
          <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
            {t('tenants.cancel')}
          </Button>
          <Button size="sm" disabled={!valid || busy} onClick={() => void submit()}>
            {busy ? t('tenants.saving') : t('tenants.create.submit')}
          </Button>
        </>
      }
    >
      <Field label={t('tenants.fields.name')}>
        <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Acme Corp" />
      </Field>
      <Field label={t('tenants.fields.domain')} hint={t('tenants.fields.domainHint')}>
        <Input
          value={domain}
          onChange={(e) => setDomain(e.target.value)}
          className="font-mono"
          placeholder="acme.utmstack.com"
        />
      </Field>
      <Field label={t('tenants.fields.adminEmail')} hint={t('tenants.fields.adminEmailHint')}>
        <Input
          type="email"
          value={adminEmail}
          onChange={(e) => setAdminEmail(e.target.value)}
          placeholder="admin@acme.com"
        />
      </Field>
    </Modal>
  )
}

const STATUSES: TenantStatus[] = ['ACTIVE', 'SUSPENDED']

function EditTenantDialog({
  tenant,
  onClose,
  onSaved,
}: {
  tenant: Tenant
  onClose: () => void
  onSaved: () => void
}) {
  const { t } = useTranslation()
  const [name, setName] = useState(tenant.name)
  const [domain, setDomain] = useState(tenant.domain)
  const [status, setStatus] = useState<TenantStatus>(tenant.status)
  const [capped, setCapped] = useState(tenant.limits.maxAIRequests != null)
  const [maxAI, setMaxAI] = useState(String(tenant.limits.maxAIRequests ?? ''))
  const [busy, setBusy] = useState(false)

  const parsedMax = Number(maxAI)
  const limitValid = !capped || (maxAI.trim() !== '' && Number.isInteger(parsedMax) && parsedMax >= 0)
  const valid = name.trim().length >= 2 && domain.trim().length >= 3 && limitValid

  const submit = async () => {
    if (!valid || busy) return
    setBusy(true)
    try {
      await tenantsHttpService.update(tenant.id, {
        name: name.trim(),
        domain: domain.trim().toLowerCase(),
        status,
        // Clearing the cap sends an explicit null: omitting the field would
        // mean "leave it as it is", which is the opposite of the intent.
        maxAIRequests: capped ? parsedMax : null,
      })
      toast.success(t('tenants.toast.updated'))
      onSaved()
    } catch (err) {
      toast.error(tenantError(err, t))
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal
      title={t('tenants.edit.title')}
      subtitle={t('tenants.edit.subtitle')}
      icon={Pencil}
      onClose={onClose}
      footer={
        <>
          <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
            {t('tenants.cancel')}
          </Button>
          <Button size="sm" disabled={!valid || busy} onClick={() => void submit()}>
            {busy ? t('tenants.saving') : t('tenants.edit.submit')}
          </Button>
        </>
      }
    >
      <Field label={t('tenants.fields.name')}>
        <Input value={name} onChange={(e) => setName(e.target.value)} />
      </Field>
      <Field label={t('tenants.fields.domain')} hint={t('tenants.fields.domainHint')}>
        <Input value={domain} onChange={(e) => setDomain(e.target.value)} className="font-mono" />
      </Field>

      <Field label={t('tenants.fields.status')}>
        <div className="flex gap-2">
          {STATUSES.map((s) => (
            <button
              key={s}
              type="button"
              onClick={() => setStatus(s)}
              className={cn(
                'flex-1 rounded-md border px-3 py-2 text-xs font-medium transition-colors',
                status === s ? 'border-primary/40 bg-primary/5 text-foreground' : 'border-border text-muted-foreground hover:bg-muted/40'
              )}
            >
              {t(`tenants.status.${s}`, { defaultValue: s })}
            </button>
          ))}
        </div>
        {status === 'SUSPENDED' && (
          <p className="text-[11px] text-amber-600 dark:text-amber-300">
            {t('tenants.fields.suspendedWarning')}
          </p>
        )}
      </Field>

      <Field label={t('tenants.fields.aiLimit')} hint={t('tenants.fields.aiLimitHint')}>
        <button
          type="button"
          onClick={() => setCapped((c) => !c)}
          className="flex w-full items-center gap-2 rounded-md border border-border px-3 py-2 text-left text-xs hover:bg-muted/40"
        >
          <span
            className={cn(
              'flex h-4 w-4 shrink-0 items-center justify-center rounded border',
              capped ? 'border-primary bg-primary text-primary-foreground' : 'border-input'
            )}
          >
            {capped && <Check size={11} strokeWidth={3} />}
          </span>
          {t('tenants.fields.aiLimitToggle')}
        </button>
        {capped && (
          <Input
            type="number"
            min={0}
            value={maxAI}
            onChange={(e) => setMaxAI(e.target.value)}
            placeholder="1000"
            className="mt-2"
          />
        )}
      </Field>
    </Modal>
  )
}

function TerminateDialog({
  tenant,
  onClose,
  onDone,
}: {
  tenant: Tenant
  onClose: () => void
  onDone: () => void
}) {
  const { t } = useTranslation()
  const [confirmation, setConfirmation] = useState('')
  const [busy, setBusy] = useState(false)

  // Terminating takes a whole customer offline, so it asks for the name rather
  // than a single click that a mis-aimed cursor could produce.
  const valid = confirmation.trim() === tenant.name

  const submit = async () => {
    if (!valid || busy) return
    setBusy(true)
    try {
      await tenantsHttpService.terminate(tenant.id)
      toast.success(t('tenants.toast.terminated'))
      onDone()
    } catch (err) {
      toast.error(tenantError(err, t))
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal
      title={t('tenants.terminate.title')}
      icon={Trash2}
      onClose={onClose}
      footer={
        <>
          <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
            {t('tenants.cancel')}
          </Button>
          <Button variant="destructive" size="sm" disabled={!valid || busy} onClick={() => void submit()}>
            {busy ? t('tenants.saving') : t('tenants.terminate.submit')}
          </Button>
        </>
      }
    >
      <p className="text-sm text-muted-foreground">
        {t('tenants.terminate.body', { name: tenant.name })}
      </p>
      <Field label={t('tenants.terminate.confirmLabel', { name: tenant.name })}>
        <Input
          value={confirmation}
          onChange={(e) => setConfirmation(e.target.value)}
          placeholder={tenant.name}
        />
      </Field>
    </Modal>
  )
}

function tenantError(err: unknown, t: TFunction): string {
  if (err instanceof TenantsHttpError) {
    if (err.status === 409) return t('tenants.toast.domainInUse')
    if (err.status === 403) return t('tenants.toast.noPermission')
    if (err.status === 404) return t('tenants.toast.notFound')
    if (err.status === 400) return err.message || t('tenants.toast.invalidRequest')
  }
  return err instanceof Error ? err.message : t('tenants.toast.operationFailed')
}
