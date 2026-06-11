import { useTranslation } from 'react-i18next'
import {
  BarChart3,
  LineChart,
  AreaChart,
  PieChart,
  Hash,
  Table as TableIcon,
  type LucideIcon,
} from 'lucide-react'
import { CHART_TYPES } from '@/features/dashboard/constants'
import { cn } from '@/shared/lib/utils'
import type { ChartTypeId } from '@/features/dashboard/types'

const ICONS: Record<string, LucideIcon> = {
  bar: BarChart3,
  line: LineChart,
  area: AreaChart,
  pie: PieChart,
  metric: Hash,
  table: TableIcon,
}

export function ChartTypePicker({
  value,
  onChange,
}: {
  value: ChartTypeId
  onChange: (next: ChartTypeId) => void
}) {
  const { t } = useTranslation()
  return (
    <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-6">
      {CHART_TYPES.map((c) => {
        const Icon = ICONS[c.icon] ?? BarChart3
        const active = value === c.id
        return (
          <button
            key={c.id}
            type="button"
            onClick={() => onChange(c.id)}
            className={cn(
              'flex flex-col items-start gap-1 rounded-md border px-3 py-2.5 text-left transition-colors',
              active
                ? 'border-primary bg-primary/5 text-foreground'
                : 'border-border hover:bg-muted'
            )}
          >
            <Icon size={18} className={active ? 'text-primary' : 'text-muted-foreground'} />
            <span className="text-sm font-medium">
              {t(`dashboards.editor.chartTypes.${c.id}.label`)}
            </span>
            <span className="text-[10px] leading-tight text-muted-foreground">
              {t(`dashboards.editor.chartTypes.${c.id}.description`)}
            </span>
          </button>
        )
      })}
    </div>
  )
}
