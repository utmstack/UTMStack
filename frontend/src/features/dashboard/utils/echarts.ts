import type { Row } from '@/features/dashboard/service/opensearch.service'

export interface ParsedChartConfig {
  option: Record<string, unknown> | null
  error: string | null
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
  const grid = option.grid ?? (hasAxes ? DEFAULT_GRID : undefined)

  return {
    ...option,
    ...(grid ? { grid } : {}),
    dataset: {
      ...existingDataset,
      source: rows,
    },
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === 'object' && !Array.isArray(value)
}
