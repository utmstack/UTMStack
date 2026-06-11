import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { useVisualizationMutations } from '@/features/dashboard/hooks/useVisualizations'

const DEFAULT_CONFIG = `{
  "title": { "text": "New visualization" },
  "tooltip": {},
  "xAxis": { "type": "category", "data": [] },
  "yAxis": { "type": "value" },
  "series": [{ "type": "bar", "data": [] }]
}`

export function NewVisualizationPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [sqlQuery, setSqlQuery] = useState('')
  const [config, setConfig] = useState(DEFAULT_CONFIG)
  const [configError, setConfigError] = useState<string | null>(null)

  const { createVisualization } = useVisualizationMutations()
  const busy = createVisualization.isPending

  const validateConfig = (raw: string): boolean => {
    try {
      JSON.parse(raw)
      setConfigError(null)
      return true
    } catch (err) {
      setConfigError((err as Error).message)
      return false
    }
  }

  const valid = name.trim().length > 0 && sqlQuery.trim().length > 0 && configError === null

  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!validateConfig(config)) return
    if (!valid || busy) return
    createVisualization.mutate(
      {
        name: name.trim(),
        description: description.trim() || undefined,
        sqlQuery: sqlQuery.trim(),
        config: config.trim(),
      },
      {
        onSuccess: () => {
          toast.success(t('dashboards.toast.visualizationCreated'))
          navigate('/dashboards/visualizations')
        },
        onError: (err) =>
          toast.error(err.message ?? t('dashboards.toast.visualizationCreateFailed')),
      }
    )
  }

  return (
    <div className="mx-auto flex h-full w-full max-w-4xl flex-col gap-4 px-6 py-6">
      <header>
        <h1 className="text-xl font-semibold">{t('dashboards.newVisualization.title')}</h1>
        <p className="text-sm text-muted-foreground">
          {t('dashboards.newVisualization.subtitle')}
        </p>
      </header>

      <form
        onSubmit={submit}
        className="flex flex-col gap-4 rounded-lg border border-border bg-card p-6"
      >
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
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

        <div>
          <label className="mb-1.5 block text-xs font-medium text-foreground/80">
            {t('dashboards.newVisualization.sqlQuery')}
          </label>
          <textarea
            value={sqlQuery}
            onChange={(e) => setSqlQuery(e.target.value)}
            placeholder={t('dashboards.newVisualization.sqlQueryPlaceholder') ?? ''}
            spellCheck={false}
            rows={5}
            className="w-full rounded-md border border-input bg-background px-3 py-2 font-mono text-xs leading-relaxed shadow-sm focus:border-ring focus:outline-none focus:ring-1 focus:ring-ring"
          />
        </div>

        <div>
          <div className="mb-1.5 flex items-center justify-between">
            <label className="block text-xs font-medium text-foreground/80">
              {t('dashboards.newVisualization.config')}
            </label>
            <span className="text-[10px] text-muted-foreground">
              {t('dashboards.newVisualization.configHint')}
            </span>
          </div>
          <textarea
            value={config}
            onChange={(e) => {
              setConfig(e.target.value)
              validateConfig(e.target.value)
            }}
            spellCheck={false}
            rows={14}
            className="w-full rounded-md border border-input bg-background px-3 py-2 font-mono text-xs leading-relaxed shadow-sm focus:border-ring focus:outline-none focus:ring-1 focus:ring-ring"
          />
          {configError && (
            <p className="mt-1 text-xs text-destructive">
              {t('dashboards.newVisualization.configError', { error: configError })}
            </p>
          )}
        </div>

        <div className="flex items-center justify-end gap-2 border-t border-border pt-4">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => navigate('/dashboards/visualizations')}
            disabled={busy}
          >
            {t('dashboards.form.cancel')}
          </Button>
          <Button type="submit" size="sm" disabled={!valid || busy}>
            {busy && <Loader2 size={14} className="mr-1 animate-spin" />}
            {t('dashboards.form.create')}
          </Button>
        </div>
      </form>
    </div>
  )
}

export default NewVisualizationPage
