import { useTranslation } from 'react-i18next'
import { normalizeType } from '@/features/dashboard/utils/field-types'
import type { IndexProperty } from '@/features/dashboard/types'

export function FieldSelect({
  value,
  onChange,
  fields,
  loading,
  disabled,
  placeholder,
  allowEmpty = true,
}: {
  value: string | null
  onChange: (next: string) => void
  fields: IndexProperty[]
  loading?: boolean
  disabled?: boolean
  placeholder?: string
  allowEmpty?: boolean
}) {
  const { t } = useTranslation()
  return (
    <select
      value={value ?? ''}
      onChange={(e) => onChange(e.target.value)}
      disabled={disabled || loading}
      className="h-9 w-full rounded-md border border-input bg-background px-2 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-60"
    >
      {allowEmpty && (
        <option value="">
          {loading
            ? t('dashboards.editor.fieldSelect.loading')
            : placeholder ?? t('dashboards.editor.fieldSelect.choose')}
        </option>
      )}
      {fields.map((f) => (
        <option key={f.name} value={f.name}>
          {f.name} {f.type ? `(${normalizeType(f.type)})` : ''}
        </option>
      ))}
    </select>
  )
}
