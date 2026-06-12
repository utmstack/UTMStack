import { useTranslation } from 'react-i18next'

export function AlertsVolumeCard({ values, total }: { values: number[]; total: number }) {
  const { t } = useTranslation()
  const max = Math.max(1, ...values)
  const w = 1000
  const h = 90
  const n = values.length || 1
  const slot = w / n
  const bw = Math.max(1, slot - 2)
  return (
    <div className="flex h-full flex-col rounded-xl border border-border bg-card p-6">
      <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{t('alerts.overview.overTime')}</div>
      <div className="mt-1 flex items-baseline gap-2">
        <span className="text-3xl font-semibold tabular-nums">{total.toLocaleString()}</span>
        <span className="text-sm text-muted-foreground">{t('alerts.overview.matchingAlerts')}</span>
      </div>
      {values.length === 0 ? (
        <div className="mt-4 flex min-h-[100px] flex-1 items-center justify-center text-xs text-muted-foreground">
          {t('alerts.overview.noData')}
        </div>
      ) : (
        <div className="mt-4 min-h-[100px] flex-1">
          <svg viewBox={`0 0 ${w} ${h}`} className="h-full w-full" preserveAspectRatio="none">
            {values.map((v, i) => {
              if (v <= 0) return null
              const bh = Math.max(2, (v / max) * h)
              return (
                <rect
                  key={i}
                  x={i * slot + (slot - bw) / 2}
                  y={h - bh}
                  width={bw}
                  height={bh}
                  rx={1.5}
                  className="fill-primary/55"
                />
              )
            })}
          </svg>
        </div>
      )}
    </div>
  )
}
