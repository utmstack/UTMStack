import type { CSSProperties, ReactNode } from 'react'
import type { ColSize, useResizableColumns } from '@/shared/hooks/useResizableColumns'
import { cn } from '@/shared/lib/utils'
import { ColumnResizeHandle } from './column-resize-handle'

export interface ResizableTableHeaderCell {
  content: ReactNode
  className?: string
}

interface ResizableTableHeaderProps {
  cells: ResizableTableHeaderCell[]
  widths: ColSize[]
  startDrag: ReturnType<typeof useResizableColumns>['startDrag']
  className?: string
  rowClassName?: string
  cellClassName?: string
}

const widthStyle = (width: ColSize | undefined): CSSProperties | undefined => {
  if (width == null) return undefined
  return { width: typeof width === 'number' ? `${width}px` : width }
}

export function ResizableTableHeader({
  cells,
  widths,
  startDrag,
  className,
  rowClassName,
  cellClassName,
}: ResizableTableHeaderProps) {
  return (
    <thead className={className}>
      <tr className={rowClassName}>
        {cells.map((cell, index) => (
          <th
            key={index}
            data-resizable-col
            className={cn('relative', cellClassName, cell.className)}
            style={widthStyle(widths[index])}
          >
            {cell.content}
            {index < cells.length - 1 && <ColumnResizeHandle onMouseDown={startDrag(index)} />}
          </th>
        ))}
      </tr>
    </thead>
  )
}
