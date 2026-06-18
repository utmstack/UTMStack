import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { ST_META, STATUSES } from '../lib/incident-meta'
import type { IncidentStatus } from '../types/incident.types'

export function IncidentsStatCards({
  counts,
  statusFilter,
  onToggle,
}: {
  counts: Record<IncidentStatus, number>
  statusFilter: IncidentStatus | 'all'
  onToggle: (s: IncidentStatus) => void
}) {
  const { t } = useTranslation()
  return (
    <div className="mt-4 grid grid-cols-2 gap-3 lg:grid-cols-4">
      {STATUSES.map((s) => (
        <button
          key={s}
          onClick={() => onToggle(s)}
          className={cn(
            'rounded-xl border bg-card p-4 text-left transition-colors',
            statusFilter === s ? 'border-primary' : 'border-border hover:bg-muted/30'
          )}
        >
          <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
            <span className={cn('h-2 w-2 rounded-full', ST_META[s].dot)} /> {t(`incidents.status.${s}`)}
          </div>
          <div className="mt-1 text-2xl font-semibold tabular-nums">{counts[s]}</div>
        </button>
      ))}
    </div>
  )
}
