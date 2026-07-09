import { Trash2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { cn } from '@/shared/lib/utils'
import { sevKey } from '../lib/incident-meta'
import { useIncidentAlertsTab } from '../hooks/use-incident-alerts-tab'
import { TabEmpty, TabError, TabLoader } from './ui-primitives'

export function IncidentAlertsTab({ incidentId, onChanged }: { incidentId: number; onChanged: () => void }) {
  const { t } = useTranslation()
  const { rows, error, reload, remove } = useIncidentAlertsTab(incidentId, onChanged)

  if (error) return <TabError onRetry={reload} />
  if (rows === null) return <TabLoader />
  if (rows.length === 0) return <TabEmpty>{t('incidents.alerts.empty')}</TabEmpty>
  return (
    <div className="overflow-hidden rounded-lg border border-border bg-card">
      {rows.map((a) => (
        <Link
          to="/threat-management/alerts"
          state={{
            socaiFilters: [
              { field: 'id', operator: 'IS', value: a.alertId },
              { field: '@timestamp', operator: 'IS_BETWEEN', value: ['now-30d', 'now'] },
            ],
            socaiTime: 'now-30d',
          }}
          key={a.id}
          className="group flex items-center gap-3 border-b border-border/60 px-4 py-2.5 text-xs last:border-b-0"
        >
          <span
            className={cn(
              'h-4 w-1 shrink-0 rounded-full',
              a.alertSeverity === 3 ? 'bg-red-500' : a.alertSeverity === 2 ? 'bg-amber-500' : 'bg-sky-500'
            )}
          />
          <span className="min-w-0 flex-1 truncate font-medium" title={a.alertName}>
            {a.alertName}
          </span>
          <span className="shrink-0 font-mono text-[10px] text-muted-foreground">
            {t(`incidents.sev.${sevKey(a.alertSeverity)}`)}
          </span>
          <button
            onClick={() => void remove(a)}
            title={t('incidents.alerts.remove')}
            className="shrink-0 rounded p-1 text-muted-foreground opacity-0 transition hover:bg-muted hover:text-red-500 group-hover:opacity-100"
          >
            <Trash2 size={13} />
          </button>
        </Link>
      ))}
    </div>
  )
}
