import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Loader2 } from 'lucide-react'
import { EChartsRenderer } from '@/features/dashboard/components/EChartsRenderer'
import { MetricRenderer } from '@/features/dashboard/components/renderers/MetricRenderer'
import { TableRenderer } from '@/features/dashboard/components/renderers/TableRenderer'
import { RegionMapRenderer } from '@/features/dashboard/components/renderers/RegionMapRenderer'
import { TextRenderer } from '@/features/dashboard/components/renderers/TextRenderer'
import { useVisualizationData } from '@/features/dashboard/hooks/useVisualizationData'
import { mergeRowsIntoOption, parseChartConfig } from '@/features/dashboard/utils/echarts'
import { parseBuilderConfig } from '@/features/dashboard/utils/builder-config'
import { parseSpec, specIsComplete } from '@/features/dashboard/utils/spec'
import { getChartTypeMeta, type ChartRenderer } from '@/features/dashboard/constants'
import type { FilterType, Visualization } from '@/features/dashboard/types'
import type { TimeRange } from '@/shared/components/ui/time-range-picker'

export function WidgetRenderer({
  visualization,
  time,
  filters,
  refreshMs,
}: {
  visualization: Visualization
  time: TimeRange
  filters?: FilterType[]
  refreshMs?: number | null
}) {
  const { t } = useTranslation()

  const parsed = useMemo(() => parseChartConfig(visualization.config), [visualization.config])
  const builderParsed = useMemo(
    () => parseBuilderConfig(visualization.config),
    [visualization.config]
  )
  const renderer: ChartRenderer = builderParsed.builder
    ? getChartTypeMeta(builderParsed.builder.chartType).renderer
    : 'echarts'

  const spec = useMemo(() => parseSpec(visualization.spec), [visualization.spec])
  const answerable = spec != null && specIsComplete(spec)
  const query = useVisualizationData(answerable ? visualization : null, time, filters, refreshMs ?? null)

  if (renderer === 'echarts' && (parsed.error || !parsed.option)) {
    return (
      <ErrorPanel
        message={t('dashboards.widget.renderError')}
        detail={parsed.error ?? undefined}
      />
    )
  }

  if (answerable && query.isLoading) {
    return (
      <div className="flex h-full w-full items-center justify-center gap-2 text-xs text-muted-foreground">
        <Loader2 size={14} className="animate-spin" />
        {t('dashboards.widget.loading')}
      </div>
    )
  }

  if (answerable && query.isError) {
    return (
      <ErrorPanel
        message={t('dashboards.widget.queryError')}
        detail={query.error instanceof Error ? query.error.message : undefined}
      />
    )
  }

  const rows = query.data?.rows ?? []

  if (!answerable) {
    return <ErrorPanel message={t('dashboards.widget.incompleteSpec')} />
  }

  if (rows.length === 0 && !query.isLoading) {
    return (
      <div className="flex h-full w-full items-center justify-center text-xs text-muted-foreground">
        {t('dashboards.widget.noData')}
      </div>
    )
  }

  if (renderer === 'table') {
    return <TableRenderer rows={rows} />
  }
  if (renderer === 'metric') {
    return <MetricRenderer rows={rows} />
  }
  if (renderer === 'region_map') {
    return <RegionMapRenderer rows={rows} />
  }
  if (renderer === 'text') {
    return <TextRenderer rows={rows} />
  }

  const option = parsed.option ? mergeRowsIntoOption(parsed.option, rows) : parsed.option!
  return <EChartsRenderer option={option} />
}

function ErrorPanel({ message, detail }: { message: string; detail?: string }) {
  return (
    <div className="flex h-full w-full flex-col items-center justify-center gap-2 px-4 text-center text-xs text-muted-foreground">
      <AlertTriangle size={18} className="text-amber-500" />
      <span>{message}</span>
      {detail && <span className="text-[10px] text-muted-foreground/70">{detail}</span>}
    </div>
  )
}
