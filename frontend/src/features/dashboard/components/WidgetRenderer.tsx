import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { AlertTriangle, Loader2 } from 'lucide-react'
import { EChartsRenderer } from '@/features/dashboard/components/EChartsRenderer'
import { MetricRenderer } from '@/features/dashboard/components/renderers/MetricRenderer'
import { TableRenderer } from '@/features/dashboard/components/renderers/TableRenderer'
import { RegionMapRenderer } from '@/features/dashboard/components/renderers/RegionMapRenderer'
import { TextRenderer } from '@/features/dashboard/components/renderers/TextRenderer'
import { useVisualizationData } from '@/features/dashboard/hooks/useVisualizationData'
import { defaultChartSkeleton, mergeRowsIntoOption, parseChartConfig } from '@/features/dashboard/utils/echarts'
import { parseBuilderConfig } from '@/features/dashboard/utils/builder-config'
import { parseSpec, specIsComplete } from '@/features/dashboard/utils/spec'
import { resolveRangeToISO } from '@/features/dashboard/utils/datemath'
import { termFilter, timeBucketRange, widgetBaseFilters } from '@/features/dashboard/utils/click-through'
import { getChartTypeMeta, type ChartRenderer } from '@/features/dashboard/constants'
import type { FilterType, Visualization } from '@/features/dashboard/types'
import type { TimeRange } from '@/shared/components/ui/time-range-picker'

export function WidgetRenderer({
  visualization,
  time,
  filters,
  refreshMs,
  interactive = true,
  title,
}: {
  visualization: Visualization
  time: TimeRange
  filters?: FilterType[]
  refreshMs?: number | null
  interactive?: boolean
  title?: string
}) {
  const { t } = useTranslation()
  const navigate = useNavigate()

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
  const rows = query.data?.rows ?? []
  const clickBase = useMemo(() => {
    if (!spec) return null
    const { fromISO, toISO } = resolveRangeToISO(time.from, time.to)
    return {
      dataset: spec.dataset,
      dataType: spec.dataType,
      filters: widgetBaseFilters(spec, filters),
      range: { from: fromISO, to: toISO, interval: time.interval },
    }
  }, [spec, filters, time.from, time.to, time.interval])

  const openSearch = (
    additionalFilters: FilterType[] = [],
    range = clickBase?.range
  ) => {
    if (!clickBase) return
    const seed = {
      ...clickBase,
      filters: [...clickBase.filters, ...additionalFilters],
      ...(range ? { range } : {}),
    }
    if (spec?.dataset === 'alerts') {
      navigate('/threat-management/alerts', { state: { dashboardSearch: seed } })
      return
    }
    navigate('/log-explorer', { state: { dashboardSearch: seed } })
  }

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
        onRetry={() => void query.refetch()}
        retryLabel={t('common.actions.retry')}
      />
    )
  }

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
    const topValues = spec?.chart === 'category' && builderParsed.builder?.chartType === 'table'
    return (
      <TableRenderer
        rows={rows}
        dimension={topValues ? spec?.dimension : undefined}
        label={topValues ? title || builderParsed.builder?.title : undefined}
        onRowClick={
          interactive && clickBase
            ? (row) => {
              if (topValues && spec?.dimension) {
                const selected = termFilter(spec.dimension, row[spec.dimension])
                if (selected) openSearch([selected])
              } else if (row.id != null) {
                openSearch([{ field: 'id', operator: 'IS_ONE_OF_TERMS', value: [row.id] }])
                } else {
                  const rowFilters = Object.entries(row)
                    .filter(([field, value]) => field !== 'count' && value != null)
                    .map(([field, value]) => ({
                      field,
                      operator: 'IS_ONE_OF_TERMS' as const,
                      value: [value],
                    }))
                  openSearch(rowFilters)
                }
              }
            : undefined
        }
      />
    )
  }
  if (renderer === 'metric') {
    return <MetricRenderer rows={rows} onClick={interactive ? () => openSearch() : undefined} />
  }
  if (renderer === 'region_map') {
    return <RegionMapRenderer rows={rows} />
  }
  if (renderer === 'text') {
    return <TextRenderer rows={rows} />
  }

  const skeleton = defaultChartSkeleton(builderParsed.builder?.chartType, spec?.chart)
  const baseOption = { ...skeleton, ...builderParsed.option }
  const option = parsed.option ? mergeRowsIntoOption(baseOption, rows) : parsed.option!
  return (
    <EChartsRenderer
      option={option}
      onMarkClick={
        interactive && clickBase
          ? ({ dataIndex, seriesName }) => {
              if (dataIndex == null || dataIndex < 0 || dataIndex >= rows.length || !spec) return
              const row = rows[dataIndex]
              if (spec.chart === 'category' && spec.dimension && row[spec.dimension] != null) {
                const selected = termFilter(spec.dimension, row[spec.dimension])
                if (selected) openSearch([selected])
                return
              }

              if (spec.chart === 'time') {
                const splitFilters =
                  spec.dimension && seriesName
                    ? [{ field: spec.dimension, operator: 'IS_ONE_OF_TERMS' as const, value: [seriesName] }]
                    : []
                const adjacent = rows[dataIndex + 1]?.at ?? rows[dataIndex - 1]?.at
                const inferredDuration =
                  typeof row.at === 'string' && typeof adjacent === 'string'
                    ? Math.abs(
                        new Date(adjacent.replace(' ', 'T')).getTime() -
                          new Date(row.at.replace(' ', 'T')).getTime(),
                      )
                    : undefined
                const bucket = timeBucketRange(row.at, spec.interval, inferredDuration)
                const timeFilters =
                  bucket?.from && bucket.to
                    ? [{ field: '@timestamp', operator: 'IS_BETWEEN' as const, value: [bucket.from, bucket.to] }]
                    : []
                openSearch([...splitFilters, ...timeFilters])
              }
            }
          : undefined
      }
    />
  )
}

function ErrorPanel({
  message,
  detail,
  onRetry,
  retryLabel,
}: {
  message: string
  detail?: string
  onRetry?: () => void
  retryLabel?: string
}) {
  return (
    <div className="flex h-full w-full flex-col items-center justify-center gap-2 px-4 text-center text-xs text-muted-foreground">
      <AlertTriangle size={18} className="text-amber-500" />
      <span>{message}</span>
      {detail && <span className="text-[10px] text-muted-foreground/70">{detail}</span>}
      {onRetry && (
        <button onClick={onRetry} className="text-primary hover:underline">
          {retryLabel}
        </button>
      )}
    </div>
  )
}
