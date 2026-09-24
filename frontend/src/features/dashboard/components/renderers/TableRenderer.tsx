import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import type { ColumnDef } from '@tanstack/react-table'
import { ResizableDataTable } from '@/shared/components/ui/resizable-data-table'
import type { Row } from '@/features/dashboard/types'

const TH = 'whitespace-nowrap px-3 py-2 text-left align-middle font-medium'
const TD = 'whitespace-nowrap px-3 py-2 align-middle text-xs text-foreground/90'

export function TableRenderer({ rows }: { rows: Row[] }) {
  const { t } = useTranslation()
  const fields = useMemo(() => (rows.length === 0 ? [] : Object.keys(rows[0])), [rows])
  const columns = useMemo<ColumnDef<Row>[]>(
    () =>
      fields.map((field) => ({
        id: field,
        header: field,
        size: 160,
        minSize: 60,
        meta: { headerClassName: TH, cellClassName: TD, cellProps: (row) => ({ title: formatCell(row[field]) }) },
        cell: ({ row }) => <span className="block truncate">{formatCell(row.original[field])}</span>,
      })),
    [fields],
  )

  if (rows.length === 0) {
    return (
      <div className="flex h-full w-full items-center justify-center text-xs text-muted-foreground">
        {t('dashboards.widget.noData')}
      </div>
    )
  }

  return (
    <div className="h-full w-full overflow-auto">
      {/* Keyed by the field set: widths are remembered per set of columns, and a
          query edited in the editor preview changes which columns there are. */}
      <ResizableDataTable
        key={fields.join('|')}
        columns={columns}
        data={rows}
        flexColumnId={fields[fields.length - 1]}
        storageKey={`dashboard-table-sizing:${fields.join('|')}`}
        headerClassName="bg-card"
        rowClassName={() => 'last:border-b-0'}
      />
    </div>
  )
}

function formatCell(value: unknown): string {
  if (value == null) return '—'
  if (typeof value === 'object') {
    try {
      return JSON.stringify(value)
    } catch {
      return String(value)
    }
  }
  return String(value)
}
