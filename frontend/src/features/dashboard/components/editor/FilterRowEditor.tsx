import { useTranslation } from 'react-i18next'
import { X } from 'lucide-react'
import { Input } from '@/shared/components/ui/input'
import { FieldCombobox } from '@/features/dashboard/components/editor/FieldCombobox'
import { OPERATORS, getOperatorMeta } from '@/features/dashboard/constants'
import type { FilterOperatorId, FilterRow, IndexPatternField } from '@/features/dashboard/types'

export function FilterRowEditor({
  row,
  fields,
  onChange,
  onRemove,
}: {
  row: FilterRow
  fields: IndexPatternField[]
  onChange: (next: FilterRow) => void
  onRemove: () => void
}) {
  const { t } = useTranslation()
  const meta = getOperatorMeta(row.operator)

  const setOperator = (op: FilterOperatorId) => {
    const nextMeta = getOperatorMeta(op)
    let value: FilterRow['value'] = null
    if (nextMeta.valueShape === 'single') value = ''
    else if (nextMeta.valueShape === 'pair') value = ['', '']
    else if (nextMeta.valueShape === 'list') value = []
    onChange({ ...row, operator: op, value })
  }

  return (
    <div className="flex flex-col gap-2 rounded-md border border-border bg-background/40 p-3 md:flex-row md:items-start">
      <div className="md:w-1/3">
        <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          {t('dashboards.editor.filters.field')}
        </label>
        <FieldCombobox
          value={row.field}
          onChange={(v) => onChange({ ...row, field: v })}
          fields={fields}
          placeholder={t('dashboards.editor.filters.fieldPlaceholder') ?? ''}
        />
      </div>
      <div className="md:w-44">
        <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          {t('dashboards.editor.filters.operator')}
        </label>
        <select
          value={row.operator}
          onChange={(e) => setOperator(e.target.value as FilterOperatorId)}
          className="h-9 w-full rounded-md border border-input bg-background px-2 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        >
          {OPERATORS.map((o) => (
            <option key={o.id} value={o.id}>
              {t(`dashboards.editor.operators.${o.id}`)}
            </option>
          ))}
        </select>
      </div>
      <div className="flex-1">
        <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          {t('dashboards.editor.filters.value')}
        </label>
        <ValueInput row={row} meta={meta} onChange={onChange} />
      </div>
      <div className="flex items-end justify-end md:pt-5">
        <button
          type="button"
          onClick={onRemove}
          className="flex h-9 w-9 items-center justify-center rounded-md text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
          aria-label={t('dashboards.editor.filters.remove') ?? 'Remove'}
        >
          <X size={14} />
        </button>
      </div>
    </div>
  )
}

function ValueInput({
  row,
  meta,
  onChange,
}: {
  row: FilterRow
  meta: ReturnType<typeof getOperatorMeta>
  onChange: (next: FilterRow) => void
}) {
  const { t } = useTranslation()

  if (meta.valueShape === 'none') {
    return (
      <div className="flex h-9 items-center text-xs text-muted-foreground">
        {t('dashboards.editor.filters.noValue')}
      </div>
    )
  }

  if (meta.valueShape === 'pair') {
    const v = Array.isArray(row.value) ? row.value : ['', '']
    const [a, b] = [String(v[0] ?? ''), String(v[1] ?? '')]
    return (
      <div className="flex items-center gap-2">
        <Input
          value={a}
          onChange={(e) => onChange({ ...row, value: [e.target.value, b] })}
          placeholder={t('dashboards.editor.filters.from') ?? ''}
          className="h-9"
        />
        <span className="text-xs text-muted-foreground">—</span>
        <Input
          value={b}
          onChange={(e) => onChange({ ...row, value: [a, e.target.value] })}
          placeholder={t('dashboards.editor.filters.to') ?? ''}
          className="h-9"
        />
      </div>
    )
  }

  if (meta.valueShape === 'list') {
    const items = Array.isArray(row.value) ? row.value.map((x) => String(x)) : []
    return (
      <div className="space-y-1.5">
        <Input
          value={items.join(', ')}
          onChange={(e) =>
            onChange({
              ...row,
              value: e.target.value
                .split(',')
                .map((s) => s.trim())
                .filter(Boolean),
            })
          }
          placeholder={t('dashboards.editor.filters.listPlaceholder') ?? ''}
          className="h-9"
        />
        {items.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {items.map((it, i) => (
              <span
                key={`${it}-${i}`}
                className="rounded bg-muted px-1.5 py-0.5 text-[10px] text-foreground/80"
              >
                {it}
              </span>
            ))}
          </div>
        )}
      </div>
    )
  }

  const v = typeof row.value === 'string' ? row.value : ''
  return (
    <Input
      value={v}
      onChange={(e) => onChange({ ...row, value: e.target.value })}
      placeholder={t('dashboards.editor.filters.valuePlaceholder') ?? ''}
      className="h-9"
    />
  )
}
