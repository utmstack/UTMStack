import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { PILL_BASE } from '@/features/alerts/lib/alert-meta'
import { ST_META, statusKey } from '../lib/incident-meta'
import type { IncidentStatus } from '../types/incident.types'

export function IncidentStatusPill({ status }: { status: IncidentStatus }) {
  const { t } = useTranslation()
  return <span className={cn(PILL_BASE, ST_META[status].pill)}>{t(`incidents.status.${statusKey(status)}`)}</span>
}
