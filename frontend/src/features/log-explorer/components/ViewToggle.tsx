import { BarChart3, Table as TableIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'

export function ViewToggle({ mode, onChange }: { mode: 'table' | 'chart'; onChange: (m: 'table' | 'chart') => void }) {
  const { t } = useTranslation()
  const opts = [
    { id: 'table' as const, icon: TableIcon, label: t('logExplorer.view.table') },
    { id: 'chart' as const, icon: BarChart3, label: t('logExplorer.view.chart') },
  ]
  return (
    <div className="inline-flex shrink-0 items-center rounded-md border border-border bg-card p-0.5 text-xs">
      {opts.map(({ id, icon: Icon, label }) => (
        <button
          key={id}
          onClick={() => onChange(id)}
          className={cn(
            'flex items-center gap-1.5 rounded px-2.5 py-1 transition-colors',
            mode === id ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'
          )}
        >
          <Icon size={13} /> {label}
        </button>
      ))}
    </div>
  )
}
