import { createApiClient } from '@/shared/lib/api-client'
import type { EdrBulkAction, EdrEndpoint, EdrEndpointsResponse, EdrOverview } from '../types/edr.types'

const api = createApiClient()

const asRecord = (value: unknown): Record<string, unknown> =>
  value && typeof value === 'object' ? value as Record<string, unknown> : {}

const first = (record: Record<string, unknown>, keys: string[], fallback: unknown = ''): unknown => {
  for (const key of keys) {
    const value = record[key]
    if (value !== undefined && value !== null && value !== '') return value
  }
  return fallback
}

const numberValue = (value: unknown, fallback = 0): number => {
  const parsed = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(parsed) ? parsed : fallback
}

const stringValue = (value: unknown, fallback = '') => String(value ?? fallback)

export function normalizeEndpoint(value: unknown, index: number): EdrEndpoint {
  const item = asRecord(value)
  const sensors = first(item, ['sensors', 'enabledSensors', 'sensorState'], [])
  const health = stringValue(first(item, ['engineHealth', 'health', 'engineStatus'], 'unknown')).toLowerCase()
  const isolation = stringValue(first(item, ['isolationState', 'isolation', 'isolateStatus'], 'normal')).toLowerCase()
  return {
    ...item,
    id: stringValue(first(item, ['id', 'endpointId', 'agentId'], `endpoint-${index}`)),
    hostName: stringValue(first(item, ['hostName', 'hostname', 'host', 'name'], '—')),
    operatingSystem: stringValue(first(item, ['operatingSystem', 'os', 'platform'], '—')),
    moduleVersion: stringValue(first(item, ['moduleVersion', 'version', 'agentVersion'], '—')),
    serviceState: stringValue(first(item, ['serviceState', 'serviceStatus', 'status'], 'unknown')),
    engineHealth: ['healthy', 'unhealthy', 'stale', 'degraded', 'off'].includes(health) ? health as EdrEndpoint['engineHealth'] : 'unknown',
    signatureAge: first(item, ['signatureAge', 'signatureAgeDays', 'definitionsAge']) === ''
      ? null
      : numberValue(first(item, ['signatureAge', 'signatureAgeDays', 'definitionsAge']), 0),
    sensors: Array.isArray(sensors) ? sensors.map(String) : Object.entries(asRecord(sensors)).filter(([, enabled]) => Boolean(enabled)).map(([name]) => name),
    assignedPolicy: stringValue(first(item, ['assignedPolicy', 'policyName', 'policy'], '—')),
    policyDrift: Boolean(first(item, ['policyDrift', 'drifted', 'policyOutOfDate'], false)),
    detectionsLastWeek: numberValue(first(item, ['detectionsLastWeek', 'detections7d', 'weeklyDetections'])),
    quarantinedFiles: numberValue(first(item, ['quarantinedFiles', 'quarantineCount', 'quarantined'])),
    isolationState: ['alert', 'responding', 'isolated', 'normal'].includes(isolation) ? isolation as EdrEndpoint['isolationState'] : 'unknown',
    lastReportTime: stringValue(first(item, ['lastReportTime', 'lastSeen', 'lastCheckIn', 'reportedAt'], '')),
  }
}

export function normalizeEndpoints(payload: EdrEndpointsResponse | unknown): EdrEndpoint[] {
  const record = asRecord(payload)
  const raw = Array.isArray(payload) ? payload : first(record, ['items', 'endpoints', 'content', 'data'], [])
  return Array.isArray(raw) ? raw.map(normalizeEndpoint) : []
}

export const edrHttpService = {
  overview: () => api.get<EdrOverview>('/edr/overview'),
  endpoints: async () => normalizeEndpoints(await api.get<EdrEndpointsResponse>('/edr/endpoints')),
  bulkAction: (action: EdrBulkAction, endpointIds: string[], policy?: string) =>
    api.post('/edr/endpoints/actions', { action, endpointIds, policy }),
}
