import { createApiClient } from '@/shared/lib/api-client'

const api = createApiClient()
const BASE = '/soar/incident-variables'

/** A SOAR command variable. Secret values come back masked as "****". */
export interface SoarVariable {
  id: number
  variableName?: string | null
  variableValue?: string | null
  variableDescription?: string | null
  isSecret: boolean
  lastModifiedDate?: string | null
}

export interface CreateVariableInput {
  variableName: string
  variableDescription?: string
  variableValue?: string
  isSecret: boolean
}

export interface UpdateVariableInput extends CreateVariableInput {
  id: number
}

export const soarVariablesService = {
  // List returns the array as the body (X-Total-Count carries the total).
  list: () => api.get<SoarVariable[]>(BASE),
  create: (input: CreateVariableInput) => api.post<SoarVariable>(BASE, input),
  // Secrets are mask-preserving: omit/blank the value to keep the stored one.
  update: (input: UpdateVariableInput) => api.put<SoarVariable>(BASE, input),
  remove: (id: number) => api.delete<{ message: string }>(`${BASE}/${id}`),
}
