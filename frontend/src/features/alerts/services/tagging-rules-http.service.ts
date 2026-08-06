import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type {
  CreateTaggingRuleInput,
  TaggingRule,
  TaggingRuleListParams,
  UpdateTaggingRuleInput,
} from '../types/tagging-rule.types'

const api = createApiClient()

export { ApiError as TaggingRulesHttpError }

function buildListQuery(p: TaggingRuleListParams): string {
  const q = new URLSearchParams()
  q.set('page', String(p.page))
  q.set('size', String(p.size))
  if (p.name) q.set('name', p.name)
  if (p.conditionField) q.set('conditionField', p.conditionField)
  if (p.conditionValue) q.set('conditionValue', p.conditionValue)
  if (p.ruleActive !== undefined) q.set('ruleActive', String(p.ruleActive))
  if (p.ruleDeleted !== undefined) q.set('ruleDeleted', String(p.ruleDeleted))
  if (p.tagIds?.length) q.set('tagIds', p.tagIds.join(','))
  return q.toString()
}

/** REST client for /alert-tag-rules. Tag catalog management is shared with the
 * alerts feature — see alertsHttpService.tags/createTag/updateTag/deleteTag. */
export const taggingRulesHttpService = {
  list: (params: TaggingRuleListParams) =>
    api.getPaged<TaggingRule[]>(`/alert-tag-rules?${buildListQuery(params)}`),

  getById: (id: string) => api.get<TaggingRule | null>(`/alert-tag-rules/${id}`),

  create: (input: CreateTaggingRuleInput) => api.post<TaggingRule>('/alert-tag-rules', input),

  update: (input: UpdateTaggingRuleInput) => api.put<TaggingRule>('/alert-tag-rules', input),

  delete: (id: string) => api.delete<void>(`/alert-tag-rules/${id}`),
}
