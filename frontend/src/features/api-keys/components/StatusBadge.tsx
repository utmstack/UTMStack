import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'

const EXPIRING_WINDOW_MS = 7 * 86_400_000

function keyStatus(expiresAt: string | null): 'active' | 'expiring' | 'expired' {
  if (!expiresAt) return 'active'
  const exp = new Date(expiresAt).getTime()
  const now = Date.now()
  if (exp <= now) return 'expired'
  if (exp - now <= EXPIRING_WINDOW_MS) return 'expiring'
  return 'active'
}

export function StatusBadge({ expiresAt }: { expiresAt: string | null }) {
  const { t } = useTranslation()
  const status = keyStatus(expiresAt)
  const meta = {
    active: { label: t('apiKeys.status.active'), cls: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300', dot: 'bg-emerald-500' },
    expiring: { label: t('apiKeys.status.expiring'), cls: 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300', dot: 'bg-amber-500' },
    expired: { label: t('apiKeys.status.expired'), cls: 'bg-red-500/15 text-red-500 ring-red-500/30 dark:text-red-300', dot: 'bg-red-500' },
  }[status]
  return (
    <span className={cn('inline-flex items-center gap-1.5 rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset', meta.cls)}>
      <span className={cn('h-1.5 w-1.5 rounded-full', meta.dot)} />
      {meta.label}
    </span>
  )
}
