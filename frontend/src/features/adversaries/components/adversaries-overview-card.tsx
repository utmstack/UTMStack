import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import type { Adversary } from '../types/adversary.types'

export function AdversariesOverviewCard({ adversaries }: { adversaries: Adversary[] }) {
  const { t } = useTranslation()
  const summed = useMemo(() => {
    const b = new Array(24).fill(0)
    for (const a of adversaries) for (let i = 0; i < 24; i++) b[i] += a.activity[i] ?? 0
    return b
  }, [adversaries])
  const high = adversaries.filter((a) => a.maxSeverity >= 3).length
  const countries = new Set(adversaries.map((a) => a.geo?.country).filter(Boolean)).size
  const totalAlerts = adversaries.reduce((s, a) => s + a.alertsCount, 0)

  const w = 1000
  const h = 100
  const max = Math.max(...summed, 1) * 1.15
  const xs = summed.map((_, i) => (i * w) / (summed.length - 1))
  const ys = summed.map((v) => h - (v / max) * h)
  const line = summed.reduce((acc, _, i) => {
    if (i === 0) return `M ${xs[i]} ${ys[i]}`
    const cx1 = xs[i - 1] + (xs[i] - xs[i - 1]) / 2
    const cx2 = xs[i] - (xs[i] - xs[i - 1]) / 2
    return `${acc} C ${cx1} ${ys[i - 1]}, ${cx2} ${ys[i]}, ${xs[i]} ${ys[i]}`
  }, '')
  const area = `${line} L ${xs[xs.length - 1]} ${h} L ${xs[0]} ${h} Z`

  return (
    <div className="rounded-xl border border-border bg-card p-6">
      <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{t('adversaries.overview.title')}</div>
      <div className="mt-1 flex items-baseline gap-3">
        <span className="text-3xl font-semibold tabular-nums">{adversaries.length}</span>
        <span className="text-sm text-muted-foreground">
          <span className="text-foreground">{high}</span> {t('adversaries.overview.high')}
          <span className="mx-2 text-border">·</span>
          <span className="text-foreground">{countries}</span> {t('adversaries.overview.countries')}
          <span className="mx-2 text-border">·</span>
          <span className="text-foreground">{totalAlerts}</span> {t('adversaries.overview.alerts')}
        </span>
      </div>
      <svg viewBox={`0 0 ${w} ${h}`} className="mt-4 h-24 w-full" preserveAspectRatio="none">
        <defs>
          <linearGradient id="advGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="rgb(239 68 68)" stopOpacity="0.25" />
            <stop offset="100%" stopColor="rgb(239 68 68)" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d={area} fill="url(#advGrad)" />
        <path
          d={line}
          fill="none"
          stroke="rgb(239 68 68)"
          strokeOpacity="0.8"
          strokeWidth="1.75"
          strokeLinejoin="round"
          strokeLinecap="round"
        />
      </svg>
      <div className="mt-2 flex justify-between text-[10px] text-muted-foreground">
        <span>{t('adversaries.overview.ago24')}</span>
        <span>{t('adversaries.overview.ago12')}</span>
        <span>{t('adversaries.overview.now')}</span>
      </div>
    </div>
  )
}
