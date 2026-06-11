import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Loader2 } from 'lucide-react'
import { EChartsRenderer } from '@/features/dashboard/components/EChartsRenderer'
import { MetricRenderer } from '@/features/dashboard/components/renderers/MetricRenderer'
import { TableRenderer } from '@/features/dashboard/components/renderers/TableRenderer'
import { TagCloudRenderer } from '@/features/dashboard/components/renderers/TagCloudRenderer'
import { RegionMapRenderer } from '@/features/dashboard/components/renderers/RegionMapRenderer'
import { TextRenderer } from '@/features/dashboard/components/renderers/TextRenderer'
import { useStagedVisualizationData } from '@/features/dashboard/hooks/useStagedVisualizationData'
import { mergeRowsIntoOption } from '@/features/dashboard/utils/echarts'
import { presetRange, TimeRangePicker, type TimeRange } from '@/shared/components/ui/time-range-picker'
import type { ChartRenderer } from '@/features/dashboard/constants'

function useDebouncedValue<T>(value: T, delayMs: number): T {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const id = window.setTimeout(() => setDebounced(value), delayMs)
    return () => window.clearTimeout(id)
  }, [value, delayMs])
  return debounced
}

export function ChartPreviewPanel({
  sql,
  option,
  renderer,
  label,
}: {
  sql: string
  option: Record<string, unknown>
  renderer: ChartRenderer
  label?: string
}) {
  const { t } = useTranslation()
  const [time, setTime] = useState<TimeRange>(() => presetRange('24h'))
  const debouncedSql = useDebouncedValue(sql, 600)
  const debouncedOption = useDebouncedValue(option, 600)

  const query = useStagedVisualizationData(debouncedSql, time)
  const rows = query.data?.rows ?? []

  const echartsOption = useMemo(
    () => mergeRowsIntoOption(debouncedOption, rows),
    [debouncedOption, rows]
  )

  return (
    <div className="flex h-full min-h-[320px] flex-col gap-2">
      <div className="flex items-center justify-between gap-2">
        <span className="text-xs font-medium text-foreground/80">
          {t('dashboards.editor.chartPreview.title')}
        </span>
        <TimeRangePicker value={time} onChange={setTime} align="right" />
      </div>
      <div className="min-h-[280px] flex-1 rounded-md border border-border bg-background/30 p-2">
        <Body
          sql={debouncedSql}
          renderer={renderer}
          option={echartsOption}
          rows={rows}
          label={label}
          isLoading={query.isLoading || query.isFetching}
          isError={query.isError}
          errorMessage={query.error instanceof Error ? query.error.message : undefined}
        />
      </div>
    </div>
  )
}

function Body({
  sql,
  renderer,
  option,
  rows,
  label,
  isLoading,
  isError,
  errorMessage,
}: {
  sql: string
  renderer: ChartRenderer
  option: Record<string, unknown>
  rows: Array<Record<string, unknown>>
  label?: string
  isLoading: boolean
  isError: boolean
  errorMessage?: string
}) {
  const { t } = useTranslation()

  if (!sql.trim()) {
    return (
      <div className="flex h-full w-full items-center justify-center text-xs text-muted-foreground">
        {t('dashboards.editor.chartPreview.noQuery')}
      </div>
    )
  }
  if (isLoading) {
    return (
      <div className="flex h-full w-full items-center justify-center gap-2 text-xs text-muted-foreground">
        <Loader2 size={14} className="animate-spin" />
        {t('dashboards.widget.loading')}
      </div>
    )
  }
  if (isError) {
    return (
      <div className="flex h-full w-full flex-col items-center justify-center gap-2 px-4 text-center text-xs text-muted-foreground">
        <AlertTriangle size={18} className="text-amber-500" />
        <span>{t('dashboards.editor.chartPreview.error')}</span>
        {errorMessage && (
          <span className="text-[10px] text-muted-foreground/70">{errorMessage}</span>
        )}
      </div>
    )
  }
  if (rows.length === 0) {
    return (
      <div className="flex h-full w-full items-center justify-center text-xs text-muted-foreground">
        {t('dashboards.widget.noData')}
      </div>
    )
  }

  if (renderer === 'table') return <TableRenderer rows={rows} />
  if (renderer === 'metric') return <MetricRenderer rows={rows} label={label} />
  if (renderer === 'tag_cloud') return <TagCloudRenderer rows={rows} />
  if (renderer === 'region_map') return <RegionMapRenderer rows={rows} />
  if (renderer === 'text') return <TextRenderer rows={rows} />
  return <EChartsRenderer option={option} />
}
