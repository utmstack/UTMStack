import { useTranslation } from 'react-i18next'
import type { ChartView } from '../types/log-explorer.types'

export function TermsChart({ data }: { data: ChartView }) {
  const { t } = useTranslation()
  const max = Math.max(1, ...data.values)
  return (
    <div className="space-y-2">
      {data.categories.map((cat, i) => {
        const v = data.values[i] ?? 0
        const pct = (v / max) * 100
        return (
          <div key={`${cat}-${i}`} className="flex items-center gap-3 text-[13px]">
            <div className="w-56 shrink-0 truncate font-mono text-foreground" title={cat}>
              {cat || t('logExplorer.fields.empty')}
            </div>
            <div className="relative h-6 flex-1 overflow-hidden rounded bg-muted/50">
              <div className="h-full rounded bg-primary/40" style={{ width: `${Math.max(1.5, pct)}%` }} />
            </div>
            <div className="w-24 shrink-0 text-right font-mono tabular-nums text-muted-foreground">
              {v.toLocaleString()}
            </div>
          </div>
        )
      })}
    </div>
  )
}
