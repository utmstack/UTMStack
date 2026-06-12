import { useCallback } from 'react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import {
  taggingRulesHttpService as svc,
  TaggingRulesHttpError,
} from '../services/tagging-rules-http.service'
import type {
  CreateTaggingRuleInput,
  TaggingRule,
  UpdateTaggingRuleInput,
} from '../types/tagging-rule.types'

export interface UseTaggingRuleMutationsResult {
  createRule: (input: CreateTaggingRuleInput) => Promise<TaggingRule | null>
  updateRule: (input: UpdateTaggingRuleInput) => Promise<TaggingRule | null>
  deleteRule: (id: number, name: string) => Promise<boolean>
}

export function useTaggingRuleMutations(refresh: () => void): UseTaggingRuleMutationsResult {
  const { t } = useTranslation()

  const createRule = useCallback(
    async (input: CreateTaggingRuleInput) => {
      try {
        const rule = await svc.create(input)
        toast.success(t('taggingRules.toast.created', { name: input.name }))
        refresh()
        return rule
      } catch (e) {
        toast.error(e instanceof TaggingRulesHttpError ? e.message : t('taggingRules.toast.createError'))
        return null
      }
    },
    [t, refresh]
  )

  const updateRule = useCallback(
    async (input: UpdateTaggingRuleInput) => {
      try {
        const rule = await svc.update(input)
        toast.success(t('taggingRules.toast.updated', { name: input.name }))
        refresh()
        return rule
      } catch (e) {
        toast.error(e instanceof TaggingRulesHttpError ? e.message : t('taggingRules.toast.updateError'))
        return null
      }
    },
    [t, refresh]
  )

  const deleteRule = useCallback(
    async (id: number, name: string) => {
      try {
        await svc.delete(id)
        toast.success(t('taggingRules.toast.deleted', { name }))
        refresh()
        return true
      } catch (e) {
        toast.error(e instanceof TaggingRulesHttpError ? e.message : t('taggingRules.toast.deleteError'))
        return false
      }
    },
    [t, refresh]
  )

  return { createRule, updateRule, deleteRule }
}
