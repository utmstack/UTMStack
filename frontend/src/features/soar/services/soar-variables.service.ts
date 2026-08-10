import { createApiClient } from '@/shared/lib/api-client'

const api = createApiClient()
const BASE = '/soar/variables'

/** A SOAR command variable. Secret values come back masked as "****". */
export interface SoarVariable {
  id: string // uuid
  variableName: string
  variableValue: string
  variableDescription?: string | null
  isSecret: boolean
  createdBy: string
  createdAt: string
  lastModifiedBy?: string | null
  lastModifiedDate?: string | null
}

export interface CreateVariableInput {
  variableName: string
  /** Required: a variable with no value leaves its $[variables.NAME] unresolved
   *  in the command that reaches the agent, so the backend refuses it. */
  variableValue: string
  variableDescription?: string
  isSecret: boolean
}

export interface UpdateVariableInput {
  id: string
  variableName?: string
  /** Omit, or send the "****" mask, to keep the stored value. */
  variableValue?: string
  variableDescription?: string
  isSecret: boolean
}

export const soarVariablesService = {
  // List returns the array as the body (X-Total-Count carries the total).
  list: () => api.get<SoarVariable[]>(BASE),
  create: (input: CreateVariableInput) => api.post<SoarVariable>(BASE, input),
  // Secrets are mask-preserving: omit/blank the value to keep the stored one.
  update: (input: UpdateVariableInput) => api.put<SoarVariable>(BASE, input),
  remove: (id: string) => api.delete<{ message: string }>(`${BASE}/${id}`),
}
