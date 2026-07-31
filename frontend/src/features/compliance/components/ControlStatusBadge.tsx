import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronDown } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { complianceService, ComplianceHttpError } from '../services/compliance-http.service'
import { CONTROL_STATUSES, type ControlStatus, type ReportControlRow } from '../types/compliance.types'
import { STATUS_TONE } from './ReportView'
import { StatusChangeReasonModal } from './StatusChangeReasonModal'

/**
 * Status pill that doubles as a manual-override selector. Native <select> styled
 * as a badge — click opens the OS picker. Used inline on report rows and inside
 * the control detail drawer. Stops propagation so it works inside clickable rows.
 * Setting a status opens a modal that requires a reason before submitting.
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
  const [pending, setPending] = useState<ControlStatus | null>(null)

  const errorMessage = (e: unknown) =>
    e instanceof ComplianceHttpError ? e.message : t('compliance.overrideError', { defaultValue: 'Failed to update status' })

  const submit = async (next: ControlStatus, reason: string) => {
    setBusy(true)
    try {
      await complianceService.setControlStatusOverride(frameworkKey, row.controlId, next, reason)
      setPending(null)
      onChanged?.()
    } catch (e) {
      toast.error(errorMessage(e))
    } finally {
      setBusy(false)
    }
  }

  const clear = async () => {
    if (busy) return
    setBusy(true)
    try {
      // ponytail: no reason on clear — backend DELETE endpoint doesn't accept one
      await complianceService.clearControlStatusOverride(frameworkKey, row.controlId)
      onChanged?.()
    } catch (e) {
      toast.error(errorMessage(e))
    } finally {
      setBusy(false)
    }
  }

  const change = (next: ControlStatus | '') => {
    if (busy) return
    if (next === '') void clear()
    else setPending(next)
  }

  return (
    <>
      <span className={cn('relative inline-flex items-center', className)}>
        <select
          value={row.overridden ? row.status : ''}
          disabled={busy}
          onClick={(e) => e.stopPropagation()}
          onChange={(e) => change(e.target.value as ControlStatus | '')}
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
      {pending && (
        <StatusChangeReasonModal
          busy={busy}
          onCancel={() => { if (!busy) setPending(null) }}
          onConfirm={(reason) => void submit(pending, reason)}
        />
      )}
    </>
  )
}
