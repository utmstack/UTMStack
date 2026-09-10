import { useTranslation } from 'react-i18next'
import { ResizableGridHeader } from '@/shared/components/ui/resizable-grid-header'
import { colMins, useResizableColumns } from '@/shared/hooks/useResizableColumns'
import { cn } from '@/shared/lib/utils'
import { useDateFormat } from '@/shared/lib/datetime'
import { SEV_TONE, TABLE_COLS, sevKey } from '../lib/incident-meta'
import type { Incident } from '../types/incident.types'
import { IncidentStatusPill } from './incident-status-pill'
import { IncidentAssignee } from './incident-assignee'

export function IncidentsTable({ incidents, onOpen }: { incidents: Incident[]; onOpen: (i: Incident) => void }) {
  const { t } = useTranslation()
  const df = useDateFormat()
  const incidentsHeaders = [
    t('incidents.table.name'),
    t('incidents.table.status'),
    t('incidents.table.severity'),
    t('incidents.table.assignee'),
    t('incidents.table.alerts'),
    t('incidents.table.created'),
  ]
  const { template: tableCols, startDrag } = useResizableColumns(TABLE_COLS, {
    min: colMins(incidentsHeaders),
    storageKey: 'incidents-table-columns',
  })
  return (
    <div className="mt-4 min-h-0 flex-1 overflow-x-auto overflow-y-auto rounded-xl border border-border">
      <ResizableGridHeader
        headers={incidentsHeaders}
        tableCols={tableCols}
        startDrag={startDrag}
        className="bg-muted/30 py-2.5 font-medium"
        cellClassName="[&:nth-child(5)]:text-center"
      />
      {incidents.map((i) => (
        <button
          key={i.id}
          onClick={() => onOpen(i)}
          className="grid w-max min-w-full items-center gap-3 border-b border-border/60 px-4 py-3 text-left text-sm transition-colors last:border-b-0 hover:bg-muted/30"
          style={{ gridTemplateColumns: tableCols }}
        >
          <div className="min-w-0">
            <div className="truncate font-medium">{i.incidentName}</div>
            {i.incidentDescription && <div className="truncate text-xs text-muted-foreground">{i.incidentDescription}</div>}
          </div>
          <div>
            <IncidentStatusPill status={i.incidentStatus} />
          </div>
          <div className={cn('text-xs font-medium', SEV_TONE[sevKey(i.incidentSeverity)])}>
            {t(`incidents.sev.${sevKey(i.incidentSeverity)}`)}
          </div>
          <div className="min-w-0">
            <IncidentAssignee login={i.incidentAssignedTo} />
          </div>
          <div className="text-center font-mono tabular-nums text-muted-foreground">{i.alertCount}</div>
          <div className="font-mono text-xs text-muted-foreground">{df.formatDate(i.incidentCreatedDate)}</div>
        </button>
      ))}
    </div>
  )
}
