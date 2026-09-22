import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2, Sparkles, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { Textarea } from '@/shared/components/ui/textarea'
import { useSocAi } from '@/features/soc-ai/SocAiProvider'
import { useSocAiConfigured } from '@/features/soc-ai/lib/useSocAiConfig'
import { useBackdropDismiss } from '@/shared/hooks/useBackdropDismiss'
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
  const [creationMode, setCreationMode] = useState<'manual' | 'ai'>('manual')
  const aiConfigured = useSocAiConfigured()
  const { openPanel, submit: submitToAssistant } = useSocAi()

  useEffect(() => {
    if (open) {
      setName(initial?.name ?? '')
      setDescription(initial?.description ?? '')
      setCreationMode('manual')
    }
  }, [open, initial])

  const backdrop = useBackdropDismiss(onClose)

  if (!open) return null

  const useAiMode = mode === 'create' && creationMode === 'ai'
  const valid = useAiMode ? name.trim().length > 0 && description.trim().length > 0 : name.trim().length > 0

  // Manual mode creates the dashboard directly, same as always. AI mode
  // instead hands the name/description off as the assistant's opening task —
  // it builds the dashboard itself (dashboards.create / visualizations.create
  // are just more MCP tools it already has) in its own isolated thread
  // ('dashboard-create'), not the general Ask panel's conversation.
  const submit = () => {
    if (!valid || busy) return
    if (useAiMode) {
      const task = t('dashboards.newDashboard.aiOpenerWithDetails', {
        name: name.trim(),
        description: description.trim(),
      })
      onClose()
      openPanel('dashboard-create')
      submitToAssistant(task, { scope: 'dashboard-create' })
      return
    }
    onSubmit({ name: name.trim(), description: description.trim() || undefined })
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
      {...backdrop}
    >
      <div className="flex w-full max-w-md flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl">
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

        {mode === 'create' && (
          <div className="flex gap-2 border-b border-border px-6 pt-4">
            <button
              type="button"
              onClick={() => setCreationMode('manual')}
              className={cn(
                'rounded-md border px-3 py-1.5 text-sm font-medium transition-colors',
                creationMode === 'manual'
                  ? 'border-primary/30 bg-primary/10 text-primary'
                  : 'border-border text-muted-foreground hover:bg-muted hover:text-foreground'
              )}
            >
              {t('dashboards.newDashboard.modeManual')}
            </button>
            <button
              type="button"
              onClick={() => aiConfigured && setCreationMode('ai')}
              disabled={!aiConfigured}
              title={aiConfigured ? undefined : t('dashboards.newDashboard.aiLocked')}
              className={cn(
                'flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-sm font-medium transition-colors',
                !aiConfigured
                  ? 'cursor-not-allowed border-border text-muted-foreground/50'
                  : creationMode === 'ai'
                    ? 'border-primary/30 bg-primary/10 text-primary'
                    : 'border-border text-muted-foreground hover:bg-muted hover:text-foreground'
              )}
            >
              <Sparkles size={14} />
              {t('dashboards.newDashboard.modeAi')}
            </button>
          </div>
        )}

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
              {useAiMode ? t('dashboards.newDashboard.aiDescriptionLabel') : t('dashboards.form.description')}
            </label>
            <Textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder={
                (useAiMode
                  ? t('dashboards.newDashboard.aiDescriptionPlaceholder')
                  : t('dashboards.form.descriptionPlaceholder')) ?? ''
              }
              maxRows={8}
            />
          </div>
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-border bg-muted/40 px-6 py-3">
          <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
            {t('dashboards.form.cancel')}
          </Button>
          <Button size="sm" onClick={submit} disabled={!valid || busy}>
            {busy && <Loader2 size={14} className="mr-1 animate-spin" />}
            {useAiMode
              ? t('dashboards.newDashboard.aiCreate')
              : mode === 'create'
                ? t('dashboards.form.create')
                : t('dashboards.form.save')}
          </Button>
        </footer>
      </div>
    </div>
  )
}
