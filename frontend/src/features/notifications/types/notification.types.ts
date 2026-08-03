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
  /** Exact message (used to pull the notifications of a specific group). */
  message?: string
  from?: string
  to?: string
  /** "field,asc|desc"; backend defaults to created_at desc. */
  sort?: string
}

/** One row of the /notifications/grouped response — a stack of same source/type/message. */
export interface NotificationGroup {
  source: string
  type: NotificationType
  message: string
  count: number
  unreadCount: number
  lastCreated: string
}
