import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { FieldSelect } from '@/features/dashboard/components/editor/FieldSelect'
import { AGGREGATIONS } from '@/features/dashboard/constants'
import { fieldsForAggregation } from '@/features/dashboard/utils/field-types'
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

  // Only the fields this aggregation can actually use — e.g. SUM/AVG offer
  // numeric fields only, so the visual builder can never generate a broken query.
  const fieldOptions = useMemo(
    () => fieldsForAggregation(fields, value.agg),
    [fields, value.agg]
  )
  const noCompatible = needsField && !loading && fieldOptions.length === 0

  return (
    <div className="flex flex-col gap-2 md:flex-row md:items-end">
      <div className="md:w-48">
        <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          {t('dashboards.editor.metric.aggregation')}
        </label>
        <select
          value={value.agg}
          onChange={(e) => {
            const nextAgg = e.target.value as AggregationId
            // Drop the field if it's not valid for the new aggregation (e.g.
            // moving a text field from COUNT DISTINCT to SUM).
            const valid = fieldsForAggregation(fields, nextAgg)
            const keepField =
              value.field && valid.some((f) => f.name === value.field) ? value.field : null
            onChange({ agg: nextAgg, field: keepField })
          }}
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
          fields={fieldOptions}
          loading={loading}
          disabled={!needsField}
          placeholder={
            !needsField
              ? t('dashboards.editor.metric.fieldNotRequired') ?? undefined
              : noCompatible
                ? t('dashboards.editor.metric.noCompatibleFields', {
                    defaultValue: 'No compatible fields for this aggregation',
                  })
                : t('dashboards.editor.metric.fieldPlaceholder') ?? undefined
          }
        />
      </div>
    </div>
  )
}
