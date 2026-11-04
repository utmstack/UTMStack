import type { TriageRow } from '../lib/triage.types'

function csvCell(value: string | number): string {
  const text = String(value ?? '')
  return /[",\n\r]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text
}

/** CSV of a triage result, RFC 4180 CRLF. */
export function toCsv(rows: TriageRow[]): string {
  const header = ['indicator', 'type', 'verdict', 'reputation', 'accuracy', 'tags', 'last_seen', 'entity_id']
  const lines = rows.map(row =>
    [
      csvCell(row.input),
      csvCell(row.type),
      csvCell(row.verdict),
      csvCell(row.entity?.reputation ?? ''),
      csvCell(row.entity?.accuracy ?? ''),
      csvCell((row.entity?.tags ?? []).join(' ')),
      csvCell(row.entity?.lastSeen ?? ''),
      csvCell(row.entity?.id ?? ''),
    ].join(','),
  )
  return [header.join(','), ...lines].join('\r\n')
}
