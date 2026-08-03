import { useEffect, useRef, useState } from 'react'
import { X } from 'lucide-react'
import { AddCustomFilterButton } from './AddCustomFilterButton'
import { FilterEditorPanel } from './FilterEditorPanel'
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
  onUpdate,
  onRemove,
  onClear,
  fields,
  operators,
  fetchValues,
  labels,
}: {
  filters: CustomFilter[]
  onAdd: (f: CustomFilter) => void
  onUpdate?: (i: number, f: CustomFilter) => void
  onRemove: (i: number) => void
  onClear: () => void
  fields: FilterFieldDef[]
  operators: FilterOpDef[]
  fetchValues?: (field: string) => Promise<FilterValue[]>
  labels: FilterBarLabels
}) {
  const [editing, setEditing] = useState<number | null>(null)
  const editRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (editing === null) return
    const onDoc = (e: MouseEvent) => editRef.current && !editRef.current.contains(e.target as Node) && setEditing(null)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [editing])

  return (
    <div className="flex flex-wrap items-center gap-1.5">
      {filters.map((f, i) => {
        const op = operators.find((o) => o.id === f.operator)
        const isEditing = editing === i
        return (
          <div key={i} className="relative" ref={isEditing ? editRef : undefined}>
            <span
              className="inline-flex items-center gap-1.5 rounded-full border border-primary/25 bg-primary/5 py-1 pl-3 pr-1.5 text-xs"
            >
              <button
                onClick={() => (onUpdate ? setEditing((cur) => (cur === i ? null : i)) : undefined)}
                disabled={!onUpdate}
                className="inline-flex items-center gap-1.5 rounded hover:bg-primary/10 disabled:cursor-default disabled:hover:bg-transparent"
              >
                <span className="text-muted-foreground">{f.label}</span>
                <span className="text-[11px] text-muted-foreground/70">{op?.label ?? f.operator}</span>
                {op?.needsValue && <span className="max-w-[200px] truncate font-mono font-medium">{f.value}</span>}
              </button>
              <button
                onClick={() => {
                  if (isEditing) setEditing(null)
                  onRemove(i)
                }}
                className="flex h-5 w-5 items-center justify-center rounded-full text-muted-foreground hover:bg-foreground/10 hover:text-foreground"
              >
                <X size={12} />
              </button>
            </span>
            {isEditing && onUpdate && (
              <div className="absolute left-0 top-full z-30 mt-1">
                <FilterEditorPanel
                  initial={f}
                  fields={fields}
                  operators={operators}
                  fetchValues={fetchValues}
                  labels={labels}
                  onCancel={() => setEditing(null)}
                  onSubmit={(next) => {
                    onUpdate(i, next)
                    setEditing(null)
                  }}
                />
              </div>
            )}
          </div>
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
