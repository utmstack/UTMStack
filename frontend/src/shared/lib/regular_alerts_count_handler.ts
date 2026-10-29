import { toast } from 'sonner'
import i18n from '@/shared/i18n'
export const NEW_ALERT_NOTIFICATION_INDEX = 'NEW_ALERT_NOTIFICATION_INDEX'
/**
 * Default new-alert handler: toast "{n} new alerts"; click navigates to the
 * alerts page. Registered once at initial page load (see
 * `AlertCountProvider` in `src/features/alerts/hooks/use-alert-count`).
 */
export const alertNotificationHandler = (
  newCount: number,
): void => {


  const go_alerts = () => window.location.href='/threat-management/alerts'

  toast(i18n.t('alerts.toast.newAlertsToast_title'), {
    description: i18n.t('alerts.toast.newAlertsToast_count', { count: newCount }),
    action: go_alerts
      ? { label: i18n.t('alerts.toast.newAlertsToast_view'), onClick: go_alerts }
      : undefined,
  })
}
