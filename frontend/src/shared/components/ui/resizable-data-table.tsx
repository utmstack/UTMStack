import {
  Fragment,
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  type MouseEvent,
  type ReactNode,
  type TdHTMLAttributes,
} from 'react'
import {
  type Column,
  type ColumnDef,
  type ColumnDefTemplate,
  type ColumnSizingState,
  type Table,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { cn } from '@/shared/lib/utils'
import { ColumnResizeHandle } from './column-resize-handle'

declare module '@tanstack/react-table' {
  interface ColumnMeta<TData, TValue> {
    /** Keep resizable columns left-aligned: the handle sits on the column's left edge, so a centered label drifts away from it as the column widens. */
    headerClassName?: string
    cellClassName?: string
    /** Extra attributes for this column's <td> in a given row (title, onClick…). */
    cellProps?: (row: TData) => TdHTMLAttributes<HTMLTableCellElement>
  }
}

export interface ResizableDataTableProps<T> {
  columns: ColumnDef<T, any>[]
  data: T[]
  getRowId?: (row: T, index: number) => string
  onRowClick?: (row: T, index: number) => void
  rowClassName?: (row: T, index: number) => string
  headerClassName?: string
  loading?: boolean
  loadingContent?: ReactNode
  error?: boolean
  errorContent?: ReactNode
  emptyContent?: ReactNode
  /** Content for a full-width row rendered right below `row`; return null for none. */
  renderExpandedRow?: (row: T, index: number) => ReactNode
  /** localStorage key under which the user's column widths are kept. */
  storageKey?: string
  /**
   * Column that absorbs whatever width the others don't need, so the table
   * fills the screen. It renders with no width of its own (only `minSize`);
   * table-layout:fixed hands it the leftover space.
   *
   * With a flex column, dragging a handle only trades width between the two
   * columns it sits between — the sum stays constant, so no other column
   * moves. When the container is narrower than the columns' `size`s, every
   * column (this one included) shrinks toward its `minSize` in proportion to
   * its slack, so the table only scrolls once all of them hit their floor.
   * Without a flex column, the handle resizes the column to its left and the
   * table grows or shrinks with it.
   */
  flexColumnId?: string
}

/**
 * A resizable table built on TanStack Table's column-sizing feature instead
 * of a hand-rolled drag implementation for every column. Each column keeps
 * an independent width — resizing one never touches another — except for the
 * one column named by `flexColumnId`, which absorbs the table's leftover
 * width (see that prop's doc comment).
 *
 * Headless by design: this only owns the <table> chrome (sticky header,
 * resize handles, loading/error/empty states). Column content, styling and
 * row shape are entirely up to the caller's `columns`/`data`.
 */
export function ResizableDataTable<T>({
  columns,
  data,
  getRowId,
  onRowClick,
  rowClassName,
  headerClassName,
  loading,
  loadingContent,
  error,
  errorContent,
  emptyContent,
  renderExpandedRow,
  storageKey,
  flexColumnId,
}: ResizableDataTableProps<T>) {
  const [columnSizing, setColumnSizing] = useState<ColumnSizingState>(() => readSizing(storageKey))
  useEffect(() => {
    if (!storageKey) return
    try {
      localStorage.setItem(storageKey, JSON.stringify(columnSizing))
    } catch {
      /* storage unavailable — widths just won't persist */
    }
  }, [storageKey, columnSizing])

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getRowId,
    columnResizeMode: 'onChange',
    enableColumnResizing: true,
    state: { columnSizing },
    onColumnSizingChange: setColumnSizing,
  })

  const tableRef = useRef<HTMLTableElement>(null)
  const [available, setAvailable] = useState(0)
  useLayoutEffect(() => {
    const container = tableRef.current?.parentElement
    if (!container) return
    const measure = () => setAvailable(container.clientWidth)
    measure()
    const observer = new ResizeObserver(measure)
    observer.observe(container)
    return () => observer.disconnect()
  }, [])

  const startDrag = useBoundaryDrag(table, flexColumnId)

  const headerGroups = table.getHeaderGroups()
  const rows = table.getRowModel().rows
  const colSpan = columns.length
  const leaves = table.getVisibleLeafColumns()
  const fitted = flexColumnId && available > 0 ? fitWidths(leaves, available) : null
  const widthOf = (column: Column<T, unknown>) => fitted?.get(column.id) ?? column.getSize()
  const tableStyle = flexColumnId
    ? {
        minWidth: leaves.reduce(
          (sum, c) => sum + (c.id === flexColumnId ? minSizeOf(c) : widthOf(c)),
          0,
        ),
      }
    : { width: table.getTotalSize() }

  return (
    <table
      ref={tableRef}
      className={cn('border-collapse table-fixed', flexColumnId && 'w-full')}
      style={tableStyle}
    >
      <thead
        className={cn(
          'sticky top-0 z-10 bg-muted/90 text-[10px] uppercase tracking-wider text-muted-foreground',
          headerClassName,
        )}
      >
        {headerGroups.map((headerGroup) => (
          <tr key={headerGroup.id} className="border-b border-border">
            {headerGroup.headers.map((header, index, headers) => {
              const isFlex = header.column.id === flexColumnId
              const prev = headers[index - 1]?.column
              const canAdjust = (c: Column<T, unknown>) => c.id === flexColumnId || c.getCanResize()
              const showHandle = prev && canAdjust(prev) && (!flexColumnId || canAdjust(header.column))
              return (
                <th
                  key={header.id}
                  className={cn('relative overflow-hidden', header.column.columnDef.meta?.headerClassName)}
                  style={isFlex ? { minWidth: minSizeOf(header.column) } : { width: widthOf(header.column) }}
                >
                  {showHandle && <ColumnResizeHandle onMouseDown={startDrag(prev, header.column)} />}
                  {header.isPlaceholder ? null : renderTemplate(header.column.columnDef.header, header.getContext())}
                </th>
              )
            })}
          </tr>
        ))}
      </thead>
      <tbody>
        {loading ? (
          <tr>
            <td colSpan={colSpan}>{loadingContent}</td>
          </tr>
        ) : error ? (
          <tr>
            <td colSpan={colSpan}>{errorContent}</td>
          </tr>
        ) : rows.length === 0 ? (
          <tr>
            <td colSpan={colSpan}>{emptyContent}</td>
          </tr>
        ) : (
          rows.map((row) => {
            const expanded = renderExpandedRow?.(row.original, row.index)
            return (
              <Fragment key={row.id}>
                <tr
                  className={cn(
                    'group border-b border-border text-sm transition-colors last:border-0',
                    onRowClick && 'cursor-pointer hover:bg-muted/40',
                    rowClassName?.(row.original, row.index),
                  )}
                  onClick={() => onRowClick?.(row.original, row.index)}
                >
                  {row.getVisibleCells().map((cell) => {
                    const meta = cell.column.columnDef.meta
                    const props = meta?.cellProps?.(row.original)
                    return (
                      <td key={cell.id} {...props} className={cn('overflow-hidden', meta?.cellClassName, props?.className)}>
                        {renderTemplate(cell.column.columnDef.cell, cell.getContext())}
                      </td>
                    )
                  })}
                </tr>
                {expanded && (
                  <tr>
                    <td colSpan={colSpan} className="border-b border-border/50 p-0">
                      {expanded}
                    </td>
                  </tr>
                )}
              </Fragment>
            )
          })
        )}
      </tbody>
    </table>
  )
}

/**
 * TanStack's flexRender mounts `header`/`cell` functions as components, and
 * callers build them inline, so every render would hand React a new component
 * type and remount every cell (losing menu/focus state). Calling them as
 * plain functions keeps the element tree stable. The flip side: they can't
 * call hooks — put stateful bits in a component they render.
 */
function renderTemplate<P extends object>(template: ColumnDefTemplate<P> | undefined, props: P): ReactNode {
  return typeof template === 'function' ? (template as (p: P) => ReactNode)(props) : template
}

function readSizing(storageKey?: string): ColumnSizingState {
  if (!storageKey) return {}
  try {
    const stored: unknown = JSON.parse(localStorage.getItem(storageKey) ?? 'null')
    if (stored && typeof stored === 'object' && !Array.isArray(stored)) {
      return Object.fromEntries(Object.entries(stored).filter(([, width]) => typeof width === 'number' && Number.isFinite(width)))
    }
  } catch {
    /* corrupt or unavailable — start from the defaults */
  }
  return {}
}

const minSizeOf = (column: Column<any, unknown>) => column.columnDef.minSize ?? 20

/**
 * Widths that fit `available` px: each column's `size` when there's room,
 * otherwise every column gives up the same fraction of its slack (size minus
 * minSize), floored so the flex column's leftover never overflows by a
 * fractional pixel.
 */
function fitWidths(columns: Column<any, unknown>[], available: number) {
  const prefs = columns.map((c) => ({ id: c.id, min: minSizeOf(c), size: Math.max(minSizeOf(c), c.getSize()) }))
  const total = prefs.reduce((sum, c) => sum + c.size, 0)
  const slack = prefs.reduce((sum, c) => sum + c.size - c.min, 0)
  const shrink = total > available && slack > 0 ? Math.min(1, (total - available) / slack) : 0
  return new Map(prefs.map((c) => [c.id, Math.floor(c.size - shrink * (c.size - c.min))]))
}

function useBoundaryDrag<T>(table: Table<T>, flexColumnId?: string) {
  return useCallback(
    (left: Column<T, unknown>, right: Column<T, unknown>) => (event: MouseEvent<HTMLElement>) => {
      event.preventDefault()
      event.stopPropagation()

      const leaves = table.getVisibleLeafColumns()
      const cells = Array.from(event.currentTarget.closest('tr')?.children ?? []) as HTMLElement[]
      const rendered = new Map(leaves.map((c, i) => [c.id, cells[i]?.offsetWidth ?? c.getSize()]))
      const flexPref = flexColumnId ? (table.getColumn(flexColumnId)?.getSize() ?? 0) : 0
      const conserve = flexColumnId !== undefined
      const lowest = minSizeOf(left) - rendered.get(left.id)!
      const highest = conserve ? rendered.get(right.id)! - minSizeOf(right) : Infinity
      const startX = event.clientX

      // Every column is pinned to what's on screen, so a fitted (shrunk)
      // layout doesn't re-fit around the two columns being dragged. The flex
      // column keeps its own size unless it has less room than that.
      const onMove = (e: globalThis.MouseEvent) => {
        const delta = Math.min(highest, Math.max(lowest, e.clientX - startX))
        const target = new Map(rendered)
        target.set(left.id, rendered.get(left.id)! + delta)
        if (conserve) target.set(right.id, rendered.get(right.id)! - delta)
        table.setColumnSizing(
          Object.fromEntries(
            leaves.map((c) => {
              const width = target.get(c.id)!
              return [c.id, c.id === flexColumnId ? Math.min(flexPref, width) : width]
            }),
          ),
        )
      }
      const onUp = () => {
        document.removeEventListener('mousemove', onMove)
        document.removeEventListener('mouseup', onUp)
      }
      document.addEventListener('mousemove', onMove)
      document.addEventListener('mouseup', onUp)
    },
    [table, flexColumnId],
  )
}
