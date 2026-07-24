/* Mirrors backend iam dto/api_keys.go. */

export interface ApiKey {
  id: number
  name: string
  allowed_ip: string[]
  created_at: string
  generated_at: string
  expires_at: string | null
}

export interface ApiKeyPageInfo {
  page: number
  page_size: number
  total_items: number
  total_pages: number
}

export interface ApiKeyListResponse {
  data: ApiKey[]
  page_info: ApiKeyPageInfo
}

export interface ApiKeyUpsertRequest {
  name: string
  allowed_ip: string[]
  expires_at: string | null
}

/** Response of POST /api-keys/:id/generate — the secret, shown once. */
export interface ApiKeyGenerateResponse {
  api_key: string
}
