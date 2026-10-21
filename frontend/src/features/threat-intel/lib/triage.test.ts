import { describe, expect, it } from 'vitest'
import { parseIndicators, refang, sortRows, toStix, type TriageRow } from './triage.types'
import { detectType } from '../services/triage.service'

describe('refang', () => {
  it('undoes common defanging', () => {
    expect(refang('1.2.3[.]4')).toBe('1.2.3.4')
    expect(refang('hxxp://bad.example/path')).toBe('http://bad.example/path')
    expect(refang('user[at]example.com')).toBe('user@example.com')
    expect(refang("<'bad.example'>")).toBe('bad.example')
  })
})

describe('parseIndicators', () => {
  it('splits on whitespace/comma/semicolon, dedupes case-insensitively', () => {
    expect(parseIndicators('1.2.3[.]4, 4.3.2.1\nBAd.Example bad.example')).toEqual(['1.2.3.4', '4.3.2.1', 'BAd.Example'])
  })
})

describe('detectType', () => {
  it('maps shapes to CM entity types', () => {
    expect(detectType('1.2.3.4')).toBe('ip')
    expect(detectType('bad.example.com')).toBe('domain')
    expect(detectType('http://bad.example')).toBe('url')
    expect(detectType('a'.repeat(32))).toBe('md5')
    expect(detectType('b'.repeat(64))).toBe('sha256')
    expect(detectType('CVE-2024-1234')).toBe('cve')
    expect(detectType('user@bad.example')).toBe('email-address')
    expect(detectType('totally not an indicator')).toBe('')
  })
})

function row(verdict: TriageRow['verdict'], input: string, reputation = 0): TriageRow {
  return { key: input, input, type: 'domain', verdict, entity: null, ...(reputation ? { entity: { id: input, type: 'domain', attributes: { domain: input }, tags: [], lastSeen: '', reputation, accuracy: 1 } as never } : {}) }
}

describe('sortRows', () => {
  it('orders worst first, then by reputation', () => {
    const rows = [row('benign', 'c', 1), row('malicious', 'a', -3), row('malicious', 'b', -2)]
    expect(sortRows(rows).map(r => r.input)).toEqual(['a', 'b', 'c'])
  })
})

describe('toStix', () => {
  it('exports only malicious/suspicious with stable ids', () => {
    const rows = [row('malicious', 'bad.example.com'), row('benign', 'good.example.com')]
    const bundle = JSON.parse(toStix(rows)) as { objects: { id: string }[] }
    expect(bundle.objects).toHaveLength(1)
    expect(bundle.objects[0].id).toBe(bundle.objects[0].id) // stable across calls
    expect(toStix(rows)).toBe(toStix(rows))
  })
})
