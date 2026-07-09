import { useTranslation } from 'react-i18next'
import { fmtCount, shortTime } from './helpers'

export function EventsChart({ points, loading }: { points: { t: string; count: number }[]; loading: boolean }) {
  const { t } = useTranslation()
  const w = 700
  const h = 180
  const pl = 44
  const pr = 16
  const pt = 10
  const pb = 24
  const innerW = w - pl - pr
  const innerH = h - pt - pb

  if (loading) {
    return <div className="mx-4 h-44 animate-pulse rounded-lg bg-muted" />
  }
  if (points.length < 2) {
    return (
      <div className="flex h-44 items-center justify-center text-xs text-muted-foreground">
        {t('home.datasources.noData')}
      </div>
    )
  }

  const counts = points.map((p) => p.count)
  const max = Math.max(...counts, 1)
  const xs = points.map((_, i) => pl + (i * innerW) / (points.length - 1))
  const ys = points.map((v) => pt + innerH - (v.count / max) * innerH)
  const linePath = points.map((_, i) => `${i === 0 ? 'M' : 'L'} ${xs[i]} ${ys[i]}`).join(' ')
  const areaPath = `${linePath} L ${xs[xs.length - 1]} ${pt + innerH} L ${xs[0]} ${pt + innerH} Z`

  // Two reference gridlines (half + full scale) with compact labels.
  const grids = [max / 2, max]
  // A handful of evenly-spaced x labels so the axis stays readable.
  const labelEvery = Math.max(1, Math.ceil(points.length / 6))

  return (
    <svg viewBox={`0 0 ${w} ${h}`} className="h-44 w-full" preserveAspectRatio="none">
      <defs>
        <pattern id="hatch" width="6" height="6" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
          <line x1="0" y1="0" x2="0" y2="6" className="stroke-foreground/5" strokeWidth="1" />
        </pattern>
        <linearGradient id="lineGrad" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0" stopColor="#22d3ee" />
          <stop offset="1" stopColor="#1a8cff" />
        </linearGradient>
      </defs>
      <rect x={pl} y={pt} width={innerW} height={innerH} fill="url(#hatch)" />
      <g className="fill-muted-foreground" fontSize="10">
        {grids.map((v) => (
          <text key={v} x={pl - 6} y={pt + innerH - (v / max) * innerH + 3} textAnchor="end">
            {fmtCount(Math.round(v))}
          </text>
        ))}
      </g>
      {grids.map((v) => (
        <line
          key={v}
          x1={pl}
          x2={pl + innerW}
          y1={pt + innerH - (v / max) * innerH}
          y2={pt + innerH - (v / max) * innerH}
          className="stroke-border/40"
          strokeDasharray="2 4"
        />
      ))}
      <path d={areaPath} className="fill-primary/10" />
      <path d={linePath} fill="none" stroke="url(#lineGrad)" strokeWidth="2" strokeLinejoin="round" />
      <g className="fill-muted-foreground" fontSize="10">
        {points.map((p, i) =>
          i % labelEvery === 0 ? (
            <text key={i} x={xs[i]} y={h - 6} textAnchor="middle">
              {shortTime(p.t)}
            </text>
          ) : null,
        )}
      </g>
    </svg>
  )
}
