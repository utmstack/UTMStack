import { BarChart3, Rows3 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { ViewSwitch } from '@/shared/components/ui/view-switch'
import type { AlertsView } from './alerts-header'

export function ViewToggle({ view, onView }: { view: AlertsView; onView: (v: AlertsView) => void }) {
  const { t } = useTranslation()
  return (
    <ViewSwitch
      value={view}
      onChange={onView}
      options={[
        { id: 'alerts', icon: Rows3, label: t('alerts.view.alerts') },
        { id: 'overview', icon: BarChart3, label: t('alerts.view.overview') },
      ]}
    />
  )
}
