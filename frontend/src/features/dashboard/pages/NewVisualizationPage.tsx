import { useEffect, useState } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { ChartTypeModal } from '@/features/dashboard/components/editor/ChartTypeModal'
import { VisualizationEditor } from '@/features/dashboard/components/VisualizationEditor'
import type { ChartTypeId } from '@/features/dashboard/types'

export function NewVisualizationPage() {
  const navigate = useNavigate()
  const { dashboardId: dashboardIdParam } = useParams<{ dashboardId: string }>()
  const dashboardId = Number(dashboardIdParam)
  const validDashboardId = Number.isFinite(dashboardId) && dashboardId > 0
  const location = useLocation()
  const initialLayout = (location.state as { layout?: string } | null)?.layout
  const [chartType, setChartType] = useState<ChartTypeId | null>(null)

  useEffect(() => {
    if (!validDashboardId) navigate('/dashboards/list', { replace: true })
  }, [validDashboardId, navigate])

  if (!validDashboardId) return null

  if (!chartType) {
    return (
      <ChartTypeModal
        open
        initial="bar"
        onConfirm={(picked) => setChartType(picked)}
        onClose={() => navigate('/dashboards/list', { state: { selectDashboardId: dashboardId } })}
      />
    )
  }

  return (
    <VisualizationEditor
      initial={null}
      initialChartType={chartType}
      dashboardId={dashboardId}
      initialLayout={initialLayout}
    />
  )
}

export default NewVisualizationPage
