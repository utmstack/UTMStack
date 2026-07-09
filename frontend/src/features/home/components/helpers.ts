import { useTranslation } from 'react-i18next'

export type TFn = ReturnType<typeof useTranslation>['t']

export function fmtCount(n: number): string {
  if (n >= 1_000_000_000) return `${(n / 1_000_000_000).toFixed(1)}B`
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`
  return n.toLocaleString()
}

export function spanLabel(t: TFn, from?: string, to?: string): string {
  if (!from || !to) return t('home.span.recent')
  const ms = new Date(to).getTime() - new Date(from).getTime()
  if (!Number.isFinite(ms) || ms <= 0) return t('home.span.recent')
  const days = Math.round(ms / 86_400_000)
  if (days >= 1) return t('home.span.days', { count: days })
  const hours = Math.max(1, Math.round(ms / 3_600_000))
  return t('home.span.hours', { n: hours })
}

export function shortTime(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleString(undefined, { month: 'short', day: 'numeric' })
}
