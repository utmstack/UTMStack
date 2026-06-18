import { useEffect, useState } from 'react'
import { Calendar as CalendarIcon, X } from 'lucide-react'
import { DayPicker, type DateRange } from 'react-day-picker'
import 'react-day-picker/style.css'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'

export interface DateRangeValue {
  from: Date | null
  to: Date | null
}

export const EMPTY_RANGE: DateRangeValue = { from: null, to: null }

export function DateRangePickerDialog({
  open,
  value,
  title,
  onClose,
  onConfirm,
}: {
  open: boolean
  value: DateRangeValue
  title?: string
  onClose: () => void
  onConfirm: (next: DateRangeValue) => void
}) {
  const { t } = useTranslation()
  const [local, setLocal] = useState<DateRangeValue>(value)

  useEffect(() => {
    if (open) setLocal(value)
  }, [open, value])

  if (!open) return null

  const selected: DateRange | undefined =
    local.from || local.to
      ? { from: local.from ?? undefined, to: local.to ?? undefined }
      : undefined

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="flex w-full max-w-2xl flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-center justify-between gap-4 border-b border-border px-6 py-4">
          <h2 className="flex items-center gap-2 text-base font-semibold">
            <CalendarIcon size={17} strokeWidth={1.75} />
            {title ?? t('common.dateRange.pick')}
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            aria-label={t('common.actions.cancel') ?? 'Cancel'}
          >
            <X size={16} />
          </button>
        </header>

        <div className="flex items-center justify-center px-2 py-4">
          <DayPicker
            mode="range"
            numberOfMonths={2}
            selected={selected}
            onSelect={(range) =>
              setLocal({ from: range?.from ?? null, to: range?.to ?? null })
            }
            captionLayout="dropdown"
            showOutsideDays
          />
        </div>

        <footer className="flex items-center justify-between gap-2 border-t border-border bg-muted/40 px-6 py-3">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => setLocal(EMPTY_RANGE)}
            disabled={!local.from && !local.to}
          >
            {t('common.dateRange.clear')}
          </Button>
          <div className="flex items-center gap-2">
            <Button type="button" variant="outline" size="sm" onClick={onClose}>
              {t('common.actions.cancel')}
            </Button>
            <Button type="button" size="sm" onClick={() => onConfirm(local)}>
              {t('common.dateRange.apply')}
            </Button>
          </div>
        </footer>
      </div>
    </div>
  )
}

export function formatRange(value: DateRangeValue, fallback?: string): string {
  if (!value.from && !value.to) return fallback ?? ''
  const fmt = (d: Date | null) =>
    d ? d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' }) : '…'
  if (value.from && value.to && sameDay(value.from, value.to)) return fmt(value.from)
  return `${fmt(value.from)} – ${fmt(value.to)}`
}

function sameDay(a: Date, b: Date): boolean {
  return (
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  )
}
