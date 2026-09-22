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
        const isResizable = item.resizable ?? true
        return (
          <div key={index} data-resizable-col className={cn('relative min-w-0 pr-2 last:pr-0', cellClassName)}>
            {item.content}
            {isResizable && index < headers.length - 1 && <ColumnResizeHandle onMouseDown={startDrag(index)} />}
          </div>
        )
      })}
    </div>
  )
}
