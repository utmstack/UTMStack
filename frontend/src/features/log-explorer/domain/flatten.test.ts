import { describe, expect, test } from 'vitest'
import { MSG_FIELDS, flattenDoc, pick, previewText } from './flatten'

describe('flattenDoc', () => {
  test('nested objects become dotted paths', () => {
    const flat = flattenDoc({ origin: { ip: '1.2.3.4', geolocation: { country: 'RU' } } })
    expect(flat['origin.ip']).toBe('1.2.3.4')
    expect(flat['origin.geolocation.country']).toBe('RU')
  })

  test('an array of strings reads as a list', () => {
    expect(flattenDoc({ tags: ['a', 'b'] }).tags).toBe('a, b')
  })

  // An alert carries `events` and `history` as arrays of objects. join() calls
  // String() on each element, which is where "[object Object]" comes from.
  test('an array of objects is rendered, not coerced', () => {
    const flat = flattenDoc({
      events: [
        { id: 'e1', timestamp: '2026-01-01T00:00:00Z' },
        { id: 'e2', timestamp: '2026-01-01T00:01:00Z' },
      ],
    })
    expect(String(flat.events)).not.toContain('[object Object]')
    expect(String(flat.events)).toContain('e1')
    expect(String(flat.events)).toContain('e2')
  })

  test('a mixed array keeps its primitives readable', () => {
    expect(String(flattenDoc({ xs: ['a', { b: 1 }] }).xs)).toBe('a, {"b":1}')
  })
})

// A cloudtrail record is an event name and an ARN; a firewall record is five
// tuple fields. Neither has a message, which is why the message column falls
// back to a summary — and why reading MSG_FIELDS alone exported a blank column.
describe('previewText', () => {
  const cloudtrail = {
    '@timestamp': '2026-01-01T00:00:00Z',
    tenantId: 't1',
    dataType: 'cloudtrail',
    log: { eventName: 'ConsoleLogin', sourceIPAddress: '1.2.3.4' },
  }

  test('a record with no message field still summarises', () => {
    const flat = flattenDoc(cloudtrail)
    expect(pick(flat, MSG_FIELDS)).toBeUndefined()
    const text = previewText(flat)
    expect(text).toContain('eventName=ConsoleLogin')
    expect(text).toContain('sourceIPAddress=1.2.3.4')
  })

  test('the summary drops the metadata every row repeats', () => {
    const text = previewText(flattenDoc(cloudtrail))
    expect(text).not.toContain('tenantId')
    expect(text).not.toContain('dataType')
  })
})
