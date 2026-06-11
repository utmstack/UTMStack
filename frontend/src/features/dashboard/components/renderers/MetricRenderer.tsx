import { useTranslation } from 'react-i18next'
import type { Row } from '@/features/dashboard/service/opensearch.service'

export function MetricRenderer({ rows, label }: { rows: Row[]; label?: string }) {
  const { t } = useTranslation()
  if (rows.length === 0) {
    return (
      <div className="flex h-full w-full items-center justify-center text-xs text-muted-foreground">
        {t('dashboards.widget.noData')}
      </div>
    )
  }

  const first = rows[0]
  const keys = Object.keys(first)
  let value: unknown
  if (keys.length === 1) {
    value = first[keys[0]]
  } else if ('y' in first) {
    value = (first as Record<string, unknown>)['y']
  } else if ('value' in first) {
    value = (first as Record<string, unknown>)['value']
  } else {
    value = first[keys[0]]
  }

  return (
    <div className="flex h-full w-full flex-col items-center justify-center gap-1 px-3 text-center">
      {label && (
        <span className="text-xs uppercase tracking-wider text-muted-foreground">{label}</span>
      )}
      <span className="text-4xl font-semibold tabular-nums">{formatValue(value)}</span>
    </div>
  )
}

function formatValue(v: unknown): string {
  if (v == null) return '—'
  if (typeof v === 'number') return v.toLocaleString()
  return String(v)
}
