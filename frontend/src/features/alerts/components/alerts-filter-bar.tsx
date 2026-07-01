import { X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { FILTER_OPS, fieldKey } from '../lib/alert-meta'
import type { CustomFilter } from '../types/alert.types'
import { AddFilterButton } from './add-filter-button'

export function AlertsFilterBar({
  filters,
  onAdd,
  onRemove,
  onClear,
}: {
  filters: CustomFilter[]
  onAdd: (f: CustomFilter) => void
  onRemove: (i: number) => void
  onClear: () => void
}) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-wrap items-center gap-1.5">
      {filters.map((f, i) => {
        const op = FILTER_OPS.find((o) => o.id === f.operator)
        return (
          <span
            key={i}
            className="inline-flex items-center gap-1.5 rounded-full border border-primary/25 bg-primary/5 py-1 pl-3 pr-1.5 text-xs"
          >
            <span className="text-muted-foreground">{t(`alerts.fields.${fieldKey(f.field)}`)}</span>
            <span className="text-[11px] text-muted-foreground/70">{op ? t(`alerts.ops.${op.id}`) : f.operator}</span>
            {op?.needsValue && <span className="max-w-[200px] truncate font-mono font-medium">{f.value}</span>}
            <button
              onClick={() => onRemove(i)}
              className="flex h-5 w-5 items-center justify-center rounded-full text-muted-foreground hover:bg-foreground/10 hover:text-foreground"
            >
              <X size={12} />
            </button>
          </span>
        )
      })}
      <AddFilterButton onAdd={onAdd} />
      {filters.length > 1 && (
        <button
          onClick={onClear}
          className="px-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground hover:underline"
        >
          {t('alerts.filters.clearAll')}
        </button>
      )}
    </div>
  )
}
