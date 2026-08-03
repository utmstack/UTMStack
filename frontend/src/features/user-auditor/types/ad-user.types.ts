/* Types mirror the backend adaudit DTOs (modules/adaudit/dto, domain.ADUser). */

export type ADUserSource = 'windows' | 'linux'

/** One user account observed in ingested Windows or Linux logs. */
export interface ADUser {
  id: number
  tenantId: string
  source: ADUserSource
  /* Windows-only fields */
  sid?: string
  samAccountName?: string
  domain?: string
  /* Linux-only fields */
  machineId?: string
  uidNumber?: string
  hostname?: string
  username?: string
  /* Shared lifecycle */
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
  source?: ADUserSource
  status?: ADUserStatus
  sort?: ADUserSort
  page?: number // zero-based
  size?: number
}

export interface DomainCount {
  domain: string
  count: number
}

export interface SourceCount {
  windows: number
  linux: number
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
  by_source: SourceCount
  by_domain: DomainCount[]
  tenants: string[]
}
