import { AGGREGATIONS } from '@/features/dashboard/constants'
import type { AggregationId, IndexProperty } from '@/features/dashboard/types'

// OpenSearch numeric mapping types.
const NUMERIC_TYPES = new Set([
  'long',
  'integer',
  'short',
  'byte',
  'double',
  'float',
  'half_float',
  'scaled_float',
  'unsigned_long',
])

const DATE_TYPES = new Set(['date', 'date_nanos'])

const norm = (type: string | undefined | null) => (type ?? '').trim().toLowerCase()

export function isNumericType(type: string | undefined | null): boolean {
  return NUMERIC_TYPES.has(norm(type))
}

export function isDateType(type: string | undefined | null): boolean {
  return DATE_TYPES.has(norm(type))
}

/**
 * Fields valid for a given aggregation in the visual builder.
 *
 * The builder is dropdown-driven and we own the generated SQL, so we must never
 * offer a combination that produces a broken/meaningless query — e.g. SUM/AVG
 * over a `text` field. This narrows the field list to the types each aggregation
 * actually supports.
 */
export function fieldsForAggregation(
  fields: IndexProperty[],
  agg: AggregationId
): IndexProperty[] {
  const kind = AGGREGATIONS.find((a) => a.id === agg)?.fieldKind ?? 'any'
  switch (kind) {
    case 'none':
      return []
    case 'numeric':
      return fields.filter((f) => isNumericType(f.type))
    case 'orderable':
      return fields.filter((f) => isNumericType(f.type) || isDateType(f.type))
    case 'any':
    default:
      return fields
  }
}
