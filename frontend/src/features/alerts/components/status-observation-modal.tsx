import { useState } from 'react'
import { X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'

export function StatusObservationModal({
  title,
  initialValue = '',
  onCancel,
  onConfirm,
}: {
  title: string
  initialValue?: string
  onCancel: () => void
  onConfirm: (observation: string) => void
}) {
  const { t } = useTranslation()
  const [observation, setObservation] = useState(initialValue)
  const canSubmit = observation.trim().length > 0

  return (
    <div
      className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
      onClick={onCancel}
    >
      <div
        className="flex w-full max-w-md flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-center justify-between gap-4 border-b border-border px-5 py-4">
          <h2 className="text-base font-semibold">{title}</h2>
          <button
            onClick={onCancel}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>

        <div className="px-5 py-4">
          <label className="mb-1 block text-xs font-medium text-foreground/80">
            {t('alerts.drawer.observationLabel')}
          </label>
          <textarea
            value={observation}
            onChange={(e) => setObservation(e.target.value)}
            rows={4}
            autoFocus
            placeholder={t('alerts.drawer.observationPlaceholder')}
            className="w-full rounded-md border border-input bg-background/40 p-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
          />
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-border bg-muted/20 px-5 py-3">
          <Button variant="outline" size="sm" onClick={onCancel}>
            {t('alerts.drawer.cancel')}
          </Button>
          <Button size="sm" disabled={!canSubmit} onClick={() => onConfirm(observation.trim())}>
            {t('alerts.drawer.continue')}
          </Button>
        </footer>
      </div>
    </div>
  )
}
