import { useEffect, useRef, useState } from 'react'
import { Crosshair, Flame, Target } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { absTime } from '../lib/alert-meta'
import type { Alert } from '../types/alert.types'

/**
 * Row-level incident affordance. Two states:
 * - Alert has no incident: "open" crosshair → click opens the create-incident modal.
 * - Alert is part of an incident: "closed" bullseye → click reveals a popover
 *   summarizing the linked incident (name, id, creator, creation date).
 */
export function AlertIncidentTarget({
  alert: a,
  onIncident,
}: {
  alert: Alert
  onIncident: (a: Alert) => void
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) =>
      ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  if (!a.isIncident) {
    return (
      <button
        onClick={(e) => {
          e.stopPropagation()
          onIncident(a)
        }}
        title={t('alerts.row.createIncident')}
        aria-label={t('alerts.row.createIncident')}
        className="flex h-7 w-7 items-center justify-center rounded text-muted-foreground/60 transition hover:bg-background hover:text-red-500"
      >
        <Crosshair size={13} />
      </button>
    )
  }

  const d = a.incidentDetail
  return (
    <div className="relative" ref={ref} onClick={(e) => e.stopPropagation()}>
      <button
        onClick={() => setOpen((v) => !v)}
        title={t('alerts.row.viewIncident')}
        aria-label={t('alerts.row.viewIncident')}
        className="flex h-7 w-7 items-center justify-center rounded text-red-500 transition hover:bg-red-500/10"
      >
        <Target size={13} />
      </button>
      {open && (
        <div className="absolute left-0 top-full z-40 mt-1 w-64 rounded-md border border-border bg-popover p-3 shadow-lg">
          <div className="mb-2 flex items-center gap-1.5 text-xs font-medium text-red-600 dark:text-red-300">
            <Flame size={12} /> {t('alerts.drawer.partOfIncident')}
          </div>
          <dl className="grid grid-cols-[80px_1fr] gap-y-1 text-[11px]">
            {d?.incidentName && (
              <>
                <dt className="text-muted-foreground">{t('alerts.drawer.incidentDetail.incident')}</dt>
                <dd className="truncate" title={d.incidentName}>{d.incidentName}</dd>
              </>
            )}
            {d?.incidentId != null && (
              <>
                <dt className="text-muted-foreground">{t('alerts.drawer.incidentDetail.id')}</dt>
                <dd className="font-mono">#{String(d.incidentId)}</dd>
              </>
            )}
            {d?.createdBy && (
              <>
                <dt className="text-muted-foreground">{t('alerts.drawer.incidentDetail.createdBy')}</dt>
                <dd className="truncate" title={d.createdBy}>{d.createdBy}</dd>
              </>
            )}
            {d?.creationDate && (
              <>
                <dt className="text-muted-foreground">{t('alerts.drawer.incidentDetail.created')}</dt>
                <dd>{absTime(d.creationDate)}</dd>
              </>
            )}
          </dl>
        </div>
      )}
    </div>
  )
}
