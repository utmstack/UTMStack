import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2, X } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import type { Dashboard } from '@/features/dashboard/types'

export function DashboardFormDialog({
  open,
  mode,
  initial,
  busy,
  onClose,
  onSubmit,
}: {
  open: boolean
  mode: 'create' | 'rename'
  initial?: Dashboard | null
  busy: boolean
  onClose: () => void
  onSubmit: (data: { name: string; description?: string }) => void
}) {
  const { t } = useTranslation()
  const [name, setName] = useState(initial?.name ?? '')
  const [description, setDescription] = useState(initial?.description ?? '')

  useEffect(() => {
    if (open) {
      setName(initial?.name ?? '')
      setDescription(initial?.description ?? '')
    }
  }, [open, initial])

  if (!open) return null

  const valid = name.trim().length > 0

  const submit = () => {
    if (!valid || busy) return
    onSubmit({ name: name.trim(), description: description.trim() || undefined })
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="flex w-full max-w-md flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <h2 className="text-lg font-semibold">
            {mode === 'create' ? t('dashboards.form.createTitle') : t('dashboards.form.renameTitle')}
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>

        <div className="space-y-4 px-6 py-5">
          <div>
            <label className="mb-1.5 block text-xs font-medium text-foreground/80">
              {t('dashboards.form.name')}
            </label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={t('dashboards.form.namePlaceholder') ?? ''}
              autoFocus
            />
          </div>
          <div>
            <label className="mb-1.5 block text-xs font-medium text-foreground/80">
              {t('dashboards.form.description')}
            </label>
            <Input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder={t('dashboards.form.descriptionPlaceholder') ?? ''}
            />
          </div>
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-border bg-muted/40 px-6 py-3">
          <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
            {t('dashboards.form.cancel')}
          </Button>
          <Button size="sm" onClick={submit} disabled={!valid || busy}>
            {busy && <Loader2 size={14} className="mr-1 animate-spin" />}
            {mode === 'create' ? t('dashboards.form.create') : t('dashboards.form.save')}
          </Button>
        </footer>
      </div>
    </div>
  )
}
