import { AlertCircle, AlertTriangle, Info, type LucideIcon } from 'lucide-react'
import type { NotificationType } from './types/notification.types'

export const TYPE_META: Record<NotificationType, { icon: LucideIcon; tone: string }> = {
  INFO: { icon: Info, tone: 'text-sky-500' },
  WARNING: { icon: AlertTriangle, tone: 'text-amber-500' },
  ERROR: { icon: AlertCircle, tone: 'text-red-500' },
}

/** Compact relative time ("just now", "5m ago", "3h ago", "2d ago", then a date). */
export function timeAgo(iso: string): string {
  const then = new Date(iso).getTime()
  if (Number.isNaN(then)) return ''
  const s = Math.floor((Date.now() - then) / 1000)
  if (s < 60) return 'just now'
  const m = Math.floor(s / 60)
  if (m < 60) return `${m}m ago`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h ago`
  const d = Math.floor(h / 24)
  if (d < 7) return `${d}d ago`
  return new Date(iso).toLocaleDateString()
}
