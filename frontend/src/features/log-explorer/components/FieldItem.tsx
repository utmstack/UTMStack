import { useEffect, useState } from 'react'
import { Check, ChevronRight, Columns3, Loader2, Minus, Plus } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { logExplorerHttpService as svc } from '../services/log-explorer-http.service'
import type { FilterType, IndexField, IndexPattern, TopValues } from '../types/log-explorer.types'
import { TypeBadge } from './TypeBadge'

export function FieldItem({
  field,
  pattern,
  filters,
  isColumn,
  open,
  onToggle,
  onAdd,
  onToggleColumn,
}: {
  field: IndexField
  pattern: IndexPattern | null
  filters: FilterType[]
  isColumn: boolean
  open: boolean
  onToggle: () => void
  onAdd: (f: FilterType) => void
  onToggleColumn: () => void
}) {
  const { t } = useTranslation()
  const [top, setTop] = useState<TopValues | null>(null)
  const [loading, setLoading] = useState(false)

  // Aggregations need the keyword sub-field for text types.
  const aggField =
    field.type === 'text' && !field.name.endsWith('.keyword') ? `${field.name}.keyword` : field.name

  useEffect(() => {
    if (!open || !pattern || top) return
    setLoading(true)
    svc
      .topValues(pattern.pattern, aggField, filters, 5)
      .then(setTop)
      .catch(() => setTop({ total: 0, top: [] }))
      .finally(() => setLoading(false))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  return (
    <div className={cn('group/field rounded-md', open && 'bg-card shadow-sm ring-1 ring-border/70')}>
      <div className={cn('flex items-center gap-2.5 rounded-md px-2 py-2', !open && 'hover:bg-card/70')}>
        <button onClick={onToggle} className="flex min-w-0 flex-1 items-center gap-2.5 text-left">
          <TypeBadge type={field.type} />
          <span className="flex-1 truncate font-mono text-xs" title={field.name}>
            {field.name}
          </span>
        </button>
        <button
          onClick={onToggleColumn}
          title={isColumn ? t('logExplorer.fields.removeColumn') : t('logExplorer.fields.addColumn')}
          className={cn(
            'flex h-6 w-6 shrink-0 items-center justify-center rounded transition-colors',
            isColumn
              ? 'text-primary hover:bg-primary/10'
              : 'text-muted-foreground opacity-0 hover:bg-muted group-hover/field:opacity-100'
          )}
        >
          {isColumn ? <Check size={13} /> : <Columns3 size={13} />}
        </button>
        <button onClick={onToggle} className="shrink-0">
          <ChevronRight size={13} className={cn('text-muted-foreground/60 transition-transform', open && 'rotate-90')} />
        </button>
      </div>
      {open && (
        <div className="px-2 pb-2.5 pt-1">
          <div className="mb-1.5 px-1 text-[10px] uppercase tracking-wider text-muted-foreground/60">
            {t('logExplorer.fields.topValues', { count: Math.min(5, top?.top.length ?? 0) })}
          </div>
          {loading ? (
            <div className="flex items-center gap-1.5 px-1 py-2 text-[11px] text-muted-foreground">
              <Loader2 className="h-3.5 w-3.5 animate-spin" /> {t('logExplorer.fields.loading')}
            </div>
          ) : !top || top.top.length === 0 ? (
            <div className="px-1 py-2 text-[11px] text-muted-foreground">{t('logExplorer.fields.noValues')}</div>
          ) : (
            <div className="space-y-1">
              {top.top.slice(0, 5).map((v) => (
                <div key={v.value} className="group rounded px-1.5 py-1.5 hover:bg-muted/50">
                  <div className="flex items-center justify-between gap-2">
                    <span className="truncate font-mono text-[11px]" title={v.value}>
                      {v.value || t('logExplorer.fields.empty')}
                    </span>
                    <span className="shrink-0 font-mono text-[10px] text-muted-foreground group-hover:hidden">
                      {Math.round(v.percent)}%
                    </span>
                    <div className="hidden shrink-0 items-center gap-1 group-hover:flex">
                      <button
                        title={t('logExplorer.fields.filterFor')}
                        onClick={() => onAdd({ field: field.name, operator: 'IS', value: v.value })}
                        className="flex h-5 w-5 items-center justify-center rounded text-emerald-500 hover:bg-emerald-500/15"
                      >
                        <Plus size={12} />
                      </button>
                      <button
                        title={t('logExplorer.fields.filterOut')}
                        onClick={() => onAdd({ field: field.name, operator: 'IS_NOT', value: v.value })}
                        className="flex h-5 w-5 items-center justify-center rounded text-red-500 hover:bg-red-500/15"
                      >
                        <Minus size={12} />
                      </button>
                    </div>
                  </div>
                  <div className="mt-1.5 h-1 overflow-hidden rounded-full bg-muted">
                    <div className="h-full rounded-full bg-primary/50" style={{ width: `${Math.max(3, v.percent)}%` }} />
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
