import { useCallback } from 'react'
import { incidentsHttpService as svc } from '../services/incidents-http.service'
import type { Incident } from '../types/incident.types'

/**
 * After a mutation, refresh the list and re-pull the drawer's incident from
 * the server so its status/assignee/solution mirror the change. The drawer
 * setter receives a patcher that no-ops when the open incident has changed.
 */
export function useIncidentSync(
  refresh: () => void,
  setOpen: (updater: (prev: Incident | null) => Incident | null) => void
) {
  return useCallback(
    async (id: number) => {
      refresh()
      try {
        const fresh = await svc.getById(id)
        setOpen((prev) => (prev && prev.id === id ? { ...prev, ...fresh } : prev))
      } catch {
        /* keep current */
      }
    },
    [refresh, setOpen]
  )
}
