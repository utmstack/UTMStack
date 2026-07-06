import { useTranslation } from 'react-i18next'
import { RefreshCw } from 'lucide-react'

// Options in seconds. 0 = auto-refresh disabled.
export const REFRESH_OPTIONS: { value: number; labelKey: string }[] = [
  { value: 0, labelKey: 'dashboards.refresh.off' },
  { value: 10, labelKey: 'dashboards.refresh.10s' },
  { value: 30, labelKey: 'dashboards.refresh.30s' },
  { value: 60, labelKey: 'dashboards.refresh.1m' },
  { value: 300, labelKey: 'dashboards.refresh.5m' },
  { value: 900, labelKey: 'dashboards.refresh.15m' },
]

export function DashboardRefreshSelect({
  value,
  onChange,
  disabled,
}: {
  value: number
  onChange: (next: number) => void
  disabled?: boolean
}) {
  const { t } = useTranslation()
  return (
    <label
      className="flex h-8 items-center gap-1.5 rounded-md border border-input bg-background pl-2 pr-1 text-xs"
      title={t('dashboards.refresh.tooltip') ?? ''}
    >
      <RefreshCw size={12} className="text-muted-foreground" />
      <select
        value={String(value)}
        onChange={(e) => onChange(Number(e.target.value))}
        disabled={disabled}
        className="h-full bg-transparent pr-1 text-xs focus:outline-none"
      >
        {REFRESH_OPTIONS.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {t(opt.labelKey)}
          </option>
        ))}
      </select>
    </label>
  )
}
