import { cn } from '@/shared/lib/utils'

export function Toggle({ on, onChange }: { on: boolean; onChange: (v: boolean) => void }) {
  return (
    <button
      onClick={(e) => { e.stopPropagation(); onChange(!on) }}
      className={cn('relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors', on ? 'bg-emerald-500' : 'bg-muted-foreground/30')}
      role="switch"
      aria-checked={on}
    >
      <span className={cn('inline-block h-4 w-4 transform rounded-full bg-white transition-transform', on ? 'translate-x-4' : 'translate-x-0.5')} />
    </button>
  )
}
