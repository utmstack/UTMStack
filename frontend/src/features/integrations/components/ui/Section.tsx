import { AlertTriangle } from 'lucide-react'
import { cn } from '@/shared/lib/utils'

export function Section({
  id,
  title,
  step,
  variant,
  children,
}: {
  id?: string
  title: string
  step?: number
  variant?: 'warning'
  children: React.ReactNode
}) {
  const isStep = step != null
  const isWarning = variant === 'warning'

  return (
    <div
      id={id}
      className={cn(
        'rounded-lg border border-border bg-card p-4 scroll-mt-6',
        isStep && 'border-l-[3px] border-l-primary/60',
        isWarning && 'border-l-[3px] border-l-amber-500/70',
      )}
    >
      <div className="mb-3 flex items-center gap-2.5">
        {isStep && (
          <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-primary/15 text-[10px] font-bold tabular-nums text-primary ring-1 ring-inset ring-primary/25">
            {step}
          </span>
        )}
        {isWarning && (
          <AlertTriangle size={14} className="shrink-0 text-amber-500" />
        )}
        <h4 className="text-sm font-semibold leading-tight">{title}</h4>
      </div>
      {children}
    </div>
  )
}
