import { memo, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import type { ColumnDef } from '@tanstack/react-table'
import { ChevronRight, X } from 'lucide-react'
import { ResizableDataTable } from '@/shared/components/ui/resizable-data-table'
import { cn } from '@/shared/lib/utils'
import type { FilterType, LogDocument } from '../types/log-explorer.types'
import { SRC_FIELDS, pick } from '../domain/flatten'
import {
  FIELD_COL,
  LAST_FIELD_COL,
  LEVEL_FIELDS,
  LEVEL_TONE,
  LogDetail,
  colValue,
  fieldColumnSize,
  fieldLabel,
  flatOf,
  shortTime,
} from './log-results'

const TS = '@timestamp'
const TH = 'whitespace-nowrap px-3 py-1.5 text-left align-middle font-medium'
const TD = 'whitespace-nowrap px-3 py-1 align-middle'
const FIELD_MIN = 80

// A picked column may be any field of the document, `@timestamp` and `source`
// included, so their ids cannot be the bare field names.
const fieldId = (name: string) => `field:${name}`

interface Layout {
  columns: string[]
  autoColumns: string[]
  expanded: number | null
  onRemoveColumn?: (field: string) => void
}

function buildColumns(t: TFunction, { columns, autoColumns, expanded, onRemoveColumn }: Layout): ColumnDef<LogDocument>[] {
  const field = (name: string, isLast: boolean, removable: boolean): ColumnDef<LogDocument> => ({
    id: fieldId(name),
    size: fieldColumnSize(name, isLast),
    minSize: FIELD_MIN,
    header: () => (
      <div className="group flex min-w-0 items-center gap-1">
        <span className="truncate" title={name}>
          {fieldLabel(name)}
        </span>
        {removable && onRemoveColumn && (
          <button
            onClick={() => onRemoveColumn(name)}
            title={t('logExplorer.results.removeColumn', { field: name })}
            className="shrink-0 opacity-0 transition-opacity hover:text-foreground group-hover:opacity-100"
          >
            <X size={11} />
          </button>
        )}
      </div>
    ),
    meta: {
      headerClassName: TH,
      cellClassName: `${TD} font-mono`,
      cellProps: (doc) => ({ title: colValue(flatOf(doc), name) }),
    },
    cell: ({ row }) => {
      const value = colValue(flatOf(row.original), name)
      return (
        <span className={cn('block truncate', value === '—' ? 'text-muted-foreground/40' : 'text-foreground/85')}>{value}</span>
      )
    },
  })

  const manual = columns.length > 0
  const data = manual
    ? columns.map((c, i) => field(c, i === columns.length - 1, true))
    : [
        {
          id: 'source',
          size: autoColumns.length === 0 ? LAST_FIELD_COL : FIELD_COL,
          minSize: FIELD_MIN,
          header: () => <span className="truncate">{t('logExplorer.results.source')}</span>,
          meta: { headerClassName: TH, cellClassName: `${TD} font-mono` },
          cell: ({ row }) => (
            <span className="block truncate text-foreground/70">{pick(flatOf(row.original), SRC_FIELDS) ?? '—'}</span>
          ),
        } satisfies ColumnDef<LogDocument>,
        ...autoColumns.map((c, i) => field(c, i === autoColumns.length - 1, false)),
      ]

  return [
    {
      id: 'expand',
      header: () => null,
      size: 32,
      minSize: 32,
      enableResizing: false,
      meta: { headerClassName: `${TH} px-2`, cellClassName: 'whitespace-nowrap px-2 py-1 align-middle' },
      cell: ({ row }) => (
        <ChevronRight
          size={13}
          className={cn('text-muted-foreground/60 transition-transform', expanded === row.index && 'rotate-90 text-foreground')}
        />
      ),
    },
    {
      id: 'level',
      header: () => null,
      size: 20,
      minSize: 20,
      enableResizing: false,
      meta: { headerClassName: `${TH} px-1`, cellClassName: 'whitespace-nowrap px-1 py-1 align-middle' },
      cell: ({ row }) => {
        const level = (pick(flatOf(row.original), LEVEL_FIELDS) ?? '').toLowerCase()
        const tone = LEVEL_TONE[level] ?? { dot: 'bg-muted-foreground/50', tone: '' }
        return <span className={cn('block h-3.5 w-[3px] rounded-full', tone.dot)} />
      },
    },
    {
      id: 'time',
      header: t('logExplorer.results.time'),
      size: 168,
      minSize: 120,
      meta: {
        headerClassName: TH,
        cellClassName: `${TD} font-mono tabular-nums text-muted-foreground`,
      },
      cell: ({ row }) => {
        const ts = flatOf(row.original)[TS] as string | undefined
        return ts ? shortTime(ts) : '—'
      },
    },
    ...data,
  ]
}

/**
 * The results grid of the log explorer: a resizable table with one expandable
 * row. Memoized on purpose — the explorer re-renders on every keystroke in the
 * query box, and a table of thousands of rows must not with it, so everything
 * it takes is a stable value or callback.
 */
function LogTableImpl({
  docs,
  columns,
  autoColumns,
  expanded,
  onToggle,
  onAdd,
  onSurrounding,
  onRemoveColumn,
}: {
  docs: LogDocument[]
  columns: string[]
  autoColumns: string[]
  expanded: number | null
  onToggle: (index: number) => void
  onAdd?: (f: FilterType) => void
  onSurrounding?: (ts: string, srcField?: string, srcVal?: string) => void
  onRemoveColumn?: (field: string) => void
}) {
  const { t } = useTranslation()
  const tableColumns = useMemo(
    () => buildColumns(t, { columns, autoColumns, expanded, onRemoveColumn }),
    [t, columns, autoColumns, expanded, onRemoveColumn],
  )
  // Widths are kept per column set, and the set changes what the columns are:
  // the table starts over with its own sizes whenever it does.
  const storageKey = `log-explorer-table-sizing:${columns.length > 0 ? columns.join('|') : `auto:${autoColumns.join('|')}`}`
  const flexColumn =
    columns.length > 0
      ? fieldId(columns[columns.length - 1])
      : autoColumns.length > 0
        ? fieldId(autoColumns[autoColumns.length - 1])
        : 'source'

  return (
    <ResizableDataTable
      key={storageKey}
      columns={tableColumns}
      data={docs}
      flexColumnId={flexColumn}
      storageKey={storageKey}
      headerClassName="bg-card"
      onRowClick={(_, index) => onToggle(index)}
      rowClassName={(_, index) =>
        cn('border-border/40 text-xs leading-tight last:border-b-0', expanded === index ? 'bg-muted/30' : 'hover:bg-muted/20')
      }
      renderExpandedRow={(doc, index) =>
        expanded === index ? (
          <div className="border-l-2 border-l-sky-500/50 bg-muted/15">
            <LogDetail doc={doc} onAdd={onAdd} onSurrounding={onSurrounding} />
          </div>
        ) : null
      }
    />
  )
}

export const LogTable = memo(LogTableImpl)
