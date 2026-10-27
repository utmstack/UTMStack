import { STATUS_KEYS, valueOrDash } from './format'
import type { DarkWebRecord } from '../types'

/**
 * Client-side export helpers for the dark web monitoring demo. The data is
 * synthetic for now, so reports are generated entirely in the browser; when
 * the real endpoints land these move next to the service that fetches them.
 */

export type ExportAction = 'CSV' | 'PDF'

type Translate = (key: string) => string

const EXPORT_HEADERS = [
  'status',
  'who',
  'domain',
  'email',
  'ipAddress',
  'username',
  'password',
  'hashedPassword',
  'database',
  'leakName',
  'foundIn',
  'source',
  'compromisedData',
  'organizationElement',
  'breached',
  'ingested',
] as const

export const MASKED = '••••••••'

function csvCell(value: string): string {
  // Prefix formula-triggering characters so spreadsheet apps treat the cell as text.
  const neutral = /^[=+@\-]/.test(value) ? `'${value}` : value
  return /[",\n]/.test(neutral) ? `"${neutral.replace(/"/g, '""')}"` : neutral
}

function escapeHtml(value: string): string {
  return value.replace(/[&<>"']/g, (character) => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
  })[character] ?? character)
}

export function downloadCsv(filename: string, rows: string[][]) {
  const csv = rows.map((row) => row.map(csvCell).join(',')).join('\r\n')
  const url = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8' }))
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}

/** Opens a print window for the rows. Returns false when pop-ups are blocked. */
export function printReportPdf(title: string, rows: string[][]): boolean {
  const printWindow = window.open('', '_blank')
  if (!printWindow) return false
  const header = rows[0] ?? []
  const body = rows.slice(1)
  printWindow.document.write(
    `<!doctype html><html><head><title>${escapeHtml(title)}</title><style>body{font:12px Arial,sans-serif;margin:24px}h1{font-size:18px}table{border-collapse:collapse;width:100%}th,td{border:1px solid #bbb;padding:6px;text-align:left;word-break:break-word}th{background:#eee}</style></head><body><h1>${escapeHtml(title)}</h1><table><thead><tr>${header.map((cell) => `<th>${escapeHtml(cell)}</th>`).join('')}</tr></thead><tbody>${body.map((row) => `<tr>${row.map((cell) => `<td>${escapeHtml(cell)}</td>`).join('')}</tr>`).join('')}</tbody></table><script>window.onload=()=>window.print()</script></body></html>`,
  )
  printWindow.document.close()
  return true
}

/** Rows with translated column headers and one row per record. */
export function buildExportRows(
  records: DarkWebRecord[],
  t: Translate,
  maskSensitive: boolean,
  formatDate: (value: string) => string,
): string[][] {
  const header = EXPORT_HEADERS.map((column) => t(`darkWebMonitoring.table.${column}`))
  const rows = records.map((record) => [
    t(`darkWebMonitoring.status.${STATUS_KEYS[record.breachStatus]}`),
    record.employee ? t('darkWebMonitoring.employee') : t('darkWebMonitoring.other'),
    valueOrDash(record.domain),
    valueOrDash(record.email),
    valueOrDash(record.ipAddress),
    valueOrDash(record.username),
    maskSensitive ? (record.password ? MASKED : valueOrDash(null)) : valueOrDash(record.password),
    maskSensitive ? (record.hashedPassword ? MASKED : valueOrDash(null)) : valueOrDash(record.hashedPassword),
    valueOrDash(record.databaseName),
    valueOrDash(record.leakName),
    valueOrDash(record.foundIn),
    valueOrDash(record.source),
    valueOrDash(record.compromisedData),
    record.organizationElement,
    formatDate(record.timestamp),
    formatDate(record.createdAt),
  ])
  return [header, ...rows]
}
