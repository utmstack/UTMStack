import type { ReactNode } from 'react'
import type { useResizableColumns } from '@/shared/hooks/useResizableColumns'
import { cn } from '@/shared/lib/utils'
import { ColumnResizeHandle } from './column-resize-handle'

interface ResizableGridHeaderProps {
  headers: Array<ReactNode | { content: ReactNode; resizable?: boolean }>
  tableCols: string
  startDrag: ReturnType<typeof useResizableColumns>['startDrag']
  className?: string
  cellClassName?: string
}

export function ResizableGridHeader({
  headers,
  tableCols,
  startDrag,
  className,
  cellClassName,
}: ResizableGridHeaderProps) {
  return (
    <div
      className={cn(
        'grid w-max min-w-full items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground',
        className,
      )}
      style={{ gridTemplateColumns: tableCols }}
    >
      {headers.map((header, index) => {
        const item = typeof header === 'object' && header !== null && 'content' in header ? header : { content: header, resizable: true }
        // See resizable-table-header.tsx: rendered at the start of this
        // column, so the flag that matters is the PREVIOUS column's — it's
        // declining the handle for the boundary right before this one.
        const prevItem =
          index > 0
            ? (typeof headers[index - 1] === 'object' && headers[index - 1] !== null && 'content' in (headers[index - 1] as object)
                ? (headers[index - 1] as { resizable?: boolean })
                : { resizable: true })
            : null
        const prevResizable = prevItem != null && (prevItem.resizable ?? true)
        return (
          <div key={index} data-resizable-col className={cn('relative min-w-0 pl-3 first:pl-0', cellClassName)}>
            {prevResizable && <ColumnResizeHandle onMouseDown={startDrag(index - 1)} />}
            {item.content}
          </div>
        )
      })}
    </div>
  )
}
