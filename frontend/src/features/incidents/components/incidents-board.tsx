import { Link2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { SEV_TONE, ST_META, STATUSES, sevKey } from '../lib/incident-meta'
import type { Incident } from '../types/incident.types'

export function IncidentsBoard({ incidents, onOpen }: { incidents: Incident[]; onOpen: (i: Incident) => void }) {
  const { t } = useTranslation()
  return (
    <div className="mt-4 grid min-h-0 flex-1 grid-cols-1 gap-3 overflow-y-auto md:grid-cols-2 xl:grid-cols-4">
      {STATUSES.map((s) => {
        const items = incidents.filter((i) => i.incidentStatus === s)
        return (
          <div
            key={s}
            className={cn(
              'flex min-h-0 flex-col rounded-xl border border-t-2 border-border bg-card/50',
              ST_META[s].col
            )}
          >
            <div className="flex items-center justify-between px-3 py-2 text-xs font-medium">
              <span className="flex items-center gap-1.5">
                <span className={cn('h-2 w-2 rounded-full', ST_META[s].dot)} /> {t(`incidents.status.${s}`)}
              </span>
              <span className="font-mono text-muted-foreground">{items.length}</span>
            </div>
            <div className="min-h-0 flex-1 space-y-2 overflow-y-auto p-2">
              {items.map((i) => (
                <button
                  key={i.id}
                  onClick={() => onOpen(i)}
                  className="w-full rounded-lg border border-border bg-card p-3 text-left transition-colors hover:bg-muted/40"
                >
                  <div className="line-clamp-2 text-sm font-medium">{i.incidentName}</div>
                  <div className="mt-2 flex items-center justify-between text-[11px] text-muted-foreground">
                    <span className={cn('font-medium', SEV_TONE[sevKey(i.incidentSeverity)])}>
                      {t(`incidents.sev.${sevKey(i.incidentSeverity)}`)}
                    </span>
                    <span className="flex items-center gap-1">
                      <Link2 size={11} /> {i.alertCount}
                    </span>
                  </div>
                </button>
              ))}
            </div>
          </div>
        )
      })}
    </div>
  )
}
