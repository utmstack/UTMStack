import { useCallback, useEffect, useState } from 'react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { incidentsHttpService as svc, IncidentsHttpError } from '../services/incidents-http.service'
import type { IncidentAlert } from '../types/incident.types'

export interface UseIncidentAlertsTabResult {
  rows: IncidentAlert[] | null
  error: boolean
  reload: () => Promise<void>
  remove: (a: IncidentAlert) => void
  pendingDelete: IncidentAlert | null
  deleting: boolean
  cancelDelete: () => void
  confirmDelete: () => Promise<void>
}

/** Data + remove-alert mutation for the incident drawer's "Alerts" tab. */
export function useIncidentAlertsTab(incidentId: number, onChanged: () => void): UseIncidentAlertsTabResult {
  const { t } = useTranslation()
  const [rows, setRows] = useState<IncidentAlert[] | null>(null)
  const [error, setError] = useState(false)
  const [pendingDelete, setPendingDelete] = useState<IncidentAlert | null>(null)
  const [deleting, setDeleting] = useState(false)

  const load = useCallback(async () => {
    setError(false)
    try {
      const { data } = await svc.alerts(incidentId)
      setRows(data ?? [])
    } catch {
      setError(true)
    }
  }, [incidentId])

  useEffect(() => {
    void load()
  }, [load])

  const remove = useCallback((a: IncidentAlert) => setPendingDelete(a), [])

  const cancelDelete = useCallback(() => setPendingDelete(null), [])

  const confirmDelete = useCallback(async () => {
    if (!pendingDelete) return
    setDeleting(true)
    try {
      await svc.removeAlert(pendingDelete.id)
      toast.success(t('incidents.alerts.removed'))
      await load()
      onChanged()
    } catch (e) {
      toast.error(e instanceof IncidentsHttpError ? e.message : t('incidents.alerts.removeError'))
    } finally {
      setDeleting(false)
      setPendingDelete(null)
    }
  }, [pendingDelete, load, onChanged, t])

  return { rows, error, reload: load, remove, pendingDelete, deleting, cancelDelete, confirmDelete }
}
