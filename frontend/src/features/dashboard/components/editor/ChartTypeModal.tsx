import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { X } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { ChartTypePicker } from '@/features/dashboard/components/editor/ChartTypePicker'
import type { ChartTypeId } from '@/features/dashboard/types'

export function ChartTypeModal({
  open,
  initial,
  title,
  confirmLabel,
  onConfirm,
  onClose,
}: {
  open: boolean
  initial: ChartTypeId
  title?: string
  confirmLabel?: string
  onConfirm: (chartType: ChartTypeId) => void
  onClose: () => void
}) {
  const { t } = useTranslation()
  const [selected, setSelected] = useState<ChartTypeId>(initial)

  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="flex w-full max-w-4xl flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <div>
            <h2 className="text-lg font-semibold">
              {title ?? t('dashboards.editor.chartTypeModal.title')}
            </h2>
            <p className="text-xs text-muted-foreground">
              {t('dashboards.editor.chartTypeModal.subtitle')}
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            aria-label={t('dashboards.form.cancel') ?? 'Cancel'}
          >
            <X size={16} />
          </button>
        </header>

        <div className="max-h-[60vh] overflow-y-auto px-6 py-5">
          <ChartTypePicker value={selected} onChange={setSelected} />
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-border bg-muted/40 px-6 py-3">
          <Button type="button" variant="outline" size="sm" onClick={onClose}>
            {t('dashboards.form.cancel')}
          </Button>
          <Button type="button" size="sm" onClick={() => onConfirm(selected)}>
            {confirmLabel ?? t('dashboards.editor.chartTypeModal.confirm')}
          </Button>
        </footer>
      </div>
    </div>
  )
}
