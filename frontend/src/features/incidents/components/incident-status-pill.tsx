import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { ST_META } from '../lib/incident-meta'
import type { IncidentStatus } from '../types/incident.types'

export function IncidentStatusPill({ status }: { status: IncidentStatus }) {
  const { t } = useTranslation()
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[11px] font-medium ring-1 ring-inset',
        ST_META[status].pill
      )}
    >
      <span className={cn('h-1.5 w-1.5 rounded-full', ST_META[status].dot)} /> {t(`incidents.status.${status}`)}
    </span>
  )
}
