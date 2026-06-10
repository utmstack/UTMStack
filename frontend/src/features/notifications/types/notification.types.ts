export type NotificationType = 'INFO' | 'WARNING' | 'ERROR'
export type NotificationStatus = 'ACTIVE' | 'HIDDEN' | 'DELETED'

export interface Notification {
  id: number
  source: string
  type: NotificationType
  message: string
  createdAt: string
  updatedAt?: string
  read: boolean
  status: NotificationStatus
}

export interface NotificationListQuery {
  /** 0-indexed page. */
  page?: number
  size?: number
  read?: boolean
  status?: NotificationStatus
  type?: NotificationType
  source?: string
  from?: string
  to?: string
  /** "field,asc|desc"; backend defaults to created_at desc. */
  sort?: string
}
