import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ChartTypeModal } from '@/features/dashboard/components/editor/ChartTypeModal'
import { VisualizationEditor } from '@/features/dashboard/components/VisualizationEditor'
import type { ChartTypeId } from '@/features/dashboard/types'

export function NewVisualizationPage() {
  const navigate = useNavigate()
  const [chartType, setChartType] = useState<ChartTypeId | null>(null)

  if (!chartType) {
    return (
      <ChartTypeModal
        open
        initial="bar"
        onConfirm={(picked) => setChartType(picked)}
        onClose={() => navigate('/dashboards/visualizations')}
      />
    )
  }

  return <VisualizationEditor initial={null} initialChartType={chartType} />
}

export default NewVisualizationPage
