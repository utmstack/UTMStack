import { describe, expect, it } from 'vitest'
import './collectors'
import { getCollector } from './registry'

describe('collector guide lookup', () => {
  it('resolves registered identifiers', () => {
    expect(getCollector('CISCO')?.getName()).toBe('CISCO')
    expect(getCollector('firewall-cisco-asa')?.getName()).toBe('CISCO')
  })

  it('does not route an unrelated product to a vendor guide by substring', () => {
    expect(getCollector('Cisco ISE')).toBeUndefined()
    expect(getCollector('Fortinet Analyzer')).toBeUndefined()
  })
})
