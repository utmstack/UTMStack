import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Sparkles, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { Textarea } from '@/shared/components/ui/textarea'
import { useSocAi } from '@/features/soc-ai/SocAiProvider'
import { useSocAiConfigured } from '@/features/soc-ai/lib/useSocAiConfig'
import { useBackdropDismiss } from '@/shared/hooks/useBackdropDismiss'

export function SoarCreateDialog({
  open,
  onClose,
  onManual,
}: {
  open: boolean
  onClose: () => void
  onManual: () => void
}) {
  const { t } = useTranslation()
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [mode, setMode] = useState<'manual' | 'ai'>('manual')
  const aiConfigured = useSocAiConfigured()
  const { openPanel, submit: submitToAssistant, setSoarCreateTarget } = useSocAi()

  useEffect(() => {
    if (open) {
      setName('')
      setDescription('')
      setMode('manual')
    }
  }, [open])

  const backdrop = useBackdropDismiss(onClose)

  if (!open) return null

  const valid = name.trim().length > 0 && (mode === 'manual' || description.trim().length > 0)

  const submit = () => {
    if (!valid) return
    if (mode === 'ai') {
      setSoarCreateTarget({ name: name.trim(), description: description.trim() })
      onClose()
      openPanel('soar-create')
      submitToAssistant(t('soar.create.aiOpener', { name: name.trim(), description: description.trim() }), { scope: 'soar-create' })
      return
    }
    onManual()
  }

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm" {...backdrop}>
      <div className="flex w-full max-w-md flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl">
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <h2 className="text-lg font-semibold">{t('soar.create.title')}</h2>
          <button type="button" onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
            <X size={16} />
          </button>
        </header>

        <div className="flex gap-2 border-b border-border px-6 pt-4">
          <button
            type="button"
            onClick={() => setMode('manual')}
            className={cn(
              'rounded-md border px-3 py-1.5 text-sm font-medium transition-colors',
              mode === 'manual' ? 'border-primary/30 bg-primary/10 text-primary' : 'border-border text-muted-foreground hover:bg-muted hover:text-foreground'
            )}
          >
            {t('soar.create.modeManual')}
          </button>
          <button
            type="button"
            onClick={() => aiConfigured && setMode('ai')}
            disabled={!aiConfigured}
            title={aiConfigured ? undefined : t('soar.create.aiLocked')}
            className={cn(
              'flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-sm font-medium transition-colors',
              !aiConfigured ? 'cursor-not-allowed border-border text-muted-foreground/50' : mode === 'ai' ? 'border-primary/30 bg-primary/10 text-primary' : 'border-border text-muted-foreground hover:bg-muted hover:text-foreground'
            )}
          >
            <Sparkles size={14} />
            {t('soar.create.modeAi')}
          </button>
        </div>

        <div className="space-y-4 px-6 py-5">
          <div>
            <label className="mb-1.5 block text-xs font-medium text-foreground/80">{t('soar.create.name')}</label>
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder={t('soar.create.namePlaceholder')} autoFocus />
          </div>
          {mode === 'ai' && (
            <div>
              <label className="mb-1.5 block text-xs font-medium text-foreground/80">{t('soar.create.aiDescriptionLabel')}</label>
              <Textarea value={description} onChange={(e) => setDescription(e.target.value)} placeholder={t('soar.create.aiDescriptionPlaceholder')} maxRows={8} />
            </div>
          )}
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-border bg-muted/40 px-6 py-3">
          <Button variant="outline" size="sm" onClick={onClose}>
            {t('common.cancel')}
          </Button>
          <Button size="sm" onClick={submit} disabled={!valid}>
            {mode === 'ai' ? t('soar.create.aiCreate') : t('soar.create.manualContinue')}
          </Button>
        </footer>
      </div>
    </div>
  )
}
