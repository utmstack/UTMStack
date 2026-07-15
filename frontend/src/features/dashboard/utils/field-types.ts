import { AGGREGATIONS, OPERATORS, type OperatorMeta } from '@/features/dashboard/constants'
import type { AggregationId, FilterOperatorId, IndexProperty } from '@/features/dashboard/types'

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

// Operators that need a LIKE-style string comparison — OpenSearch SQL's `LIKE`
// only accepts string operands, so pointing one at a date/numeric field errors
// (e.g. `like function expected {[STRING,STRING]}, but get [TIMESTAMP,STRING]`).
const STRING_MATCH_OPERATORS = new Set<FilterOperatorId>([
  'CONTAIN',
  'DOES_NOT_CONTAIN',
  'START_WITH',
  'ENDS_WITH',
])

// Comparison operators that only make sense on something sortable.
const ORDERABLE_OPERATORS = new Set<FilterOperatorId>([
  'IS_GREATER_THAN',
  'IS_LESS_THAN_OR_EQUALS',
  'IS_BETWEEN',
])

// Valid on any field type: equality doesn't care about representation, and
// exist/does-not-exist just checks for a value at all.
const UNIVERSAL_OPERATORS = new Set<FilterOperatorId>(['IS', 'IS_NOT', 'EXIST', 'DOES_NOT_EXIST'])

/**
 * Operators valid for a given field type in the filter builder.
 *
 * Mirrors {@link fieldsForAggregation}'s guardrail: the builder is
 * dropdown-driven and owns the generated SQL, so it must never offer a
 * field/operator combination OpenSearch will reject — e.g. `CONTAIN` (LIKE) on
 * a `date` field, or `IS_BETWEEN` on plain text.
 *
 * `type` is `undefined` while the field list is still loading or for a field
 * that isn't in the current index pattern — don't over-restrict in that case,
 * the field-staleness check elsewhere handles the latter.
 */
export function operatorsForFieldType(type: string | undefined | null): OperatorMeta[] {
  const t = norm(type)
  if (!t) return OPERATORS
  return OPERATORS.filter((o) => {
    if (UNIVERSAL_OPERATORS.has(o.id)) return true
    if (o.id === 'IS_ONE_OF_TERMS') return t !== 'text'
    if (STRING_MATCH_OPERATORS.has(o.id)) return t === 'text' || t === 'keyword'
    if (ORDERABLE_OPERATORS.has(o.id)) return isNumericType(t) || isDateType(t)
    return true
  })
}
