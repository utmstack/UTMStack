import { useCallback, useEffect, useState } from 'react'
import { incidentsHttpService as svc } from '../services/incidents-http.service'
import type { IncidentHistory } from '../types/incident.types'

export interface UseIncidentHistoryTabResult {
  rows: IncidentHistory[] | null
  error: boolean
  reload: () => Promise<void>
}

/** Data for the incident drawer's "History" tab — sorted newest-first. */
export function useIncidentHistoryTab(incidentId: string): UseIncidentHistoryTabResult {
  const [rows, setRows] = useState<IncidentHistory[] | null>(null)
  const [error, setError] = useState(false)

  const load = useCallback(async () => {
    setError(false)
    try {
      const { data } = await svc.history(incidentId)
      setRows((data ?? []).slice().sort((a, b) => +new Date(b.actionCreatedDate) - +new Date(a.actionCreatedDate)))
    } catch {
      setError(true)
    }
  }, [incidentId])

  useEffect(() => {
    void load()
  }, [load])

  return { rows, error, reload: load }
}
