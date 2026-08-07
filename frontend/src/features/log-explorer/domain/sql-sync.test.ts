import { describe, expect, test } from 'vitest'
import { buildSql, looksLikeSql, parseSql } from './sql-sync'

const NOW = Date.parse('2026-07-01T00:00:00.000Z')

describe('buildSql/parseSql roundtrip', () => {
  test('empty state produces bare SELECT', () => {
    const sql = buildSql('logs', null, { from: null, to: 'now', interval: 'day' }, [], '', NOW)
    expect(sql).toBe('SELECT * FROM logs ORDER BY `@timestamp` DESC')
  })

  test('range + freetext + IS/IS_NOT/CONTAIN/EXIST/IN roundtrip', () => {
    const filters = [
      { field: 'severity', operator: 'IS' as const, value: 'high' },
      { field: 'source.ip', operator: 'IS_NOT' as const, value: '1.2.3.4' },
      { field: 'message', operator: 'CONTAIN' as const, value: 'failed login' },
      { field: 'user.name', operator: 'EXIST' as const },
      { field: 'action', operator: 'IS_ONE_OF_TERMS' as const, value: ['login', 'logout'] },
    ]
    const sql = buildSql(
      'logs',
      'wineventlog',
      { from: 'now-1h', to: 'now', interval: 'minute' },
      filters,
      'root',
      NOW,
    )
    const parsed = parseSql(sql)
    expect(parsed).not.toBeNull()
    expect(parsed!.dataType).toBe('wineventlog')
    expect(parsed!.searchInput).toBe('root')
    expect(parsed!.range.from).toBe(new Date(NOW - 3_600_000).toISOString())
    expect(parsed!.range.to).toBe(new Date(NOW).toISOString())
    expect(parsed!.filters).toEqual(filters)
  })

  test('value with embedded quote roundtrips', () => {
    const filters = [{ field: 'user', operator: 'IS' as const, value: "o'brien" }]
    const sql = buildSql('logs', null, { from: null, to: 'now', interval: 'day' }, filters, '', NOW)
    expect(sql).toContain("toString(user) = 'o''brien'")
    const parsed = parseSql(sql)
    expect(parsed!.filters[0].value).toBe("o'brien")
  })

  test('unrecognized clause returns null', () => {
    expect(parseSql('SELECT foo FROM "idx" WHERE bar UNKNOWN 1')).toBeNull()
  })

  test('parses SQL with no WHERE', () => {
    const parsed = parseSql('SELECT * FROM logs ORDER BY `@timestamp` DESC')
    expect(parsed).toEqual({
      dataset: 'logs',
      dataType: null,
      range: { from: null, to: 'now', interval: 'hour' },
      filters: [],
      searchInput: '',
    })
  })
})

describe('looksLikeSql', () => {
  test('a statement is recognised', () => {
    for (const q of [
      'SELECT * FROM logs',
      '  select severity, count() as c from alerts group by severity  ',
      'WITH x AS (SELECT 1) SELECT * FROM x',
      'SELECT toStartOfHour(`@timestamp`) AS h, count() FROM logs GROUP BY h',
    ]) {
      expect(looksLikeSql(q)).toBe(true)
    }
  })

  // Hunting for injection attempts is a reasonable thing to search logs for, so
  // the words alone must not be enough to change what the box does.
  test('text that merely mentions SQL is still a search', () => {
    for (const q of [
      'select',
      'select *',
      'from',
      'SELECT statement blocked by WAF',
      'user selected from the menu',
      'error: select failed',
      '',
    ]) {
      expect(looksLikeSql(q)).toBe(false)
    }
  })
})
