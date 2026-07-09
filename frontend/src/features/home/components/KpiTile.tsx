import { Link } from 'react-router-dom'
import { type LucideIcon } from 'lucide-react'
import { Sparkline } from './Sparkline'

export function KpiTile({
  icon: Icon,
  label,
  value,
  sublabel,
  sparkline,
  accent,
  loading,
  href,
}: {
  icon: LucideIcon
  label: string
  value: string
  sublabel?: string
  sparkline?: number[]
  accent: string
  loading?: boolean
  href?: string
}) {
  const inner = (
    <div className="flex h-full flex-col rounded-xl border border-border bg-card p-4 transition-colors hover:border-primary/40">
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <Icon size={14} strokeWidth={1.75} className={accent} />
        {label}
      </div>
      {loading ? (
        <div className="mt-2 h-8 w-16 animate-pulse rounded bg-muted" />
      ) : (
        <div className="mt-2 text-3xl font-semibold tracking-tight tabular-nums">{value}</div>
      )}
      <div className="mt-3 -mx-1 min-h-[36px]">
        {sparkline && sparkline.length > 1 ? (
          <Sparkline data={sparkline} accent={accent} />
        ) : sublabel ? (
          <span className="text-xs text-muted-foreground">{sublabel}</span>
        ) : null}
      </div>
    </div>
  )
  return href ? (
    <Link to={href} className="block h-full">
      {inner}
    </Link>
  ) : (
    inner
  )
}
