import { useCallback } from 'react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { alertsHttpService as svc, AlertsHttpError } from '../services/alerts-http.service'
import type { Alert, FilterType } from '../types/alert.types'

export interface UseAlertMutationsArgs {
  refresh: () => void
  clearSelection: () => void
  openAlert: Alert | null
  setOpenAlert: (a: Alert | null) => void
}

export interface UseAlertMutationsResult {
  applyStatus: (ids: string[], status: number, observation?: string, fp?: boolean) => Promise<void>
  applyTags: (ids: string[], tags: string[]) => Promise<void>
  updateNotes: (alertId: string, notes: string) => Promise<void>
  exportCsv: (filters: FilterType[]) => Promise<void>
}

/**
 * Mutations that touch one or many alerts. On success they:
 *   - refresh the page (`refresh()`),
 *   - clear the bulk-selection,
 *   - patch `openAlert` locally for instant feedback, then
 *   - re-fetch the updated doc so status/tags/history reflect the change
 *     (1.2s delay lets OpenSearch make the new version searchable).
 */
export function useAlertMutations({
  refresh,
  clearSelection,
  openAlert,
  setOpenAlert,
}: UseAlertMutationsArgs): UseAlertMutationsResult {
  const { t } = useTranslation()

  const refetchOpen = useCallback(
    (id: string) => {
      setTimeout(() => {
        void svc
          .getById(id)
          .then((fresh) => {
            if (fresh && openAlert && openAlert.id === id) setOpenAlert(fresh)
          })
          .catch(() => {})
      }, 1200)
    },
    [openAlert, setOpenAlert]
  )

  const applyStatus = useCallback(
    async (ids: string[], status: number, observation = '', fp = false) => {
      try {
        await svc.updateStatus(ids, status, observation, fp)
        toast.success(t('alerts.toast.statusUpdated'))
        clearSelection()
        refresh()
        if (openAlert && ids.includes(openAlert.id)) {
          setOpenAlert({ ...openAlert, status })
          refetchOpen(openAlert.id)
        }
      } catch (e) {
        toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.toast.statusFailed'))
      }
    },
    [t, clearSelection, refresh, openAlert, setOpenAlert, refetchOpen]
  )

  const applyTags = useCallback(
    async (ids: string[], tags: string[]) => {
      try {
        await svc.updateTags(ids, tags)
        toast.success(t('alerts.toast.tagsUpdated'))
        clearSelection()
        refresh()
        if (openAlert && ids.includes(openAlert.id)) {
          setOpenAlert({ ...openAlert, tags })
          refetchOpen(openAlert.id)
        }
      } catch (e) {
        toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.toast.tagsFailed'))
      }
    },
    [t, clearSelection, refresh, openAlert, setOpenAlert, refetchOpen]
  )

  const updateNotes = useCallback(
    async (alertId: string, notes: string) => {
      try {
        await svc.updateNotes(alertId, notes)
        toast.success(t('alerts.toast.notesSaved'))
        if (openAlert && openAlert.id === alertId) {
          setOpenAlert({ ...openAlert, notes })
          refetchOpen(alertId)
        }
      } catch {
        toast.error(t('alerts.toast.notesFailed'))
      }
    },
    [t, openAlert, setOpenAlert, refetchOpen]
  )

  const exportCsv = useCallback(
    async (filters: FilterType[]) => {
      try {
        await svc.exportCsv(filters)
      } catch (e) {
        toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.toast.exportFailed'))
      }
    },
    [t]
  )

  return { applyStatus, applyTags, updateNotes, exportCsv }
}
