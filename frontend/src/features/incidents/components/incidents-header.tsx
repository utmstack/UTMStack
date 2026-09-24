import { LayoutGrid, Rows3 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { ViewSwitch } from '@/shared/components/ui/view-switch'

export type IncidentsLayout = 'table' | 'board'

export function IncidentsHeader({
  total,
  openCount,
  layout,
  onLayout,
}: {
  total: number
  openCount: number
  layout: IncidentsLayout
  onLayout: (l: IncidentsLayout) => void
}) {
  const { t } = useTranslation()
  return (
    <header className="flex flex-wrap items-center justify-between gap-3">
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <span className="font-medium text-foreground">{t('incidents.count', { count: total })}</span>
        <span className="text-muted-foreground/50">·</span>
        <span>
          <span className="font-medium text-foreground">{openCount.toLocaleString()}</span> {t('incidents.openNow')}
        </span>
      </div>
      <ViewSwitch
        value={layout}
        onChange={onLayout}
        options={[
          { id: 'table', icon: Rows3, label: t('incidents.layout.table') },
          { id: 'board', icon: LayoutGrid, label: t('incidents.layout.board') },
        ]}
      />
    </header>
  )
}
