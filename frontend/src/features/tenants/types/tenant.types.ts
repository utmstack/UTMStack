/* Types mirror the backend tenant DTOs (modules/tenant). Unlike the rest of the
 * API this module answers in camelCase, because it is serialized straight from
 * the domain struct. */

export type TenantStatus = 'ACTIVE' | 'SUSPENDED' | 'TERMINATED'

/** How much of itself a tenant has opened to the platform operator. Granted by
 * the tenant's own administrator — the platform cannot raise its own access. */
export type SupportAccess = 'NONE' | 'READ' | 'FULL'

export interface TenantLimits {
  /** null means "whatever the instance licence allows". */
  maxAIRequests: number | null
}

export interface Tenant {
  id: string
  name: string
  domain: string
  status: TenantStatus
  supportAccess: SupportAccess
  limits: TenantLimits
  createdAt: string
  updatedAt: string
}

export interface CreateTenantRequest {
  name: string
  domain: string
  /** The first administrator. They get an invitation, not a password. */
  adminEmail: string
}

export interface UpdateTenantRequest {
  name?: string
  domain?: string
  status?: TenantStatus
  maxAIRequests?: number | null
}

export interface TenantFilter {
  name?: string
  domain?: string
  status?: TenantStatus
  page?: number
  size?: number
}

/* ---- per-tenant statistics ---- */

/** What a card shows. Every field is independent: one unreachable subsystem
 * blanks its own number instead of the whole card. */
export interface TenantStats {
  users: number | null
  datasources: number | null
  openAlerts: number | null
  ai: { used: number; limit: number } | null
}

export interface AIUsage {
  limit: number
  used: number
  remaining?: number
  resetsAt: string
}
