import { createApiClient } from '@/shared/lib/api-client'

const api = createApiClient()

/** The roles a mapping may grant. The backend already limits this to the
 * tenant's own plus the ones the product ships, so whatever comes back here is
 * exactly what may be handed out. */
export interface RoleOption {
  id: string
  name: string
  display_name: string
  system: boolean
}

export const rolesHttpService = {
  list: () => api.get<RoleOption[]>('/roles'),
}
