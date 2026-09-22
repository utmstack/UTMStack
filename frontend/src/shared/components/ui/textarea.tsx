import * as React from 'react'
import { cn } from '@/shared/lib/utils'

export interface TextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  /** Grows with content up to this many rows, then scrolls instead. Default 8. */
  maxRows?: number
}

/** Textarea that grows with its content instead of staying a fixed size. */
const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, maxRows = 8, onChange, value, rows = 3, ...props }, forwardedRef) => {
    const innerRef = React.useRef<HTMLTextAreaElement>(null)
    React.useImperativeHandle(forwardedRef, () => innerRef.current as HTMLTextAreaElement)

    const resize = React.useCallback(() => {
      const el = innerRef.current
      if (!el) return
      el.style.height = 'auto'
      const lineHeight = parseFloat(getComputedStyle(el).lineHeight || '20') || 20
      const maxHeight = lineHeight * maxRows
      el.style.height = `${Math.min(el.scrollHeight, maxHeight)}px`
      el.style.overflowY = el.scrollHeight > maxHeight ? 'auto' : 'hidden'
    }, [maxRows])

    React.useLayoutEffect(() => {
      resize()
    }, [value, resize])

    return (
      <textarea
        ref={innerRef}
        value={value}
        rows={rows}
        onChange={(e) => {
          onChange?.(e)
          resize()
        }}
        className={cn(
          'flex w-full resize-none rounded-md border border-input bg-background/40 px-3 py-2 text-sm',
          'placeholder:text-muted-foreground',
          'transition-colors',
          'focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring focus-visible:border-ring',
          'disabled:cursor-not-allowed disabled:opacity-50',
          className
        )}
        {...props}
      />
    )
  }
)
Textarea.displayName = 'Textarea'

export { Textarea }
