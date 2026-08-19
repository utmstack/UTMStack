import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import {
  Building2,
  Globe,
  Lock,
  Loader2,
  LogIn,
  Pencil,
  Power,
  ShieldCheck,
  Trash2,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { setSupportTenant } from '@/shared/lib/current-tenant'
import { fetchTenantStats } from '../services/tenants-http.service'
import type { SupportAccess, Tenant, TenantStats, TenantStatus } from '../types/tenant.types'

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

// A limit of 0 or less is the backend saying "no cap of its own": the tenant
// simply spends against whatever the instance licence allows.
function aiHint(stats: TenantStats | null, t: TFunction): string | undefined {
  if (!stats?.ai) return undefined
  return stats.ai.limit > 0 ? `/ ${stats.ai.limit}` : t('tenants.stats.noLimit')
}

export function TenantCard({
  tenant,
  readable,
  onEdit,
  onActivate,
  onDeactivate,
  onDelete,
}: {
  tenant: Tenant
  readable: boolean
  onEdit: () => void
  onActivate: () => void
  onDeactivate: () => void
  onDelete: () => void
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
          {tenant.status === 'ACTIVE' ? (
            <button
              type="button"
              onClick={onDeactivate}
              title={t('tenants.card.deactivate')}
              className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-amber-500/10 hover:text-amber-600"
            >
              <Power size={13} />
            </button>
          ) : (
            <button
              type="button"
              onClick={onActivate}
              title={t('tenants.card.activate')}
              className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-emerald-500/10 hover:text-emerald-600"
            >
              <Power size={13} />
            </button>
          )}
          <button
            type="button"
            onClick={onDelete}
            title={t('tenants.card.deletePermanent')}
            className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <Trash2 size={13} />
          </button>
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
