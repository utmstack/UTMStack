import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { SEV_BADGE } from '@/features/alerts/lib/alert-meta'
import { SEVERITY_CHIP } from '@/features/alerts/components/alert-cells'
import { sevKey } from '../lib/incident-meta'
import type { IncidentSeverity } from '../types/incident.types'

/** Same chip the alerts table uses; an incident with no alerts has no severity yet. */
export function IncidentSeverityBadge({ severity }: { severity?: IncidentSeverity }) {
  const { t } = useTranslation()
  const key = sevKey(severity)
  return (
    <span className={cn(SEVERITY_CHIP, key === 'unknown' ? 'bg-muted text-muted-foreground ring-border' : SEV_BADGE[key])}>
      {t(`incidents.sev.${key}`)}
    </span>
  )
}
