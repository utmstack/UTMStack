import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronDown } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { complianceService, ComplianceHttpError } from '../services/compliance-http.service'
import { CONTROL_STATUSES, type ControlStatus, type ReportControlRow } from '../types/compliance.types'
import { STATUS_TONE } from './ReportView'

/**
 * Status pill that doubles as a manual-override selector. Native <select> styled
 * as a badge — click opens the OS picker. Used inline on report rows and inside
 * the control detail drawer. Stops propagation so it works inside clickable rows.
 */
export function ControlStatusBadge({
  frameworkKey,
  row,
  onChanged,
  className,
}: {
  frameworkKey: string
  row: ReportControlRow
  onChanged?: () => void
  className?: string
}) {
  const { t } = useTranslation()
  const [busy, setBusy] = useState(false)

  const change = async (next: ControlStatus | '') => {
    if (busy) return
    setBusy(true)
    try {
      if (next === '') {
        await complianceService.clearControlStatusOverride(frameworkKey, row.controlId)
      } else {
        await complianceService.setControlStatusOverride(frameworkKey, row.controlId, next)
      }
      onChanged?.()
    } catch (e) {
      toast.error(e instanceof ComplianceHttpError ? e.message : t('compliance.overrideError', { defaultValue: 'Failed to update status' }))
    } finally {
      setBusy(false)
    }
  }

  return (
    <span className={cn('relative inline-flex items-center', className)}>
      <select
        value={row.overridden ? row.status : ''}
        disabled={busy}
        onClick={(e) => e.stopPropagation()}
        onChange={(e) => void change(e.target.value as ControlStatus | '')}
        title={t('compliance.status.label', { defaultValue: 'Status' })}
        className={cn(
          'cursor-pointer appearance-none rounded py-0.5 pl-1.5 pr-5 text-[10px] font-semibold outline-none transition-opacity focus-visible:ring-1 focus-visible:ring-ring disabled:opacity-60',
          STATUS_TONE[row.status],
        )}
      >
        <option value="">
          {row.overridden ? t('compliance.statusAuto', { defaultValue: 'Auto (evaluated)' }) : t(`compliance.status.${row.status}`)}
        </option>
        {CONTROL_STATUSES.map((s) => (
          <option key={s} value={s}>{t(`compliance.status.${s}`)}</option>
        ))}
      </select>
      <ChevronDown size={10} className="pointer-events-none absolute right-1 opacity-70" />
    </span>
  )
}
