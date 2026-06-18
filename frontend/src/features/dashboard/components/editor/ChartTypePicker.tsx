import { useTranslation } from 'react-i18next'
import { BarChart3 } from 'lucide-react'
import { CHART_TYPES } from '@/features/dashboard/constants'
import { CHART_ICONS } from '@/features/dashboard/components/chart-icon'
import { cn } from '@/shared/lib/utils'
import type { ChartTypeId } from '@/features/dashboard/types'

export function ChartTypePicker({
  value,
  onChange,
}: {
  value: ChartTypeId
  onChange: (next: ChartTypeId) => void
}) {
  const { t } = useTranslation()
  return (
    <div className="grid auto-rows-fr grid-cols-2 gap-2.5 sm:grid-cols-3 lg:grid-cols-4">
      {CHART_TYPES.map((c) => {
        const Icon = CHART_ICONS[c.icon] ?? BarChart3
        const active = value === c.id
        return (
          <button
            key={c.id}
            type="button"
            onClick={() => onChange(c.id)}
            className={cn(
              'flex h-full flex-col items-start gap-1.5 rounded-lg border p-3 text-left transition-colors',
              active
                ? 'border-primary bg-primary/5 text-foreground ring-1 ring-primary/30'
                : 'border-border hover:border-border hover:bg-muted'
            )}
          >
            <Icon size={18} className={cn('shrink-0', active ? 'text-primary' : 'text-muted-foreground')} />
            <span className="text-sm font-medium leading-tight">
              {t(`dashboards.editor.chartTypes.${c.id}.label`)}
            </span>
            <span className="line-clamp-2 text-[11px] leading-snug text-muted-foreground">
              {t(`dashboards.editor.chartTypes.${c.id}.description`)}
            </span>
          </button>
        )
      })}
    </div>
  )
}
