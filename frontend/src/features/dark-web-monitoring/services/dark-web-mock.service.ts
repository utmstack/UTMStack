import type { BreachStatus, DarkWebRecord, OrganizationBreachStats } from '../types'
import { DEMO_RECORDS } from '../data/demo-records'
import { DEMO_ORGANIZATION_BREACH_STATS } from '../data/demo-organization-stats'

function clone<T>(value: T): T {
  return typeof structuredClone === 'function' ? structuredClone(value) : JSON.parse(JSON.stringify(value))
}

// Resolve on a microtask (not setTimeout) so the mock is safe under
// vi.useFakeTimers() and tests can flush it with a single `await`.
const tick = () => Promise.resolve()

interface MockDarkWebService {
  listRecords(): Promise<DarkWebRecord[]>
  getOrganizationStats(): Promise<OrganizationBreachStats>
  updateRecordStatus(id: string, breachStatus: BreachStatus): Promise<DarkWebRecord>
}

/**
 * Builds an in-memory stand-in for the dark web monitoring API. Each instance
 * owns its record list, so tests that exercise status writes get a clean copy
 * and cannot leak state into the UI's instance.
 *
 * Returned records are cloned so callers never mutate the service's state, and
 * a status change is reflected in later `listRecords` calls — the same
 * behaviour the real endpoint contract will need.
 */
export function createDarkWebMockService(): MockDarkWebService {
  const records = clone(DEMO_RECORDS)
  return {
    async listRecords() {
      await tick()
      return clone(records)
    },
    async getOrganizationStats() {
      await tick()
      return clone(DEMO_ORGANIZATION_BREACH_STATS)
    },
    async updateRecordStatus(id, breachStatus) {
      await tick()
      const record = records.find((item) => item.id === id)
      if (!record) throw new Error(`Unknown dark web record: ${id}`)
      record.breachStatus = breachStatus
      return clone(record)
    },
  }
}

/** Shared instance serving the dark web page (see `services/index.ts`). */
export const darkWebMockService: MockDarkWebService = createDarkWebMockService()
