import { useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { Trash2, X, type LucideIcon } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'

export interface ConfirmDialogProps {
  open: boolean
  title: string
  body: ReactNode
  confirmLabel?: string
  cancelLabel?: string
  danger?: boolean
  icon?: LucideIcon
  busy?: boolean
  hideCancel?: boolean
  onClose: () => void
  onConfirm: () => void | Promise<void>
}

export function ConfirmDialog({
  open,
  title,
  body,
  confirmLabel,
  cancelLabel,
  danger,
  icon: Icon = Trash2,
  busy: externalBusy,
  hideCancel,
  onClose,
  onConfirm,
}: ConfirmDialogProps) {
  const { t } = useTranslation()
  const [internalBusy, setInternalBusy] = useState(false)
  const busy = externalBusy || internalBusy

  if (!open) return null

  const run = async () => {
    if (busy) return
    try {
      setInternalBusy(true)
      await onConfirm()
    } finally {
      setInternalBusy(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
      onClick={() => !busy && onClose()}
    >
      <div
        className="flex w-full max-w-lg flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-center justify-between gap-4 border-b border-border px-6 py-4">
          <h2 className="flex items-center gap-2 text-base font-semibold">
            <Icon size={17} strokeWidth={1.75} />
            {title}
          </h2>
          <button
            type="button"
            onClick={onClose}
            disabled={busy}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
            aria-label={cancelLabel ?? t('common.actions.cancel') ?? 'Cancel'}
          >
            <X size={16} />
          </button>
        </header>

        <div className="px-6 py-5 text-sm text-muted-foreground">{body}</div>

        <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
          {!hideCancel && (
            <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
              {cancelLabel ?? t('common.actions.cancel')}
            </Button>
          )}
          <Button
            size="sm"
            variant={danger ? 'destructive' : 'default'}
            disabled={busy}
            onClick={() => void run()}
          >
            {busy
              ? t('common.confirm.working') ?? 'Working…'
              : confirmLabel ?? t('common.confirm.delete') ?? 'Delete'}
          </Button>
        </footer>
      </div>
    </div>
  )
}
