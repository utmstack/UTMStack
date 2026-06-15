import type {
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
    case 'sum':
      return metric.field ? `SUM(${metric.field})` : 'SUM(*)'
    case 'avg':
      return metric.field ? `AVG(${metric.field})` : 'AVG(*)'
    case 'min':
      return metric.field ? `MIN(${metric.field})` : 'MIN(*)'
    case 'max':
      return metric.field ? `MAX(${metric.field})` : 'MAX(*)'
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
  return `WHERE ${userClauses}{{timeFilter}}`
}

function hasWhereClause(sql: string): boolean {
  return /\bWHERE\b/i.test(sql)
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
