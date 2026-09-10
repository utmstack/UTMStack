import type { MouseEvent } from 'react'
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
        'absolute right-0 top-0 z-20 h-full w-2 cursor-col-resize select-none touch-none',
        'bg-transparent hover:bg-primary/40 active:bg-primary',
        'transition-colors',
        className,
      )}
      aria-label="Resize column"
    />
  )
}
