import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { AlertTriangle, X } from 'lucide-react'
import { datasourcesHttpService as svc } from '../services/datasources-http.service'
import type { Usage } from '../types/datasource.types'

/**
 * App-wide warning shown when the deployment is over its licensed datasource cap.
 * Logs from datasources beyond the limit are dropped at the event processor, so
 * this needs to be loud and visible — but an analyst staring at it 8h/day shouldn't
 * be forced to keep it. It's dismissible, and the dismissal is keyed to the current
 * count: if the overage grows worse, the banner comes back on its own.
 */
const DISMISS_KEY = 'utmstack-dslimit-dismissed-count'

export function DatasourceLimitBanner() {
  const { t } = useTranslation()
  const [usage, setUsage] = useState<Usage | null>(null)
  const [dismissedCount, setDismissedCount] = useState<number | null>(() => {
    if (typeof window === 'undefined') return null
    const raw = window.localStorage.getItem(DISMISS_KEY)
    const n = raw == null ? NaN : Number(raw)
    return Number.isFinite(n) ? n : null
  })

  useEffect(() => {
    let cancelled = false
    svc
      .usage()
      .then((u) => {
        if (!cancelled) setUsage(u)
      })
      .catch(() => {
        /* best-effort; no banner if usage can't be read */
      })
    return () => {
      cancelled = true
    }
  }, [])

  if (!usage?.overLimit) return null
  // Already acknowledged at this (or a higher) count → stay hidden until it worsens.
  if (dismissedCount != null && usage.count <= dismissedCount) return null

  const dismiss = () => {
    setDismissedCount(usage.count)
    try {
      window.localStorage.setItem(DISMISS_KEY, String(usage.count))
    } catch {
      /* ignore quota/availability errors */
    }
  }

  return (
    <div className="flex items-center gap-x-2 border-b border-amber-400 bg-amber-100 px-6 py-2 text-sm text-slate-800 dark:border-amber-500/40 dark:bg-amber-500/10 dark:text-amber-300">
      <AlertTriangle size={16} className="shrink-0 text-amber-600 dark:text-amber-400" />
      <span className="min-w-0 flex-1 text-center">
        <span className="font-bold">{t('datasources.limitBanner.title')}</span>{' '}
        {t('datasources.limitBanner.message', { count: usage.count, limit: usage.limit })}{' '}
        <Link
          to="/settings/license"
          className="font-bold text-slate-900 underline underline-offset-2 hover:opacity-80 dark:text-amber-200 dark:hover:text-amber-100"
        >
          {t('datasources.limitBanner.cta')}
        </Link>
      </span>
      <button
        onClick={dismiss}
        title={t('datasources.limitBanner.dismiss')}
        aria-label={t('datasources.limitBanner.dismiss')}
        className="shrink-0 rounded p-1 text-amber-700/70 transition-colors hover:bg-amber-500/20 hover:text-amber-900 dark:text-amber-300/70 dark:hover:text-amber-200"
      >
        <X size={15} />
      </button>
    </div>
  )
}
