import { useTranslation } from 'react-i18next'
import { useDateFormat } from '@/shared/lib/datetime'
import { SEV_TONE, sevKey } from '../lib/incident-meta'
import type { Incident } from '../types/incident.types'
import { DescRow, Section } from './ui-primitives'
import { IncidentAssignee } from './incident-assignee'
import { IncidentStatusPill } from './incident-status-pill'

export function IncidentOverviewTab({
  incident,
  solution,
  onSolutionChange,
}: {
  incident: Incident
  solution: string
  onSolutionChange: (v: string) => void
}) {
  const { t } = useTranslation()
  const df = useDateFormat()
  return (
    <div className="space-y-4">
      {incident.incidentDescription && (
        <Section title={t('incidents.drawer.section.description')}>
          <p className="whitespace-pre-wrap text-xs leading-relaxed text-muted-foreground">
            {incident.incidentDescription}
          </p>
        </Section>
      )}
      <Section title={t('incidents.drawer.section.solution')}>
        <textarea
          value={solution}
          onChange={(e) => onSolutionChange(e.target.value)}
          rows={3}
          placeholder={t('incidents.drawer.solutionPlaceholder')}
          className="w-full rounded-md border border-input bg-background/40 p-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        />
        <p className="mt-1 text-[11px] text-muted-foreground">{t('incidents.drawer.solutionHint')}</p>
      </Section>
      <Section title={t('incidents.drawer.section.details')}>
        <dl className="grid grid-cols-[140px_1fr] gap-y-2 text-xs">
          <DescRow k={t('incidents.drawer.details.id')}>
            <span className="font-mono">#{incident.id}</span>
          </DescRow>
          <DescRow k={t('incidents.drawer.details.status')}>
            <IncidentStatusPill status={incident.incidentStatus} />
          </DescRow>
          <DescRow k={t('incidents.drawer.details.severity')}>
            <span className={SEV_TONE[sevKey(incident.incidentSeverity)]}>
              {t(`incidents.sev.${sevKey(incident.incidentSeverity)}`)}
            </span>
          </DescRow>
          <DescRow k={t('incidents.drawer.details.assignee')}>
            <IncidentAssignee login={incident.incidentAssignedTo} />
          </DescRow>
          <DescRow k={t('incidents.drawer.details.created')}>{df.formatDateTime(incident.incidentCreatedDate)}</DescRow>
          <DescRow k={t('incidents.drawer.details.alerts')}>{incident.alertCount}</DescRow>
        </dl>
      </Section>
    </div>
  )
}
