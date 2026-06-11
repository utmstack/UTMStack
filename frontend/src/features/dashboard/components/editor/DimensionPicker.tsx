import { useTranslation } from 'react-i18next'
import { FieldCombobox } from '@/features/dashboard/components/editor/FieldCombobox'
import type { IndexPatternField } from '@/features/dashboard/types'

export function DimensionPicker({
  value,
  fields,
  onChange,
}: {
  value: string | null
  fields: IndexPatternField[]
  onChange: (next: string | null) => void
}) {
  const { t } = useTranslation()
  return (
    <div>
      <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
        {t('dashboards.editor.dimension.field')}
      </label>
      <FieldCombobox
        value={value ?? ''}
        onChange={(v) => onChange(v || null)}
        fields={fields}
        placeholder={t('dashboards.editor.dimension.placeholder') ?? ''}
      />
    </div>
  )
}
