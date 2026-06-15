import { useTranslation } from 'react-i18next'
import { X } from 'lucide-react'
import { FieldSelect } from '@/features/dashboard/components/editor/FieldSelect'
import type { IndexProperty } from '@/features/dashboard/types'

/**
 * Multi-column picker for table widgets: chips for selected columns + a field
 * dropdown to add more. Empty selection means "all columns" (SELECT *).
 */
export function ColumnsPicker({
  value,
  fields,
  loading,
  onChange,
}: {
  value: string[]
  fields: IndexProperty[]
  loading?: boolean
  onChange: (next: string[]) => void
}) {
  const { t } = useTranslation()
  const selected = value ?? []
  const available = fields.filter((f) => !selected.includes(f.name))

  const add = (name: string) => {
    if (!name || selected.includes(name)) return
    onChange([...selected, name])
  }
  const remove = (name: string) => onChange(selected.filter((c) => c !== name))

  return (
    <div className="flex flex-col gap-2">
      {selected.length > 0 ? (
        <div className="flex flex-wrap gap-1.5">
          {selected.map((c) => (
            <span
              key={c}
              className="inline-flex items-center gap-1 rounded-md bg-muted px-2 py-1 font-mono text-xs ring-1 ring-inset ring-border"
            >
              {c}
              <button
                type="button"
                onClick={() => remove(c)}
                className="text-muted-foreground hover:text-foreground"
                aria-label={t('dashboards.editor.table.removeColumn') ?? 'Remove'}
              >
                <X size={12} />
              </button>
            </span>
          ))}
        </div>
      ) : (
        <p className="text-xs text-muted-foreground">{t('dashboards.editor.table.allColumns')}</p>
      )}
      <FieldSelect
        value=""
        onChange={add}
        fields={available}
        loading={loading}
        placeholder={t('dashboards.editor.table.addColumn') ?? undefined}
      />
    </div>
  )
}
