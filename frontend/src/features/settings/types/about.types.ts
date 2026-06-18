/* Mirrors the backend endpoints that feed the About page. */

/** GET /billing/version — deploy version + edition (from version.json/instance-config). */
export interface VersionInfo {
  version: string
  edition: string // community | enterprise
  changelog?: string
  server?: string // hostname from instance-config
  instanceId?: string
}

/** GET /billing/license — license edition, limits, expiry. */
export interface LicenseInfo {
  edition: string // community | enterprise
  mssp: boolean
  datasources: number // cap; 0 = unlimited
  type: string // online | offline ("" when community)
  expiresAt?: string // RFC3339, omitted when no expiry
}

/** GET /datasources/usage — datasource count vs cap. */
export interface DatasourceUsage {
  count: number
  limit: number // effective cap; 0 when unlimited
  unlimited: boolean
  overLimit: boolean
}

/** GET /opensearch/cluster/status — OpenSearch cluster health (null on 204). */
export interface ClusterHealth {
  cluster_name: string
  status: string // green | yellow | red
  number_of_nodes: number
  number_of_data_nodes: number
  active_shards: number
}

/** GET /mcp/health — SOC AI / MCP server status. */
export interface McpHealth {
  enabled: boolean
  tools_registered: number
  server_version: string
  server_name: string
}
