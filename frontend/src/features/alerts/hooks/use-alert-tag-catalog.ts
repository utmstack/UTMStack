import { useCallback, useEffect, useState } from 'react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { alertsHttpService as svc, AlertsHttpError } from '../services/alerts-http.service'
import type { AlertTag } from '../types/alert.types'

export interface UseAlertTagCatalogResult {
  tagCatalog: AlertTag[]
  refresh: () => Promise<void>
  createTag: (tagName: string, tagColor: string) => Promise<AlertTag | null>
  updateTag: (id: number, tagName: string, tagColor: string) => Promise<void>
  deleteTag: (id: number, tagName: string) => Promise<void>
}

/**
 * Tag catalog state + CRUD. System-owned tags are rejected by the backend with
 * 403 on update/delete; we surface the message via toast.
 *
 * `onTagDeleted` lets the caller react to a deletion (e.g. drop it from the
 * active tag filter) so this hook stays presentation-agnostic.
 */
export function useAlertTagCatalog(onTagDeleted?: (tagName: string) => void): UseAlertTagCatalogResult {
  const { t } = useTranslation()
  const [tagCatalog, setTagCatalog] = useState<AlertTag[]>([])

  const refresh = useCallback(async () => {
    try {
      setTagCatalog((await svc.tags()) ?? [])
    } catch {
      /* leave prior catalog */
    }
  }, [])

  useEffect(() => {
    void refresh()
  }, [refresh])

  const createTag = useCallback(
    async (tagName: string, tagColor: string): Promise<AlertTag | null> => {
      try {
        const tag = await svc.createTag(tagName, tagColor)
        await refresh()
        return tag
      } catch (e) {
        toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.toast.tagCreateFailed'))
        return null
      }
    },
    [refresh, t]
  )

  const updateTag = useCallback(
    async (id: number, tagName: string, tagColor: string) => {
      try {
        await svc.updateTag(id, tagName, tagColor)
        await refresh()
        toast.success(t('alerts.toast.tagUpdated'))
      } catch (e) {
        toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.toast.tagUpdateFailed'))
      }
    },
    [refresh, t]
  )

  const deleteTag = useCallback(
    async (id: number, tagName: string) => {
      try {
        await svc.deleteTag(id)
        await refresh()
        onTagDeleted?.(tagName)
        toast.success(t('alerts.toast.tagDeleted'))
      } catch (e) {
        toast.error(e instanceof AlertsHttpError ? e.message : t('alerts.toast.tagDeleteFailed'))
      }
    },
    [refresh, t, onTagDeleted]
  )

  return { tagCatalog, refresh, createTag, updateTag, deleteTag }
}
