import { cn } from '@/shared/lib/utils'

interface StatProps {
  label: string
  value: string
  tone?: string
}

export function Stat({ label, value, tone }: StatProps) {
  return (
    <div className="rounded-md border border-border bg-background/40 px-2 py-1.5">
      <div className="text-[9px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className={cn('mt-0.5 font-semibold tabular-nums', tone)}>{value}</div>
    </div>
  )
}
