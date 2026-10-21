export type EdrHealth = 'healthy' | 'unhealthy' | 'stale' | 'degraded' | 'off' | 'unknown'
export type EdrProtectionState = 'alert' | 'responding' | 'isolated' | 'normal' | 'unknown'

export interface EdrOverview {
  coverage?: { protected?: number; unprotected?: number; total?: number }
  health?: {
    healthy?: number
    unhealthyEngine?: number
    staleSignatures?: number
    degradedNetwork?: number
    sensorsOff?: number
    policyDrift?: number
  }
  protection?: { alertMode?: number; responding?: number; isolated?: number }
  activity?: {
    detectionsOverTime?: Array<{ timestamp?: string; source?: string; count?: number }>
    ransomwareContainments?: number
    blockedConnections?: number
    topDetections?: Array<{ name?: string; count?: number }>
    affectedEndpoints?: Array<{ hostName?: string; count?: number }>
  }
}

export interface EdrEndpoint {
  id: string
  hostName: string
  operatingSystem: string
  moduleVersion: string
  serviceState: string
  engineHealth: EdrHealth
  signatureAge: number | null
  sensors: string[]
  assignedPolicy: string
  policyDrift: boolean
  detectionsLastWeek: number
  quarantinedFiles: number
  isolationState: EdrProtectionState
  lastReportTime: string
  [key: string]: unknown
}

export interface EdrEndpointsResponse {
  items?: unknown[]
  endpoints?: unknown[]
  content?: unknown[]
  total?: number
}

export type EdrBulkAction = 'apply-policy' | 'scan' | 'isolate'
