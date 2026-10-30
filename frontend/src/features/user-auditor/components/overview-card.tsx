import type { TFunction } from 'i18next'
import { Clock, UserMinus, Users } from 'lucide-react'
import type { ADUserStats } from '../types/ad-user.types'

export function OverviewCard({ stats, t }: { stats: ADUserStats | null; t: TFunction }) {
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
        <div className="mt-4 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
          <span className="inline-flex items-center gap-2">
            <Clock size={13} />
            {t('userAuditor.overview.seen24h', { count: stats?.seen_24h ?? 0 })}
          </span>
          <span className="inline-flex items-center gap-2">
            <Users size={13} />
            {t('userAuditor.overview.grouped', { count: stats?.identities ?? 0 })}
          </span>
          <span className="inline-flex items-center gap-2">
            <UserMinus size={13} />
            {t('userAuditor.overview.unassigned', { count: stats?.unassigned ?? 0 })}
          </span>
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
