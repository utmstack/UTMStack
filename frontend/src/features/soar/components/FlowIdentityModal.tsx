import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { X } from 'lucide-react'
import { Input } from '@/shared/components/ui/input'

interface Props {
  name: string
  description: string
  maxDepth: number
  readOnly?: boolean
  onChange: (patch: { name?: string; description?: string; maxDepth?: number }) => void
  onClose: () => void
}

// Floating modal for the flow's identity fields — used to be a top-level
// SectionCard; now opened from a pencil icon next to the flow title in the
// header. Live-commits on every keystroke via onChange.
// ponytail: no shared Dialog primitive here, just overlay + card + Esc handler.
export function FlowIdentityModal({ name, description, maxDepth, readOnly, onChange, onClose }: Props) {
  const { t } = useTranslation()

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [onClose])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/60 backdrop-blur-sm">
      <div className="relative w-[min(560px,92vw)] rounded-xl border border-border bg-card p-5 shadow-xl">
        <div className="mb-3 flex items-center justify-between">
          <h3 className="text-sm font-semibold">{t('soar.editor.identityModalTitle')}</h3>
          <button
            onClick={onClose}
            className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={14} />
          </button>
        </div>
        <div className="space-y-3">
          <div className="grid grid-cols-[1fr_120px] gap-3">
            <div className="space-y-1">
              <label className="block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                {t('soar.editor.nameLabel')}
                {!readOnly && <span className="ml-0.5 text-red-500">*</span>}
              </label>
              <Input
                value={name}
                readOnly={readOnly}
                onChange={(e) => onChange({ name: e.target.value })}
                placeholder={t('soar.editor.namePlaceholder')}
                className="text-base font-semibold"
                autoFocus
              />
            </div>
            <div className="space-y-1">
              <label className="block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                {t('soar.editor.maxDepth', 'Max depth')}
              </label>
              <Input
                type="number"
                min={1}
                max={1000}
                value={maxDepth}
                readOnly={readOnly}
                onChange={(e) => onChange({ maxDepth: Number(e.target.value) || 50 })}
                className="text-sm"
              />
            </div>
          </div>
          <div className="space-y-1">
            <label className="block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
              {t('soar.editor.descriptionLabel')}
            </label>
            <textarea
              value={description}
              readOnly={readOnly}
              onChange={(e) => onChange({ description: e.target.value })}
              rows={3}
              placeholder={t('soar.editor.descriptionHint')}
              className="w-full resize-none rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
          </div>
        </div>
      </div>
    </div>
  )
}
