import { useCallback, useState } from 'react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { incidentsHttpService as svc, IncidentsHttpError } from '../services/incidents-http.service'
import type { Incident, IncidentStatus } from '../types/incident.types'

export interface UseIncidentStatusResult {
  busy: boolean
  changeStatus: (status: IncidentStatus, solution?: string) => Promise<void>
}

/** Status transitions for a single incident. */
export function useIncidentStatus(incident: Incident, onChanged: (id: string) => void): UseIncidentStatusResult {
  const { t } = useTranslation()
  const [busy, setBusy] = useState(false)

  const changeStatus = useCallback(
    async (status: IncidentStatus, solution?: string) => {
      if (busy || status === incident.incidentStatus) return
      setBusy(true)
      try {
        await svc.changeStatus({
          id: incident.id,
          incidentName: incident.incidentName,
          incidentStatus: status,
          incidentSolution:
            status === 'Completed' && solution && solution.trim() ? solution.trim() : incident.incidentSolution,
        })
        toast.success(t('incidents.toast.statusChanged'))
        onChanged(incident.id)
      } catch (e) {
        toast.error(e instanceof IncidentsHttpError ? e.message : t('incidents.toast.statusError'))
      } finally {
        setBusy(false)
      }
    },
    [busy, incident, onChanged, t]
  )

  return { busy, changeStatus }
}
