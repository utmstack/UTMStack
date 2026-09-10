import { useTranslation } from 'react-i18next'
import { ResizableTableHeader } from '@/shared/components/ui/resizable-table-header'
import { colMins, useResizableColumns } from '@/shared/hooks/useResizableColumns'

const TH = 'whitespace-nowrap px-3 py-2.5 text-left align-middle font-medium'
const INCIDENT_ALERTS_TABLE_COLS = [6, 36, 360, 90, 160]

export function IncidentAlertsPickerHeader({
  allChecked,
  onTogglePage,
}: {
  allChecked: boolean
  onTogglePage: () => void
}) {
  const { t } = useTranslation()
  const pickerLabelMins = colMins([
    t('alerts.table.alert'),
    t('alerts.table.severity'),
    t('alerts.table.time'),
  ])
  const { widths, startDrag } = useResizableColumns(INCIDENT_ALERTS_TABLE_COLS, {
    min: [6, 36, ...pickerLabelMins],
    storageKey: 'incident-alerts-picker-table-columns',
  })
  return (
    <ResizableTableHeader
      cells={[
        { content: null, className: `${TH} w-[6px] p-0` },
        { content: <button onClick={onTogglePage} className="flex h-4 w-4 items-center justify-center rounded border border-input">{allChecked && <span className="h-2 w-2 rounded-sm bg-primary" />}</button>, className: `${TH} w-px` },
        { content: t('alerts.table.alert'), className: TH },
        { content: t('alerts.table.severity'), className: `${TH} text-center` },
        { content: t('alerts.table.time'), className: `${TH} text-center` },
      ]}
      widths={widths}
      startDrag={startDrag}
      className="sticky top-0 z-10 bg-muted/90 text-[10px] uppercase tracking-wider text-muted-foreground"
      rowClassName="border-b border-border"
    />
  )
}
