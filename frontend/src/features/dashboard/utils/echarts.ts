import { getChartTypeMeta } from '@/features/dashboard/constants'
import type { ChartTypeId, Row, SpecChart } from '@/features/dashboard/types'

export interface ParsedChartConfig {
  option: Record<string, unknown> | null
  error: string | null
}

/**
 * mergeRowsIntoOption never invents a `series` (see below) — a config that
 * names a chart type but never actually wrote the ECharts series/axes
 * skeleton merges into a legend with nothing to plot. This is that skeleton,
 * picked from the builder's chart type when one was saved, or failing that
 * from the spec's chart shape (category/time only — metric and table never
 * reach this renderer, see WidgetRenderer). A config that already carries
 * its own series/axes overwrites this when spread on top, so it only ever
 * fills a real gap.
 */
export function defaultChartSkeleton(
  builderChartType: ChartTypeId | undefined,
  specChart: SpecChart | undefined
): Record<string, unknown> {
  if (builderChartType) return getChartTypeMeta(builderChartType).defaultConfig
  if (specChart === 'category') return getChartTypeMeta('bar').defaultConfig
  if (specChart === 'time') return getChartTypeMeta('line').defaultConfig
  return {}
}

export function parseChartConfig(raw: string): ParsedChartConfig {
  if (!raw || !raw.trim()) {
    return { option: null, error: 'empty config' }
  }
  try {
    const parsed = JSON.parse(raw)
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
      return { option: null, error: 'config is not an object' }
    }
    return { option: parsed as Record<string, unknown>, error: null }
  } catch (err) {
    return { option: null, error: err instanceof Error ? err.message : 'invalid JSON' }
  }
}

// Sensible plot margins for cartesian charts. ECharts' built-in grid leaves ~60px
// top/bottom, which squashes the plot into a thin band on short containers (e.g.
// the editor preview). `containLabel` still grows margins to fit axis labels.
const DEFAULT_GRID = { left: 8, right: 16, top: 16, bottom: 16, containLabel: true }

export function mergeRowsIntoOption(
  option: Record<string, unknown>,
  rows: Row[]
): Record<string, unknown> {
  const existingDataset = isRecord(option.dataset) ? option.dataset : {}

  // Only cartesian charts (with x/y axes) use a grid; pie/gauge/etc. don't.
  const hasAxes =
    'xAxis' in option || 'yAxis' in option

  // A bottom legend on every chart, so categories/series are identifiable at a
  // glance. Skipped only when the saved config already defines its own legend.
  const addLegend = !('legend' in option)
  const legend = {
    show: true,
    type: 'scroll' as const,
    bottom: 4,
    left: 'center' as const,
    icon: 'circle',
    itemWidth: 9,
    itemHeight: 9,
    itemGap: 12,
    textStyle: { fontSize: 11 },
  }

  // Cartesian charts need extra bottom room so the legend doesn't overlap the
  // x-axis labels.
  const grid =
    option.grid ??
    (hasAxes ? { ...DEFAULT_GRID, bottom: addLegend ? 40 : DEFAULT_GRID.bottom } : undefined)

  const series = normalizeSeries(expandSeries(option.series, rows))
  // Pie charts have no axes, so the only way to read a slice is on hover.
  const tooltip = option.tooltip ?? (seriesHasType(series, 'pie') ? { trigger: 'item' } : undefined)

  return {
    ...option,
    ...(grid ? { grid } : {}),
    ...(series !== undefined ? { series } : {}),
    ...(tooltip ? { tooltip } : {}),
    ...(addLegend ? { legend } : {}),
    dataset: {
      ...existingDataset,
      source: rows,
    },
  }
}

/**
 * A split-by-field answer arrives as one column per series (`at`, "windows",
 * "linux"…), while a saved config carries a single series drawn from the second
 * column. Repeat that series once per value column so every line is drawn; the
 * column name becomes the series name, which is what the legend shows.
 */
function expandSeries(series: unknown, rows: Row[]): unknown {
  if (rows.length === 0) return series
  const columns = Object.keys(rows[0])
  if (columns.length <= 2) return series

  const template = Array.isArray(series) ? series[0] : series
  if (!isRecord(template) || Array.isArray(series) && series.length > 1) return series
  if (!isRecord(template.encode) || !('x' in (template.encode as Record<string, unknown>))) {
    return series
  }

  return columns.slice(1).map((name, i) => ({
    ...template,
    name,
    encode: { ...(template.encode as Record<string, unknown>), y: i + 1 },
  }))
}

// Pie slices get cluttered, overlapping leader-line labels by default (one per
// category). Hide the always-on labels + lines and instead reveal the slice
// name — in white — only when the slice is hovered. Respect an explicit `label`
// if the saved config set one.
function normalizePieSeries(series: unknown): unknown {
  if (!isRecord(series) || series.type !== 'pie') return series
  const next: Record<string, unknown> = { ...series }
  // Hide the always-on labels + leader lines; reveal the slice name (white) on
  // hover. Respect an explicit `label` if the saved config set one.
  if (!('label' in next)) {
    const emphasis = isRecord(next.emphasis) ? next.emphasis : {}
    next.label = { show: false }
    next.labelLine = { show: false }
    next.emphasis = {
      ...emphasis,
      label: {
        show: true,
        color: '#ffffff',
        fontWeight: 'bold',
        ...(isRecord(emphasis.label) ? emphasis.label : {}),
      },
    }
  }
  // Lift the pie slightly so it doesn't overlap the bottom legend.
  if (!('center' in next)) next.center = ['50%', '45%']
  return next
}

function normalizeSeries(series: unknown): unknown {
  if (Array.isArray(series)) return series.map(normalizePieSeries)
  if (isRecord(series)) return normalizePieSeries(series)
  return series
}

function seriesHasType(series: unknown, type: string): boolean {
  if (Array.isArray(series)) return series.some((s) => isRecord(s) && s.type === type)
  return isRecord(series) && series.type === type
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === 'object' && !Array.isArray(value)
}
