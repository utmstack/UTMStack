import type { FilterType } from '../types/log-explorer.types'
import type { TimeRange } from '@/shared/components/ui/time-range-picker'

const TS = '@timestamp'
const UNIT_MS: Record<string, number> = {
  m: 60_000,
  h: 3_600_000,
  d: 86_400_000,
  w: 604_800_000,
  M: 2_592_000_000,
}

// A statement, not a phrase: it opens with SELECT or WITH *and* names a source.
// Requiring both is what keeps "select" — a reasonable thing to search logs for
// when hunting injection — from being mistaken for a query.
const SQL_SHAPE = /^\s*(select|with)\b[\s\S]*\bfrom\b/i

/** Reports whether the text in the search box is meant as SQL. */
export function looksLikeSql(text: string): boolean {
  return SQL_SHAPE.test(text.trim())
}

const quote = (s: string) => `'${String(s).replace(/'/g, "''")}'`
const unquote = (s: string) => s.replace(/''/g, "'")

// The picker carries relative tokens; SQL takes instants. Non-tokens pass through.
function resolveDate(v: string, nowMs: number = Date.now()): string {
  if (v === 'now') return new Date(nowMs).toISOString()
  const m = /^now-(\d+)([mhdwM])$/.exec(v)
  if (!m) return v
  const ms = UNIT_MS[m[2]]
  return new Date(nowMs - Number(m[1]) * ms).toISOString()
}

// A path inside a JSON column reads as Dynamic, which cannot be compared to a
// string directly — the declared columns can. Wrapping every field is simpler
// than knowing which is which, and costs nothing on a declared one.
const col = (field: string) => (field === TS ? '`@timestamp`' : `toString(${field})`)

function filterToSql(f: FilterType): string | null {
  const field = col(f.field)
  const v = f.value
  switch (f.operator) {
    case 'IS':      return `${field} = ${quote(String(v ?? ''))}`
    case 'IS_NOT':  return `${field} <> ${quote(String(v ?? ''))}`
    case 'CONTAIN': return `${field} LIKE ${quote('%' + String(v ?? '') + '%')}`
    case 'EXIST':   return `${field} IS NOT NULL`
    case 'IS_ONE_OF_TERMS': {
      const arr = Array.isArray(v) ? v : v == null ? [] : [v]
      if (!arr.length) return null
      return `${field} IN (${arr.map((x) => quote(String(x))).join(', ')})`
    }
    default: return null // IS_BETWEEN / IS_IN_FIELDS handled outside as range/search
  }
}

/** Build the ClickHouse SQL the explorer's state describes. `logs` is the
 *  tenant-scoped dataset the backend hands the query. */
export function buildSql(
  dataset: string,
  dataType: string | null,
  range: TimeRange,
  filters: FilterType[],
  searchInput: string,
  nowMs?: number,
): string {
  const parts: string[] = []
  if (dataType) parts.push(`dataType = ${quote(dataType)}`)
  if (range.from) {
    parts.push(
      '`@timestamp` >= ' + quote(resolveDate(range.from, nowMs)) +
        ' AND `@timestamp` <= ' + quote(resolveDate(range.to, nowMs)),
    )
  }
  // Free text is a substring of the record as it arrived, which is what `raw`
  // holds. There is no full-text operator to reach for.
  const search = searchInput.trim()
  if (search) parts.push(`positionCaseInsensitive(raw, ${quote(search)}) > 0`)
  for (const f of filters) {
    if (f.field === TS && f.operator === 'IS_BETWEEN') continue
    if (f.operator === 'IS_IN_FIELDS') continue
    const s = filterToSql(f)
    if (s) parts.push(s)
  }
  const where = parts.length ? ` WHERE ${parts.join(' AND ')}` : ''
  return `SELECT * FROM ${dataset}${where} ORDER BY \`@timestamp\` DESC`
}

export interface ParsedSql {
  dataset: string
  dataType: string | null
  range: TimeRange
  filters: FilterType[]
  searchInput: string
}

// Split "a AND b AND c" respecting quoted strings and parentheses.
function splitAnd(s: string): string[] {
  const out: string[] = []
  let depth = 0
  let inStr = false
  let start = 0
  for (let i = 0; i < s.length; i++) {
    const ch = s[i]
    if (ch === "'") {
      if (inStr && s[i + 1] === "'") { i++; continue }
      inStr = !inStr
      continue
    }
    if (inStr) continue
    if (ch === '(') depth++
    else if (ch === ')') depth--
    else if (
      depth === 0 &&
      (ch === 'A' || ch === 'a') &&
      /^and(\s|$)/i.test(s.slice(i, i + 4))
    ) {
      out.push(s.slice(start, i))
      i += 2
      start = i + 1
    }
  }
  out.push(s.slice(start))
  return out.map((c) => c.trim()).filter(Boolean)
}

type Clause =
  | { kind: 'dataType'; value: string }
  | { kind: 'timeFrom'; value: string }
  | { kind: 'timeTo'; value: string }
  | { kind: 'search'; value: string }
  | { kind: 'filter'; filter: FilterType }

// Field references come back through toString(...) or backticks; strip either.
const bare = (f: string) => {
  const m = /^toString\((.+)\)$/.exec(f)
  return (m ? m[1] : f).replace(/^`|`$/g, '')
}

function parseClause(c: string): Clause | null {
  let m: RegExpExecArray | null
  if ((m = /^dataType\s*=\s*'((?:[^']|'')*)'$/i.exec(c)))
    return { kind: 'dataType', value: unquote(m[1]) }
  if ((m = /^`@timestamp`\s*>=\s*'((?:[^']|'')*)'$/i.exec(c)))
    return { kind: 'timeFrom', value: unquote(m[1]) }
  if ((m = /^`@timestamp`\s*<=\s*'((?:[^']|'')*)'$/i.exec(c)))
    return { kind: 'timeTo', value: unquote(m[1]) }
  if ((m = /^positionCaseInsensitive\(\s*raw\s*,\s*'((?:[^']|'')*)'\s*\)\s*>\s*0$/i.exec(c)))
    return { kind: 'search', value: unquote(m[1]) }
  if ((m = /^(\S+)\s+IS\s+NOT\s+NULL$/i.exec(c)))
    return { kind: 'filter', filter: { field: bare(m[1]), operator: 'EXIST' } }
  if ((m = /^(\S+)\s+LIKE\s+'%((?:[^']|'')*)%'$/i.exec(c)))
    return { kind: 'filter', filter: { field: bare(m[1]), operator: 'CONTAIN', value: unquote(m[2]) } }
  if ((m = /^(\S+)\s*<>\s*'((?:[^']|'')*)'$/.exec(c)))
    return { kind: 'filter', filter: { field: bare(m[1]), operator: 'IS_NOT', value: unquote(m[2]) } }
  if ((m = /^(\S+)\s*!=\s*'((?:[^']|'')*)'$/.exec(c)))
    return { kind: 'filter', filter: { field: bare(m[1]), operator: 'IS_NOT', value: unquote(m[2]) } }
  if ((m = /^(\S+)\s*=\s*'((?:[^']|'')*)'$/.exec(c)))
    return { kind: 'filter', filter: { field: bare(m[1]), operator: 'IS', value: unquote(m[2]) } }
  if ((m = /^(\S+)\s+IN\s+\((.+)\)$/i.exec(c))) {
    const vals = [...m[2].matchAll(/'((?:[^']|'')*)'/g)].map((x) => unquote(x[1]))
    if (!vals.length) return null
    return { kind: 'filter', filter: { field: bare(m[1]), operator: 'IS_ONE_OF_TERMS', value: vals } }
  }
  return null
}

/**
 * Best-effort parse of SQL matching buildSql's shape.
 * Returns null on any unrecognized clause — caller keeps prior state.
 */
export function parseSql(sql: string, fallbackInterval = 'hour'): ParsedSql | null {
  const from = /\bFROM\s+(logs|alerts)\b/i.exec(sql)
  if (!from) return null
  const dataset = from[1].toLowerCase()

  const whereMatch = /\bWHERE\b\s+([\s\S]+?)(?:\s+ORDER\s+BY|\s+LIMIT|\s*$)/i.exec(sql)
  const whereRaw = whereMatch?.[1]?.trim() ?? ''

  const range: TimeRange = { from: null, to: 'now', interval: fallbackInterval }
  const filters: FilterType[] = []
  let searchInput = ''
  let dataType: string | null = null

  if (whereRaw) {
    for (const raw of splitAnd(whereRaw)) {
      const p = parseClause(raw)
      if (!p) return null
      if (p.kind === 'dataType') dataType = p.value
      else if (p.kind === 'timeFrom') range.from = p.value
      else if (p.kind === 'timeTo') range.to = p.value
      else if (p.kind === 'search') searchInput = p.value
      else filters.push(p.filter)
    }
  }
  return { dataset, dataType, range, filters, searchInput }
}
