import { flattenDoc } from '@/shared/lib/flatten'
import type {
  BuilderState,
  ChartTypeId,
  FilterOperatorId,
  FilterRow,
  FilterType,
  IntervalId,
  QueryResult,
  Row,
  SpecChart,
  SpecFilter,
  SpecOp,
  VizSpec,
} from '@/features/dashboard/types'

/**
 * A widget asks one question and the store answers it. This is both
 * directions of that: the builder's state becomes a spec, and the answer
 * becomes the rows every renderer already reads.
 */

// The filter vocabulary the UI speaks, in the store's own words. Anchored
// matching is kept apart from "contains" because a prefix is not the same
// question, and answering one as the other returns records nobody asked for.
const OPS: Record<FilterOperatorId, SpecOp> = {
  IS: 'eq',
  IS_NOT: 'not_eq',
  IS_ONE_OF_TERMS: 'in',
  IS_GREATER_THAN: 'gt',
  IS_LESS_THAN_OR_EQUALS: 'lte',
  IS_BETWEEN: 'between',
  CONTAIN: 'contains',
  DOES_NOT_CONTAIN: 'not_contains',
  START_WITH: 'starts_with',
  ENDS_WITH: 'ends_with',
  EXIST: 'exists',
  DOES_NOT_EXIST: 'not_exists',
}

const VALUELESS = new Set<SpecOp>(['exists', 'not_exists'])

function toSpecFilter(field: string, operator: FilterOperatorId, value: unknown): SpecFilter | null {
  const op = OPS[operator]
  if (!op || !field.trim()) return null
  if (VALUELESS.has(op)) return { field, op }
  if (value == null || value === '') return null
  if (Array.isArray(value) && value.length === 0) return null
  return { field, op, value }
}

/** The builder's own filter rows — the ones saved with the widget. */
export function filterRowsToSpec(rows: FilterRow[]): SpecFilter[] {
  return rows
    .map((r) => toSpecFilter(r.field, r.operator, r.value))
    .filter((f): f is SpecFilter => f != null)
}

/** The dashboard's filter bar — chosen at view time, not saved. */
export function filterTypesToSpec(filters: FilterType[]): SpecFilter[] {
  return filters
    .map((f) => toSpecFilter(f.field, f.operator, f.value))
    .filter((f): f is SpecFilter => f != null)
}

/**
 * Which shape of answer a chart type needs. A metric is one number, a table is
 * records; everything else plots a series, over time or over the top values of
 * a field. The map is drawn from records, so it asks for them.
 */
export function specChartFor(chartType: ChartTypeId, breakdown: BuilderState['breakdown']): SpecChart {
  switch (chartType) {
    case 'metric':
      return 'metric'
    case 'table':
    case 'list':
    case 'region_map':
    case 'text':
      return 'table'
    default:
      return breakdown === 'time' ? 'time' : 'category'
  }
}

export function builderToSpec(builder: BuilderState): VizSpec {
  const chart = specChartFor(builder.chartType, builder.breakdown)
  const spec: VizSpec = {
    dataset: builder.dataset,
    chart,
    metric: { agg: 'count' },
  }

  const filters = filterRowsToSpec(builder.filters)
  if (filters.length > 0) spec.filters = filters
  if (builder.limit != null) spec.limit = builder.limit

  // A category chart is broken down by the field; a time chart may be split by
  // one, and is a single line when it is not.
  if (chart === 'category' || chart === 'time') {
    if (builder.dimension) spec.dimension = builder.dimension
  }
  if (chart === 'time' && builder.interval) spec.interval = builder.interval
  if (chart === 'table' && builder.columns.length > 0) spec.columns = builder.columns

  return spec
}

/** What the editor reopens with. `chartType` lives in the widget's config. */
export function specToBuilder(spec: VizSpec): Partial<BuilderState> {
  return {
    dataset: spec.dataset ?? '',
    breakdown: spec.chart === 'time' ? 'time' : 'field',
    dimension: spec.dimension ?? null,
    interval: (spec.interval ?? '') as IntervalId,
    limit: spec.limit ?? null,
    columns: spec.columns ?? [],
  }
}

export function parseSpec(raw: string | null | undefined): VizSpec | null {
  if (!raw || !raw.trim()) return null
  try {
    const parsed = JSON.parse(raw)
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return null
    return parsed as VizSpec
  } catch {
    return null
  }
}

/** A spec is answerable once it names a dataset and whatever its chart needs. */
export function specIsComplete(spec: VizSpec): boolean {
  if (!spec.dataset?.trim()) return false
  if (spec.chart === 'category' && !spec.dimension?.trim()) return false
  return true
}

/**
 * The answer as rows. Every renderer and every saved ECharts config reads a
 * flat array of records — keeping that shape is what lets a spec-driven widget
 * use the chart configs that already exist.
 */
export function resultToRows(result: QueryResult | undefined, spec: VizSpec): Row[] {
  if (!result) return []

  switch (spec.chart) {
    case 'metric':
      return result.total == null ? [] : [{ value: result.total }]

    case 'category': {
      const label = spec.dimension || 'key'
      return (result.buckets ?? []).map((b) => ({ [label]: b.key, count: b.count }))
    }

    case 'time': {
      const fmt = timeFormatter(spec.interval)
      if (result.series?.length) return seriesToRows(result.series, fmt)
      return (result.points ?? []).map((p) => ({ at: fmt(p.at), count: p.count }))
    }

    case 'table':
      return (result.rows ?? []).map((doc) => project(flattenDoc(doc), spec.columns))
  }

  return []
}

/**
 * Split series become one row per instant with a column per series, because
 * that is what an ECharts dataset draws as several lines. A series missing an
 * instant means no records then, which is zero — leaving it out would make the
 * line jump the gap.
 */
function seriesToRows(
  series: NonNullable<QueryResult['series']>,
  fmt: (iso: string) => string
): Row[] {
  const byInstant = new Map<string, Row>()
  const instants: string[] = []

  for (const s of series) {
    for (const p of s.points) {
      if (!byInstant.has(p.at)) {
        byInstant.set(p.at, { at: fmt(p.at) })
        instants.push(p.at)
      }
    }
  }
  instants.sort()

  for (const s of series) {
    const seen = new Set<string>()
    for (const p of s.points) {
      const row = byInstant.get(p.at)
      if (row) row[s.key] = p.count
      seen.add(p.at)
    }
    for (const at of instants) {
      if (!seen.has(at)) {
        const row = byInstant.get(at)
        if (row) row[s.key] = 0
      }
    }
  }

  return instants.map((at) => byInstant.get(at)!).filter(Boolean)
}

/**
 * A record's fields under the names the widget asked for. The store returns
 * nested documents, so a column is a path (`origin.geolocation.country`); the
 * header shows its last segment unless two columns would then read the same.
 */
function project(flat: Record<string, unknown>, columns: string[] | undefined): Row {
  if (!columns || columns.length === 0) return flat

  const leaves = columns.map((c) => c.split('.').pop() ?? c)
  const out: Row = {}
  for (let i = 0; i < columns.length; i++) {
    const leaf = leaves[i]
    const unique = leaves.filter((l) => l === leaf).length === 1
    out[unique ? leaf : columns[i]] = flat[columns[i]] ?? null
  }
  return out
}

// Buckets a day or wider are read as dates; anything finer needs the time, and
// the year is noise on a chart whose x axis spans hours.
function timeFormatter(interval: string | undefined): (iso: string) => string {
  const dateOnly = interval === '1d' || interval === '1w'
  return (iso: string) => {
    const d = new Date(iso)
    if (Number.isNaN(d.getTime())) return iso
    const pad = (n: number) => String(n).padStart(2, '0')
    const date = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
    return dateOnly ? date : `${date} ${pad(d.getHours())}:${pad(d.getMinutes())}`
  }
}
