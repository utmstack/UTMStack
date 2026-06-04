import { createApiClient } from '@/shared/lib/api-client'
import type {
  CreateWorkspaceInput,
  UpdateWorkspaceInput,
  Workspace,
} from '../types/workspace.types'

const api = createApiClient()

export const workspaceHttpService = {
  list: () => api.get<Workspace[]>('/workspaces'),
  get: (id: number) => api.get<Workspace>(`/workspaces/${id}`),
  create: (input: CreateWorkspaceInput) => api.post<Workspace>('/workspaces', input),
  update: (id: number, input: UpdateWorkspaceInput) =>
    api.put<Workspace>(`/workspaces/${id}`, input),
  remove: (id: number) => api.delete<void>(`/workspaces/${id}`),
}
