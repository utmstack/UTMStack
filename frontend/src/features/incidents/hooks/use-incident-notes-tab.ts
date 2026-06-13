import { useCallback, useEffect, useState } from 'react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { incidentsHttpService as svc, IncidentsHttpError } from '../services/incidents-http.service'
import type { IncidentNote } from '../types/incident.types'

export interface UseIncidentNotesTabResult {
  rows: IncidentNote[] | null
  error: boolean
  busy: boolean
  reload: () => Promise<void>
  add: (text: string) => Promise<void>
}

/** Data + add-note mutation for the incident drawer's "Notes" tab. */
export function useIncidentNotesTab(incidentId: number): UseIncidentNotesTabResult {
  const { t } = useTranslation()
  const [rows, setRows] = useState<IncidentNote[] | null>(null)
  const [error, setError] = useState(false)
  const [busy, setBusy] = useState(false)

  const load = useCallback(async () => {
    setError(false)
    try {
      const { data } = await svc.notes(incidentId)
      setRows((data ?? []).slice().sort((a, b) => +new Date(b.noteSendDate) - +new Date(a.noteSendDate)))
    } catch {
      setError(true)
    }
  }, [incidentId])

  useEffect(() => {
    void load()
  }, [load])

  const add = useCallback(
    async (text: string) => {
      const v = text.trim()
      if (!v || busy) return
      setBusy(true)
      try {
        await svc.addNote(incidentId, v)
        await load()
      } catch (e) {
        toast.error(e instanceof IncidentsHttpError ? e.message : t('incidents.notes.addError'))
      } finally {
        setBusy(false)
      }
    },
    [busy, incidentId, load, t]
  )

  return { rows, error, busy, reload: load, add }
}
