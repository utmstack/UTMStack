import { useTranslation } from 'react-i18next'
import { ViewToggle } from './view-toggle'

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
