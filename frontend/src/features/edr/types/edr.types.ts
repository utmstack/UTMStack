// Types mirror backend/modules/edr/dto/*.go exactly. Fields the backend does
// not produce (isolationState, blocklist feed detail) have no counterpart and
// are intentionally absent here.

export interface EdrOverview {
  endpoints: number
  withModule: number
  healthy: number
  unhealthy: number
  staleSignatures: number
  drifted: number
  detectionsToday: number
}

export interface EdrEndpoint {
  agentId: number
  hostname: string
  ip: string
  os: string
  platform: string
  version: string
  status: string
  lastSeen: string
  hasModule: boolean
  moduleHealthy: boolean
  sigdbVersion?: string
  staleSignature: boolean
  policyVersion?: string
  drift: boolean
  lastReport?: string | null
}

export interface EdrActionResult {
  agentId: number
  action: string
  cmdId: string
  dispatched: boolean
  error?: string
}

export interface EdrCommandRow {
  cmdId: string
  action: string
  status: string
  result?: string
  executedBy: string
  createdAt: string
  updatedAt: string
}

export interface EdrQuarantineItem {
  agentId: number
  hostname: string
  id: string
  originalPath: string
  sha256: string
  detection: string
  quarantinedAt: string
  restorable: boolean
}

export interface EdrAllowlistEntry {
  category: 'allow_paths' | 'allow_processes' | 'allow_commands'
  value: string
}

export interface EdrBlocklistStatus {
  indicators: number
  activeFilters: number
  feedStale: boolean
  feedAgeSec: number
  recentBlocks: number
  enforced: boolean
  enforcementDegraded: boolean
}

export type EdrBulkAction = 'apply-policy' | 'scan' | 'isolate'
