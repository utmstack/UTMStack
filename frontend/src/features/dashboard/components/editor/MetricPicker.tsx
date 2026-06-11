import { useTranslation } from 'react-i18next'
import { FieldSelect } from '@/features/dashboard/components/editor/FieldSelect'
import { AGGREGATIONS } from '@/features/dashboard/constants'
import type {
  AggregationId,
  BuilderMetric,
  IndexProperty,
} from '@/features/dashboard/types'

export function MetricPicker({
  value,
  fields,
  loading,
  onChange,
}: {
  value: BuilderMetric
  fields: IndexProperty[]
  loading?: boolean
  onChange: (next: BuilderMetric) => void
}) {
  const { t } = useTranslation()
  const meta = AGGREGATIONS.find((a) => a.id === value.agg) ?? AGGREGATIONS[0]
  const needsField = meta.requiresField

  return (
    <div className="flex flex-col gap-2 md:flex-row md:items-end">
      <div className="md:w-48">
        <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          {t('dashboards.editor.metric.aggregation')}
        </label>
        <select
          value={value.agg}
          onChange={(e) =>
            onChange({ ...value, agg: e.target.value as AggregationId })
          }
          className="h-9 w-full rounded-md border border-input bg-background px-2 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        >
          {AGGREGATIONS.map((a) => (
            <option key={a.id} value={a.id}>
              {t(`dashboards.editor.aggregations.${a.id}`)}
            </option>
          ))}
        </select>
      </div>
      <div className="flex-1">
        <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          {t('dashboards.editor.metric.field')}
        </label>
        <FieldSelect
          value={value.field}
          onChange={(v) => onChange({ ...value, field: v || null })}
          fields={fields}
          loading={loading}
          disabled={!needsField}
          placeholder={
            needsField
              ? t('dashboards.editor.metric.fieldPlaceholder') ?? undefined
              : t('dashboards.editor.metric.fieldNotRequired') ?? undefined
          }
        />
      </div>
    </div>
  )
}
