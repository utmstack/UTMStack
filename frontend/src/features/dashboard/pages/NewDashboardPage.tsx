import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Loader2, Sparkles } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { useDashboards } from '@/features/dashboard/hooks/useDashboards'
import { useSocAi } from '@/features/soc-ai/SocAiProvider'
import { useSocAiConfigured } from '@/features/soc-ai/lib/useSocAiConfig'

export function NewDashboardPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [mode, setMode] = useState<'manual' | 'ai'>('manual')
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const aiConfigured = useSocAiConfigured()
  const { openPanel, submit: submitToAssistant } = useSocAi()

  const { createDashboard } = useDashboards({ size: 0 })

  const aiMode = mode === 'ai'
  const valid = aiMode ? name.trim().length > 0 && description.trim().length > 0 : name.trim().length > 0
  const busy = createDashboard.isPending

  // Manual mode creates the dashboard directly. AI mode instead hands the
  // name/description off as the assistant's opening task, in its own
  // isolated thread ('dashboard-create') — it builds the dashboard itself
  // (dashboards.create / visualizations.create are just more MCP tools it
  // already has), same flow as the "Add new dashboard" modal.
  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!valid || busy) return

    if (aiMode) {
      const task = t('dashboards.newDashboard.aiOpenerWithDetails', {
        name: name.trim(),
        description: description.trim(),
      })
      openPanel('dashboard-create')
      submitToAssistant(task, { scope: 'dashboard-create' })
      navigate('/dashboards/list')
      return
    }

    createDashboard.mutate(
      {
        name: name.trim(),
        description: description.trim() || undefined,
        config: '',
      },
      {
        onSuccess: () => {
          toast.success(t('dashboards.toast.created'))
          navigate('/dashboards/list')
        },
        onError: (err) => toast.error(err.message ?? t('dashboards.toast.createFailed')),
      }
    )
  }

  return (
    <div className="mx-auto flex h-full w-full max-w-3xl flex-col gap-4 px-6 py-6">
      <header>
        <h1 className="text-xl font-semibold">{t('dashboards.newDashboard.title')}</h1>
        <p className="text-sm text-muted-foreground">{t('dashboards.newDashboard.subtitle')}</p>
      </header>

      <div className="flex gap-2">
        <button
          type="button"
          onClick={() => setMode('manual')}
          className={cn(
            'rounded-md border px-3 py-1.5 text-sm font-medium transition-colors',
            mode === 'manual'
              ? 'border-primary/30 bg-primary/10 text-primary'
              : 'border-border text-muted-foreground hover:bg-muted hover:text-foreground'
          )}
        >
          {t('dashboards.newDashboard.modeManual')}
        </button>
        <button
          type="button"
          onClick={() => aiConfigured && setMode('ai')}
          disabled={!aiConfigured}
          title={aiConfigured ? undefined : t('dashboards.newDashboard.aiLocked')}
          className={cn(
            'flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-sm font-medium transition-colors',
            !aiConfigured
              ? 'cursor-not-allowed border-border text-muted-foreground/50'
              : mode === 'ai'
                ? 'border-primary/30 bg-primary/10 text-primary'
                : 'border-border text-muted-foreground hover:bg-muted hover:text-foreground'
          )}
        >
          <Sparkles size={14} />
          {t('dashboards.newDashboard.modeAi')}
        </button>
      </div>

      <form onSubmit={submit} className="flex flex-col gap-4 rounded-lg border border-border bg-card p-6">
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
            {aiMode ? t('dashboards.newDashboard.aiDescriptionLabel') : t('dashboards.form.description')}
          </label>
          <Input
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder={
              (aiMode
                ? t('dashboards.newDashboard.aiDescriptionPlaceholder')
                : t('dashboards.form.descriptionPlaceholder')) ?? ''
            }
          />
        </div>

        <div className="flex items-center justify-end gap-2 border-t border-border pt-4">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => navigate('/dashboards/list')}
            disabled={busy}
          >
            {t('dashboards.form.cancel')}
          </Button>
          <Button type="submit" size="sm" disabled={!valid || busy}>
            {busy && <Loader2 size={14} className="mr-1 animate-spin" />}
            {aiMode ? t('dashboards.newDashboard.aiCreate') : t('dashboards.form.create')}
          </Button>
        </div>
      </form>
    </div>
  )
}

export default NewDashboardPage
