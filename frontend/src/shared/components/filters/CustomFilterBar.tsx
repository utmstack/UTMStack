import { X } from 'lucide-react'
import { AddCustomFilterButton } from './AddCustomFilterButton'
import type {
  CustomFilter,
  FilterBarLabels,
  FilterFieldDef,
  FilterOpDef,
  FilterValue,
} from './custom-filter.types'

export function CustomFilterBar({
  filters,
  onAdd,
  onRemove,
  onClear,
  fields,
  operators,
  fetchValues,
  labels,
}: {
  filters: CustomFilter[]
  onAdd: (f: CustomFilter) => void
  onRemove: (i: number) => void
  onClear: () => void
  fields: FilterFieldDef[]
  operators: FilterOpDef[]
  fetchValues?: (field: string) => Promise<FilterValue[]>
  labels: FilterBarLabels
}) {
  return (
    <div className="flex flex-wrap items-center gap-1.5">
      {filters.map((f, i) => {
        const op = operators.find((o) => o.id === f.operator)
        return (
          <span
            key={i}
            className="inline-flex items-center gap-1.5 rounded-full border border-primary/25 bg-primary/5 py-1 pl-3 pr-1.5 text-xs"
          >
            <span className="text-muted-foreground">{f.label}</span>
            <span className="text-[11px] text-muted-foreground/70">{op?.label ?? f.operator}</span>
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
      <AddCustomFilterButton onAdd={onAdd} fields={fields} operators={operators} fetchValues={fetchValues} labels={labels} />
      {filters.length > 1 && (
        <button
          onClick={onClear}
          className="px-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground hover:underline"
        >
          {labels.clearAll}
        </button>
      )}
    </div>
  )
}
