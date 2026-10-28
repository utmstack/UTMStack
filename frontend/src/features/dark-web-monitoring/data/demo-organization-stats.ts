import type { OrganizationBreachStats } from '../types'

/**
 * Headline numbers for the demo. `totalBreaches` is the grand total the page
 * shows in its summary cards — it must equal the sum of the per-surface counts
 * (2 emails + 1 phone + 7 domains + 1 ip + 1 username = 12, matching the 12
 * demo records) or the summary disagrees with the table.
 */
export const DEMO_ORGANIZATION_BREACH_STATS: OrganizationBreachStats = {
  id: 10901,
  orgName: 'test01',
  totalEmailsBreaches: 2,
  totalPhonesBreaches: 1,
  totalDomainBreaches: 7,
  totalIpsBreaches: 1,
  totalUsernamesBreaches: 1,
  totalBreaches: 12,
}
