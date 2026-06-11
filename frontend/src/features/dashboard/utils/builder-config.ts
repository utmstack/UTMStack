import { CHART_TYPES, DEFAULT_BUILDER_METRIC } from '@/features/dashboard/constants'
import type { BuilderState, ChartTypeId } from '@/features/dashboard/types'

export interface ParsedBuilderConfig {
  option: Record<string, unknown>
  builder: BuilderState | null
}

const BUILDER_KEY = '__builder'

function isRecord(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === 'object' && !Array.isArray(value)
}

export function parseBuilderConfig(raw: string | undefined | null): ParsedBuilderConfig {
  if (!raw || !raw.trim()) {
    return { option: {}, builder: null }
  }
  try {
    const parsed = JSON.parse(raw)
    if (!isRecord(parsed)) return { option: {}, builder: null }
    const rest: Record<string, unknown> = {}
    let builder: BuilderState | null = null
    for (const [k, v] of Object.entries(parsed)) {
      if (k === BUILDER_KEY) {
        if (isRecord(v)) builder = sanitizeBuilder(v)
      } else {
        rest[k] = v
      }
    }
    return { option: rest, builder }
  } catch {
    return { option: {}, builder: null }
  }
}

export function serializeBuilderConfig(
  option: Record<string, unknown>,
  builder: BuilderState | null
): string {
  const merged: Record<string, unknown> = { ...option }
  if (builder) merged[BUILDER_KEY] = builder
  return JSON.stringify(merged, null, 2)
}

function sanitizeBuilder(raw: Record<string, unknown>): BuilderState {
  const chartType = (raw.chartType as ChartTypeId) ?? 'bar'
  const validChart = CHART_TYPES.some((c) => c.id === chartType) ? chartType : 'bar'
  const filtersRaw = Array.isArray(raw.filters) ? raw.filters : []
  const filters = filtersRaw
    .filter(isRecord)
    .map((f, idx) => ({
      id: typeof f.id === 'string' ? f.id : `f-${idx}-${Math.random().toString(36).slice(2, 8)}`,
      field: typeof f.field === 'string' ? f.field : '',
      operator: (typeof f.operator === 'string' ? f.operator : 'IS') as BuilderState['filters'][number]['operator'],
      value: (f.value ?? null) as BuilderState['filters'][number]['value'],
    }))
  const metricRaw = isRecord(raw.metric) ? raw.metric : null
  const metric = metricRaw
    ? {
        agg: (metricRaw.agg as BuilderState['metric']['agg']) ?? 'count',
        field: typeof metricRaw.field === 'string' ? metricRaw.field : null,
      }
    : { ...DEFAULT_BUILDER_METRIC }

  return {
    chartType: validChart as ChartTypeId,
    indexPattern: typeof raw.indexPattern === 'string' ? raw.indexPattern : '',
    rawMode: !!raw.rawMode,
    filters,
    metric,
    dimension: typeof raw.dimension === 'string' ? raw.dimension : null,
    advancedSelect: typeof raw.advancedSelect === 'string' ? raw.advancedSelect : null,
    rawSql: typeof raw.rawSql === 'string' ? raw.rawSql : null,
    configTouched: !!raw.configTouched,
  }
}

export function makeInitialBuilder(): BuilderState {
  return {
    chartType: 'bar',
    indexPattern: '',
    rawMode: false,
    filters: [],
    metric: { ...DEFAULT_BUILDER_METRIC },
    dimension: null,
    advancedSelect: null,
    rawSql: null,
    configTouched: false,
  }
}

export function hashString(value: string): number {
  let h = 0
  for (let i = 0; i < value.length; i++) {
    h = (h << 5) - h + value.charCodeAt(i)
    h |= 0
  }
  return h
}
