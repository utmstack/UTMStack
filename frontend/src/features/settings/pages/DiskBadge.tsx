import { useTranslation } from 'react-i18next'
import type { StoreHealth } from '../types/storage.types'

export function DiskBadge({ health }: { health: StoreHealth }) {
  const { t } = useTranslation()
  const warn = health.status !== 'ok'
  return (
    <span
      className={
        'rounded-full px-2 py-0.5 text-[11px] ' +
        (warn
          ? 'bg-amber-500/10 text-amber-700 dark:text-amber-300'
          : 'bg-muted text-muted-foreground')
      }
      title={health.message}
    >
      {t('dataRetention.diskUsed', { pct: health.diskUsedPct.toFixed(1) })}
    </span>
  )
}
