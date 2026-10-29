import { describe, expect, it } from 'vitest'
import { parseFilter, serializeFilter } from './filter-model'

function serializeKv(config: Record<string, unknown>): string {
  return serializeFilter({
    entries: [{ dataTypes: ['generic'], order: 100, steps: [{ kind: 'kv', config }] }],
  })
}

describe('kv fieldSplit normalization', () => {
  it('normalizes an empty fieldSplit to a space', () => {
    const out = serializeKv({ source: 'log', fieldSplit: '', valueSplit: '=' })
    expect(out).toContain('fieldSplit: " "')
    expect(out).toContain('valueSplit: "="')
  })

  it('normalizes a missing fieldSplit to a space', () => {
    const out = serializeKv({ source: 'log', valueSplit: '=' })
    expect(out).toContain('fieldSplit: " "')
    expect(out).toContain('valueSplit: "="')
  })

  it('normalizes a whitespace-only fieldSplit to a single space', () => {
    const out = serializeKv({ source: 'log', fieldSplit: ' ', valueSplit: '=' })
    expect(out).toContain('fieldSplit: " "')
  })

  it('round-trips the space through parse', () => {
    const out = serializeKv({ source: 'log', fieldSplit: ' ', valueSplit: '=' })
    const r = parseFilter(out)
    expect(r.ok).toBe(true)
    if (!r.ok) return
    const cfg = (r.model.entries[0].steps[0].config as Record<string, unknown>)
    expect(cfg.fieldSplit).toBe(' ')
    expect(cfg.valueSplit).toBe('=')
  })

  it('passes a non-space fieldSplit through unchanged', () => {
    const out = serializeKv({ source: 'log', fieldSplit: ', ', valueSplit: ':' })
    expect(out).toContain('fieldSplit: ", "')
    const r = parseFilter(out)
    expect(r.ok).toBe(true)
    if (!r.ok) return
    const cfg = (r.model.entries[0].steps[0].config as Record<string, unknown>)
    expect(cfg.fieldSplit).toBe(', ')
    expect(cfg.valueSplit).toBe(':')
  })

  it('still prunes empty strings for other keys (e.g. where)', () => {
    const out = serializeKv({ source: 'log', fieldSplit: ' ', where: '' })
    expect(out).not.toContain('where')
  })

  it('quotes a csv separator too', () => {
    const out = serializeFilter({
      entries: [{ dataTypes: ['generic'], order: 100, steps: [{ kind: 'csv', config: { source: 'log', separator: ', ' } }] }],
    })
    expect(out).toContain('separator: ", "')
  })
})
