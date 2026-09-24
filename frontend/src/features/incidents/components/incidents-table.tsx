import { useMemo } from 'react'
import { AlertTriangle, Loader2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import type { ColumnDef } from '@tanstack/react-table'
import { ResizableDataTable } from '@/shared/components/ui/resizable-data-table'
import { useDateFormat } from '@/shared/lib/datetime'
import type { Incident } from '../types/incident.types'
import { IncidentSeverityBadge } from './incident-severity-badge'
import { IncidentStatusMenu } from './incident-status-menu'
import { IncidentAssignee } from './incident-assignee'

function Message({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">{children}</div>
}

const TH = 'whitespace-nowrap px-3 py-2.5 text-left align-middle font-medium'
const TD = 'whitespace-nowrap px-3 py-3 align-middle'

const stopRowClick = (e: React.MouseEvent) => e.stopPropagation()

export function IncidentsTable({
  incidents,
  onOpen,
  onChanged,
  loading,
  error,
  onRetry,
}: {
  incidents: Incident[]
  onOpen: (i: Incident) => void
  onChanged: (id: string) => void
  loading: boolean
  error: boolean
  onRetry: () => void
}) {
  const { t } = useTranslation()
  const df = useDateFormat()
  const columns = useMemo(() => buildColumns(t, df.formatDate, onChanged), [t, df.formatDate, onChanged])
  return (
    <ResizableDataTable
      columns={columns}
      data={incidents}
      flexColumnId="name"
      storageKey="incidents-table-sizing"
      getRowId={(i) => i.id}
      onRowClick={onOpen}
      rowClassName={() => 'border-border/60 last:border-b-0 hover:bg-muted/30'}
      loading={loading && incidents.length === 0}
      loadingContent={
        <Message>
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        </Message>
      }
      error={error}
      errorContent={
        <Message>
          <AlertTriangle size={16} className="text-amber-500" /> {t('incidents.loadError')}
          <button onClick={onRetry} className="ml-2 text-primary hover:underline">
            {t('incidents.retry')}
          </button>
        </Message>
      }
      emptyContent={<Message>{t('incidents.empty')}</Message>}
    />
  )
}

function buildColumns(
  t: TFunction,
  formatDate: (date: Incident['incidentCreatedDate']) => string,
  onChanged: (id: string) => void,
): ColumnDef<Incident>[] {
  return [
    {
      id: 'name',
      header: t('incidents.table.name'),
      size: 320,
      minSize: 120,
      meta: { headerClassName: `${TH} pl-4`, cellClassName: `${TD} pl-4` },
      cell: ({ row }) => (
        <>
          <div className="truncate font-medium">{row.original.incidentName}</div>
          {row.original.incidentDescription && (
            <div className="truncate text-xs text-muted-foreground">{row.original.incidentDescription}</div>
          )}
        </>
      ),
    },
    {
      id: 'status',
      header: t('incidents.table.status'),
      size: 120,
      minSize: 100,
      meta: { headerClassName: TH, cellClassName: TD, cellProps: () => ({ onClick: stopRowClick }) },
      cell: ({ row }) => <IncidentStatusMenu incident={row.original} onChanged={onChanged} />,
    },
    {
      id: 'severity',
      header: t('incidents.table.severity'),
      size: 100,
      minSize: 88,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => <IncidentSeverityBadge severity={row.original.incidentSeverity} />,
    },
    {
      id: 'assignee',
      header: t('incidents.table.assignee'),
      size: 150,
      minSize: 80,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => <IncidentAssignee login={row.original.incidentAssignedTo} />,
    },
    {
      id: 'alerts',
      header: t('incidents.table.alerts'),
      size: 70,
      minSize: 60,
      meta: { headerClassName: TH, cellClassName: `${TD} font-mono tabular-nums text-muted-foreground` },
      cell: ({ row }) => row.original.alertCount,
    },
    {
      id: 'created',
      header: t('incidents.table.created'),
      size: 110,
      minSize: 90,
      meta: { headerClassName: TH, cellClassName: `${TD} font-mono text-xs text-muted-foreground` },
      cell: ({ row }) => formatDate(row.original.incidentCreatedDate),
    },
  ]
}
