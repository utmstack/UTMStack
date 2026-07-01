import { BarChart3, Rows3 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import type { AlertsView } from './alerts-header'

export function ViewToggle({ view, onView }: { view: AlertsView; onView: (v: AlertsView) => void }) {
  const { t } = useTranslation()
  const opts = [
    { id: 'alerts' as const, icon: Rows3, label: t('alerts.view.alerts') },
    { id: 'overview' as const, icon: BarChart3, label: t('alerts.view.overview') },
  ]
  return (
    <div className="inline-flex shrink-0 items-center rounded-md border border-border bg-card p-0.5 text-xs">
      {opts.map(({ id, icon: Icon, label }) => (
        <button
          key={id}
          onClick={() => onView(id)}
          className={cn(
            'flex items-center gap-1.5 rounded px-2.5 py-1 transition-colors',
            view === id ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'
          )}
        >
          <Icon size={13} /> {label}
        </button>
      ))}
    </div>
  )
}
