import { useTranslation } from 'react-i18next'
import { Plus } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { FilterRowEditor } from '@/features/dashboard/components/editor/FilterRowEditor'
import type { FilterRow, IndexPatternField } from '@/features/dashboard/types'

export function FilterBuilder({
  filters,
  fields,
  onChange,
}: {
  filters: FilterRow[]
  fields: IndexPatternField[]
  onChange: (next: FilterRow[]) => void
}) {
  const { t } = useTranslation()

  const add = () => {
    onChange([
      ...filters,
      {
        id: `f-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
        field: '',
        operator: 'IS',
        value: '',
      },
    ])
  }

  const updateRow = (id: string, next: FilterRow) => {
    onChange(filters.map((f) => (f.id === id ? next : f)))
  }
  const removeRow = (id: string) => {
    onChange(filters.filter((f) => f.id !== id))
  }

  return (
    <div className="flex flex-col gap-2">
      {filters.length === 0 ? (
        <p className="text-xs text-muted-foreground">{t('dashboards.editor.filters.empty')}</p>
      ) : (
        filters.map((row) => (
          <FilterRowEditor
            key={row.id}
            row={row}
            fields={fields}
            onChange={(next) => updateRow(row.id, next)}
            onRemove={() => removeRow(row.id)}
          />
        ))
      )}
      <div>
        <Button type="button" variant="outline" size="sm" onClick={add}>
          <Plus size={14} className="mr-1" />
          {t('dashboards.editor.filters.add')}
        </Button>
      </div>
    </div>
  )
}
