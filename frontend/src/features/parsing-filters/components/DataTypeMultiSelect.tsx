import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Check, ChevronDown, Plus, X } from 'lucide-react'
import type { DataTypeOption } from '@/features/data-processing/types/data-processing.types'

interface Props {
  values: string[]
  options: DataTypeOption[]
  readOnly?: boolean
  onChange: (values: string[]) => void
}

/**
 * Multi-select for a filter block's dataTypes: pick from the integrations
 * catalog with search, and add a custom dataType the catalog doesn't know yet
 * (filters can target arbitrary dataTypes).
 */
export function DataTypeMultiSelect({ values, options, readOnly, onChange }: Props) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [q, setQ] = useState('')
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const toggle = (dt: string) =>
    onChange(values.includes(dt) ? values.filter((x) => x !== dt) : [...values, dt])

  const query = q.trim()
  const filtered = options.filter((o) =>
    query ? (o.dataType + ' ' + o.name).toLowerCase().includes(query.toLowerCase()) : true,
  )
  // Offer "add custom" when the typed value isn't already a known option or selected.
  const canAddCustom =
    query !== '' &&
    !options.some((o) => o.dataType.toLowerCase() === query.toLowerCase()) &&
    !values.some((v) => v.toLowerCase() === query.toLowerCase())

  const addCustom = () => {
    onChange([...values, query])
    setQ('')
  }

  return (
    <div ref={ref} className="relative">
      <div className="flex min-h-9 flex-wrap items-center gap-1 rounded-md border border-input bg-background/40 p-1.5">
        {values.map((v) => (
          <span
            key={v}
            className="inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5 font-mono text-[11px]"
          >
            {v}
            {!readOnly && (
              <button
                type="button"
                onClick={() => toggle(v)}
                className="text-muted-foreground hover:text-red-500"
              >
                <X size={11} />
              </button>
            )}
          </span>
        ))}
        {!readOnly && (
          <button
            type="button"
            onClick={() => setOpen((o) => !o)}
            className="inline-flex items-center gap-1 px-1 text-[11px] text-muted-foreground hover:text-foreground"
          >
            <Plus size={11} /> {values.length ? '' : t('parsingFilters.visual.dataTypesPick')}
            <ChevronDown size={10} className="opacity-60" />
          </button>
        )}
      </div>
      {open && !readOnly && (
        <div className="absolute left-0 top-full z-30 mt-1 w-72 rounded-md border border-border bg-popover py-1 shadow-lg">
          <div className="px-2 pb-1.5 pt-1">
            <input
              value={q}
              onChange={(e) => setQ(e.target.value)}
              autoFocus
              onKeyDown={(e) => {
                if (e.key === 'Enter' && canAddCustom) {
                  e.preventDefault()
                  addCustom()
                }
              }}
              placeholder={t('parsingFilters.visual.dataTypesSearch')}
              className="h-7 w-full rounded-md border border-input bg-background px-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
          </div>
          <div className="max-h-56 overflow-y-auto">
            {filtered.length === 0 && !canAddCustom && (
              <div className="px-3 py-1.5 text-xs text-muted-foreground">
                {t('parsingFilters.visual.dataTypesNone')}
              </div>
            )}
            {filtered.map((o) => {
              const on = values.includes(o.dataType)
              return (
                <button
                  key={o.dataType}
                  type="button"
                  onClick={() => toggle(o.dataType)}
                  className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-xs hover:bg-muted"
                >
                  <span className="min-w-0 flex-1 truncate font-mono">{o.dataType}</span>
                  {o.name && o.name !== o.dataType && (
                    <span className="shrink-0 truncate text-[10px] text-muted-foreground">{o.name}</span>
                  )}
                  {on && <Check size={13} className="shrink-0 text-primary" />}
                </button>
              )
            })}
            {canAddCustom && (
              <button
                type="button"
                onClick={addCustom}
                className="flex w-full items-center gap-2 border-t border-border px-3 py-1.5 text-left text-xs text-primary hover:bg-muted"
              >
                <Plus size={13} className="shrink-0" />
                <span className="truncate">{t('parsingFilters.visual.dataTypesAdd', { value: query })}</span>
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
