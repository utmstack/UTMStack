import { useTranslation } from 'react-i18next'

type Row = Record<string, unknown>

export function TextRenderer({ rows }: { rows: Row[] }) {
  const { t } = useTranslation()

  if (rows.length === 0) {
    return (
      <div className="flex h-full w-full items-center justify-center text-xs text-muted-foreground">
        {t('dashboards.widget.noData')}
      </div>
    )
  }

  return (
    <div className="flex h-full w-full flex-col gap-2 overflow-auto px-4 py-3 text-sm leading-relaxed text-foreground/90">
      {rows.map((r, i) => {
        const keys = Object.keys(r)
        const value = keys.length === 1 ? r[keys[0]] : keys.map((k) => `${k}: ${formatValue(r[k])}`).join(' · ')
        return (
          <p key={i} className="whitespace-pre-wrap break-words">
            {formatValue(value)}
          </p>
        )
      })}
    </div>
  )
}

function formatValue(value: unknown): string {
  if (value == null) return ''
  if (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') {
    return String(value)
  }
  try {
    return JSON.stringify(value)
  } catch {
    return ''
  }
}
