import { useTranslation } from 'react-i18next'

export function AlertsVolumeCard({ values, total }: { values: number[]; total: number }) {
  const { t } = useTranslation()
  const max = Math.max(1, ...values)

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
        // Flex columns instead of a stretched SVG. Each bar's height is a % of the
        // tallest bucket; a per-bar max-width + centered row means even a single
        // populated bucket reads as one slim column, never a full-width solid block.
        <div className="mt-4 flex min-h-[120px] flex-1 items-end justify-center gap-[2px]">
          {values.map((v, i) => (
            <div
              key={i}
              className="h-full max-w-[16px] flex-1 self-stretch flex items-end"
            >
              <div
                className="w-full rounded-t-sm bg-primary/55 transition-colors hover:bg-primary"
                style={{ height: `${v <= 0 ? 0 : Math.max(3, (v / max) * 100)}%` }}
                title={v.toLocaleString()}
              />
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
