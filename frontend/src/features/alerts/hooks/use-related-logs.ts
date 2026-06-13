import { useCallback, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { alertsHttpService as svc, AlertsHttpError } from '../services/alerts-http.service'
import type { Alert } from '../types/alert.types'

export interface UseRelatedLogsResult {
  loading: boolean
  viewRelatedLogs: (alert: Alert) => Promise<void>
}

/**
 * Reproduces the Event Processor's correlation search server-side (uncapped)
 * and deep-links into the Log Explorer with the matching ids preloaded.
 */
export function useRelatedLogs(): UseRelatedLogsResult {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)

  const viewRelatedLogs = useCallback(
    async (alert: Alert) => {
      setLoading(true)
      try {
        const r = await svc.relatedLogs(alert.id)
        if (!r.ids?.length) {
          toast.info(t('alerts.toast.noRelatedLogs'))
          return
        }
        navigate('/log-explorer', {
          state: {
            relatedLogs: {
              ids: r.ids,
              indexPattern: r.indexPattern,
              timeFrom: r.timeFrom,
              timeTo: r.timeTo,
              alertName: alert.name,
              truncated: r.truncated,
            },
          },
        })
      } catch (e) {
        toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.toast.relatedLogsFailed'))
      } finally {
        setLoading(false)
      }
    },
    [navigate, t]
  )

  return { loading, viewRelatedLogs }
}
