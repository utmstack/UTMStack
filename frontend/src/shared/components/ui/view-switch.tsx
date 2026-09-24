import type { LucideIcon } from 'lucide-react'
import { cn } from '@/shared/lib/utils'

export interface ViewSwitchOption<T extends string> {
  id: T
  icon: LucideIcon
  label: string
}

/** The segmented icon+label switch that sits at the top right of a page. */
export function ViewSwitch<T extends string>({
  value,
  onChange,
  options,
}: {
  value: T
  onChange: (v: T) => void
  options: ViewSwitchOption<T>[]
}) {
  return (
    <div className="inline-flex shrink-0 items-center rounded-md border border-border bg-card p-0.5 text-xs">
      {options.map(({ id, icon: Icon, label }) => (
        <button
          key={id}
          onClick={() => onChange(id)}
          className={cn(
            'flex items-center gap-1.5 rounded px-2.5 py-1 transition-colors',
            value === id ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'
          )}
        >
          <Icon size={13} /> {label}
        </button>
      ))}
    </div>
  )
}
