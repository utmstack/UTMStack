import type {
  AggregationId,
  BuilderMetric,
  BuilderState,
  FilterRow,
  FilterOperatorId,
} from '@/features/dashboard/types'

function escapeSqlString(value: string): string {
  return value.replace(/'/g, "''")
}

function quote(value: string): string {
  return `'${escapeSqlString(value)}'`
}

function asString(value: unknown): string {
  return value == null ? '' : String(value)
}

export function aggregationToSelect(metric: BuilderMetric): string {
  switch (metric.agg) {
    case 'count':
      return 'COUNT(*)'
    case 'count_distinct':
      return metric.field ? `COUNT(DISTINCT ${metric.field})` : 'COUNT(*)'
    // SUM/AVG/MIN/MAX need a column — `SUM(*)` is invalid SQL. If no field is
    // selected yet, fall back to COUNT(*) so the builder never emits broken SQL.
    case 'sum':
      return metric.field ? `SUM(${metric.field})` : 'COUNT(*)'
    case 'avg':
      return metric.field ? `AVG(${metric.field})` : 'COUNT(*)'
    case 'min':
      return metric.field ? `MIN(${metric.field})` : 'COUNT(*)'
    case 'max':
      return metric.field ? `MAX(${metric.field})` : 'COUNT(*)'
    default:
      return 'COUNT(*)'
  }
}

export function filterRowToWhere(row: FilterRow): string | null {
  const field = row.field?.trim()
  if (!field) return null
  const op: FilterOperatorId = row.operator

  switch (op) {
    case 'EXIST':
      return `${field} IS NOT NULL`
    case 'DOES_NOT_EXIST':
      return `${field} IS NULL`
    case 'IS_BETWEEN': {
      const v = row.value
      if (!Array.isArray(v) || v.length < 2) return null
      const a = asString(v[0]).trim()
      const b = asString(v[1]).trim()
      if (!a || !b) return null
      return `${field} BETWEEN ${quote(a)} AND ${quote(b)}`
    }
    case 'IS_ONE_OF_TERMS': {
      const list = Array.isArray(row.value) ? row.value : []
      const items = list.map((x) => asString(x).trim()).filter(Boolean)
      if (items.length === 0) return null
      return `${field} IN (${items.map(quote).join(',')})`
    }
    default: {
      const value = asString(row.value).trim()
      if (!value) return null
      switch (op) {
        case 'IS':
          return `${field} = ${quote(value)}`
        case 'IS_NOT':
          return `${field} <> ${quote(value)}`
        case 'IS_GREATER_THAN':
          return `${field} > ${quote(value)}`
        case 'IS_LESS_THAN_OR_EQUALS':
          return `${field} <= ${quote(value)}`
        case 'CONTAIN':
          return `${field} LIKE ${quote(`%${value}%`)}`
        case 'DOES_NOT_CONTAIN':
          return `${field} NOT LIKE ${quote(`%${value}%`)}`
        case 'START_WITH':
          return `${field} LIKE ${quote(`${value}%`)}`
        case 'ENDS_WITH':
          return `${field} LIKE ${quote(`%${value}`)}`
        default:
          return null
      }
    }
  }
}

function buildWhereClause(filters: FilterRow[]): string {
  const parts: string[] = []
  for (const f of filters) {
    const fragment = filterRowToWhere(f)
    if (fragment) parts.push(fragment)
  }
  const userClauses = parts.length > 0 ? `${parts.join(' AND ')} AND ` : ''
  return `WHERE ${userClauses}{{dashboardFilters}}{{timeFilter}}`
}

function hasWhereClause(sql: string): boolean {
  return /\bWHERE\b/i.test(sql)
}

/**
 * Best-effort inverse of {@link composeSql}. Recognises the exact shapes that
 * {@link composeSql} emits (metric / dimension-grouped / table SELECTs) and
 * returns a partial builder patch. Anything the parser doesn't understand — a
 * hand-rolled WHERE clause, unusual quoting, custom joins — is left alone so
 * the caller keeps its existing builder state for those fields.
 *
 * Filters and chart type are intentionally NOT reversed: WHERE parsing is
 * fragile and chartType would flip the UI in surprising ways.
 */
export function parseSqlToBuilder(sql: string | null | undefined): Partial<BuilderState> | null {
  if (!sql || !sql.trim()) return null
  const s = sql.trim()

  const patch: Partial<BuilderState> = {}

  const fromMatch = s.match(/\bFROM\s+([^\s\n;]+)/i)
  if (fromMatch) {
    patch.indexPattern = fromMatch[1].replace(/^[`"']|[`"']$/g, '')
  }

  const selectMatch = s.match(/\bSELECT\s+([\s\S]*?)\s+FROM\b/i)
  if (!selectMatch) return patch
  const selectExpr = selectMatch[1].trim()

  const hasGroupBy = /\bGROUP\s+BY\b/i.test(s)
  const hasAsY = /\bAS\s+y\b/i.test(selectExpr)

  if (hasGroupBy && hasAsY) {
    const m = selectExpr.match(/^(.+?)\s+AS\s+x\s*,\s*(.+?)\s+AS\s+y$/i)
    if (m) {
      const dimExpr = m[1].trim()
      patch.dimension = dimExpr && dimExpr.toUpperCase() !== 'NULL' ? dimExpr : null
      const metric = parseMetricExpr(m[2].trim())
      if (metric) patch.metric = metric
    }
  } else if (hasAsY) {
    const m = selectExpr.match(/^(.+?)\s+AS\s+y$/i)
    if (m) {
      const metric = parseMetricExpr(m[1].trim())
      if (metric) patch.metric = metric
    }
  } else if (!hasGroupBy) {
    // Plain projection — treat as a table SELECT column list.
    if (selectExpr === '*') {
      patch.columns = []
    } else {
      const cols = selectExpr.split(',').map((c) => c.trim()).filter(Boolean)
      if (cols.length > 0) patch.columns = cols
    }
  }

  return patch
}

function parseMetricExpr(expr: string): BuilderMetric | null {
  const clean = expr.trim()
  if (/^COUNT\(\s*\*\s*\)$/i.test(clean)) return { agg: 'count', field: null }
  const distinct = clean.match(/^COUNT\(\s*DISTINCT\s+([^)]+?)\s*\)$/i)
  if (distinct) return { agg: 'count_distinct', field: distinct[1].trim() }
  const other = clean.match(/^(SUM|AVG|MIN|MAX)\(\s*([^)]+?)\s*\)$/i)
  if (other) {
    const agg = other[1].toLowerCase() as AggregationId
    return { agg, field: other[2].trim() }
  }
  return null
}

export function composeSql(state: BuilderState): string {
  if (state.rawMode && state.rawSql && state.rawSql.trim()) {
    return state.rawSql.trim()
  }
  const indexPattern = state.indexPattern?.trim()
  if (!indexPattern) return ''

  const where = buildWhereClause(state.filters)
  const metricExpr = aggregationToSelect(state.metric)

  if (state.chartType === 'metric') {
    return [`SELECT ${metricExpr} AS y`, `FROM ${indexPattern}`, where].join('\n')
  }

  if (state.chartType === 'table') {
    // Legacy raw SELECT (kept for widgets saved before the columns picker).
    const advanced = state.advancedSelect?.trim()
    if (advanced) {
      if (hasWhereClause(advanced)) return advanced
      return `${advanced}\n${where}\nLIMIT 100`
    }
    const cols = (state.columns ?? []).map((c) => c.trim()).filter(Boolean)
    const select = cols.length > 0 ? cols.join(', ') : '*'
    return [`SELECT ${select}`, `FROM ${indexPattern}`, where, `LIMIT 100`].join('\n')
  }

  const dimensionExpr = (state.dimension?.trim() || 'NULL').trim()
  return [
    `SELECT ${dimensionExpr} AS x, ${metricExpr} AS y`,
    `FROM ${indexPattern}`,
    where,
    `GROUP BY x`,
    `ORDER BY y DESC`,
    `LIMIT 100`,
  ].join('\n')
}
