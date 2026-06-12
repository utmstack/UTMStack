/* Mirrors backend /alert-tag-rules DTOs. Conditions are a conjunction (AND) of
   FilterType rows; tags are full catalog refs (the backend stores them as a CSV
   of IDs and resolves them back on read). */

import type { FilterType, AlertTag } from './alert.types'

export type { FilterType, AlertTag }

/** GET /alert-tag-rules item / POST·PUT response. */
export interface TaggingRule {
  id: number
  name: string
  description: string
  conditions: FilterType[]
  tags: AlertTag[]
  active: boolean
  deleted: boolean
  createdBy: string
  createdDate: string
  lastModifiedBy?: string
  lastModifiedDate?: string
}

/** POST /alert-tag-rules body. */
export interface CreateTaggingRuleInput {
  name: string
  description: string
  conditions: FilterType[]
  tags: AlertTag[]
}

/** PUT /alert-tag-rules body. */
export interface UpdateTaggingRuleInput extends CreateTaggingRuleInput {
  id: number
}

/** Query params for GET /alert-tag-rules. */
export interface TaggingRuleListParams {
  page: number // 1-based, matches backend
  size: number
  name?: string
  conditionField?: string
  conditionValue?: string
  ruleActive?: boolean
  ruleDeleted?: boolean
  tagIds?: number[]
}
