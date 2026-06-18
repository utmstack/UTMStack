import { ShieldAlert } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { SEV_META } from '../lib/alert-meta'
import { SEVERITY_INT, STATUS_INT, type SeverityKey, type StatusKey } from '../types/alert.types'

export function AlertsBreakdownCard({
  sevCounts,
  statusCounts,
}: {
  sevCounts: Record<string, number>
  statusCounts: Record<string, number>
}) {
  const { t } = useTranslation()
  const sevRows = (['high', 'medium', 'low'] as SeverityKey[]).map((k) => ({
    k,
    label: t(`alerts.severity.${k}`),
    bar: SEV_META[k].bar,
    count: sevCounts[String(SEVERITY_INT[k])] ?? 0,
  }))
  const stRows = (['open', 'in_review', 'completed', 'auto'] as StatusKey[]).map((k) => ({
    k,
    label: t(`alerts.status.${k}`),
    count: statusCounts[String(STATUS_INT[k])] ?? 0,
  }))
  const sevMax = Math.max(1, ...sevRows.map((r) => r.count))
  return (
    <div className="flex h-full flex-col gap-4 rounded-xl border border-border bg-card p-5">
      <div>
        <div className="mb-2 flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
          <ShieldAlert size={13} /> {t('alerts.overview.bySeverity')}
        </div>
        <div className="space-y-2">
          {sevRows.map((r) => (
            <div key={r.k} className="flex items-center gap-2 text-xs">
              <span className="w-14 text-muted-foreground">{r.label}</span>
              <div className="h-2 flex-1 overflow-hidden rounded-full bg-muted">
                <div
                  className={cn('h-full rounded-full', r.bar)}
                  style={{ width: `${Math.max(2, (r.count / sevMax) * 100)}%` }}
                />
              </div>
              <span className="w-12 text-right font-mono tabular-nums">{r.count.toLocaleString()}</span>
            </div>
          ))}
        </div>
      </div>
      <div className="border-t border-border pt-3">
        <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">{t('alerts.overview.byStatus')}</div>
        <div className="grid grid-cols-2 gap-2">
          {stRows.map((r) => (
            <div key={r.k} className="rounded-md border border-border bg-background/40 px-2.5 py-1.5">
              <div className="truncate text-[10px] text-muted-foreground">{r.label}</div>
              <div className="font-semibold tabular-nums">{r.count.toLocaleString()}</div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
