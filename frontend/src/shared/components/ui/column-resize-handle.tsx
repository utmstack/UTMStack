import type { MouseEvent } from 'react'
import { GripVertical } from 'lucide-react'
import { cn } from '@/shared/lib/utils'

interface Props {
  onMouseDown: (e: MouseEvent<HTMLElement>) => void
  className?: string
}

export function ColumnResizeHandle({ onMouseDown, className }: Props) {
  return (
    <span
      role="separator"
      aria-orientation="vertical"
      onMouseDown={onMouseDown}
      onClick={(e) => e.stopPropagation()}
      className={cn(
        'group absolute right-0 top-0 z-20 flex h-full w-3 cursor-col-resize select-none items-center justify-center touch-none',
        'bg-transparent hover:bg-primary/20 active:bg-primary/30',
        'transition-colors',
        className,
      )}
      aria-label="Resize column"
    >
      <GripVertical className="h-3 w-3 text-muted-foreground/45 opacity-70 transition-colors group-hover:text-primary group-hover:opacity-100" strokeWidth={1.75} />
    </span>
  )
}
