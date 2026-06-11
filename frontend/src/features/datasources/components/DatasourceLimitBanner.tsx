import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { AlertTriangle } from 'lucide-react'
import { datasourcesHttpService as svc } from '../services/datasources-http.service'
import type { Usage } from '../types/datasource.types'

/**
 * App-wide warning shown when the deployment is over its licensed datasource cap.
 * Logs from datasources beyond the limit are dropped at the event processor, so
 * this needs to be loud and visible everywhere, not just on the datasources page.
 */
export function DatasourceLimitBanner() {
  const { t } = useTranslation()
  const [usage, setUsage] = useState<Usage | null>(null)

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

  return (
    <div className="flex flex-wrap items-center justify-center gap-x-2 gap-y-1 border-b border-amber-500/40 bg-amber-500/10 px-6 py-2 text-center text-sm text-amber-700 dark:text-amber-300">
      <AlertTriangle size={16} className="shrink-0" />
      <span>
        <span className="font-semibold">{t('datasources.limitBanner.title')}</span>{' '}
        {t('datasources.limitBanner.message', { count: usage.count, limit: usage.limit })}
      </span>
      <Link to="/settings/license" className="font-medium underline-offset-2 hover:underline">
        {t('datasources.limitBanner.cta')}
      </Link>
    </div>
  )
}
