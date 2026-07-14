import { useParams, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Loader2 } from 'lucide-react'
import { VisualizationEditor } from '@/features/dashboard/components/VisualizationEditor'
import { useVisualization } from '@/features/dashboard/hooks/useVisualizations'

export function EditVisualizationPage() {
  const { t } = useTranslation()
  const params = useParams<{ id: string; dashboardId: string }>()
  const id = Number(params.id)
  const dashboardId = Number(params.dashboardId)
  const query = useVisualization(Number.isFinite(id) && id > 0 ? id : null)

  if (query.isLoading) {
    return (
      <div className="flex h-full w-full items-center justify-center gap-2 text-xs text-muted-foreground">
        <Loader2 size={14} className="animate-spin" />
        {t('dashboards.editor.loading')}
      </div>
    )
  }

  if (query.isError || !query.data) {
    return (
      <div className="mx-auto flex h-full w-full max-w-3xl flex-col items-center justify-center gap-3 px-6 py-10 text-center">
        <h1 className="text-lg font-semibold">{t('dashboards.editor.notFoundTitle')}</h1>
        <p className="text-sm text-muted-foreground">
          {t('dashboards.editor.notFoundSubtitle')}
        </p>
        <Link
          to="/dashboards/list"
          state={Number.isFinite(dashboardId) ? { selectDashboardId: dashboardId } : undefined}
          className="text-sm text-primary underline-offset-4 hover:underline"
        >
          {t('dashboards.editor.backToDashboard')}
        </Link>
      </div>
    )
  }

  return <VisualizationEditor initial={query.data} dashboardId={query.data.dashboardId} />
}

export default EditVisualizationPage
