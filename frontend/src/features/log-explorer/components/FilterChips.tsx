import { Filter, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import type { FilterType } from '../types/log-explorer.types'
import { OP_KEY } from './log-explorer.constants'

export function FilterChips({
  filters,
  onRemove,
  onClear,
}: {
  filters: FilterType[]
  onRemove: (i: number) => void
  onClear: () => void
}) {
  const { t } = useTranslation()
  return (
    <div className="mt-3 flex flex-wrap items-center gap-2">
      <span className="inline-flex items-center gap-1.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/70">
        <Filter size={12} /> {t('logExplorer.filters.label')}
      </span>
      {filters.map((f, i) => {
        const neg = f.operator === 'IS_NOT'
        return (
          <span
            key={i}
            className={cn(
              'inline-flex items-center gap-1.5 rounded-full border py-1 pl-3 pr-1.5 text-xs',
              neg ? 'border-red-500/30 bg-red-500/10' : 'border-primary/25 bg-primary/5'
            )}
          >
            <span className="font-mono text-muted-foreground">{f.field}</span>
            <span className={cn('text-[11px]', neg ? 'text-red-500' : 'text-muted-foreground/70')}>
              {OP_KEY[f.operator] ? t(`logExplorer.ops.${OP_KEY[f.operator]}`) : f.operator}
            </span>
            {f.value != null && f.operator !== 'EXIST' && (
              <span className="max-w-[220px] truncate font-mono font-medium">
                {Array.isArray(f.value) ? t('logExplorer.related.nLogs', { count: f.value.length }) : String(f.value)}
              </span>
            )}
            <button
              onClick={() => onRemove(i)}
              className="flex h-5 w-5 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-foreground/10 hover:text-foreground"
            >
              <X size={12} />
            </button>
          </span>
        )
      })}
      {filters.length > 1 && (
        <button
          onClick={onClear}
          className="px-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground hover:underline"
        >
          {t('logExplorer.filters.clearAll')}
        </button>
      )}
    </div>
  )
}
