import { useMemo } from 'react'

export interface MatchOverviewCardProps {
  total?: number
}

export function MatchOverviewCard({ total = 0 }: MatchOverviewCardProps) {
  // ponytail: hardcoded sparkline until CM exposes a timeseries endpoint
  const buckets = 48
  const data = useMemo(
    () =>
      Array.from({ length: buckets }, (_, i) =>
        Math.max(
          0,
          Math.round(Math.abs(Math.sin(i / 5)) * 18 + Math.abs(Math.sin(i / 11)) * 8 + (i === 38 ? 12 : 0) + 3)
        )
      ),
    []
  )
  const w = 1000
  const h = 100
  const max = Math.max(...data) * 1.15
  const xs = data.map((_, i) => (i * w) / (data.length - 1))
  const ys = data.map((v) => h - (v / max) * h)
  const linePath = data.reduce((acc, _, i) => {
    if (i === 0) return `M ${xs[i]} ${ys[i]}`
    const prevX = xs[i - 1]
    const prevY = ys[i - 1]
    const cx1 = prevX + (xs[i] - prevX) / 2
    const cx2 = xs[i] - (xs[i] - prevX) / 2
    return `${acc} C ${cx1} ${prevY}, ${cx2} ${ys[i]}, ${xs[i]} ${ys[i]}`
  }, '')
  const areaPath = `${linePath} L ${xs[xs.length - 1]} ${h} L ${xs[0]} ${h} Z`

  return (
    <div className="rounded-xl border border-border bg-card p-6">
      <div className="flex items-baseline justify-between">
        <div>
          <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
            IOC matches · last 24 hours
          </div>
          <div className="mt-1 flex items-baseline gap-3">
            <span className="text-3xl font-semibold tabular-nums">{total.toLocaleString()}</span>
            <span className="text-sm text-muted-foreground">total indicators</span>
          </div>
        </div>
      </div>

      <svg viewBox={`0 0 ${w} ${h}`} className="mt-4 h-24 w-full" preserveAspectRatio="none">
        <defs>
          <linearGradient id="iocGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="rgb(168 85 247)" stopOpacity="0.28" />
            <stop offset="100%" stopColor="rgb(168 85 247)" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d={areaPath} fill="url(#iocGrad)" />
        <path
          d={linePath}
          fill="none"
          stroke="rgb(168 85 247)"
          strokeOpacity="0.85"
          strokeWidth="1.75"
          strokeLinejoin="round"
          strokeLinecap="round"
        />
      </svg>

      <div className="mt-2 flex justify-between text-[10px] text-muted-foreground">
        <span>24h ago</span>
        <span>12h ago</span>
        <span>now</span>
      </div>
    </div>
  )
}
