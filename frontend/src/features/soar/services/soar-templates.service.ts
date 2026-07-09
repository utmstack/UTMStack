import { createApiClient } from '@/shared/lib/api-client'
import type { ActionTemplate, ActionTemplateListQuery } from '../types/soar.types'

const api = createApiClient()
const BASE = '/soar/action-templates'

export const soarTemplatesService = {
  // Returns { data: ActionTemplate[], total } — total from X-Total-Count.
  list: (q: ActionTemplateListQuery = {}) => {
    const p = new URLSearchParams()
    p.set('page', String(q.page ?? 0))
    p.set('size', String(q.size ?? 20))
    return api.getPaged<ActionTemplate[]>(`${BASE}?${p.toString()}`)
  },
}
