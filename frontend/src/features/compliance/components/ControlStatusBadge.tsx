import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Check, ChevronDown } from 'lucide-react'
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
  const [menuOpen, setMenuOpen] = useState(false)
  const rootRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!menuOpen) return
    const onDown = (e: MouseEvent) => {
      if (!rootRef.current?.contains(e.target as Node)) setMenuOpen(false)
    }
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setMenuOpen(false)
    }
    document.addEventListener('mousedown', onDown)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('mousedown', onDown)
      document.removeEventListener('keydown', onKey)
    }
  }, [menuOpen])

  const errorMessage = (e: unknown) =>
    e instanceof ComplianceHttpError
      ? e.message
      : t('compliance.overrideError', { defaultValue: 'Failed to update status' })

  const submit = async (note: string, evidences: string[]) => {
    if (!pending) return
    setBusy(true)
    try {
      const input: { status: ComplianceStatus; note: string; evidences?: string[] } = { status: pending, note }
      if (evidences.length > 0) input.evidences = evidences
      const updated = await complianceService.editControl(frameworkKey, row.controlId, input)
      setPending(null)
      onChanged?.(updated)
    } catch (e) {
      toast.error(errorMessage(e))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div ref={rootRef} className={cn('relative inline-flex shrink-0', className)} onClick={(e) => e.stopPropagation()}>
      <button
        type="button"
        aria-haspopup="listbox"
        aria-expanded={menuOpen}
        aria-label={t('compliance.setStatus', { defaultValue: 'Set status' })}
        disabled={busy}
        onClick={() => setMenuOpen((v) => !v)}
        className={cn(
          'inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] font-semibold',
          'cursor-pointer hover:opacity-80',
          STATUS_TONE[row.status],
          busy && 'opacity-60',
        )}
      >
        {t(`compliance.status.${row.status}`)}
        {/* An overridden row says so; the engine's own verdict stays on record.
            A plain note is not an override and does not get the mark. */}
        {isOverridden(row) && <span title={row.note}>*</span>}
        <ChevronDown size={11} className={cn('opacity-70 transition-transform', menuOpen && 'rotate-180')} />
      </button>

      {menuOpen && (
        <div
          role="listbox"
          aria-label={t('compliance.setStatus', { defaultValue: 'Set status' })}
          className="absolute left-0 top-full z-50 mt-1 max-h-64 w-44 overflow-y-auto rounded-md border border-border bg-popover p-1 shadow-md"
        >
          {COMPLIANCE_STATUSES.map((s) => (
            <button
              key={s}
              type="button"
              role="option"
              aria-selected={s === row.status}
              onClick={() => {
                setMenuOpen(false)
                setPending(s)
              }}
              className={cn(
                'flex w-full items-center justify-between rounded-sm px-2 py-1.5 text-left text-[11px] transition-colors',
                'hover:bg-accent hover:text-accent-foreground',
                s === row.status && 'font-semibold',
                STATUS_TONE[s].split(' ').slice(1).join(' '),
              )}
            >
              {t(`compliance.status.${s}`)}
              {s === row.status && <Check size={12} />}
            </button>
          ))}
        </div>
      )}

      {pending && <StatusChangeReasonModal busy={busy} onCancel={() => setPending(null)} onConfirm={submit} />}
    </div>
  )
}
