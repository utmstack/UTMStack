export interface AuditLog {
  id: number
  timestamp: string
  user_id?: number
  user_login?: string
  ip?: string
  user_agent?: string
  session_id?: number
  action: string
  status: 'success' | 'failure'
  error_message?: string
  resource_type?: string
  resource_id?: string
  metadata?: Record<string, unknown>
  prev_hash?: string
  hash: string
}

export interface AuditPageInfo {
  page: number
  page_size: number
  total_items: number
  total_pages: number
  has_next: boolean
  has_prev: boolean
}

export interface AuditListResponse {
  data: AuditLog[]
  page_info: AuditPageInfo
}

export interface AuditListQuery {
  user_id?: number
  action?: string
  status?: 'success' | 'failure' | ''
  resource_type?: string
  resource_id?: string
  from?: string
  to?: string
  page?: number
  page_size?: number
}
