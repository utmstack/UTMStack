import { useTranslation } from 'react-i18next'
import { ResizableTableHeader } from '@/shared/components/ui/resizable-table-header'
import type { ColSize, useResizableColumns } from '@/shared/hooks/useResizableColumns'

const TH = 'whitespace-nowrap px-3 py-2.5 text-left align-middle font-medium'
export const ALERTS_TABLE_COLS = [6, 36, 38, 38, 360, 130, 180, 160, 160, 90, 90, 160]

export function AlertsTableHeader({
  allChecked,
  widths,
  startDrag,
  onTogglePage,
}: {
  allChecked: boolean
  widths: ColSize[]
  startDrag: ReturnType<typeof useResizableColumns>['startDrag']
  onTogglePage: () => void
}) {
  const { t } = useTranslation()
  return (
    <ResizableTableHeader
      cells={[
        { content: null, className: `${TH} w-[6px] p-0` },
        { content: <button onClick={onTogglePage} className="flex h-4 w-4 items-center justify-center rounded border border-input">{allChecked && <span className="h-2 w-2 rounded-sm bg-primary" />}</button>, className: `${TH} w-px` },
        { content: t('alerts.table.actions'), className: `${TH} text-center` },
        { content: null, className: `${TH} text-center` },
        { content: t('alerts.table.alert'), className: TH },
        { content: t('alerts.table.status'), className: TH },
        { content: t('alerts.table.technique'), className: TH },
        { content: t('alerts.table.source'), className: TH },
        { content: t('alerts.table.adversary'), className: TH },
        { content: t('alerts.table.severity'), className: `${TH} text-center` },
        { content: t('alerts.table.echoes'), className: `${TH} text-center` },
        { content: t('alerts.table.time'), className: `${TH} text-center` },
      ]}
      widths={widths}
      startDrag={startDrag}
      className="sticky top-0 z-10 bg-muted/90 text-[10px] uppercase tracking-wider text-muted-foreground"
      rowClassName="border-b border-border"
    />
  )
}

export const ALERTS_TABLE_COLUMN_COUNT = 12
