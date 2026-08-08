import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronDown } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { complianceService, ComplianceHttpError } from '../services/compliance-http.service'
import {
  COMPLIANCE_STATUSES,
  isOverridden,
  type ComplianceStatus,
  type ControlRow,
  type Report,
} from '../types/compliance.types'
import { STATUS_TONE } from './ReportView'
import { StatusChangeReasonModal } from './StatusChangeReasonModal'

/**
 * Status pill that doubles as the verdict selector — the only editable level in
 * a report. Choosing a status opens a modal for the justification, which the
 * backend requires: an override no one can explain is worth nothing to whoever
 * reads the report later.
 *
 * The edit returns the whole recomputed report, so the caller replaces its copy
 * rather than refetching.
 */
export function ControlStatusBadge({
  frameworkKey,
  row,
  onChanged,
  className,
}: {
  frameworkKey: string
  row: ControlRow
  onChanged?: (updated: Report) => void
  className?: string
}) {
  const { t } = useTranslation()
  const [busy, setBusy] = useState(false)
  const [pending, setPending] = useState<ComplianceStatus | null>(null)

  const errorMessage = (e: unknown) =>
    e instanceof ComplianceHttpError
      ? e.message
      : t('compliance.overrideError', { defaultValue: 'Failed to update status' })

  const submit = async (note: string) => {
    if (!pending) return
    setBusy(true)
    try {
      const updated = await complianceService.editControl(frameworkKey, row.controlId, { status: pending, note })
      setPending(null)
      onChanged?.(updated)
    } catch (e) {
      toast.error(errorMessage(e))
    } finally {
      setBusy(false)
    }
  }

  return (
    <>
      <div className={cn('relative inline-flex shrink-0', className)} onClick={(e) => e.stopPropagation()}>
        <span
          className={cn(
            'inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] font-semibold',
            STATUS_TONE[row.status],
            busy && 'opacity-60',
          )}
        >
          {t(`compliance.status.${row.status}`)}
          {/* An overridden row says so; the engine's own verdict stays on record.
              A plain note is not an override and does not get the mark. */}
          {isOverridden(row) && <span title={row.note}>*</span>}
          <ChevronDown size={11} className="opacity-70" />
        </span>
        <select
          aria-label={t('compliance.setStatus', { defaultValue: 'Set status' })}
          value={row.status}
          disabled={busy}
          onChange={(e) => setPending(e.target.value as ComplianceStatus)}
          className="absolute inset-0 cursor-pointer opacity-0"
        >
          {COMPLIANCE_STATUSES.map((s) => (
            <option key={s} value={s}>
              {t(`compliance.status.${s}`)}
            </option>
          ))}
        </select>
      </div>

      {pending && <StatusChangeReasonModal busy={busy} onCancel={() => setPending(null)} onConfirm={submit} />}
    </>
  )
}
