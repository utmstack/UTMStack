import { createApiClient } from '@/shared/lib/api-client'
import type {
  Notification,
  NotificationListQuery,
  NotificationStatus,
} from '../types/notification.types'

const api = createApiClient()

function toQuery(q: NotificationListQuery): string {
  const params = new URLSearchParams()
  if (q.page != null) params.set('page', String(q.page))
  if (q.size != null) params.set('size', String(q.size))
  if (q.read != null) params.set('read', String(q.read))
  if (q.status) params.set('status', q.status)
  if (q.type) params.set('type', q.type)
  if (q.source) params.set('source', q.source)
  if (q.from) params.set('from', q.from)
  if (q.to) params.set('to', q.to)
  if (q.sort) params.set('sort', q.sort)
  const s = params.toString()
  return s ? `?${s}` : ''
}

export const notificationsHttpService = {
  /** List (bare array; newest first). Used by the infinite-scroll bell dropdown. */
  list: (q: NotificationListQuery = {}) => api.get<Notification[]>(`/notifications${toQuery(q)}`),
  /** Same list but with the total count (X-Total-Count) for classic pagination. */
  listPaged: (q: NotificationListQuery = {}) =>
    api.getPaged<Notification[]>(`/notifications${toQuery(q)}`),
  /** Count of unread active notifications (for the bell badge). */
  unreadCount: () => api.get<number>('/notifications/unread-count'),
  getById: (id: number) => api.get<Notification>(`/notifications/${id}`),
  /** Mark a single notification read/unread. */
  markRead: (id: number, read: boolean) =>
    api.put<Notification>(`/notifications/${id}/read?read=${read}`),
  /** Mark every notification read. */
  markAllRead: () => api.put<void>('/notifications/read-all'),
  /** Soft state change (ACTIVE | HIDDEN | DELETED). */
  updateStatus: (id: number, status: NotificationStatus) =>
    api.put<Notification>(`/notifications/${id}/status?status=${status}`),
  /** Hard delete a notification. */
  remove: (id: number) => api.delete<void>(`/notifications/${id}`),
}
