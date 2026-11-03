import { deriveStatus, type DerivedStatus } from './status'
import type { Datasource } from '../types/datasource.types'

export type DatasourceReportStatus = 'all' | DerivedStatus

export interface DatasourceReportLabels {
  title: string
  generatedAt: string
  sourceCount: string
  noSources: string
  columns: {
    name: string
    type: string
    kind: string
    status: string
    ip: string
    lastSeen: string
    discovered: string
    labels: string
  }
  statuses: Record<DerivedStatus, string>
  kinds: Record<string, string>
}

export function filterDatasourceReport(
  sources: Datasource[],
  status: DatasourceReportStatus,
  offlineTab = false,
): Datasource[] {
  return sources.filter((source) => {
    const sourceStatus = deriveStatus(source.lastPingAt)
    if (offlineTab && sourceStatus !== 'offline') return false
    return status === 'all' || sourceStatus === status
  })
}

function escapeHtml(value: unknown): string {
  return String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function formatDate(value?: string): string {
  if (!value) return '—'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

export function buildDatasourceReportHtml(
  sources: Datasource[],
  labels: DatasourceReportLabels,
  generatedAt: string,
): string {
  const rows = sources.map((source) => {
    const status = deriveStatus(source.lastPingAt)
    const kind = source.sourceKind ? labels.kinds[source.sourceKind] ?? source.sourceKind : '—'
    const values = [
      source.name,
      source.dataType || '—',
      kind,
      labels.statuses[status],
      source.ip || '—',
      formatDate(source.lastPingAt),
      formatDate(source.discoveredAt),
      source.labels || '—',
    ]
    return `<tr>${values.map((value) => `<td>${escapeHtml(value)}</td>`).join('')}</tr>`
  }).join('')

  const columns = Object.values(labels.columns)
  const table = sources.length
    ? `<table><thead><tr>${columns.map((column) => `<th>${escapeHtml(column)}</th>`).join('')}</tr></thead><tbody>${rows}</tbody></table>`
    : `<p class="empty">${escapeHtml(labels.noSources)}</p>`

  return `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>${escapeHtml(labels.title)}</title>
<style>
  @page { size: landscape; margin: 14mm; }
  body { color: #172033; font: 12px Arial, sans-serif; }
  h1 { margin: 0 0 6px; font-size: 22px; }
  .meta { margin: 0 0 18px; color: #536078; }
  table { width: 100%; border-collapse: collapse; }
  th, td { padding: 7px 8px; border: 1px solid #d5dbe5; text-align: left; vertical-align: top; overflow-wrap: anywhere; }
  th { background: #eef2f7; font-weight: 700; }
  tr { break-inside: avoid; }
  .empty { margin-top: 24px; color: #536078; }
</style>
</head>
<body>
  <h1>${escapeHtml(labels.title)}</h1>
  <p class="meta">${escapeHtml(generatedAt)} · ${escapeHtml(labels.sourceCount.replace('{{count}}', String(sources.length)))}</p>
  ${table}
  <script>window.addEventListener('load', () => window.print())</script>
</body>
</html>`
}
