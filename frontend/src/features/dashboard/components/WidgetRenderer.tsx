import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle } from 'lucide-react'
import { EChartsRenderer } from '@/features/dashboard/components/EChartsRenderer'
import { parseChartConfig } from '@/features/dashboard/utils/echarts'
import type { Visualization } from '@/features/dashboard/types'
import type { TimeRange } from '@/shared/components/ui/time-range-picker'

export function WidgetRenderer({
  visualization,
  time,
}: {
  visualization: Visualization
  time: TimeRange
}) {
  const { t } = useTranslation()
  const { option, error } = useMemo(
    () => parseChartConfig(visualization.config, time),
    [visualization.config, time]
  )

  if (error || !option) {
    return (
      <div className="flex h-full w-full flex-col items-center justify-center gap-2 px-4 text-center text-xs text-muted-foreground">
        <AlertTriangle size={18} className="text-amber-500" />
        <span>{t('dashboards.widget.renderError')}</span>
        {error && <span className="text-[10px] text-muted-foreground/70">{error}</span>}
      </div>
    )
  }

  return <EChartsRenderer option={option} />
}
