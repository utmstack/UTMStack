export interface TeamUser {
  id: number;
  login: string;
  email?: string;
  first_name?: string;
  last_name?: string;
  image_url?: string;
  activated: boolean;
  pending: boolean;
  tfa_enabled: boolean;
  created_by?: string;
  created_at: string;
}

export interface TeamUserPageInfo {
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

export interface TeamUserListResponse {
  data: TeamUser[];
  page_info: TeamUserPageInfo;
}

export interface TeamUserCreatePayload {
  login: string;
  email: string;
  first_name: string;
  last_name: string;
  lang_key: string;
}

export interface TeamUserUpdatePayload {
  email: string;
  first_name: string;
  last_name: string;
  activated?: boolean;
}

export interface TeamUserListQuery {
  page?: number;
  page_size?: number;
  search?: string;
}
