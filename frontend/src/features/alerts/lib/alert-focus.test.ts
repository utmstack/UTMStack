import { describe, expect, it } from 'vitest'
import type { Alert } from '../types/alert.types'
import { alertFocus } from './alert-focus'

const alert: Alert = {
  id: 'abc-123',
  name: 'Windows: Firewall Rule Deleted',
  severity: 'high',
  status: 'Open',
  category: 'Defense Evasion',
  technique: 'T1562.004',
  dataSource: 'WIN-01',
  tags: ['False positive'],
  assignee: 'analyst@example.com',
}

describe('alertFocus', () => {
  const focus = alertFocus(alert)

  it('names the alert and what the chip should show', () => {
    expect(focus.kind).toBe('alert')
    expect(focus.id).toBe('abc-123')
    expect(focus.label).toBe('Windows: Firewall Rule Deleted')
  })

  it('tells the agent what is open and how to read the rest, without asking for an id', () => {
    expect(focus.context).toContain('id abc-123')
    expect(focus.context).toContain('severity high')
    expect(focus.context).toContain('data source WIN-01')
    expect(focus.context).toContain('do not ask for its id')
    expect(focus.context).toContain('store.search')
    expect(focus.context).toContain('alerts.score')
  })

  it('keeps the assignee out of the prompt', () => {
    expect(focus.context).not.toContain('analyst@example.com')
  })

  it('falls back to the id when the alert has no name', () => {
    expect(alertFocus({ id: 'x-1' }).label).toBe('x-1')
  })
})
