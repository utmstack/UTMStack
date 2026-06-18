import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { useDashboards } from '@/features/dashboard/hooks/useDashboards'

export function NewDashboardPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')

  const { createDashboard } = useDashboards({ size: 0 })

  const valid = name.trim().length > 0
  const busy = createDashboard.isPending

  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!valid || busy) return
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

      <form
        onSubmit={submit}
        className="flex flex-col gap-4 rounded-lg border border-border bg-card p-6"
      >
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
            {t('dashboards.form.create')}
          </Button>
        </div>
      </form>
    </div>
  )
}

export default NewDashboardPage
