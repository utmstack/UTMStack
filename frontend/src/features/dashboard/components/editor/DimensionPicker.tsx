import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { FieldSelect } from '@/features/dashboard/components/editor/FieldSelect'
import type { IndexProperty } from '@/features/dashboard/types'

export function DimensionPicker({
  value,
  fields,
  loading,
  onChange,
}: {
  value: string | null
  fields: IndexProperty[]
  loading?: boolean
  onChange: (next: string | null) => void
}) {
  const { t } = useTranslation()

  // Same staleness guard as MetricPicker: a raw-SQL round trip or an index
  // pattern change can leave `value` pointing at a field that's no longer
  // groupable, while the select silently falls back to its placeholder.
  useEffect(() => {
    if (loading) return
    if (!value) return
    if (fields.some((f) => f.name === value)) return
    onChange(null)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [fields, loading])

  return (
    <div>
      <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
        {t('dashboards.editor.dimension.field')}
      </label>
      <FieldSelect
        value={value}
        onChange={(v) => onChange(v || null)}
        fields={fields}
        loading={loading}
        placeholder={t('dashboards.editor.dimension.placeholder') ?? undefined}
      />
    </div>
  )
}
