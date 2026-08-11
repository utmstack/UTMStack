import { useEffect, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { X } from 'lucide-react'
import { Input } from '@/shared/components/ui/input'
import { FieldSelect } from '@/features/dashboard/components/editor/FieldSelect'
import { getOperatorMeta } from '@/features/dashboard/constants'
import { operatorsForFieldType } from '@/features/dashboard/utils/field-types'
import type { FilterOperatorId, FilterRow, IndexProperty } from '@/features/dashboard/types'

function defaultValueForOperator(op: FilterOperatorId): FilterRow['value'] {
  const meta = getOperatorMeta(op)
  if (meta.valueShape === 'single') return ''
  if (meta.valueShape === 'pair') return ['', '']
  if (meta.valueShape === 'list') return []
  return null
}

export function FilterRowEditor({
  row,
  fields,
  loading,
  onChange,
  onRemove,
}: {
  row: FilterRow
  fields: IndexProperty[]
  loading?: boolean
  onChange: (next: FilterRow) => void
  onRemove: () => void
}) {
  const { t } = useTranslation()
  const meta = getOperatorMeta(row.operator)
  const fieldType = fields.find((f) => f.name === row.field)?.type
  const validOperators = useMemo(() => operatorsForFieldType(fieldType), [fieldType])

  const setOperator = (op: FilterOperatorId) => {
    onChange({ ...row, operator: op, value: defaultValueForOperator(op) })
  }

  // Same staleness guard as Metric/DimensionPicker, applied to a filter row:
  // if the field no longer exists on this index pattern (switched patterns),
  // clear the whole row; if it still exists but the chosen operator no longer
  // fits its type (e.g. a raw-SQL round trip left `CONTAIN` on a date field),
  // fall back to the first operator that does. Otherwise the operator select
  // keeps showing an option that isn't even in its own list, and the query
  // composer keeps emitting a combination the store will reject.
  useEffect(() => {
    if (loading) return
    if (!row.field) return
    const fieldStillExists = fields.some((f) => f.name === row.field)
    if (!fieldStillExists) {
      onChange({ ...row, field: '', operator: 'IS', value: '' })
      return
    }
    if (validOperators.some((o) => o.id === row.operator)) return
    const nextOp = validOperators[0]?.id ?? 'IS'
    onChange({ ...row, operator: nextOp, value: defaultValueForOperator(nextOp) })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [row.field, fields, loading, validOperators])

  return (
    <div className="flex flex-col gap-2 rounded-md border border-border bg-background/40 p-3 md:flex-row md:items-start">
      <div className="md:w-1/3">
        <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          {t('dashboards.editor.filters.field')}
        </label>
        <FieldSelect
          value={row.field}
          onChange={(v) => onChange({ ...row, field: v })}
          fields={fields}
          loading={loading}
          placeholder={t('dashboards.editor.filters.fieldPlaceholder') ?? undefined}
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
          {validOperators.map((o) => (
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
