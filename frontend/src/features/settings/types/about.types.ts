/* Mirrors the backend endpoints that feed the About page. */

/** GET /billing/version — deploy version + edition (from version.json/instance-config). */
export interface VersionInfo {
  version: string
  edition: string // community | enterprise
  changelog?: string
  server?: string // hostname from instance-config
  instanceId?: string
}

/**
 * GET /billing/license — edition and, for the default tenant only, the terms.
 * The backend withholds the commercial fields from every other tenant, so they
 * arrive absent rather than zero.
 */
export interface LicenseInfo {
  edition: string // community | enterprise
  mssp: boolean
  ingestGbPerMonth?: number // contracted GB/month; 0 = unlimited
  type?: string // online | offline ("" when community)
  expiresAt?: string // RFC3339, omitted when no expiry
}

/** GET /datasources/count — number of configured datasources. */
export interface DatasourceUsage {
  count: number
}

/** GET /mcp/health — SOC AI / MCP server status. */
export interface McpHealth {
  enabled: boolean
  tools_registered: number
  server_version: string
  server_name: string
}
