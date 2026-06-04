export interface Workspace {
  id: number
  slug: string
  name: string
  description?: string
  is_default: boolean
  created_at: string
  updated_at: string
  created_by?: string
}

export interface CreateWorkspaceInput {
  name: string
  slug?: string
  description?: string
}

export interface UpdateWorkspaceInput {
  name?: string
  description?: string
}
