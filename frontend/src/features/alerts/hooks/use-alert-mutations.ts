import { useCallback, useEffect, useRef, useState } from 'react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { alertsHttpService as svc, AlertsHttpError } from '../services/alerts-http.service'
import { type Alert, type FilterType } from '../types/alert.types'
import { waitForAiSummary } from '../lib/ai-analysis'

export interface UseAlertMutationsArgs {
  refresh: () => void
  clearSelection: () => void
  openAlert: Alert | null
  setOpenAlert: (a: Alert | null) => void
}

export interface UseAlertMutationsResult {
  applyStatus: (ids: string[], status: string, observation?: string, fp?: boolean) => Promise<void>
  applyTags: (ids: string[], tags: string[]) => Promise<void>
  updateNotes: (alertId: string, notes: string) => Promise<void>
  updateAssignee: (alertId: string, assignee: string) => Promise<void>
  exportCsv: (filters: FilterType[]) => Promise<void>
  /** Asks SOC-AI for a summary of the alert and waits until it is on the alert. */
  generateAiSummary: (alert: Alert) => Promise<void>
  /** Ids of the alerts whose summary is being generated. */
  aiGenerating: ReadonlySet<string>
}

/**
 * Mutations that touch one or many alerts. On success they:
 *   - refresh the page (`refresh()`),
 *   - clear the bulk-selection,
 *   - patch `openAlert` locally for instant feedback, then
 *   - re-fetch the updated doc so status/tags/history reflect the change
 *     (1.2s delay lets the store make the new version readable).
 */
export function useAlertMutations({
  refresh,
  clearSelection,
  openAlert,
  setOpenAlert,
}: UseAlertMutationsArgs): UseAlertMutationsResult {
  const { t } = useTranslation()
  const [aiGenerating, setAiGenerating] = useState<ReadonlySet<string>>(new Set())
  // The wait outlives the drawer: closing it must not lose the result, but
  // leaving the page must stop it.
  const mounted = useRef(true)
  const openAlertRef = useRef(openAlert)
  openAlertRef.current = openAlert
  useEffect(() => {
    mounted.current = true
    return () => {
      mounted.current = false
    }
  }, [])

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
    async (ids: string[], status: string, observation = '', fp = false) => {
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

  const updateAssignee = useCallback(
    async (alertId: string, assignee: string) => {
      try {
        await svc.updateAssignee(alertId, assignee)
        toast.success(assignee ? t('alerts.toast.assigned') : t('alerts.toast.unassigned'))
        if (openAlert && openAlert.id === alertId) {
          setOpenAlert({ ...openAlert, assignee })
          refetchOpen(alertId)
        }
        refresh()
      } catch {
        toast.error(t('alerts.toast.assignFailed'))
      }
    },
    [t, openAlert, setOpenAlert, refetchOpen, refresh]
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

  const generateAiSummary = useCallback(
    async (alert: Alert) => {
      const setGenerating = (on: boolean) =>
        setAiGenerating((prev) => {
          const next = new Set(prev)
          if (on) next.add(alert.id)
          else next.delete(alert.id)
          return next
        })

      setGenerating(true)
      try {
        await svc.analyzeWithAi(alert)
      } catch (e) {
        setGenerating(false)
        toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.ai.failed'))
        return
      }

      const fresh = await waitForAiSummary(() => svc.getById(alert.id), { isCancelled: () => !mounted.current })
      if (!mounted.current) return
      setGenerating(false)
      if (!fresh) {
        toast.message(t('alerts.ai.slow'))
        return
      }
      toast.success(t('alerts.ai.ready'))
      refresh()
      if (openAlertRef.current?.id === alert.id) setOpenAlert(fresh)
    },
    [t, refresh, setOpenAlert]
  )

  return { applyStatus, applyTags, updateNotes, updateAssignee, exportCsv, generateAiSummary, aiGenerating }
}
