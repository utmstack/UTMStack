/* Types mirror the backend adaudit DTOs (modules/adaudit/dto, domain.ADUser). */

/** One Active Directory account observed in ingested Windows security logs. */
export interface ADUser {
  id: number
  tenantId: string
  sid: string
  samAccountName: string
  domain: string
  active: boolean
  accountCreatedAt?: string
  lastLogon?: string
  accountDeletedAt?: string
  lastSeen?: string
}

/** Derived lifecycle bucket the backend can filter by (overrides `active`). */
export type ADUserStatus = 'active' | 'disabled' | 'deleted' | 'stale' | 'service'

export type ADUserSort = 'recent' | 'name'

export interface ADUserListQuery {
  search?: string
  tenantId?: string
  status?: ADUserStatus
  sort?: ADUserSort
  page?: number // zero-based
  size?: number
}

export interface DomainCount {
  domain: string
  count: number
}

/** GET /ad-audit/stats — inventory roll-up for the overview. */
export interface ADUserStats {
  total: number
  active: number
  disabled: number
  deleted: number
  stale: number
  service: number
  seen_24h: number
  by_domain: DomainCount[]
  tenants: string[]
}
