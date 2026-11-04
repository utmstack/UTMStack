import type { FilterType, SpecFilter, SpecOp, VizSpec } from '@/features/dashboard/types'
import type { TimeRange } from '@/shared/components/ui/time-range-picker'

const SPEC_OPERATOR_TO_FILTER: Record<SpecOp, FilterType['operator']> = {
  eq: 'IS_ONE_OF_TERMS',
  not_eq: 'IS_NOT',
  in: 'IS_ONE_OF_TERMS',
  not_in: 'IS_NOT_ONE_OF_TERMS',
  gt: 'IS_GREATER_THAN',
  gte: 'IS_GREATER_THAN_OR_EQUALS',
  lt: 'IS_LESS_THAN',
  lte: 'IS_LESS_THAN_OR_EQUALS',
  between: 'IS_BETWEEN',
  not_between: 'IS_NOT_BETWEEN',
  contains: 'CONTAIN',
  not_contains: 'DOES_NOT_CONTAIN',
  starts_with: 'START_WITH',
  not_starts_with: 'DOES_NOT_START_WITH',
  ends_with: 'ENDS_WITH',
  not_ends_with: 'DOES_NOT_END_WITH',
  exists: 'EXIST',
  not_exists: 'DOES_NOT_EXIST',
}

export function specFiltersToLogFilters(filters: SpecFilter[] = []): FilterType[] {
  return filters.map((filter) => {
    const value =
      filter.op === 'eq' && filter.value !== undefined && !Array.isArray(filter.value)
        ? [filter.value]
        : filter.value
    return {
      field: filter.field,
      operator: SPEC_OPERATOR_TO_FILTER[filter.op],
      value,
    }
  })
}

export function widgetBaseFilters(spec: VizSpec, dashboardFilters: FilterType[] = []): FilterType[] {
  const filters = specFiltersToLogFilters(spec.filters)
  if (spec.dataType) {
    filters.unshift({ field: 'dataType', operator: 'IS_ONE_OF_TERMS', value: [spec.dataType] })
  }
  return [...filters, ...dashboardFilters]
}

export function termFilter(field: string, value: unknown): FilterType | null {
  if (value == null) return null
  return { field, operator: 'IS_ONE_OF_TERMS', value: [value] }
}

export function timeBucketRange(
  label: unknown,
  interval: string | undefined,
  inferredDurationMs?: number,
): TimeRange | null {
  if (typeof label !== 'string') return null
  const start = new Date(label.replace(' ', 'T'))
  if (Number.isNaN(start.getTime())) return null

  const duration = interval ? intervalDuration(interval) : inferredDurationMs ?? intervalDuration(undefined)
  const from = start.toISOString()
  const to = new Date(start.getTime() + duration - 1).toISOString()
  return { from, to, interval: interval === '1d' || interval === '1w' ? 'day' : 'hour' }
}

function intervalDuration(interval: string | undefined): number {
  switch (interval) {
    case '1m': return 60_000
    case '5m': return 5 * 60_000
    case '15m': return 15 * 60_000
    case '1d': return 24 * 60 * 60_000
    case '1w': return 7 * 24 * 60 * 60_000
    case '1h':
    default: return 60 * 60_000
  }
}
