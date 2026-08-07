import { useTranslation } from 'react-i18next'
import { useDateFormat } from '@/shared/lib/datetime'
import { useIncidentHistoryTab } from '../hooks/use-incident-history-tab'
import { TabEmpty, TabError, TabLoader } from './ui-primitives'

export function IncidentHistoryTab({ incidentId }: { incidentId: string }) {
  const { t } = useTranslation()
  const df = useDateFormat()
  const { rows, error, reload } = useIncidentHistoryTab(incidentId)

  if (error) return <TabError onRetry={reload} />
  if (rows === null) return <TabLoader />
  if (rows.length === 0) return <TabEmpty>{t('incidents.history.empty')}</TabEmpty>
  return (
    <ol className="relative ml-2 space-y-4 border-l border-border pl-5">
      {rows.map((h) => (
        <li key={h.id} className="relative">
          <span className="absolute -left-[27px] top-0.5 flex h-4 w-4 items-center justify-center rounded-full border border-border bg-card">
            <span className="h-1.5 w-1.5 rounded-full bg-primary" />
          </span>
          {/* The row stores the action code; the readable label is a translation
              of it. An unknown code falls back to itself rather than to a blank
              line, so a new action shows up as something instead of nothing. */}
          <div className="text-xs font-medium">{t(`incidents.action.${h.action}`, h.action)}</div>
          <div className="mt-0.5 text-[10px] text-muted-foreground">
            {h.actionCreatedBy || t('incidents.unknownUser')} · {df.formatDateTime(h.actionCreatedDate)}
          </div>
        </li>
      ))}
    </ol>
  )
}
