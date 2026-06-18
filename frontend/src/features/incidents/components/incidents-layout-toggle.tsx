import { LayoutGrid, Rows3 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'

export type IncidentsLayout = 'table' | 'board'

export function IncidentsLayoutToggle({
  value,
  onChange,
}: {
  value: IncidentsLayout
  onChange: (v: IncidentsLayout) => void
}) {
  const { t } = useTranslation()
  return (
    <div className="inline-flex rounded-md border border-border p-0.5">
      <button
        onClick={() => onChange('table')}
        title={t('incidents.layout.table')}
        className={cn(
          'flex h-8 w-8 items-center justify-center rounded transition-colors',
          value === 'table' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'
        )}
      >
        <Rows3 size={15} />
      </button>
      <button
        onClick={() => onChange('board')}
        title={t('incidents.layout.board')}
        className={cn(
          'flex h-8 w-8 items-center justify-center rounded transition-colors',
          value === 'board' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'
        )}
      >
        <LayoutGrid size={15} />
      </button>
    </div>
  )
}
