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
