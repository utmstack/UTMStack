/* Types mirror the backend adaudit DTOs (modules/adaudit/dto, domain.ADUser). */

/* Cloud: windows | linux | o365 | azure | gcp | aws */
export type ADUserSource = 'windows' | 'linux' | 'o365' | 'azure' | 'gcp' | 'aws'

export const CLOUD_SOURCES: ADUserSource[] = ['o365', 'azure', 'gcp', 'aws']

/** One user account observed in ingested Windows, Linux or cloud logs. */
export interface ADUser {
  id: string
  tenantId: string
  source: ADUserSource
  /* Windows-only fields */
  sid?: string
  samAccountName?: string
  domain?: string
  /* Linux-only fields */
  machineId?: string
  uidNumber?: number
  hostname?: string
  username?: string
  /* Cloud-only fields */
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
  identity?: string // alias uuid, or the literal 'none'
  page?: number // zero-based
  size?: number
}

/** Row returned by GET /ad-audit/users: the account plus its assigned identity alias name (null when unassigned). */
export type ADUserWithIdentity = ADUser & { identity?: string | null }

/** GET /ad-audit/identities — an operator-named identity alias. */
export interface Identity {
  id: string
  name: string
  createdAt: string
}

/** GET /ad-audit/identities list item: the alias plus how many accounts it covers. */
export interface IdentityListItem extends Identity {
  count: number
}

/** GET /ad-audit/identities/:id — the alias and every account assigned to it. */
export interface IdentityDetail extends Identity {
  users: ADUserWithIdentity[]
}

export interface DomainCount {
  domain: string
  count: number
}

export interface SourceCount {
  windows: number
  linux: number
  o365: number
  azure: number
  gcp: number
  aws: number
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
  identities: number
  unassigned: number
  by_source: SourceCount
  by_domain: DomainCount[]
  tenants: string[]
}
