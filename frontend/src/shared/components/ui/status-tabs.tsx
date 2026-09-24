import type { ReactNode } from 'react'
import { cn } from '@/shared/lib/utils'

export interface StatusTab<T extends string> {
  id: T
  label: string
  count?: number
}

/** Underlined tab strip with an optional slot aligned to its right end. */
export function StatusTabs<T extends string>({
  tabs,
  current,
  onChange,
  actions,
}: {
  tabs: StatusTab<T>[]
  current: T
  onChange: (id: T) => void
  actions?: ReactNode
}) {
  return (
    <div className="flex flex-wrap items-center gap-1 border-b border-border">
      {tabs.map(({ id, label, count }) => {
        const active = current === id
        return (
          <button
            key={id}
            onClick={() => onChange(id)}
            className={cn(
              'relative flex items-center gap-2 px-3 py-2 text-xs transition-colors',
              active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
            )}
          >
            {label}
            {count != null && (
              <span
                className={cn(
                  'rounded-md px-1.5 py-0.5 font-mono text-[10px] tabular-nums',
                  active ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground'
                )}
              >
                {count}
              </span>
            )}
            {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
          </button>
        )
      })}
      {actions && <div className="ml-auto pb-1">{actions}</div>}
    </div>
  )
}
