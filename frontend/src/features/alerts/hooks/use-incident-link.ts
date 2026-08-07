import { useCallback, useEffect, useMemo, useState } from 'react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { alertsHttpService as svc, AlertsHttpError } from '../services/alerts-http.service'
import { incidentsHttpService } from '@/features/incidents/services/incidents-http.service'
import type { AlertLinkItem, Incident } from '@/features/incidents/types/incident.types'
import { STATUS_BY_VALUE, STATUS_VALUE } from '../types/alert.types'
import type { Alert } from '../types/alert.types'

export type IncidentMode = 'new' | 'existing'

export interface UseIncidentLinkArgs {
  alerts: Alert[]
  mode: IncidentMode
  onDone: () => void
}

export interface UseIncidentLinkResult {
  incidents: Incident[]
  loadingIncidents: boolean
  busy: boolean
  alertList: AlertLinkItem[]
  submit: (input: { name: string; description: string; incidentId: string }) => Promise<void>
}

/**
 * Drives the "add to incident" modal: lazy-fetches the incident list when the
 * user picks the "existing" tab, and submits either a new incident or an
 * append to an existing one — followed by the convert-to-incident flag that
 * stamps the alert docs.
 */
export function useIncidentLink({ alerts, mode, onDone }: UseIncidentLinkArgs): UseIncidentLinkResult {
  const { t } = useTranslation()
  const [incidents, setIncidents] = useState<Incident[]>([])
  const [loadingIncidents, setLoadingIncidents] = useState(false)
  const [busy, setBusy] = useState(false)

  const alertList = useMemo<AlertLinkItem[]>(
    () =>
      alerts.map((a) => ({
        alertId: a.id,
        alertName: a.name || a.id,
        alertSeverity: a.severity ?? 'low',
        alertStatus: STATUS_VALUE[STATUS_BY_VALUE[a.status ?? ''] ?? 'open'],
      })),
    [alerts]
  )

  useEffect(() => {
    if (mode !== 'existing' || incidents.length) return
    setLoadingIncidents(true)
    incidentsHttpService
      .list()
      .then((r) => setIncidents(r.data ?? []))
      .catch(() => setIncidents([]))
      .finally(() => setLoadingIncidents(false))
  }, [mode, incidents.length])

  const submit = useCallback(
    async ({ name, description, incidentId }: { name: string; description: string; incidentId: string }) => {
      if (busy) return
      setBusy(true)
      try {
        const inc =
          mode === 'new'
            ? await incidentsHttpService.create({
                incidentName: name.trim(),
                incidentDescription: description.trim() || undefined,
                alertList,
              })
            : await incidentsHttpService.addAlerts(incidentId, alertList)
        await svc.convertToIncident(
          alerts.map((a) => a.id),
          inc.incidentName,
          inc.id
        )
        toast.success(mode === 'new' ? t('alerts.toast.incidentCreated') : t('alerts.toast.alertsAdded'))
        onDone()
      } catch (e) {
        toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.toast.incidentFailed'))
        setBusy(false)
      }
    },
    [busy, mode, alertList, alerts, t, onDone]
  )

  return { incidents, loadingIncidents, busy, alertList, submit }
}
