import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ExternalLink, Loader2 } from 'lucide-react'
import { Link } from 'react-router-dom'
import { LogResults } from '@/features/log-explorer/components/log-results'
import { alertsHttpService } from '@/features/alerts/services/alerts-http.service'
import type { Alert } from '@/features/alerts/types/alert.types'

/**
 * The linked alert, opened where it is listed. Working an incident is reading
 * its alerts, and sending the analyst to another page for each one loses the
 * incident they were working.
 *
 * Read-only on purpose: the alert's own actions — status, tags, assignee — live
 * in the alerts drawer, and the link below leads there. Duplicating them here
 * would mean two places that change an alert and can disagree about it.
 */
export function IncidentAlertPreview({ alertId }: { alertId: string }) {
  const { t } = useTranslation()
  const [alert, setAlert] = useState<Alert | null>(null)
  const [state, setState] = useState<'loading' | 'ready' | 'missing' | 'error'>('loading')

  useEffect(() => {
    let cancelled = false
    setState('loading')
    alertsHttpService
      .getById(alertId)
      .then((a) => {
        if (cancelled) return
        setAlert(a)
        // The link survives the alert: retention drops the document long before
        // anyone deletes the incident that points at it.
        setState(a ? 'ready' : 'missing')
      })
      .catch(() => !cancelled && setState('error'))
    return () => {
      cancelled = true
    }
  }, [alertId])

  if (state === 'loading')
    return (
      <div className="flex items-center gap-2 px-4 py-3 text-xs text-muted-foreground">
        <Loader2 className="h-3.5 w-3.5 animate-spin" /> {t('incidents.alerts.loadingAlert')}
      </div>
    )

  if (state !== 'ready' || !alert)
    return (
      <div className="px-4 py-3 text-xs text-muted-foreground">
        {state === 'missing' ? t('incidents.alerts.alertGone') : t('incidents.alerts.alertFailed')}
      </div>
    )

  return (
    <div className="border-t border-border/60 bg-muted/20">
      <LogResults docs={[alert as unknown as Record<string, unknown>]} />
      <div className="flex justify-end px-4 py-2">
        <Link
          to="/threat-management/alerts"
          state={{
            socaiFilters: [{ field: 'id', operator: 'IS', value: alertId }],
            socaiTime: 'now-30d',
          }}
          className="inline-flex items-center gap-1 text-[11px] text-muted-foreground hover:text-foreground"
        >
          <ExternalLink size={11} /> {t('incidents.alerts.openInAlerts')}
        </Link>
      </div>
    </div>
  )
}
