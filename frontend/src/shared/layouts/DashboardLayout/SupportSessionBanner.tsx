import { useTranslation } from 'react-i18next'
import { Eye, LogOut, ShieldAlert } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { setSupportTenant, useSupportTenant } from '@/shared/lib/current-tenant'

/**
 * Sits above everything while an operator is working inside a customer's tenant.
 *
 * Deliberately loud and never dismissible: every screen below it is someone
 * else's data, and mistaking it for your own is the one failure this whole
 * feature has to prevent. Leaving reloads for the same reason entering does —
 * so nothing of theirs survives into the operator's own session.
 */
export function SupportSessionBanner() {
  const { t } = useTranslation()
  const tenant = useSupportTenant()
  if (!tenant) return null

  const readOnly = tenant.access === 'READ'

  const leave = () => {
    setSupportTenant(null)
    window.location.assign('/tenants')
  }

  // Tinted rather than filled: this sits above every screen for as long as the
  // session lasts, so it has to be noticeable without being painful to work
  // under. The left edge carries the colour, the text stays readable.
  return (
    <div
      className={cn(
        'flex flex-wrap items-center gap-x-3 gap-y-1 border-b border-l-4 px-4 py-1.5 text-xs',
        readOnly
          ? 'border-b-sky-500/25 border-l-sky-500 bg-sky-500/10 text-sky-900 dark:text-sky-200'
          : 'border-b-amber-500/25 border-l-amber-500 bg-amber-500/10 text-amber-900 dark:text-amber-200'
      )}
    >
      {readOnly ? (
        <Eye size={14} className="shrink-0 text-sky-600 dark:text-sky-400" />
      ) : (
        <ShieldAlert size={14} className="shrink-0 text-amber-600 dark:text-amber-400" />
      )}
      <span className="font-medium">{t('supportSession.inside', { name: tenant.name })}</span>
      <span className="opacity-70">
        {readOnly ? t('supportSession.readOnly') : t('supportSession.full')}
      </span>
      <button
        onClick={leave}
        className="ml-auto inline-flex items-center gap-1 rounded border border-current/25 px-2 py-0.5 font-medium transition-colors hover:bg-current/10"
      >
        <LogOut size={12} />
        {t('supportSession.leave')}
      </button>
    </div>
  )
}
