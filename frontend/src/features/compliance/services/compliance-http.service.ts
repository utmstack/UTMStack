import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type {
  ComplianceSchedule,
  Control,
  Framework,
  Report,
  ReportSnapshot,
  ReportSnapshotMeta,
  SaveScheduleInput,
} from '../types/compliance.types'

const api = createApiClient()

export { ApiError as ComplianceHttpError }

export const complianceService = {
  // ── Frameworks ──────────────────────────────────────────────────────────
  listFrameworks: () => api.get<Framework[]>('/compliance/frameworks'),
  getFramework: (key: string) => api.get<Framework>(`/compliance/frameworks/${encodeURIComponent(key)}`),
  createFramework: (f: Framework) => api.post<Framework>('/compliance/frameworks', f),
  updateFramework: (key: string, f: Framework) => api.put<Framework>(`/compliance/frameworks/${encodeURIComponent(key)}`, f),
  deleteFramework: (key: string) => api.delete<void>(`/compliance/frameworks/${encodeURIComponent(key)}`),
  setFrameworkEnabled: (key: string, enabled: boolean) =>
    api.put<void>(`/compliance/frameworks/${encodeURIComponent(key)}/enabled`, { enabled }),

  // ── Controls (library) ──────────────────────────────────────────────────
  listControls: () => api.get<Control[]>('/compliance/controls'),
  getControl: (id: string) => api.get<Control>(`/compliance/controls/${encodeURIComponent(id)}`),
  // Index-pattern catalog (OpenSearch) — to populate the check's indexPattern select.
  listIndexPatterns: () => api.getPaged<{ pattern: string; isActive?: boolean }[]>('/opensearch/index-patterns?page=0&size=500&sort=pattern,asc'),

  createControl: (c: Control) => api.post<Control>('/compliance/controls', c),
  updateControl: (id: string, c: Control) => api.put<Control>(`/compliance/controls/${encodeURIComponent(id)}`, c),
  deleteControl: (id: string) => api.delete<void>(`/compliance/controls/${encodeURIComponent(id)}`),
  setControlEnabled: (id: string, enabled: boolean) =>
    api.put<void>(`/compliance/controls/${encodeURIComponent(id)}/enabled`, { enabled }),

  // ── Reports (live evaluation + stored snapshots) ──────────────────────────
  liveReport: (key: string) => api.get<Report>(`/compliance/frameworks/${encodeURIComponent(key)}/report`),
  generateReport: (key: string) => api.post<Report>(`/compliance/frameworks/${encodeURIComponent(key)}/report`),
  listSnapshots: (frameworkKey?: string, limit = 50) => {
    const p = new URLSearchParams()
    if (frameworkKey) p.set('framework', frameworkKey)
    p.set('limit', String(limit))
    return api.get<ReportSnapshotMeta[]>(`/compliance/reports?${p.toString()}`)
  },
  getSnapshot: (id: string) => api.get<ReportSnapshot>(`/compliance/reports/${encodeURIComponent(id)}`),
  deleteSnapshot: (id: string) => api.delete<void>(`/compliance/reports/${encodeURIComponent(id)}`),
  // Backend-generated PDF (same renderer the scheduled email reports use).
  downloadFrameworkPdf: (key: string) =>
    api.get<Blob>(`/compliance/frameworks/${encodeURIComponent(key)}/report.pdf`, { responseType: 'blob' }),
  downloadSnapshotPdf: (id: string) =>
    api.get<Blob>(`/compliance/reports/${encodeURIComponent(id)}/pdf`, { responseType: 'blob' }),

  // Manual (framework, control) status overrides — applied on live evaluations.
  setControlStatusOverride: (frameworkKey: string, controlId: string, status: string, reason?: string) =>
    api.put<void>(
      `/compliance/frameworks/${encodeURIComponent(frameworkKey)}/controls/${encodeURIComponent(controlId)}/status`,
      { status, reason: reason ?? '' },
    ),
  clearControlStatusOverride: (frameworkKey: string, controlId: string) =>
    api.delete<void>(`/compliance/frameworks/${encodeURIComponent(frameworkKey)}/controls/${encodeURIComponent(controlId)}/status`),

  // User notes on (framework, control) — freeform text surfaced on report rows.
  setControlNote: (frameworkKey: string, controlId: string, note: string) =>
    api.put<void>(
      `/compliance/frameworks/${encodeURIComponent(frameworkKey)}/controls/${encodeURIComponent(controlId)}/note`,
      { note },
    ),
  clearControlNote: (frameworkKey: string, controlId: string) =>
    api.delete<void>(`/compliance/frameworks/${encodeURIComponent(frameworkKey)}/controls/${encodeURIComponent(controlId)}/note`),

  // ── Schedules (Postgres) ─────────────────────────────────────────────────
  listSchedules: (frameworkKey?: string) => {
    const p = new URLSearchParams()
    if (frameworkKey) p.set('frameworkKey', frameworkKey)
    p.set('page', '0')
    p.set('size', '200')
    return api.getPaged<ComplianceSchedule[]>(`/compliance-report-schedules/by-user?${p.toString()}`)
  },
  createSchedule: (input: SaveScheduleInput) => api.post<ComplianceSchedule>('/compliance-report-schedules', input),
  updateSchedule: (input: SaveScheduleInput) => api.put<ComplianceSchedule>('/compliance-report-schedules', input),
  deleteSchedule: (id: number) => api.delete<void>(`/compliance-report-schedules/${id}`),
}
