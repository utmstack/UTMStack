import { useEffect, useRef, useState } from 'react'
import { ListFilter } from 'lucide-react'
import { FilterEditorPanel } from './FilterEditorPanel'
import type {
  CustomFilter,
  FilterBarLabels,
  FilterFieldDef,
  FilterOpDef,
  FilterValue,
} from './custom-filter.types'

export function AddCustomFilterButton({
  onAdd,
  fields,
  operators,
  fetchValues,
  labels,
}: {
  onAdd: (f: CustomFilter) => void
  fields: FilterFieldDef[]
  operators: FilterOpDef[]
  fetchValues?: (field: string) => Promise<FilterValue[]>
  labels: FilterBarLabels
}) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        className="inline-flex items-center gap-1.5 rounded-full border border-dashed border-border px-3 py-1 text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
      >
        <ListFilter size={12} /> {labels.add}
      </button>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1">
          <FilterEditorPanel
            fields={fields}
            operators={operators}
            fetchValues={fetchValues}
            labels={labels}
            onCancel={() => setOpen(false)}
            onSubmit={(f) => {
              onAdd(f)
              setOpen(false)
            }}
          />
        </div>
      )}
    </div>
  )
}
