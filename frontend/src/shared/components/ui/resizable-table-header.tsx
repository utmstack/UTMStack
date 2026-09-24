import type { CSSProperties, ReactNode } from 'react'
import type { ColSize, useResizableColumns } from '@/shared/hooks/useResizableColumns'
import { cn } from '@/shared/lib/utils'
import { ColumnResizeHandle } from './column-resize-handle'

export interface ResizableTableHeaderCell {
  content: ReactNode
  className?: string
  // Declines the resize handle for the boundary right before this column
  // (i.e. between the previous column and this one), not this column's own
  // right edge — matches where that boundary's handle actually renders now.
  resizable?: boolean
}

interface ResizableTableHeaderProps {
  cells: ResizableTableHeaderCell[]
  widths: Array<ColSize | undefined>
  // Per-column CSS min-width, used only for a column whose width is
  // `undefined` (the flex column) so it can't be squeezed to nothing by its
  // neighbors — pass the `mins` array useResizableColumns already returns.
  mins?: number[]
  startDrag: ReturnType<typeof useResizableColumns>['startDrag']
  className?: string
  rowClassName?: string
  cellClassName?: string
}

const cellStyle = (width: ColSize | undefined, min: number | undefined): CSSProperties | undefined => {
  if (width == null) return min != null ? { minWidth: `${min}px` } : undefined
  return { width: typeof width === 'number' ? `${width}px` : width }
}

export function ResizableTableHeader({
  cells,
  widths,
  mins,
  startDrag,
  className,
  rowClassName,
  cellClassName,
}: ResizableTableHeaderProps) {
  return (
    <thead className={className}>
      <tr className={rowClassName}>
        {cells.map((cell, index) => {
          // Rendered at the START of this column rather than the end of the
          // previous one (ColumnResizeHandle itself sits at left-0), so it's
          // flush against this column's label instead of trailing off the
          // end of the last one — still resizes the previous column (or, if
          // that's the flex column, the one after it — see startDrag).
          const prevResizable = index > 0 && (cells[index - 1].resizable ?? true)
          return (
            <th
              key={index}
              data-resizable-col
              className={cn('relative', cellClassName, cell.className)}
              style={cellStyle(widths[index], mins?.[index])}
            >
              {prevResizable && <ColumnResizeHandle onMouseDown={startDrag(index - 1)} />}
              {cell.content}
            </th>
          )
        })}
      </tr>
    </thead>
  )
}
