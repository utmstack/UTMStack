import { BarChart3, Rows3 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'

export type AlertsView = 'alerts' | 'overview'

export function AlertsHeader({
  total,
  openCount,
  view,
  onView,
}: {
  total: number
  openCount: number | null
  view: AlertsView
  onView: (v: AlertsView) => void
}) {
  const { t } = useTranslation()
  return (
    <header className="flex flex-wrap items-center justify-between gap-3">
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <span className="font-medium text-foreground">{t('alerts.matching', { count: total.toLocaleString() })}</span>
        {openCount != null && (
          <>
            <span className="text-muted-foreground/50">·</span>
            <span>
              <span className="font-medium text-foreground">{openCount.toLocaleString()}</span> {t('alerts.openNow')}
            </span>
          </>
        )}
      </div>
      <ViewToggle view={view} onView={onView} />
    </header>
  )
}

function ViewToggle({ view, onView }: { view: AlertsView; onView: (v: AlertsView) => void }) {
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
