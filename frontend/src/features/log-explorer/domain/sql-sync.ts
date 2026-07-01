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

const quote = (s: string) => `'${String(s).replace(/'/g, "''")}'`
const unquote = (s: string) => s.replace(/''/g, "'")

// Resolve OpenSearch date-math (now, now-Nu) to absolute ISO so the SQL plugin
// gets literal timestamps. Non-tokens (absolute ISO) pass through.
function resolveDate(v: string, nowMs: number = Date.now()): string {
  if (v === 'now') return new Date(nowMs).toISOString()
  const m = /^now-(\d+)([mhdwM])$/.exec(v)
  if (!m) return v
  const ms = UNIT_MS[m[2]]
  return new Date(nowMs - Number(m[1]) * ms).toISOString()
}

function filterToSql(f: FilterType): string | null {
  const field = f.field
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

/** Build canonical OpenSearch SQL from the log-explorer state. */
export function buildSql(
  patternStr: string,
  range: TimeRange,
  filters: FilterType[],
  searchInput: string,
  nowMs?: number,
): string {
  const parts: string[] = []
  if (range.from) {
    parts.push(
      `${TS} >= ${quote(resolveDate(range.from, nowMs))} AND ${TS} <= ${quote(resolveDate(range.to, nowMs))}`,
    )
  }
  const search = searchInput.trim()
  if (search) parts.push(`QUERY(${quote('*' + search + '*')})`)
  for (const f of filters) {
    if (f.field === TS && f.operator === 'IS_BETWEEN') continue
    if (f.operator === 'IS_IN_FIELDS') continue
    const s = filterToSql(f)
    if (s) parts.push(s)
  }
  const where = parts.length ? ` WHERE ${parts.join(' AND ')}` : ''
  return `SELECT * FROM "${patternStr}"${where} ORDER BY ${TS} DESC`
}

export interface ParsedSql {
  patternStr: string
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
  | { kind: 'timeFrom'; value: string }
  | { kind: 'timeTo'; value: string }
  | { kind: 'search'; value: string }
  | { kind: 'filter'; filter: FilterType }

function parseClause(c: string): Clause | null {
  let m: RegExpExecArray | null
  if ((m = /^@timestamp\s*>=\s*'((?:[^']|'')*)'$/i.exec(c)))
    return { kind: 'timeFrom', value: unquote(m[1]) }
  if ((m = /^@timestamp\s*<=\s*'((?:[^']|'')*)'$/i.exec(c)))
    return { kind: 'timeTo', value: unquote(m[1]) }
  if ((m = /^QUERY\(\s*'((?:[^']|'')*)'\s*\)$/i.exec(c))) {
    const raw = unquote(m[1])
    return { kind: 'search', value: raw.replace(/^\*+|\*+$/g, '') }
  }
  if ((m = /^(\S+)\s+IS\s+NOT\s+NULL$/i.exec(c)))
    return { kind: 'filter', filter: { field: m[1], operator: 'EXIST' } }
  if ((m = /^(\S+)\s+LIKE\s+'%((?:[^']|'')*)%'$/i.exec(c)))
    return { kind: 'filter', filter: { field: m[1], operator: 'CONTAIN', value: unquote(m[2]) } }
  if ((m = /^(\S+)\s*<>\s*'((?:[^']|'')*)'$/.exec(c)))
    return { kind: 'filter', filter: { field: m[1], operator: 'IS_NOT', value: unquote(m[2]) } }
  if ((m = /^(\S+)\s*!=\s*'((?:[^']|'')*)'$/.exec(c)))
    return { kind: 'filter', filter: { field: m[1], operator: 'IS_NOT', value: unquote(m[2]) } }
  if ((m = /^(\S+)\s*=\s*'((?:[^']|'')*)'$/.exec(c)))
    return { kind: 'filter', filter: { field: m[1], operator: 'IS', value: unquote(m[2]) } }
  if ((m = /^(\S+)\s+IN\s+\((.+)\)$/i.exec(c))) {
    const vals = [...m[2].matchAll(/'((?:[^']|'')*)'/g)].map((x) => unquote(x[1]))
    if (!vals.length) return null
    return { kind: 'filter', filter: { field: m[1], operator: 'IS_ONE_OF_TERMS', value: vals } }
  }
  return null
}

/**
 * Best-effort parse of SQL matching buildSql's shape.
 * Returns null on any unrecognized clause — caller keeps prior state.
 */
export function parseSql(sql: string, fallbackInterval = 'hour'): ParsedSql | null {
  const fromMatch = /FROM\s+"([^"]+)"/i.exec(sql)
  if (!fromMatch) return null
  const patternStr = fromMatch[1]

  const whereMatch = /\bWHERE\b\s+([\s\S]+?)(?:\s+ORDER\s+BY|\s+LIMIT|\s*$)/i.exec(sql)
  const whereRaw = whereMatch?.[1]?.trim() ?? ''

  const range: TimeRange = { from: null, to: 'now', interval: fallbackInterval }
  const filters: FilterType[] = []
  let searchInput = ''

  if (whereRaw) {
    for (const raw of splitAnd(whereRaw)) {
      const p = parseClause(raw)
      if (!p) return null
      if (p.kind === 'timeFrom') range.from = p.value
      else if (p.kind === 'timeTo') range.to = p.value
      else if (p.kind === 'search') searchInput = p.value
      else filters.push(p.filter)
    }
  }
  return { patternStr, range, filters, searchInput }
}
