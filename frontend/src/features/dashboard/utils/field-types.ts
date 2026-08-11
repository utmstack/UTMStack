import { OPERATORS, type OperatorMeta } from '@/features/dashboard/constants'
import type { FilterOperatorId } from '@/features/dashboard/types'

/**
 * Field types as the event store reports them — ClickHouse column types, with
 * `Nullable(…)` and `LowCardinality(…)` wrapped around them. Everything here
 * unwraps first, so `LowCardinality(String)` is a string like any other.
 */
export function normalizeType(type: string | undefined | null): string {
  let t = (type ?? '').trim()
  for (;;) {
    const m = /^(Nullable|LowCardinality|SimpleAggregateFunction)\((.*)\)$/i.exec(t)
    if (!m) break
    t = m[2].trim()
  }
  return t
}

const NUMERIC = /^(U?Int\d+|Float\d+|Decimal|Bool)/i
const DATE = /^(Date|DateTime)/i
const TEXT = /^(String|FixedString|Enum|UUID|IPv[46])/i

export function isNumericType(type: string | undefined | null): boolean {
  return NUMERIC.test(normalizeType(type))
}

export function isDateType(type: string | undefined | null): boolean {
  return DATE.test(normalizeType(type))
}

export function isTextType(type: string | undefined | null): boolean {
  return TEXT.test(normalizeType(type))
}

/** A list column: `errors`, `references`. One record holds several values. */
export function isArrayType(type: string | undefined | null): boolean {
  return /^Array\(/i.test(normalizeType(type))
}

/**
 * A whole sub-document (`origin`, `log`), not a value. Its own paths are
 * reported next to it — `origin.geolocation.country` — and those are what a
 * widget can group by or filter on.
 */
export function isContainerType(type: string | undefined | null): boolean {
  const t = normalizeType(type)
  return t === '' || /^(JSON|Tuple|Map|Object)/i.test(t)
}

// Operators that match text against text. Pointing one at a date or a number
// asks a question with no answer.
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
 * The builder is dropdown-driven, so it must never offer a field/operator pair
 * that cannot be answered — `CONTAIN` on a date, `IS_BETWEEN` on plain text.
 *
 * `type` is `undefined` while the field list is still loading or for a field
 * the current dataset does not carry — don't over-restrict in that case, the
 * field-staleness check elsewhere handles the latter.
 */
export function operatorsForFieldType(type: string | undefined | null): OperatorMeta[] {
  const t = normalizeType(type)
  if (!t) return OPERATORS
  return OPERATORS.filter((o) => {
    if (UNIVERSAL_OPERATORS.has(o.id)) return true
    if (STRING_MATCH_OPERATORS.has(o.id)) return isTextType(t) || isArrayType(t)
    if (ORDERABLE_OPERATORS.has(o.id)) return isNumericType(t) || isDateType(t)
    return true
  })
}
