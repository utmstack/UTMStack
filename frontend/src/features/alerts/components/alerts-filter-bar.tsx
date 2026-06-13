import { useEffect, useRef, useState } from 'react'
import { ListFilter, Loader2, Search, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { alertsHttpService as svc } from '../services/alerts-http.service'
import { FILTER_FIELDS, FILTER_OPS, SELECT_CLS, fieldKey } from '../lib/alert-meta'
import type { CustomFilter } from '../types/alert.types'

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

function AddFilterButton({ onAdd }: { onAdd: (f: CustomFilter) => void }) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [field, setField] = useState(FILTER_FIELDS[0].field)
  const [operator, setOperator] = useState('IS')
  const [values, setValues] = useState<{ value: string; count: number }[]>([])
  const [loadingValues, setLoadingValues] = useState(false)
  const [vq, setVq] = useState('')
  const ref = useRef<HTMLDivElement>(null)

  const op = FILTER_OPS.find((o) => o.id === operator)
  const needsValue = op?.needsValue ?? true

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  // Fetch the field's real existing values when the field changes.
  useEffect(() => {
    if (!open || !needsValue) return
    setLoadingValues(true)
    setValues([])
    svc
      .fieldValues(field)
      .then(setValues)
      .catch(() => setValues([]))
      .finally(() => setLoadingValues(false))
  }, [open, field, needsValue])

  const add = (value: string) => {
    const fdef = FILTER_FIELDS.find((f) => f.field === field)!
    onAdd({ field, label: fdef.label, operator, value })
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
        <ListFilter size={12} /> {t('alerts.filters.add')}
      </button>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 w-80 rounded-md border border-border bg-popover p-3 shadow-lg">
          <div className="space-y-2">
            <select value={field} onChange={(e) => setField(e.target.value)} className={cn(SELECT_CLS, 'w-full')}>
              {FILTER_FIELDS.map((f) => (
                <option key={f.field} value={f.field}>
                  {t(`alerts.fields.${fieldKey(f.field)}`)}
                </option>
              ))}
            </select>
            <select value={operator} onChange={(e) => setOperator(e.target.value)} className={cn(SELECT_CLS, 'w-full')}>
              {FILTER_OPS.map((o) => (
                <option key={o.id} value={o.id}>
                  {t(`alerts.ops.${o.id}`)}
                </option>
              ))}
            </select>

            {needsValue ? (
              <>
                <div className="relative">
                  <Search size={13} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    value={vq}
                    onChange={(e) => setVq(e.target.value)}
                    placeholder={t('alerts.filters.filterValues')}
                    className="h-8 pl-8 text-xs"
                    autoFocus
                  />
                </div>
                <div className="max-h-48 overflow-y-auto rounded-md border border-border">
                  {loadingValues ? (
                    <div className="flex items-center gap-1.5 px-3 py-3 text-xs text-muted-foreground">
                      <Loader2 className="h-3.5 w-3.5 animate-spin" /> {t('alerts.filters.loadingValues')}
                    </div>
                  ) : filtered.length === 0 ? (
                    <div className="px-3 py-3 text-xs text-muted-foreground">{t('alerts.filters.noValues')}</div>
                  ) : (
                    filtered.map((v) => (
                      <button
                        key={v.value}
                        onClick={() => add(v.value)}
                        className="flex w-full items-center justify-between gap-2 px-3 py-1.5 text-left text-xs hover:bg-muted"
                      >
                        <span className="truncate font-mono">{v.value || t('alerts.filters.empty')}</span>
                        <span className="shrink-0 font-mono text-[10px] text-muted-foreground">{v.count.toLocaleString()}</span>
                      </button>
                    ))
                  )}
                </div>
                <p className="text-[10px] text-muted-foreground">{t('alerts.filters.pickValue')}</p>
              </>
            ) : (
              <div className="flex justify-end gap-2 pt-1">
                <Button variant="outline" size="sm" onClick={() => setOpen(false)}>
                  {t('alerts.filters.cancel')}
                </Button>
                <Button size="sm" onClick={() => add('')}>
                  {t('alerts.filters.addBtn')}
                </Button>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
