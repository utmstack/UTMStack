import type { BreachStatus, DarkWebRecord, OrganizationBreachStats } from '../types'
import { DEMO_RECORDS } from '../data/demo-records'
import { DEMO_ORGANIZATION_BREACH_STATS } from '../data/demo-organization-stats'

function clone<T>(value: T): T {
  return typeof structuredClone === 'function' ? structuredClone(value) : JSON.parse(JSON.stringify(value))
}

// Resolve on a microtask (not setTimeout) so the mock is safe under
// vi.useFakeTimers() and tests can flush it with a single `await`.
const tick = () => Promise.resolve()

/**
 * In-memory stand-in for the dark web monitoring API. It returns cloned
 * demo data so the UI never mutates the module constants, and it keeps the
 * status changes in sync with the list — exactly what the real endpoint
 * contract will need to do.
 */
const state: { records: DarkWebRecord[] } = { records: clone(DEMO_RECORDS) }

export const darkWebMockService = {
  async listRecords(): Promise<DarkWebRecord[]> {
    await tick()
    return clone(state.records)
  },
  async getOrganizationStats(): Promise<OrganizationBreachStats> {
    await tick()
    return clone(DEMO_ORGANIZATION_BREACH_STATS)
  },
  async updateRecordStatus(id: string, breachStatus: BreachStatus): Promise<DarkWebRecord> {
    await tick()
    const record = state.records.find((item) => item.id === id)
    if (!record) throw new Error(`Unknown dark web record: ${id}`)
    record.breachStatus = breachStatus
    return clone(record)
  },
}
