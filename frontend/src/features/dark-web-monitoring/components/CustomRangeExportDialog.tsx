import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { CalendarDays, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'
import type { DarkWebRecord } from '../types'
import type { ExportAction } from './export-types'
import { parsePanelDate } from './TimeRangeFilter'

interface CustomRangeExportDialogProps {
  records: DarkWebRecord[]
  onClose: () => void
  onExport: (records: DarkWebRecord[], action: ExportAction, filename: string, maskSensitive: boolean) => boolean
}

export function CustomRangeExportDialog({ records, onClose, onExport }: CustomRangeExportDialogProps) {
  const { t } = useTranslation()
  const [action, setAction] = useState<ExportAction>('PDF')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const fromInputRef = useRef<HTMLInputElement>(null)
  const toInputRef = useRef<HTMLInputElement>(null)
  const invalidRange = Boolean(from && to && from > to)

  const openPicker = (input: HTMLInputElement | null) => {
    if (!input) return
    if (typeof input.showPicker === 'function') {
      try {
        input.showPicker()
      } catch {
        input.focus()
      }
    } else {
      input.focus()
    }
  }

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKeyDown)
    return () => document.removeEventListener('keydown', onKeyDown)
  }, [onClose])

  const handleExport = () => {
    if (!from || !to || invalidRange) return
    // Local wall-clock bounds for the picked dates (a date-only input has no
    // timezone; UTC would shift the day in non-UTC timezones).
    const fromTime = parsePanelDate(from, 'start')
    const toTime = parsePanelDate(to, 'end')
    // Filter on the breached `timestamp` (what the user picked), not the
    // ingested `createdAt` — and only over records already on the page.
    const matchingRecords = records.filter((record) => {
      const timestamp = Date.parse(record.timestamp)
      return timestamp >= fromTime && timestamp <= toTime
    })
    if (onExport(matchingRecords, action, 'dark-web-monitoring-custom-range', true)) onClose()
  }

  return createPortal(
    <div
      className="fixed inset-0 z-[80] flex items-center justify-center bg-black/65 p-4"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) onClose()
      }}
    >
      <section
        role="dialog"
        aria-modal="true"
        aria-labelledby="custom-range-title"
        className="w-full max-w-md overflow-hidden rounded-xl border border-border bg-card shadow-2xl"
      >
        <header className="flex items-center justify-between border-b border-border px-5 py-4">
          <h2 id="custom-range-title" className="font-mono text-sm font-semibold">{t('darkWebMonitoring.exportDialogs.custom.title')}</h2>
          <button type="button" aria-label={t('darkWebMonitoring.exportDialogs.close')} onClick={onClose} className="text-muted-foreground hover:text-foreground">
            <X size={16} />
          </button>
        </header>
        <div className="space-y-3 px-6 py-4 font-mono text-xs">
          <label className="block space-y-1.5">
            <span className="font-semibold">{t('darkWebMonitoring.exportDialogs.action')}</span>
            <select value={action} onChange={(event) => setAction(event.target.value as ExportAction)} className="h-9 w-full rounded-md border border-input bg-background px-2 text-xs">
              <option value="PDF">PDF</option>
              <option value="CSV">CSV</option>
            </select>
          </label>
          <label className="block space-y-1.5">
            <span className="font-semibold">{t('darkWebMonitoring.exportDialogs.custom.from')}</span>
            <span className="flex h-9 overflow-hidden rounded-md border border-input bg-background">
              <input aria-label={t('darkWebMonitoring.exportDialogs.custom.from')} ref={fromInputRef} type="date" value={from} max={to || undefined} onChange={(event) => setFrom(event.target.value)} className="dwm-date-input min-w-0 flex-1 bg-transparent px-2 text-xs" />
              <button type="button" aria-hidden="true" tabIndex={-1} onClick={() => openPicker(fromInputRef.current)} className="flex w-10 items-center justify-center bg-primary/10 text-primary transition-colors hover:bg-primary/20"><CalendarDays size={14} /></button>
            </span>
          </label>
          <label className="block space-y-1.5">
            <span className="font-semibold">{t('darkWebMonitoring.exportDialogs.custom.to')}</span>
            <span className="flex h-9 overflow-hidden rounded-md border border-input bg-background">
              <input aria-label={t('darkWebMonitoring.exportDialogs.custom.to')} ref={toInputRef} type="date" value={to} min={from || undefined} onChange={(event) => setTo(event.target.value)} className="dwm-date-input min-w-0 flex-1 bg-transparent px-2 text-xs" />
              <button type="button" aria-hidden="true" tabIndex={-1} onClick={() => openPicker(toInputRef.current)} className="flex w-10 items-center justify-center bg-primary/10 text-primary transition-colors hover:bg-primary/20"><CalendarDays size={14} /></button>
            </span>
          </label>
          {invalidRange && <p className="text-xs text-destructive">{t('darkWebMonitoring.exportDialogs.custom.invalidRange')}</p>}
        </div>
        <footer className="flex justify-end gap-3 px-6 pb-5">
          <Button variant="secondary" size="sm" onClick={onClose}>{t('darkWebMonitoring.exportDialogs.cancel')}</Button>
          <Button size="sm" onClick={handleExport} disabled={!from || !to || invalidRange}>
            <CalendarDays size={13} className="mr-1.5" />
            {t('darkWebMonitoring.exportDialogs.custom.select')}
          </Button>
        </footer>
      </section>
    </div>,
    document.body,
  )
}
