import { useState } from 'react'
import { X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { useDateFormat } from '@/shared/lib/datetime'
import { SEV_TONE, ST_META, STATUSES, sevKey } from '../lib/incident-meta'
import { useIncidentStatus } from '../hooks/use-incident-status'
import type { Incident, IncidentStatus } from '../types/incident.types'
import { IncidentAssignee, IncidentAssigneePicker } from './incident-assignee'
import { IncidentStatusPill } from './incident-status-pill'
import { IncidentOverviewTab } from './incident-overview-tab'
import { IncidentAlertsTab } from './incident-alerts-tab'
import { IncidentNotesTab } from './incident-notes-tab'
import { IncidentHistoryTab } from './incident-history-tab'

type Tab = 'overview' | 'alerts' | 'notes' | 'history'

export function IncidentDrawer({
  incident,
  onClose,
  onChanged,
}: {
  incident: Incident
  onClose: () => void
  onChanged: (id: number) => void
}) {
  const { t } = useTranslation()
  const df = useDateFormat()
  const [tab, setTab] = useState<Tab>('overview')
  const [solution, setSolution] = useState(incident.incidentSolution ?? '')
  const { busy, changeStatus } = useIncidentStatus(incident, onChanged)

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[760px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                <IncidentStatusPill status={incident.incidentStatus} />
                <span className={cn('font-medium', SEV_TONE[sevKey(incident.incidentSeverity)])}>
                  {t(`incidents.sev.${sevKey(incident.incidentSeverity)}`)}
                </span>
                <span>· #{incident.id}</span>
              </div>
              <h2 className="mt-1 text-xl font-semibold">{incident.incidentName}</h2>
              <div className="mt-1 flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
                <IncidentAssignee login={incident.incidentAssignedTo} />
                <span>· {t('incidents.drawer.created', { date: df.formatDateTime(incident.incidentCreatedDate) })}</span>
                <span>· {t('incidents.drawer.alertsCount', { count: incident.alertCount })}</span>
              </div>
            </div>
            <button
              onClick={onClose}
              className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <X size={16} />
            </button>
          </div>

          {/* Status actions + assignment */}
          <div className="mt-4 flex flex-wrap items-center gap-1.5">
            {STATUSES.map((s) => (
              <button
                key={s}
                disabled={busy || s === incident.incidentStatus}
                onClick={() => void changeStatus(s, solution)}
                className={cn(
                  'inline-flex items-center gap-1.5 rounded-md border px-2.5 py-1 text-xs transition-colors disabled:cursor-default',
                  s === incident.incidentStatus
                    ? cn('ring-1 ring-inset', ST_META[s].pill)
                    : 'border-border text-muted-foreground hover:bg-muted hover:text-foreground'
                )}
              >
                <span className={cn('h-2 w-2 rounded-full', ST_META[s].dot)} /> {t(`incidents.status.${s}`)}
              </button>
            ))}
            <span className="mx-1 h-4 w-px bg-border" />
            <IncidentAssigneePicker incident={incident} onChanged={() => onChanged(incident.id)} />
          </div>
        </header>

        <nav className="flex items-center gap-1 border-b border-border px-6">
          {(['overview', 'alerts', 'notes', 'history'] as Tab[]).map((id) => (
            <button
              key={id}
              onClick={() => setTab(id)}
              className={cn(
                'relative px-3 py-2.5 text-xs transition-colors',
                tab === id ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
              )}
            >
              {t(`incidents.drawer.tabs.${id}`)}
              {tab === id && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
            </button>
          ))}
        </nav>

        <div className="flex-1 overflow-y-auto bg-muted/10 p-6">
          {tab === 'overview' && (
            <IncidentOverviewTab incident={incident} solution={solution} onSolutionChange={setSolution} />
          )}
          {tab === 'alerts' && <IncidentAlertsTab incidentId={incident.id} onChanged={() => onChanged(incident.id)} />}
          {tab === 'notes' && <IncidentNotesTab incidentId={incident.id} />}
          {tab === 'history' && <IncidentHistoryTab incidentId={incident.id} />}
        </div>
      </div>
    </div>
  )
}

/** Re-export the status type-bridge so the drawer's consumer doesn't need to drill into the types file. */
export type { IncidentStatus }
