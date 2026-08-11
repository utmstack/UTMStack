import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type {
  ComplianceSchedule,
  Control,
  EditControlInput,
  Framework,
  Report,
  ReportMeta,
  SaveScheduleInput,
  ScorePoint,
} from '../types/compliance.types'

const api = createApiClient()

export { ApiError as ComplianceHttpError }

const fw = (key: string) => `/compliance/frameworks/${encodeURIComponent(key)}`

export const complianceService = {
  // ── Frameworks ──────────────────────────────────────────────────────────
  listFrameworks: () => api.get<Framework[]>('/compliance/frameworks'),
  getFramework: (key: string) => api.get<Framework>(fw(key)),
  createFramework: (f: Framework) => api.post<Framework>('/compliance/frameworks', f),
  updateFramework: (key: string, f: Framework) => api.put<Framework>(fw(key), f),
  deleteFramework: (key: string) => api.delete<void>(fw(key)),

  // ── Controls (library) ──────────────────────────────────────────────────
  listControls: () => api.get<Control[]>('/compliance/controls'),
  getControl: (id: string) => api.get<Control>(`/compliance/controls/${encodeURIComponent(id)}`),
  createControl: (c: Control) => api.post<Control>('/compliance/controls', c),
  updateControl: (id: string, c: Control) => api.put<Control>(`/compliance/controls/${encodeURIComponent(id)}`, c),
  deleteControl: (id: string) => api.delete<void>(`/compliance/controls/${encodeURIComponent(id)}`),

  /** Data types a check can target, read from what the store actually holds. */
  listDataTypes: (dataset: 'logs' | 'alerts' = 'logs') =>
    api.get<string[]>(`/log-analyzer/datasets/${encodeURIComponent(dataset)}/data-types`),

  // ── Report ──────────────────────────────────────────────────────────────
  /** The standing report. Does not re-run the framework. */
  getReport: (key: string) => api.get<Report>(`${fw(key)}/report`),
  /** Re-runs it, keeping the edits already on it. */
  evaluate: (key: string, windowDays?: number) => {
    const q = windowDays ? `?windowDays=${windowDays}` : ''
    return api.post<Report>(`${fw(key)}/report${q}`)
  },
  listReports: () => api.get<ReportMeta[]>('/compliance/reports'),
  deleteReport: (key: string) => api.delete<void>(`${fw(key)}/report`),
  downloadReportPdf: (key: string) => api.get<Blob>(`${fw(key)}/report.pdf`, { responseType: 'blob' }),

  /**
   * Records a human verdict on one control. Everything above it — requirements,
   * sections, the score — recomputes from it, so the whole report comes back.
   */
  editControl: (key: string, controlId: string, input: EditControlInput) =>
    api.put<Report>(`${fw(key)}/controls/${encodeURIComponent(controlId)}/status`, input),

  // ── Score over time ─────────────────────────────────────────────────────
  history: (key: string, from?: string, to?: string) => {
    const p = new URLSearchParams()
    if (from) p.set('from', from)
    if (to) p.set('to', to)
    const q = p.toString()
    return api.get<ScorePoint[]>(`${fw(key)}/history${q ? `?${q}` : ''}`)
  },
  /** The document behind a point on the chart. `day` is YYYY-MM-DD. */
  downloadHistoryPdf: (key: string, day: string) =>
    api.get<Blob>(`${fw(key)}/history.pdf?day=${encodeURIComponent(day)}`, { responseType: 'blob' }),

  // ── Schedules ───────────────────────────────────────────────────────────
  listSchedules: (frameworkKey?: string) => {
    const p = new URLSearchParams()
    if (frameworkKey) p.set('frameworkKey', frameworkKey)
    p.set('page', '0')
    p.set('size', '200')
    return api.getPaged<ComplianceSchedule[]>(`/compliance-report-schedules/by-user?${p.toString()}`)
  },
  createSchedule: (input: SaveScheduleInput) => api.post<ComplianceSchedule>('/compliance-report-schedules', input),
  updateSchedule: (input: SaveScheduleInput) => api.put<ComplianceSchedule>('/compliance-report-schedules', input),
  deleteSchedule: (id: string) => api.delete<void>(`/compliance-report-schedules/${encodeURIComponent(id)}`),
}
