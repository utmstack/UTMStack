import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Map } from 'lucide-react'

type Row = Record<string, unknown>

export function RegionMapRenderer({ rows }: { rows: Row[] }) {
  const { t } = useTranslation()

  const entries = useMemo(() => parseEntries(rows), [rows])

  if (entries.length === 0) {
    return (
      <div className="flex h-full w-full flex-col items-center justify-center gap-2 text-xs text-muted-foreground">
        <Map size={20} className="text-muted-foreground/60" />
        <span>{t('dashboards.widget.noData')}</span>
      </div>
    )
  }

  return (
    <div className="flex h-full w-full flex-col gap-2 overflow-auto p-3">
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <Map size={14} />
        <span>{t('dashboards.widget.regionMapPending', { count: entries.length })}</span>
      </div>
      <ul className="grid grid-cols-2 gap-1 text-xs sm:grid-cols-3">
        {entries.map((e, i) => (
          <li
            key={`${e.region}-${i}`}
            className="flex items-center justify-between rounded-md border border-border/60 bg-background/30 px-2 py-1.5"
          >
            <span className="truncate font-medium">{e.region}</span>
            <span className="text-muted-foreground">{e.value}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}

function parseEntries(rows: Row[]): { region: string; value: number }[] {
  return rows
    .map((r) => {
      const keys = Object.keys(r)
      const regionKey = keys.find((k) => k === 'region' || k === 'country' || k === 'x') ?? keys[0]
      const valueKey = keys.find((k) => k === 'value' || k === 'y' || k === 'count') ?? keys[1]
      const region = regionKey != null ? String(r[regionKey] ?? '') : ''
      const valueRaw = valueKey != null ? r[valueKey] : 0
      const value = typeof valueRaw === 'number' ? valueRaw : Number(valueRaw)
      if (!region) return null
      return { region, value: Number.isFinite(value) ? value : 0 }
    })
    .filter((x): x is { region: string; value: number } => x != null)
}
