import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

type Row = Record<string, unknown>

export function TagCloudRenderer({ rows }: { rows: Row[] }) {
  const { t } = useTranslation()

  const entries = useMemo(() => parseEntries(rows), [rows])

  if (entries.length === 0) {
    return (
      <div className="flex h-full w-full items-center justify-center text-xs text-muted-foreground">
        {t('dashboards.widget.noData')}
      </div>
    )
  }

  const max = Math.max(...entries.map((e) => e.value))
  const min = Math.min(...entries.map((e) => e.value))
  const span = max - min || 1

  return (
    <div className="flex h-full w-full flex-wrap items-center justify-center gap-3 overflow-auto p-3">
      {entries.map((e, i) => {
        const ratio = (e.value - min) / span
        const fontPx = 12 + Math.round(ratio * 36)
        const opacity = 0.55 + ratio * 0.45
        return (
          <span
            key={`${e.label}-${i}`}
            title={`${e.label}: ${e.value}`}
            style={{ fontSize: `${fontPx}px`, opacity }}
            className="inline-block whitespace-nowrap font-semibold text-foreground"
          >
            {e.label}
          </span>
        )
      })}
    </div>
  )
}

function parseEntries(rows: Row[]): { label: string; value: number }[] {
  return rows
    .map((r) => {
      const keys = Object.keys(r)
      const labelKey = keys.find((k) => k === 'label' || k === 'x' || k === 'term') ?? keys[0]
      const valueKey = keys.find((k) => k === 'value' || k === 'y' || k === 'count') ?? keys[1]
      const labelRaw = labelKey != null ? r[labelKey] : null
      const valueRaw = valueKey != null ? r[valueKey] : null
      const label = labelRaw == null ? '' : String(labelRaw)
      const value = typeof valueRaw === 'number' ? valueRaw : Number(valueRaw)
      if (!label || !Number.isFinite(value)) return null
      return { label, value }
    })
    .filter((x): x is { label: string; value: number } => x != null)
}
