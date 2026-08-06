import { useCallback, useEffect, useState } from 'react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { incidentsHttpService as svc, IncidentsHttpError } from '../services/incidents-http.service'
import { usersHttpService } from '@/features/team/services/team-http.service'

export interface UseIncidentAssignmentResult {
  users: string[] | null // null while loading
  busy: boolean
  assign: (email: string | null) => Promise<void>
}

/**
 * Lazy-loads IAM users for the assignee picker and handles assign/unassign.
 * `enabled` is gated by the popover open state so we don't pay the users API
 * cost until the user opens the picker.
 */
export function useIncidentAssignment(
  incidentId: number,
  enabled: boolean,
  onChanged: () => void
): UseIncidentAssignmentResult {
  const { t } = useTranslation()
  const [users, setUsers] = useState<string[] | null>(null)
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    if (!enabled || users !== null) return
    usersHttpService
      .list({ page_size: 500 })
      .then((r) => setUsers((r.data ?? []).map((u) => u.email).sort()))
      .catch(() => setUsers([]))
  }, [enabled, users])

  const assign = useCallback(
    async (email: string | null) => {
      if (busy) return
      setBusy(true)
      try {
        await svc.assign(incidentId, email)
        toast.success(email ? t('incidents.assign.assigned', { user: email }) : t('incidents.assign.unassigned'))
        onChanged()
      } catch (e) {
        toast.error(e instanceof IncidentsHttpError ? e.message : t('incidents.assign.error'))
      } finally {
        setBusy(false)
      }
    },
    [busy, incidentId, onChanged, t]
  )

  return { users, busy, assign }
}
