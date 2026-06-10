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
}

/** Mirrors the backend common_models.ListResponse[LogResponse]. */
export interface AuditListResponse {
  items: AuditLog[]
  page_number: number
  page_size: number
  total_items: number
  total_pages: number
}

/**
 * UI-friendly filters. The service translates these into the backend's
 * page_number / page_size / search_query (DSL: field.op.value, &-joined) params.
 */
export interface AuditListQuery {
  user_id?: number
  user_login?: string
  action?: string
  status?: 'success' | 'failure' | ''
  resource_type?: string
  resource_id?: string
  /** RFC3339 timestamp lower bound. */
  from?: string
  /** RFC3339 timestamp upper bound. */
  to?: string
  /** 1-based page. */
  page?: number
  page_size?: number
}
