import { useEffect, useRef, useState } from 'react'
import { ChevronDown, Plus } from 'lucide-react'
import type { ApiKey } from '@/features/api-keys/types/api-key.types'

export interface ApiKeyPickerProps {
  keys: ApiKey[]
  value: number | null
  onChange: (id: number) => void
  onAddNew: () => void
  addLabel: string
  placeholder: string
  emptyLabel: string
  disabled?: boolean
}

export function ApiKeyPicker({ keys, value, onChange, onAddNew, addLabel, placeholder, emptyLabel, disabled }: ApiKeyPickerProps) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  const selected = keys.find((k) => k.id === value) ?? null

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  return (
    <div className="relative" ref={ref}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        disabled={disabled}
        className="flex h-9 w-full items-center justify-between gap-2 rounded-md border border-border bg-background px-3 text-sm text-foreground disabled:opacity-70"
      >
        <span className="truncate">{selected ? selected.name : placeholder}</span>
        <ChevronDown size={14} className="shrink-0 text-muted-foreground" />
      </button>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 w-full rounded-md border border-border bg-popover py-1 shadow-lg">
          {keys.length === 0 && (
            <span className="block px-3 py-1.5 text-sm text-muted-foreground">{emptyLabel}</span>
          )}
          {keys.map((k) => (
            <button
              key={k.id}
              type="button"
              onClick={() => { onChange(k.id); setOpen(false) }}
              className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm hover:bg-muted"
            >
              <span className="truncate">{k.name}</span>
            </button>
          ))}
          <button
            type="button"
            onClick={() => { onAddNew(); setOpen(false) }}
            className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm hover:bg-muted"
          >
            <Plus size={12} className="shrink-0 text-muted-foreground" />
            <span className="truncate">{addLabel}</span>
          </button>
        </div>
      )}
    </div>
  )
}
