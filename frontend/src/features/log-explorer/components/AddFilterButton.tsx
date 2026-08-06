import { useEffect, useMemo, useRef, useState } from 'react'
import { Filter, Loader2, Search } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { logExplorerHttpService as svc } from '../services/log-explorer-http.service'
import type {
  FilterOperator,
  FilterType,
  IndexField,
  TopValues,
} from '../types/log-explorer.types'
import { OP_KEY, SELECT_CLS } from './log-explorer.constants'

/* Explicit field/operator/value filter builder — pick a field, an operator, and
 * a value from the field's real existing values (fetched via top-x-values). */
const BUILDER_OPS: { id: FilterOperator; label: string; needsValue: boolean }[] = [
  { id: 'IS', label: 'is', needsValue: true },
  { id: 'IS_NOT', label: 'is not', needsValue: true },
  { id: 'CONTAIN', label: 'contains', needsValue: true },
  { id: 'EXIST', label: 'exists', needsValue: false },
]

export function AddFilterButton({
  pattern,
  fields,
  filters,
  onAdd,
}: {
  pattern: string
  fields: IndexField[]
  filters: FilterType[]
  onAdd: (f: FilterType) => void
}) {
  const { t } = useTranslation()
  const selectable = useMemo(
    () => fields.filter((f) => !f.name.endsWith('.keyword')).sort((a, b) => a.name.localeCompare(b.name)),
    [fields]
  )
  const [open, setOpen] = useState(false)
  const [field, setField] = useState('')
  const [operator, setOperator] = useState<FilterOperator>('IS')
  const [values, setValues] = useState<TopValues['top']>([])
  const [loadingValues, setLoadingValues] = useState(false)
  const [vq, setVq] = useState('')
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (open && !field && selectable.length) setField(selectable[0].name)
  }, [open, field, selectable])

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const op = BUILDER_OPS.find((o) => o.id === operator)
  const needsValue = op?.needsValue ?? true
  const fieldDef = selectable.find((f) => f.name === field)
  const aggField = fieldDef?.type === 'text' && !field.endsWith('.keyword') ? `${field}.keyword` : field

  useEffect(() => {
    if (!open || !needsValue || !field) return
    setLoadingValues(true)
    setValues([])
    svc
      .topValues(pattern, aggField, filters, 100)
      .then((r) => setValues(r.top ?? []))
      .catch(() => setValues([]))
      .finally(() => setLoadingValues(false))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, field, needsValue])

  const add = (value: string) => {
    onAdd({ field, operator, value: needsValue ? value : undefined })
    setOpen(false)
    setVq('')
  }
  const filtered = values.filter((v) => (vq ? String(v.value).toLowerCase().includes(vq.toLowerCase()) : true))

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        className="inline-flex items-center gap-1.5 rounded-full border border-dashed border-border px-3 py-1 text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
      >
        <Filter size={12} /> {t('logExplorer.builder.add')}
      </button>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 w-80 rounded-md border border-border bg-popover p-3 shadow-lg">
          <div className="space-y-2">
            <select value={field} onChange={(e) => setField(e.target.value)} className={cn(SELECT_CLS, 'w-full font-mono')}>
              {selectable.map((f) => (
                <option key={f.name} value={f.name}>{f.name}</option>
              ))}
            </select>
            <select value={operator} onChange={(e) => setOperator(e.target.value as FilterOperator)} className={cn(SELECT_CLS, 'w-full')}>
              {BUILDER_OPS.map((o) => (
                <option key={o.id} value={o.id}>{t(`logExplorer.ops.${OP_KEY[o.id] ?? o.id}`)}</option>
              ))}
            </select>
            {needsValue ? (
              <>
                <div className="relative">
                  <Search size={13} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
                  <Input value={vq} onChange={(e) => setVq(e.target.value)} placeholder={t('logExplorer.builder.filterValues')} className="h-8 pl-8 text-xs" autoFocus />
                </div>
                <div className="max-h-48 overflow-y-auto rounded-md border border-border">
                  {loadingValues ? (
                    <div className="flex items-center gap-1.5 px-3 py-3 text-xs text-muted-foreground"><Loader2 className="h-3.5 w-3.5 animate-spin" /> {t('logExplorer.builder.loadingValues')}</div>
                  ) : filtered.length === 0 ? (
                    <div className="px-3 py-3 text-xs text-muted-foreground">{t('logExplorer.builder.noValues')}</div>
                  ) : (
                    filtered.map((v) => (
                      <button key={v.value} onClick={() => add(v.value)} className="flex w-full items-center justify-between gap-2 px-3 py-1.5 text-left text-xs hover:bg-muted">
                        <span className="truncate font-mono">{v.value || t('logExplorer.fields.empty')}</span>
                        <span className="shrink-0 font-mono text-[10px] text-muted-foreground">{v.count.toLocaleString()}</span>
                      </button>
                    ))
                  )}
                </div>
              </>
            ) : (
              <div className="flex justify-end gap-2 pt-1">
                <Button variant="outline" size="sm" onClick={() => setOpen(false)}>{t('logExplorer.builder.cancel')}</Button>
                <Button size="sm" onClick={() => add('')}>{t('logExplorer.builder.confirm')}</Button>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
